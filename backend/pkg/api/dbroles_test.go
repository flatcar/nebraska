package api

import (
	"fmt"
	"net/url"
	"os"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/flatcar/nebraska/backend/pkg/random"
)

// The admin/runtime split, declared here independently of db/grants.sql so that
// drift in either one is caught.
var (
	adminTables = []string{
		"admin_activity", "application", "channel", "channel_package_floors",
		"flatcar_action", "groups", "package", "package_channel_blacklist",
		"package_file", "team", "users",
	}

	runtimeTables = []string{
		"activity", "event", "group_local", "instance", "instance_application",
		"instance_stats", "instance_status_history",
	}

	// Written only by migrations, or not writable at all, so neither serving role
	// may write them.
	unmanagedTables = []string{"all_activity", "database_migrations", "event_type", "instance_status"}
)

const privTestGroupID = "8c0a9a5e-0e8c-4a3c-9f2d-1b5e6a7c8d90"

// TestRoleNameBounds pins the identifier limit. Over it, PostgreSQL truncates
// silently, which used to hand two databases the same roles and leave the grant
// policy skipping a role it could not find.
func TestRoleNameBounds(t *testing.T) {
	assert.Equal(t, "nebraska_admin_nebraska", roleName(adminRolePrefix, "nebraska"))

	long := strings.Repeat("a", 60)
	sibling := long[:59] + "b"

	for _, prefix := range []string{adminRolePrefix, runtimeRolePrefix} {
		name := roleName(prefix, long)
		assert.LessOrEqual(t, len(name), maxIdentifier, "%s exceeds the identifier limit", name)
		assert.Equal(t, name, roleName(prefix, long), "the name must be stable across startups")
		assert.NotEqual(t, name, roleName(prefix, sibling), "databases sharing a prefix must not share a role")
	}

	assert.NotEqual(t, roleName(adminRolePrefix, long), roleName(runtimeRolePrefix, long))

	multibyte := roleName(adminRolePrefix, strings.Repeat("é", 40))
	assert.LessOrEqual(t, len(multibyte), maxIdentifier)
	assert.True(t, utf8.ValidString(multibyte), "truncation must not split a character")
}

func TestTableClassification(t *testing.T) {
	a := newForTest(t)
	t.Cleanup(a.Close)

	var tables []string
	require.NoError(t, a.db().Select(&tables, `select c.relname from pg_class c
		join pg_namespace n on n.oid = c.relnamespace
		where n.nspname = 'public' and c.relkind in ('r', 'v', 'm', 'p', 'f') order by 1`))

	classified := map[string]bool{}
	for _, group := range [][]string{adminTables, runtimeTables, unmanagedTables} {
		for _, tbl := range group {
			require.False(t, classified[tbl], "%s is classified twice", tbl)
			classified[tbl] = true
		}
	}

	for _, tbl := range tables {
		assert.True(t, classified[tbl],
			"relation %q is unclassified: add it to db/grants.sql and to this test", tbl)
	}
	assert.Len(t, classified, len(tables), "this test names relations that do not exist")
}

