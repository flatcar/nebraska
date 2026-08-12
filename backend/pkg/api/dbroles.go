package api

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jmoiron/sqlx"
)

// The logical roles. Role identity is cluster-global while table
// privileges are per-database, so the names carry the database to keep two
// Nebraska databases in one cluster from sharing a role. db/grants.sql is handed
// the names rather than deriving them a second time.
const (
	adminRolePrefix   = "nebraska_admin_"
	runtimeRolePrefix = "nebraska_runtime_"
)

// maxIdentifier is PostgreSQL's NAMEDATALEN - 1. Anything longer is truncated
// where it is used, which would quietly merge the roles of two databases whose
// names share a long prefix.
const maxIdentifier = 63

// provisioningLock serializes the steps below across instances starting at the
// same time: PostgreSQL has no CREATE ROLE IF NOT EXISTS, and concurrent grants
// collide on the catalog with "tuple concurrently updated". The number is
// arbitrary; advisory locks are scoped to the database.
const provisioningLock = 1375

// servingRoles are the logical roles for one database.
type servingRoles struct {
	admin   string
	runtime string
}

func newServingRoles(db *sqlx.DB) (servingRoles, error) {
	database, err := currentDatabase(db)
	if err != nil {
		return servingRoles{}, err
	}

	return servingRoles{
		admin:   roleName(adminRolePrefix, database),
		runtime: roleName(runtimeRolePrefix, database),
	}, nil
}

// roleName keeps the derived name within maxIdentifier, ending it with a digest
// of the database name so that two databases sharing a long prefix still get
// their own roles.
func roleName(prefix, database string) string {
	name := prefix + database
	if len(name) <= maxIdentifier {
		return name
	}

	sum := sha256.Sum256([]byte(database))
	suffix := "_" + hex.EncodeToString(sum[:4])

	// Cutting on a byte boundary can split a character, which PostgreSQL rejects.
	return strings.ToValidUTF8(name[:maxIdentifier-len(suffix)], "") + suffix
}

// checkSameDeployment rejects a migrations connection to another database or to
// another server. The two DSNs are configured independently, and the database
// name alone does not distinguish them because environments usually share it.
func checkSameDeployment(servingDB, migrationsDB *sqlx.DB) error {
	serving, err := currentDeployment(servingDB)
	if err != nil {
		return err
	}

	migrations, err := currentDeployment(migrationsDB)
	if err != nil {
		return err
	}

	if serving != migrations {
		return fmt.Errorf("NEBRASKA_MIGRATIONS_DB_URL uses %s but the serving connection uses %s", migrations, serving)
	}

	return nil
}

// setupServingRoles provisions the serving roles and applies the grant policy.
// Both are skipped when the serving identity is the one running migrations: it
// owns the schema and bypasses every grant anyway, which is the single-instance
// case.
func (api *API) setupServingRoles(migrationsDB *sqlx.DB) error {
	servingUser, err := currentUser(api.db())
	if err != nil {
		return err
	}

	migrationsUser, err := currentUser(migrationsDB)
	if err != nil {
		return err
	}

	if servingUser == migrationsUser {
		return nil
	}

	// The migrations pool is capped at one connection, so this lock and the work
	// it guards share a session; closing that connection would release it anyway.
	if _, err := migrationsDB.Exec("select pg_advisory_lock($1)", provisioningLock); err != nil {
		return fmt.Errorf("locking the serving role provisioning: %w", err)
	}

	defer func() {
		_, _ = migrationsDB.Exec("select pg_advisory_unlock($1)", provisioningLock)
	}()

	roles, err := newServingRoles(migrationsDB)
	if err != nil {
		return err
	}

	if err := ensureLogicalRoles(migrationsDB, roles); err != nil {
		return err
	}

	granted, err := grantServingRole(migrationsDB, servingUser, roles.admin)
	if err != nil {
		return err
	}

	if granted {
		l.Info().Str("role", roles.admin).Str("user", servingUser).Msg("granted the serving database role")
	}

	return applyGrants(migrationsDB, roles)
}

func currentUser(db *sqlx.DB) (string, error) {
	var user string
	if err := db.Get(&user, "select current_user"); err != nil {
		return "", fmt.Errorf("reading the current database user: %w", err)
	}

	return user, nil
}

func currentDatabase(db *sqlx.DB) (string, error) {
	var database string
	if err := db.Get(&database, "select current_database()"); err != nil {
		return "", fmt.Errorf("reading the current database name: %w", err)
	}

	return database, nil
}

// currentDeployment names one database on one server. system_identifier is
// assigned at initdb and is readable by an unprivileged role.
func currentDeployment(db *sqlx.DB) (string, error) {
	var deployment string
	if err := db.Get(&deployment,
		"select current_database()::text || ' on cluster ' || system_identifier::text from pg_control_system()"); err != nil {
		return "", fmt.Errorf("reading the database identity: %w", err)
	}

	return deployment, nil
}

// ensureLogicalRoles creates the roles when they are absent. The lookup and the
// create are not atomic, so callers hold provisioningLock.
func ensureLogicalRoles(db *sqlx.DB, roles servingRoles) error {
	for _, role := range []string{roles.runtime, roles.admin} {
		var exists bool
		if err := db.Get(&exists, "select exists (select 1 from pg_roles where rolname = $1)", role); err != nil {
			return fmt.Errorf("looking up the %s role: %w", role, err)
		}

		if exists {
			continue
		}

		if _, err := db.Exec("create role " + pgx.Identifier{role}.Sanitize() + " nologin"); err != nil {
			return fmt.Errorf("creating the %s role: %w; the NEBRASKA_MIGRATIONS_DB_URL role needs CREATE ROLE", role, err)
		}
	}

	return nil
}

// grantServingRole makes the serving user a member of the logical role carrying
// its privileges, and reports whether it had to. Once NEBRASKA_INSTANCE_MODE
// exists this picks the runtime role on edge nodes, and the admin role on
// control and single nodes.
func grantServingRole(db *sqlx.DB, user, role string) (bool, error) {
	var member bool
	if err := db.Get(&member, "select pg_has_role($1, $2, 'member')", user, role); err != nil {
		return false, fmt.Errorf("checking whether %s is a member of %s: %w", user, role, err)
	}

	if member {
		return false, nil
	}

	stmt := fmt.Sprintf("grant %s to %s", pgx.Identifier{role}.Sanitize(), pgx.Identifier{user}.Sanitize())
	if _, err := db.Exec(stmt); err != nil {
		return false, fmt.Errorf("granting %s to %s: %w; run that grant as a role that can administer %s", role, user, err, role)
	}

	return true, nil
}

func applyGrants(db *sqlx.DB, roles servingRoles) error {
	sqlFile, err := sqlFolder.ReadFile("db/grants.sql")
	if err != nil {
		return err
	}

	tx, err := db.Beginx()
	if err != nil {
		return fmt.Errorf("applying the serving role grants: %w", err)
	}

	defer func() { _ = tx.Rollback() }()

	// Transaction-scoped, so the policy reads the names this code derived rather
	// than deriving them again in SQL.
	if _, err := tx.Exec("select set_config('nebraska.admin_role', $1, true), set_config('nebraska.runtime_role', $2, true)",
		roles.admin, roles.runtime); err != nil {
		return fmt.Errorf("naming the serving roles for the grant policy: %w", err)
	}

	if _, err := tx.Exec(string(sqlFile)); err != nil {
		return fmt.Errorf("applying the serving role grants: %w", err)
	}

	return tx.Commit()
}
