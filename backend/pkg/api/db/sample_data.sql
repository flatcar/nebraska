-- Initial data

-- Drop data added in migration routines

-- added in 0006_initial_application.sql
delete from groups;
delete from channel;
delete from package;
delete from application;

-- added in 0005_default_team_id.sql
delete from team;

-- Default team and user (admin/admin)
insert into team (id, name) values ('d89342dc-9214-441d-a4af-bdd837a3b239', 'default');
insert into users (username, secret, team_id) values ('admin', '8b31292d4778582c0e5fa96aee5513f1', 'd89342dc-9214-441d-a4af-bdd837a3b239');

-- Flatcar Container Linux application
insert into application (id, name, description, team_id) values ('e96281a6-d1af-4bde-9a0a-97b76e56dc57', 'Flatcar Container Linux', 'Linux for massive server deployments', 'd89342dc-9214-441d-a4af-bdd837a3b239');
insert into package values ('2ba4c984-5e9b-411e-b7c3-b3eb14f7a261', 1, '766.3.0', 'https://update.release.flatcar-linux.net/amd64-usr/766.3.0/', 'flatcar_production_update.gz', NULL, '154967458', 'l4Kw7AeBLrVID9JbfyMoJeB5yKg=', '2019-08-20 00:12:37.523938', 'e96281a6-d1af-4bde-9a0a-97b76e56dc57', 1);
insert into package values ('337b3f7e-ff29-47e8-a052-f0834d25bdb5', 1, '766.4.0', 'https://update.release.flatcar-linux.net/amd64-usr/766.4.0/', 'flatcar_production_update.gz', NULL, '155018912', 'frkka+B/zTv7OPWgidY+k4SnDSg=', '2019-08-20 06:15:29.108266', 'e96281a6-d1af-4bde-9a0a-97b76e56dc57', 1);
insert into package values ('c2a36312-b989-403e-ab57-06c055a7eac2', 1, '808.0.0', 'https://update.release.flatcar-linux.net/amd64-usr/808.0.0/', 'flatcar_production_update.gz', NULL, '177717414', 'bq3fQRHP8xB3RFUjCdAf3wQYC2E=', '2019-08-20 00:09:06.839989', 'e96281a6-d1af-4bde-9a0a-97b76e56dc57', 1);
insert into package values ('43580892-cad8-468a-a0bb-eb9d0e09eca4', 1, '815.0.0', 'https://update.release.flatcar-linux.net/amd64-usr/815.0.0/', 'flatcar_production_update.gz', NULL, '178643579', 'kN4amoKYVZUG2WoSdQH1PHPzr5A=', '2019-08-25 13:55:20.741419', 'e96281a6-d1af-4bde-9a0a-97b76e56dc57', 1);
insert into package values ('284d295b-518f-4d67-999e-94968d0eed90', 1, '829.0.0', 'https://update.release.flatcar-linux.net/amd64-usr/829.0.0/', 'flatcar_production_update.gz', NULL, '186245514', '2lhoUvvnoY359pi2FnaS/xsgtig=', '2019-09-10 23:11:10.825985', 'e96281a6-d1af-4bde-9a0a-97b76e56dc57', 1);
insert into channel values ('e06064ad-4414-4904-9a6e-fd465593d1b2', 'stable', '#14b9d6', '2019-08-19 05:09:34.261241', 'e96281a6-d1af-4bde-9a0a-97b76e56dc57', '337b3f7e-ff29-47e8-a052-f0834d25bdb5', 1);
insert into channel values ('128b8c29-5058-4643-8e67-a1a0e3c641c9', 'beta', '#fc7f33', '2019-08-19 05:09:34.264334', 'e96281a6-d1af-4bde-9a0a-97b76e56dc57', '337b3f7e-ff29-47e8-a052-f0834d25bdb5', 1);
insert into channel values ('a87a03ad-4984-47a1-8dc4-3507bae91ee1', 'alpha', '#1fbb86', '2019-08-19 05:09:34.265754', 'e96281a6-d1af-4bde-9a0a-97b76e56dc57', '284d295b-518f-4d67-999e-94968d0eed90', 1);
insert into groups values ('9a2deb70-37be-4026-853f-bfdd6b347bbe', 'Stable', 'For production clusters', false, true, true, false, 'Europe/Berlin', '15 minutes', 2, '60 minutes', '2019-08-19 05:09:34.269062', 'e96281a6-d1af-4bde-9a0a-97b76e56dc57', 'e06064ad-4414-4904-9a6e-fd465593d1b2', 'stable');
insert into groups values ('3fe10490-dd73-4b49-b72a-28ac19acfcdc', 'Beta', 'Promoted alpha releases, to catch bugs specific to your configuration', true, true, true, false, 'Europe/Berlin', '15 minutes', 2, '60 minutes', '2019-08-19 05:09:34.273244', 'e96281a6-d1af-4bde-9a0a-97b76e56dc57', '128b8c29-5058-4643-8e67-a1a0e3c641c9', 'beta');
insert into groups values ('5b810680-e36a-4879-b98a-4f989e80b899', 'Alpha', 'Tracks current development work and is released frequently', false, true, true, false, 'Europe/Berlin', '15 minutes', 1, '30 minutes', '2019-08-19 05:09:34.274911', 'e96281a6-d1af-4bde-9a0a-97b76e56dc57', 'a87a03ad-4984-47a1-8dc4-3507bae91ee1', 'alpha');
insert into flatcar_action values ('b2b16e2e-57f8-4775-827f-8f0b11ae9bd2', 'postinstall', '', 'k8CB8tMe0M8DyZ5RZwzDLyTdkHjO/YgfKVn2RgUMokc=', false, false, true, '', '', '', '2019-08-20 00:12:37.532281', '2ba4c984-5e9b-411e-b7c3-b3eb14f7a261');
insert into flatcar_action values ('d5a2cbf3-b810-4e8c-88e8-6df91fc264c6', 'postinstall', '', 'QUGnmP51hp7zy+++o5fBIwElInTAms7/njnkxutn/QI=', false, false, true, '', '', '', '2019-08-20 06:15:29.11685', '337b3f7e-ff29-47e8-a052-f0834d25bdb5');
insert into flatcar_action values ('299c54d1-3344-4ae9-8ad2-5c63d56d6c14', 'postinstall', '', 'SCv89GYzx7Ix+TljqbNsd7on65ooWqBzcCrLFL4wChQ=', false, false, true, '', '', '', '2019-08-20 00:09:06.927461', 'c2a36312-b989-403e-ab57-06c055a7eac2');
insert into flatcar_action values ('748df5fc-12a5-4dad-a71e-465cc1668048', 'postinstall', '', '9HUs4whizfyvb4mgl+WaNaW3VLQYwsW1GHNHJNpcFg4=', false, false, true, '', '', '', '2019-08-25 13:55:20.825242', '43580892-cad8-468a-a0bb-eb9d0e09eca4');
insert into flatcar_action values ('9cd474c5-efa3-4989-9992-58ddb852ed84', 'postinstall', '', '1S9zQCLGjmefYnE/aFcpCjL1NsguHhQGj0UCm5f0M98=', false, false, true, '', '', '', '2019-09-10 23:11:10.913778', '284d295b-518f-4d67-999e-94968d0eed90');