func TestServingRoleGrants(t *testing.T) {
	a := newForTest(t)
	t.Cleanup(a.Close)
	requireRoleAdmin(t, a)
	roles := withServingRoles(t, a)

	can := func(role, table, priv string) bool {
		var got bool
		require.NoError(t, a.db().Get(&got,
			"select has_table_privilege($1, $2, $3)", role, "public."+table, priv))
		return got
	}

	writes := []string{"INSERT", "UPDATE", "DELETE"}

	for _, tbl := range adminTables {
		assert.True(t, can(roles.runtime, tbl, "SELECT"), "runtime must read %s", tbl)
		for _, priv := range writes {
			assert.False(t, can(roles.runtime, tbl, priv), "runtime must not %s %s", priv, tbl)
			assert.True(t, can(roles.admin, tbl, priv), "admin must %s %s", priv, tbl)
		}
	}

	for _, tbl := range runtimeTables {
		for _, priv := range writes {
			assert.True(t, can(roles.runtime, tbl, priv), "runtime must %s %s", priv, tbl)
			assert.True(t, can(roles.admin, tbl, priv), "admin must %s %s", priv, tbl)
		}
	}

	for _, tbl := range unmanagedTables {
		assert.True(t, can(roles.runtime, tbl, "SELECT"), "runtime must read %s", tbl)
		for _, priv := range writes {
			assert.False(t, can(roles.runtime, tbl, priv), "runtime must not %s %s", priv, tbl)
			assert.False(t, can(roles.admin, tbl, priv), "admin must not %s %s", priv, tbl)
		}
	}

	for _, tbl := range append(append([]string{}, adminTables...), runtimeTables...) {
		assert.False(t, can(roles.runtime, tbl, "TRUNCATE"), "no role may truncate %s", tbl)
		assert.False(t, can(roles.admin, tbl, "TRUNCATE"), "no role may truncate %s", tbl)
	}

	canSeq := func(role, seq string) bool {
		var got bool
		require.NoError(t, a.db().Get(&got, "select has_sequence_privilege($1, $2, 'USAGE')", role, seq))
		return got
	}

	assert.True(t, canSeq(roles.runtime, "event_id_seq"), "runtime inserts into event need its sequence")
	assert.False(t, canSeq(roles.runtime, "event_type_id_seq"), "event_type is not writable, nor is its sequence")
	assert.False(t, canSeq(roles.admin, "event_type_id_seq"), "event_type is not writable, nor is its sequence")
}

func TestServingRoleEnforcement(t *testing.T) {
	a := newForTest(t)
	t.Cleanup(a.Close)
	requireRoleAdmin(t, a)
	roles := withServingRoles(t, a)

	edge := loginRole(t, a, roles.runtime)

	// Inserting a zero-row select still requires the privilege, so this works
	// uniformly for every table without tripping over constraints.
	for _, tbl := range adminTables {
		requireDenied(t, edge, fmt.Sprintf("insert into %s select * from %s where false", tbl, tbl))
		requireDenied(t, edge, fmt.Sprintf("delete from %s where false", tbl))
	}

	for _, tbl := range runtimeTables {
		_, err := edge.Exec(fmt.Sprintf("insert into %s select * from %s where false", tbl, tbl))
		require.NoError(t, err, "runtime role must write %s", tbl)
	}

	var activity int
	require.NoError(t, edge.Get(&activity, "select count(*) from all_activity"))

	// A real insert, to exercise the sequence behind event.id.
	_, err := edge.Exec("insert into instance (id, ip) values ('privtest', '10.0.0.1') on conflict (id) do nothing")
	require.NoError(t, err)
	_, err = edge.Exec("insert into event (instance_id, application_id, event_type_id) values ('privtest', $1, 1)", flatcarAppID)
	require.NoError(t, err)

	requireDenied(t, edge, "create table privtest_ddl (id int)")

	// Creating a group fires the group_local trigger as the calling user, so the
	// admin role needs the runtime tables even though it never writes them itself.
	ctl := loginRole(t, a, roles.admin)

	_, err = ctl.Exec(`insert into groups (id, name, description, policy_period_interval,
		policy_max_updates_per_period, policy_update_timeout, application_id, track)
		values ($1, 'privtest', 'privtest', '15 minutes', 10, '60 minutes', $2, 'privtest')`,
		privTestGroupID, flatcarAppID)
	require.NoError(t, err)

	var sidecar int
	require.NoError(t, ctl.Get(&sidecar, "select count(*) from group_local where group_id = $1", privTestGroupID))
	assert.Equal(t, 1, sidecar, "the groups trigger must be able to write group_local")
}

