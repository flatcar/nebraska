# Instance retention

Nebraska records every machine that checks in. Those rows are never removed
unless you enable this optional pruner.

Kubernetes clusters that replace Flatcar nodes often leave years of dead
`instance` rows behind, along with `instance_application`,
`instance_status_history`, `event`, and `activity` rows. That growth is what
makes reporting queries and indexes more expensive over time.

Retention is **off by default**. Existing deployments do not change behaviour
on upgrade.

## Enable it

Pick instances by their newest `instance_application.last_check_for_updates`.
If a machine never registered an application, `instance.created_ts` is used
instead. An instance that still checks in for any application is kept.

```text
nebraska \
  -auth-mode noop \
  -instance-retention=2160h \
  -instance-retention-batch-size=500
```

`2160h` is 90 days. Use whatever window matches how long a decommissioned
node should stay visible.

Helm users can pass the same flags through `extraArgs`.

## Dry-run first

Log the would-delete count without removing anything:

```text
nebraska \
  -auth-mode noop \
  -instance-retention=2160h \
  -instance-retention-dry-run
```

Watch the log line `instance retention dry-run; no rows deleted` and the
`candidates` field. Turn dry-run off only after that number looks right.

The job runs once at startup and then every hour, next to the existing
`instance_stats` updater.

## What is deleted

Each pass deletes up to `instance-retention-batch-size` rows from `instance`
(`500` by default) and repeats until none remain. Foreign keys on related
tables are `ON DELETE CASCADE`, so history, events, and activity for those
machines go with them.

`instance_stats` is an hourly aggregate table, not per-instance, and is not
pruned by this job.

## Metric

`nebraska_instances_pruned_total` counts rows actually deleted. Dry-run does
not increment it.