-- Sample application 1
insert into application (id, name, description, team_id, product_id) values ('b6458005-8f40-4627-b33b-be70a718c48e', 'Sample application', 'Just an application to show the capabilities of Nebraska', 'd89342dc-9214-441d-a4af-bdd837a3b239','io.flatcar.sample');
insert into package (id, type, url, filename, version, application_id, arch) values ('5195d5a2-5f82-11e5-9d70-feff819cdc9f', 4, 'https://www.flatcar.org/', 'test_1.0.2', '1.0.2', 'b6458005-8f40-4627-b33b-be70a718c48e', 1);
insert into package (id, type, url, filename, version, application_id, arch) values ('12697fa4-5f83-11e5-9d70-feff819cdc9f', 4, 'https://www.flatcar.org/', 'test_1.0.3', '1.0.3', 'b6458005-8f40-4627-b33b-be70a718c48e', 1);
insert into package (id, type, url, filename, version, application_id, arch) values ('8004bad8-5f97-11e5-9d70-feff819cdc9f', 4, 'https://www.flatcar.org/', 'test_1.0.4', '1.0.4', 'b6458005-8f40-4627-b33b-be70a718c48e', 1);
insert into package (id, type, url, filename, version, application_id, arch) values ('aaaaaaaa-5f98-11e5-9d70-feff819cdc9f', 4, 'https://www.flatcar.org/', 'test_1.0.5', '1.0.5', 'b6458005-8f40-4627-b33b-be70a718c48e', 1);
insert into channel (id, name, color, application_id, package_id, arch) values ('bfe32b4a-5f8c-11e5-9d70-feff819cdc9f', 'Master', '#00CC00', 'b6458005-8f40-4627-b33b-be70a718c48e', '8004bad8-5f97-11e5-9d70-feff819cdc9f', 1);
insert into channel (id, name, color, application_id, package_id, arch) values ('cb2deea8-5f83-11e5-9d70-feff819cdc9f', 'Stable', '#0099FF', 'b6458005-8f40-4627-b33b-be70a718c48e', '12697fa4-5f83-11e5-9d70-feff819cdc9f', 1);
insert into channel (id, name, color, application_id, package_id, arch) values ('dddddddd-5f83-11e5-9d70-feff819cdc9f', 'Failing', '#AA99FF', 'b6458005-8f40-4627-b33b-be70a718c48e', 'aaaaaaaa-5f98-11e5-9d70-feff819cdc9f', 1);
insert into groups values ('bcaa68bc-5f82-11e5-9d70-feff819cdc9f', 'Prod EC2 us-west-2', 'Production servers, west coast', false, true, true, false, 'Europe/Berlin', '15 minutes', 2, '60 minutes', '2019-08-19 05:09:34.269062', 'b6458005-8f40-4627-b33b-be70a718c48e', 'cb2deea8-5f83-11e5-9d70-feff819cdc9f', 'bcaa68bc-5f82-11e5-9d70-feff819cdc9f');
insert into groups values ('7074264a-2070-4b84-96ed-8a269dba5021', 'Prod EC2 us-east-1', 'Production servers, east coast', false, true, true, false, 'Europe/Berlin', '15 minutes', 2, '60 minutes', '2019-08-19 05:09:34.269062', 'b6458005-8f40-4627-b33b-be70a718c48e', 'cb2deea8-5f83-11e5-9d70-feff819cdc9f', '7074264a-2070-4b84-96ed-8a269dba5021');
insert into groups values ('b110813a-5f82-11e5-9d70-feff819cdc9f', 'Qa-Dev', 'QA and development servers, Sydney', false, true, true, false, 'Europe/Berlin', '15 minutes', 2, '60 minutes', '2019-08-19 05:09:34.269062', 'b6458005-8f40-4627-b33b-be70a718c48e', 'bfe32b4a-5f8c-11e5-9d70-feff819cdc9f', 'b110813a-5f82-11e5-9d70-feff819cdc9f');
insert into groups values ('cccccccc-5f82-11e5-9d70-feff819cdc9f', 'Failing Qa-Dev', 'Failing QA and development servers, Sydney', false, true, true, false, 'Europe/Berlin', '15 minutes', 2, '60 minutes', '2019-08-19 05:09:34.269062', 'b6458005-8f40-4627-b33b-be70a718c48e', 'dddddddd-5f83-11e5-9d70-feff819cdc9f', 'cccccccc-5f82-11e5-9d70-feff819cdc9f');
insert into instance (id, ip) values ('instance1', '10.0.0.1');
insert into instance (id, ip) values ('instance2', '10.0.0.2');
insert into instance (id, ip) values ('instance3', '10.0.0.3');
insert into instance (id, ip) values ('instance4', '10.0.0.4');
insert into instance (id, ip) values ('instance5', '10.0.0.5');
insert into instance (id, ip) values ('instance6', '10.0.0.6');
insert into instance (id, ip) values ('instance7', '10.0.0.7');
insert into instance (id, ip) values ('instance8', '10.0.0.8');
insert into instance (id, ip) values ('instance9', '10.0.0.9');
insert into instance (id, ip) values ('instance10', '10.0.0.10');
insert into instance (id, ip) values ('instance11', '10.0.0.11');
insert into instance (id, ip) values ('instance12', '10.0.0.12');
insert into instance_application values ('1.0.3', default, 4, default, default, NULL, default, 'instance1', 'b6458005-8f40-4627-b33b-be70a718c48e', 'bcaa68bc-5f82-11e5-9d70-feff819cdc9f');
insert into instance_application values ('1.0.3', default, 4, default, default, NULL, default, 'instance2', 'b6458005-8f40-4627-b33b-be70a718c48e', 'bcaa68bc-5f82-11e5-9d70-feff819cdc9f');
insert into instance_application values ('1.0.2', default, 4, default, default, NULL, default, 'instance3', 'b6458005-8f40-4627-b33b-be70a718c48e', 'bcaa68bc-5f82-11e5-9d70-feff819cdc9f');
insert into instance_application values ('1.0.3', default, 4, default, default, NULL, default, 'instance4', 'b6458005-8f40-4627-b33b-be70a718c48e', '7074264a-2070-4b84-96ed-8a269dba5021');
insert into instance_application values ('1.0.3', default, 4, default, default, NULL, default, 'instance5', 'b6458005-8f40-4627-b33b-be70a718c48e', '7074264a-2070-4b84-96ed-8a269dba5021');
insert into instance_application values ('1.0.2', default, 4, default, default, NULL, default, 'instance6', 'b6458005-8f40-4627-b33b-be70a718c48e', '7074264a-2070-4b84-96ed-8a269dba5021');
insert into instance_application values ('1.0.1', default, 3, default, default, NULL, default, 'instance7', 'b6458005-8f40-4627-b33b-be70a718c48e', '7074264a-2070-4b84-96ed-8a269dba5021');
insert into instance_application values ('1.0.4', default, 4, default, default, NULL, default, 'instance8', 'b6458005-8f40-4627-b33b-be70a718c48e', 'b110813a-5f82-11e5-9d70-feff819cdc9f');
insert into instance_application values ('1.0.3', default, 7, default, default, NULL, default, 'instance9', 'b6458005-8f40-4627-b33b-be70a718c48e', 'b110813a-5f82-11e5-9d70-feff819cdc9f');
insert into instance_application values ('1.0.2', default, 2, default, default, NULL, default, 'instance10', 'b6458005-8f40-4627-b33b-be70a718c48e', 'b110813a-5f82-11e5-9d70-feff819cdc9f');
insert into instance_application values ('1.0.1', default, 3, default, default, NULL, default, 'instance11', 'b6458005-8f40-4627-b33b-be70a718c48e', 'b110813a-5f82-11e5-9d70-feff819cdc9f');
insert into instance_application values ('1.0.1', default, 3, default, default, NULL, default, 'instance12', 'b6458005-8f40-4627-b33b-be70a718c48e', 'cccccccc-5f82-11e5-9d70-feff819cdc9f');
insert into activity values (default, now() at time zone 'utc' - interval '3 hours', 1, 4, '1.0.3', 'b6458005-8f40-4627-b33b-be70a718c48e', '7074264a-2070-4b84-96ed-8a269dba5021', 'cb2deea8-5f83-11e5-9d70-feff819cdc9f', 'instance1');
insert into activity values (default, now() at time zone 'utc' - interval '6 hours', 5, 3, '1.0.3', 'b6458005-8f40-4627-b33b-be70a718c48e', '7074264a-2070-4b84-96ed-8a269dba5021', 'cb2deea8-5f83-11e5-9d70-feff819cdc9f', 'instance1');
insert into activity values (default, now() at time zone 'utc' - interval '12 hours', 3, 1, '1.0.3', 'b6458005-8f40-4627-b33b-be70a718c48e', '7074264a-2070-4b84-96ed-8a269dba5021', 'cb2deea8-5f83-11e5-9d70-feff819cdc9f', 'instance1');
insert into activity values (default, now() at time zone 'utc' - interval '18 hours', 4, 4, '1.0.3', 'b6458005-8f40-4627-b33b-be70a718c48e', '7074264a-2070-4b84-96ed-8a269dba5021', 'cb2deea8-5f83-11e5-9d70-feff819cdc9f', 'instance1');
insert into activity values (default, now() at time zone 'utc' - interval '24 hours', 2, 2, '1.0.3', 'b6458005-8f40-4627-b33b-be70a718c48e', '7074264a-2070-4b84-96ed-8a269dba5021', 'cb2deea8-5f83-11e5-9d70-feff819cdc9f', 'instance1');
insert into instance_status_history values (default, 4, '1.0.3', now() at time zone 'utc' - interval '8 hours', 'instance5', 'b6458005-8f40-4627-b33b-be70a718c48e', '7074264a-2070-4b84-96ed-8a269dba5021');
insert into instance_status_history values (default, 5, '1.0.3', now() at time zone 'utc' - interval '8 hours 5 minutes', 'instance5', 'b6458005-8f40-4627-b33b-be70a718c48e', '7074264a-2070-4b84-96ed-8a269dba5021');
insert into instance_status_history values (default, 6, '1.0.3', now() at time zone 'utc' - interval '9 hours', 'instance5', 'b6458005-8f40-4627-b33b-be70a718c48e', '7074264a-2070-4b84-96ed-8a269dba5021');
insert into instance_status_history values (default, 7, '1.0.3', now() at time zone 'utc' - interval '9 hours 45 minutes', 'instance5', 'b6458005-8f40-4627-b33b-be70a718c48e', '7074264a-2070-4b84-96ed-8a269dba5021');
insert into instance_status_history values (default, 2, '1.0.3', now() at time zone 'utc' - interval '9 hours 45 minutes 10 seconds', 'instance5', 'b6458005-8f40-4627-b33b-be70a718c48e', '7074264a-2070-4b84-96ed-8a269dba5021');
insert into instance_status_history values (default, 4, '1.0.2', now() at time zone 'utc' - interval '36 hours', 'instance5', 'b6458005-8f40-4627-b33b-be70a718c48e', '7074264a-2070-4b84-96ed-8a269dba5021');
insert into instance_status_history values (default, 5, '1.0.2', now() at time zone 'utc' - interval '36 hours 5 minutes', 'instance5', 'b6458005-8f40-4627-b33b-be70a718c48e', '7074264a-2070-4b84-96ed-8a269dba5021');
insert into instance_status_history values (default, 6, '1.0.2', now() at time zone 'utc' - interval '37 hours', 'instance5', 'b6458005-8f40-4627-b33b-be70a718c48e', '7074264a-2070-4b84-96ed-8a269dba5021');
insert into instance_status_history values (default, 7, '1.0.2', now() at time zone 'utc' - interval '37 hours 45 minutes', 'instance5', 'b6458005-8f40-4627-b33b-be70a718c48e', '7074264a-2070-4b84-96ed-8a269dba5021');
insert into instance_status_history values (default, 2, '1.0.2', now() at time zone 'utc' - interval '37 hours 45 minutes 10 seconds', 'instance5', 'b6458005-8f40-4627-b33b-be70a718c48e', '7074264a-2070-4b84-96ed-8a269dba5021');
insert into event (previous_version, error_code, instance_id, application_id, event_type_id) select '', '', 'instance12', 'b6458005-8f40-4627-b33b-be70a718c48e', et.id from event_type et where et.type = 3 and et.result = 0;