// TestServingRoleProvisioning drives NewWithMigrations so the wiring in
// setupServingRoles is covered, not just the helpers it calls.
func TestServingRoleProvisioning(t *testing.T) {
	a := newForTest(t)
	t.Cleanup(a.Close)
	requireRoleAdmin(t, a)
	roles := requireNoServingRoles(t, a)

	migrationsURL := testDBURL()
	name, password := createLoginRole(t, a)

	t.Setenv("NEBRASKA_MIGRATIONS_DB_URL", migrationsURL)
	t.Setenv("NEBRASKA_DB_URL", loginURL(t, name, password))

	served, err := NewWithMigrations()
	require.NoError(t, err)
	t.Cleanup(served.Close)

	for _, role := range []string{roles.admin, roles.runtime} {
		var login bool
		require.NoError(t, a.db().Get(&login,
			"select rolcanlogin or rolsuper or rolcreaterole from pg_roles where rolname = $1", role))
		assert.False(t, login, "%s must be a plain nologin group role", role)
	}

	var member bool
	require.NoError(t, a.db().Get(&member, "select pg_has_role($1, $2, 'usage')", name, roles.admin))
	assert.True(t, member, "the serving user must inherit %s", roles.admin)

	for _, priv := range []string{"SELECT", "INSERT", "UPDATE", "DELETE"} {
		var got bool
		require.NoError(t, a.db().Get(&got,
			"select has_table_privilege($1, 'public.application', $2)", name, priv))
		assert.True(t, got, "the serving user must %s application", priv)
	}
}

// TestServingRoleGrantAfterMigrationsRotation pins that rotating the migrations
// identity does not wedge startup. Re-granting a membership is not a no-op to
// PostgreSQL, it still needs ADMIN OPTION, which only the role that created the
// membership holds.
func TestServingRoleGrantAfterMigrationsRotation(t *testing.T) {
	a := newForTest(t)
	t.Cleanup(a.Close)
	requireRoleAdmin(t, a)
	roles := requireNoServingRoles(t, a)

	require.NoError(t, ensureLogicalRoles(a.db(), roles))

	serving, _ := createLoginRole(t, a)

	granted, err := grantServingRole(a.db(), serving, roles.admin)
	require.NoError(t, err)
	assert.True(t, granted, "the first start must issue the grant")

	name, password := createLoginRole(t, a)
	rotated, err := sqlx.Open("pgx", loginURL(t, name, password))
	require.NoError(t, err)
	require.NoError(t, rotated.Ping())

	t.Cleanup(func() { _ = rotated.Close() })

	granted, err = grantServingRole(rotated, serving, roles.admin)
	require.NoError(t, err, "a rotated migrations user must not re-issue the grant")
	assert.False(t, granted)

	_, err = grantServingRole(rotated, serving, roles.runtime)
	require.Error(t, err, "a membership that does not exist yet still needs ADMIN OPTION")
	assert.Contains(t, err.Error(), "administer "+roles.runtime)
}

// TestServingRoleNotProvisionedWithoutMigrationsDB pins the opt-in: the logical
// roles are cluster-global, so their presence must not be enough to grant.
func TestServingRoleNotProvisionedWithoutMigrationsDB(t *testing.T) {
	a := newForTest(t)
	t.Cleanup(a.Close)
	requireRoleAdmin(t, a)
	roles := requireNoServingRoles(t, a)

	require.NoError(t, ensureLogicalRoles(a.db(), roles))

	t.Setenv("NEBRASKA_MIGRATIONS_DB_URL", "")

	served, err := NewWithMigrations()
	require.NoError(t, err)
	t.Cleanup(served.Close)

	for _, role := range []string{roles.admin, roles.runtime} {
		var granted int
		require.NoError(t, a.db().Get(&granted, `select count(*) from pg_class c
			join pg_namespace n on n.oid = c.relnamespace
			where n.nspname = 'public' and c.relkind in ('r', 'v', 'm', 'p', 'f')
			  and has_table_privilege($1, c.oid, 'select')`, role))
		assert.Zero(t, granted, "%s must hold nothing without a migrations connection", role)
	}
}

// TestMigrationsDatabaseMismatch pins that a migrations DSN naming another
// database is rejected before any DDL reaches it.
func TestMigrationsDatabaseMismatch(t *testing.T) {
	a := newForTest(t)
	t.Cleanup(a.Close)

	serving, err := currentDatabase(a.db())
	require.NoError(t, err)

	other := "postgres"
	if serving == other {
		other = "template1"
	}

	otherURL := databaseURL(t, other)

	t.Setenv("NEBRASKA_MIGRATIONS_DB_URL", otherURL)
	t.Setenv("NEBRASKA_DB_URL", testDBURL())

	_, err = NewWithMigrations()
	require.Error(t, err)
	assert.Contains(t, err.Error(), other)
	assert.Contains(t, err.Error(), serving)

	db, err := sqlx.Open("pgx", otherURL)
	require.NoError(t, err)

	t.Cleanup(func() { _ = db.Close() })

	var absent bool
	require.NoError(t, db.Get(&absent, "select to_regclass('public."+migrationsTable+"') is null"))
	assert.True(t, absent, "the mismatch must be caught before any DDL reaches %s", other)
}

