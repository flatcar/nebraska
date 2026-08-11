-- +migrate Up

-- Widen users.secret so it can hold a bcrypt hash (60 characters, e.g.
-- "$2a$10$..."), which replaces the raw MD5 digest (32 hex characters)
-- previously stored here. Existing legacy MD5 secrets keep working
-- unchanged: bcrypt's own output always starts with "$2", so the
-- value's own shape tells the two formats apart and no separate
-- scheme-marker column is needed. See backend/pkg/api/admin/users.go.
alter table users alter column secret type text;

-- +migrate Down

-- Only safe if no bcrypt secret (>50 chars) has been written since the
-- Up migration ran.
alter table users alter column secret type varchar(50);