-- Sample application 2
insert into application (id, name, description, team_id) values ('780d6940-9a48-4414-88df-95ba63bbe9cb', 'Sample application 2', 'Another sample application, feel free to remove me', 'd89342dc-9214-441d-a4af-bdd837a3b239');
insert into package (id, type, url, filename, version, application_id, arch) values ('efb186c9-d5cb-4df2-9382-c4821e4dcc4b', 4, 'http://localhost:8000/', 'demo_v1.0.0', '1.0.0', '780d6940-9a48-4414-88df-95ba63bbe9cb', 1);
insert into package (id, type, url, filename, version, application_id, arch) values ('ba28af48-b5b9-460e-946a-eba906ce7daf', 4, 'http://localhost:8000/', 'demo_v1.0.1', '1.0.1', '780d6940-9a48-4414-88df-95ba63bbe9cb', 1);
insert into channel (id, name, color, application_id, package_id, arch) values ('a7c8c9a4-d2a3-475d-be64-911ff8d6e997', 'Master', '#14b9d6', '780d6940-9a48-4414-88df-95ba63bbe9cb', 'efb186c9-d5cb-4df2-9382-c4821e4dcc4b', 1);
insert into groups values ('51a32aa9-3552-49fc-a28c-6543bccf0069', 'Master - dev', 'The latest stuff will be always here', false, true, true, false, 'Europe/Berlin', '15 minutes', 2, '60 minutes', '2019-08-19 05:09:34.269062', '780d6940-9a48-4414-88df-95ba63bbe9cb', 'a7c8c9a4-d2a3-475d-be64-911ff8d6e997', '51a32aa9-3552-49fc-a28c-6543bccf0069');