func requireRoleAdmin(t *testing.T, a *API) {
	t.Helper()

	var ok bool
	require.NoError(t, a.db().Get(&ok,
		"select rolsuper or rolcreaterole from pg_roles where rolname = current_user"))
	if !ok {
		t.Skip("the test database user cannot create roles")
	}
}

// requireNoServingRoles skips rather than touching roles that already exist, so
// running the suite against a real deployment cannot disturb one.
func requireNoServingRoles(t *testing.T, a *API) servingRoles {
	t.Helper()

	roles, err := newServingRoles(a.db())
	require.NoError(t, err)

	for _, role := range []string{roles.admin, roles.runtime} {
		var exists bool
		require.NoError(t, a.db().Get(&exists,
			"select exists (select 1 from pg_roles where rolname = $1)", role))
		if exists {
			t.Skipf("role %s already exists in this cluster", role)
		}
	}

	t.Cleanup(func() {
		names := pgx.Identifier{roles.admin}.Sanitize() + ", " + pgx.Identifier{roles.runtime}.Sanitize()
		_, _ = a.db().Exec("drop owned by " + names)
		_, _ = a.db().Exec("drop role " + names)
	})

	return roles
}

func withServingRoles(t *testing.T, a *API) servingRoles {
	t.Helper()

	roles := requireNoServingRoles(t, a)
	require.NoError(t, ensureLogicalRoles(a.db(), roles))
	require.NoError(t, applyGrants(a.db(), roles))

	return roles
}

func createLoginRole(t *testing.T, a *API) (string, string) {
	t.Helper()

	name := "nebraska_test_" + strings.ToLower(random.String(10))
	password := strings.ToLower(random.String(24))
	quoted := pgx.Identifier{name}.Sanitize()

	_, err := a.db().Exec(fmt.Sprintf("create role %s login password '%s'", quoted, password))
	require.NoError(t, err)

	t.Cleanup(func() {
		_, _ = a.db().Exec("drop owned by " + quoted)
		_, _ = a.db().Exec("drop role " + quoted)
	})

	return name, password
}

func loginRole(t *testing.T, a *API, memberOf string) *sqlx.DB {
	t.Helper()

	name, password := createLoginRole(t, a)

	_, err := a.db().Exec("grant " + pgx.Identifier{memberOf}.Sanitize() + " to " + pgx.Identifier{name}.Sanitize())
	require.NoError(t, err)

	db, err := sqlx.Open("pgx", loginURL(t, name, password))
	require.NoError(t, err)
	require.NoError(t, db.Ping())

	t.Cleanup(func() { _ = db.Close() })

	return db
}

func loginURL(t *testing.T, user, password string) string {
	t.Helper()

	parsed, err := url.Parse(testDBURL())
	require.NoError(t, err)
	parsed.User = url.UserPassword(user, password)

	return parsed.String()
}

func databaseURL(t *testing.T, database string) string {
	t.Helper()

	parsed, err := url.Parse(testDBURL())
	require.NoError(t, err)
	parsed.Path = "/" + database

	return parsed.String()
}

func testDBURL() string {
	if raw := os.Getenv("NEBRASKA_DB_URL"); raw != "" {
		return raw
	}

	return defaultTestDbURL
}

func requireDenied(t *testing.T, db *sqlx.DB, stmt string) {
	t.Helper()

	_, err := db.Exec(stmt)
	require.Error(t, err, "expected %q to be denied", stmt)

	var pgErr *pgconn.PgError
	require.ErrorAs(t, err, &pgErr)
	assert.Equal(t, "42501", pgErr.Code, "expected insufficient_privilege for %q", stmt)
}
