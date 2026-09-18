-- Provisions the control and edge nodes used by the distributed-mode tests.
--
-- The single node uses the database the image creates from POSTGRES_DB. A
-- control or edge node needs a database of its own and two distinct logins,
-- because those modes require NEBRASKA_MIGRATIONS_DB_URL to name a different
-- user than NEBRASKA_DB_URL.
--
-- Nebraska creates the nebraska_admin_<database> and nebraska_runtime_<database>
-- group roles itself and grants the serving user whichever one its mode calls
-- for, so only the databases and the logins belong here. CREATEROLE is what
-- lets Nebraska create those group roles.

create role control_mig login password 'nebraska' createrole;
create role control_srv login password 'nebraska';
create database nebraska_control owner control_mig;

create role edge_mig login password 'nebraska' createrole;
create role edge_srv login password 'nebraska';
create database nebraska_edge owner edge_mig;
