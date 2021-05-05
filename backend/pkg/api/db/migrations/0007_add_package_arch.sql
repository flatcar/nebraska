-- +migrate Up

-- see arch.go for meaning of a numeric value of an arch

alter table package add column arch int not null default 0;
-- the constraint is not explicitly named (so the name is generated),
-- hopefully this name will work for most versions of the postgresql
alter table package drop constraint package_application_id_version_key;
alter table package add constraint package_appid_version_arch_unique unique(application_id, version, arch);

alter table channel add column arch int not null default 0;
-- the constraint is not explicitly named (so the name is generated),
-- hopefully this name will work for most versions of the postgresql
alter table channel drop constraint channel_application_id_name_key;
alter table channel add constraint channel_appid_name_arch_unique unique(application_id, name, arch);

-- update all our initial flatcar channels to amd64
update channel set arch = 1 where application_id = 'e96281a6-d1af-4bde-9a0a-97b76e56dc57';

-- update all the packages advertised by our channels to amd64
update package set arch = 1 where application_id = 'e96281a6-d1af-4bde-9a0a-97b76e56dc57';

-- +migrate Down

alter table channel drop constraint channel_appid_name_arch_unique;
alter table channel add constraint channel_application_id_name_key unique(application_id, name);
alter table channel drop column arch;

alter table package drop constraint package_appid_version_arch_unique;
alter table package add constraint package_application_id_version_key unique(application_id, version);
alter table package drop column arch;
