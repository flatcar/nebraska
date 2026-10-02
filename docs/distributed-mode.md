# WIP: Distributed Nebraska — Design

**Status:** Proposed (RFC [#1375](https://github.com/flatcar/nebraska/issues/1375), implementation in flight)
**Audience:** Nebraska maintainers and contributors

## 1. Problem

Nebraska today assumes a single instance with a single Postgres database. Hosting it as a managed service - where instances are deployed across multiple regions so each fleet of Flatcar nodes can reach the nearest one with low latency - hits two walls:

1. **Every node writes runtime data.** Omaha update checks insert into `instance`, `event`, `activity`, etc. Read replicas are not an option.
2. **A single database is a single point of failure** for the entire fleet across all regions, and adds cross-region latency for every Omaha poll.

The goal is to let Nebraska run as multiple regional nodes, each backed by its own writable Postgres, with publisher metadata (apps, channels, packages, groups) staying consistent across all of them.

## 2. Requirements

- Single-instance behaviour is preserved. An operator who does not opt in sees no change.
- One control node and N edge nodes can be configured.
- Admin metadata (apps, channels, packages, groups) is consistent across all edge nodes.
- Each edge node serves Omaha update checks and records its own runtime data locally.
- Admin API writes are accepted on exactly one node (the control node) and blocked on the rest.

## 3. Non-goals

- **Control-plane HA / failover.** Single-instance Nebraska already has the DB as a SPOF; HA-Postgres addresses it identically in both worlds. Not regressed, not solved here.
- **Multi-primary writes.** Admin writes have a single logical target.
- **Cross-region telemetry aggregation.** Each region's runtime data stays local. Fleet-wide views are an operator concern, not a Nebraska feature.
- **Auto-pause-on-failure (`safe_mode`) coordinated across nodes.** Each node owns its own brake; see §5 for how this is expressed.

## 4. Design

### 4.1 Topology

```
        Publisher (admin API)
                │
         ┌──────▼──────┐
         │ Control     │  admin writes accepted
         │ Nebraska    │  ──┐ logical replication
         │  + own PG   │    │ (admin tables only)
         └─────────────┘    │
                            ▼
            ┌──────────────────────────┐
            │ Edge Nebraska × N        │  admin writes return 403
            │  + own regional PG each  │  Omaha checks + telemetry local
            └──────────────────────────┘
```

One control node, N edge nodes. Each runs the same Nebraska binary, each owns its own Postgres. Publisher metadata flows control → edge via **PostgreSQL logical replication** of the admin tables. Runtime tables stay local per edge.

### 4.2 Table model

Nebraska's schema already splits cleanly into admin and runtime tables, with FKs pointing only from runtime to admin. Two tables today mix the two concerns and need to be split before replication is safe:

- **`groups`** mixes admin-owned policy fields with the runtime-mutated `rollout_in_progress` flag, and `policy_*` fields the runtime needs to be able to override locally (safe-mode brake). → Splits into `groups` (admin defaults, replicated) + node-local `group_local` sidecar. `group_local` holds `rollout_in_progress` plus one nullable `policy_*_override` column for each admin policy field. Reads return `COALESCE(group_local.policy_X_override, groups.policy_X)` for policy fields, and `group_local.rollout_in_progress` directly. An `AFTER INSERT` trigger on `groups` auto-creates the matching `group_local` row on every node, including during a subscriber's initial COPY sync, so the sidecar is always present. [PR #1396](https://github.com/flatcar/nebraska/pull/1396).
- **`activity`** mixes admin-originated rows (channel-package-updated) and runtime-originated rows (rollout lifecycle, instance update results). → Splits into `admin_activity` (replicated) + `activity` (local) + `all_activity` view. [PR #1398](https://github.com/flatcar/nebraska/pull/1398).

With these two splits, every table is either pure admin (replicated control → edge) or pure runtime (local per edge), or a small reference set that's seeded by migrations on every node:

| Category | Tables | Lifecycle |
|---|---|---|
| Admin (replicated control → edge) | `application`, `package`, `channel`, `groups`, `team`, `users`, `flatcar_action`, `package_channel_blacklist`, `channel_package_floors`, `package_file`, `admin_activity` | Written on the control node; flows one-way via logical replication. |
| Runtime (local per edge) | `instance`, `instance_application`, `instance_status_history`, `event`, `activity`, `group_local`, `instance_stats` | Written on whichever node observes the event; never replicated. |
| Reference / seed (local, identical everywhere) | `event_type` | Populated by migrations on every node; not replicated and not written at runtime. |

### 4.3 Code model

Go packages are reshaped to make the admin/runtime boundary a compile-time property:

- `pkg/api/admin/` — admin-table writers, used only on the control node.
- `pkg/api/runtime/` — runtime-table writers, used on every node.
- `pkg/api/dbreads/` — shared read queries, embedded by both services above.

`runtime/` cannot import `admin/` and vice versa. Cross-package helpers live in `pkg/api/internal/shared/`.

```
pkg/api/
├── admin/              # writers for admin tables (control node only)
│   ├── service.go
│   ├── applications.go
│   ├── channels.go
│   ├── groups.go
│   ├── packages.go
│   ├── teams.go
│   ├── users.go
│   ├── actions.go
│   └── admin_activity.go
├── runtime/            # writers for runtime tables (all nodes)
│   ├── service.go
│   ├── events.go
│   ├── instances.go
│   ├── updates.go
│   ├── group_local.go
│   └── activity.go
├── dbreads/            # shared read queries (all nodes)
│   ├── queries.go
│   ├── applications.go
│   ├── channels.go
│   ├── groups.go
│   ├── instances.go
│   ├── packages.go
│   └── activity.go
├── internal/shared/    # constants and helpers shared across sub-packages
├── api.go              # core types, DB connection, migrations
└── db/migrations/
```

The admin/runtime boundary is enforced at three levels:

1. **Compile-time** — Go package boundaries prevent `runtime/` from calling `admin/` methods.
2. **HTTP** — a `requirePrimary` middleware on the admin handlers returns 403 on edge nodes.
3. **Database** — the runtime PG role has no write grants on admin tables, so even bypassing the first two layers fails at the database.

### 4.4 Instance mode

A single env var, `NEBRASKA_INSTANCE_MODE`, with a small validated allowlist:

| Value | Meaning |
|---|---|
| unset or `single` | Single-instance (today's behaviour) |
| `control` | Control node |
| `edge` | Edge node |

Any other value fails the process at startup with a clear error so we are sure the node starts with an explicit setting for the node type.

Edge nodes:
- Return HTTP 403 on admin endpoints via a `requirePrimary` middleware.
- Do not start the syncer.

### 4.5 Database authentication

Nebraska does not create, drop, or rotate Postgres roles. The operator provisions all roles ahead of deployment. Nebraska reads connection strings from environment variables.

**Two-layer role model.** Migrations attach grants to a fixed set of `NOLOGIN` logical roles whose names Nebraska guarantees. The operator creates their own `LOGIN` roles with whatever names they like and adds them `IN ROLE` the logical ones; PG role inheritance carries the grants through. Nebraska only requires that the logical roles exist and that whatever login role the DSN connects as inherits from them.

**Logical roles** (fixed names, `NOLOGIN`, no `CREATEROLE`, no `SUPERUSER`):

| Logical role | Purpose | Privileges (granted by migrations) |
|---|---|---|
| `nebraska_admin` | Logical role with write access to admin tables. | `USAGE` on schema. `INSERT, UPDATE, DELETE` on every admin table. `SELECT` on every table. `USAGE, SELECT` on every sequence. Inherits `nebraska_runtime` via `GRANT nebraska_runtime TO nebraska_admin`, because the control node also serves Omaha for its own region and runs the rollout engine — both write runtime tables. |
| `nebraska_runtime` | Logical role with write access to runtime tables. | `USAGE` on schema. `INSERT, UPDATE, DELETE` on every runtime table. `SELECT` on every table. `USAGE, SELECT` on every sequence (required for `serial` PK inserts). No write privileges on admin tables. |

**Operator login roles** (operator-named, `LOGIN`):

- One login role that owns the schema (`CREATE`, `USAGE` on `public`, and owns every Nebraska table) — used by `NEBRASKA_MIGRATIONS_DB_URL`. No membership in the logical roles required.
- One login role `IN ROLE nebraska_admin` — used by `NEBRASKA_DB_URL` on the control node.
- One login role `IN ROLE nebraska_runtime` — used by `NEBRASKA_DB_URL` on each edge node.

The operator may also point multiple regional login roles at the same logical role (e.g. one `IN ROLE nebraska_runtime` login per edge region, each with its own password rotation cadence).

**Grants live next to the schema, in migrations.** Each migration that adds an admin or runtime table includes the matching `GRANT INSERT, UPDATE, DELETE ON … TO nebraska_admin | nebraska_runtime` in the same file, wrapped in `IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = …)`. Per-table grants and `ALTER DEFAULT PRIVILEGES … GRANT USAGE, SELECT ON SEQUENCES` for both serving roles are added in the same migration. The schema and the grants are added and reviewed together; drift is less likely by construction. In single-instance deployments where the logical roles do not exist, every grant is a silent no-op.

**Connection strings Nebraska reads:**

- `NEBRASKA_DB_URL` (required) — long-lived serving pool. Operator points it at their admin login role on the control node, runtime login role on edges, or any role with full privileges in single-instance mode.
- `NEBRASKA_MIGRATIONS_DB_URL` (required in distributed mode, optional in single-instance mode) — short-lived connection used to run migrations at startup, then closed. Operator points it at their schema-owner login role in distributed mode. When unset, the serving DSN is used for migrations too (today's single-instance behaviour, unchanged). Nebraska fails closed at startup if `NEBRASKA_INSTANCE_MODE` is `control` or `edge` and this is unset.


### 4.6 Schema migration discipline

Logical replication does not replicate DDL. Migrations are applied to every node, in order, with an **expand-then-contract** discipline:

- Additive changes (new tables, new nullable columns): edge nodes first, then control.
- Subtractive changes (drops, tightened constraints): control first, then edge nodes.

The invariant is that the **subscriber schema is always a superset** of what the primary currently writes — which is exactly what logical replication needs.

### 4.7 Setting up logical replication

Logical replication itself is **the operator's responsibility**, not Nebraska's — standard PostgreSQL setup, no Nebraska-specific tooling required. After Nebraska has started and applied its migrations on every node, the operator runs:

- On the **control DB**: `CREATE PUBLICATION nebraska_admin FOR TABLE <admin tables>;` (the list mirrors the admin set documented in this file).
- On **each edge DB**: `CREATE SUBSCRIPTION nebraska_admin CONNECTION '<control DSN>' PUBLICATION nebraska_admin;`

From there, replication runs in the background. The operator is also responsible for monitoring replication lag and slot retention on the control DB, and for following the migration discipline above when applying schema changes across the fleet. See <https://www.postgresql.org/docs/current/logical-replication.html> for the standard runbook.


## 5. What's different in distributed mode (operator-visible)

Single-instance Nebraska is unchanged. When `NEBRASKA_INSTANCE_MODE` is set, the operator-visible behaviour changes in four ways.

**1. Admin actions are accepted only on the control node.** Creating apps, channels, groups, or packages, editing default policies, and any other admin write returns HTTP 403 on edge nodes. There is one place to make admin changes, and that place is the control node. Admin metadata replicates to every edge automatically.

**2. The activity feed is per-node.** Admin events (e.g. setting a channel's package) replicate from the control node to every edge, so the same admin entry appears in every node's activity feed. Runtime events (rollout lifecycle, instance update results) stay on the node that observed them. So the control node's feed shows the admin events it generated plus its own runtime events; each edge's feed shows the replicated admin events plus its own local runtime events. No node sees runtime events from other nodes. Fleet-wide visibility, if needed, lives outside Nebraska.

**3. Safe-mode auto-pause is per-node.** If an edge's rollout fails enough to trip the safe-mode brake, the brake is local to that edge. Other edges and the control node keep serving updates from their own buckets. The operator clears the brake on the affected edge directly through a new runtime endpoint that is intentionally available on edges. The control node's "enable updates" action still works as a global default lever: it replicates the admin intent to every node, but does not override a local brake until the operator explicitly clears it on that edge.

**4. The group view in the UI has two panels on the control node, one on edges.** On the control node, the operator sees the admin defaults (replicated to every edge) and, separately, the control node's own local overrides. On an edge, the operator sees only that edge's local overrides. The effective value at any node is "the local override if set, otherwise the admin default". This makes it explicit which surface they're editing and what scope it has.

There is no aggregated fleet view in Nebraska itself in this iteration: the control node does not show which edges have tripped brakes, and there is no central "clear everywhere" button beyond the admin default. Fleet-wide observability is intentionally an operator-side concern.

## 6. Rollout plan (PR series)

Each PR is independently mergeable. Single-instance behaviour is preserved throughout.

1. `group_local` sidecar split. [#1396](https://github.com/flatcar/nebraska/pull/1396)
2. `activity` row-level split + `all_activity` view. [#1398](https://github.com/flatcar/nebraska/pull/1398)
3. Extract `pkg/api/dbreads/`.
4. Split writes into `pkg/api/admin/` and `pkg/api/runtime/`.
5. Conditional role grants migration + two-DSN migration/serving split (`NEBRASKA_MIGRATIONS_DB_URL`).
6. Instance mode: `NEBRASKA_INSTANCE_MODE` + `requirePrimary` middleware + validated allowlist.
7. Per-edge override management endpoint.
8. Operator documentation.

Independent track: OIDC `aud` validation patch.

## 7. References

- RFC: [#1375](https://github.com/flatcar/nebraska/issues/1375)
- PoC: [Moustafa-Moustafa/nebraska#1](https://github.com/Moustafa-Moustafa/nebraska/pull/1)
- Postgres logical replication: <https://www.postgresql.org/docs/current/logical-replication.html>
- Expand-and-contract migration pattern: <https://www.tim-wellhausen.de/papers/ExpandAndContract/ExpandAndContract.html>
