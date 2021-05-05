-- +migrate Up

-- only add data if an application does not exist yet.

-- +migrate StatementBegin

DO LANGUAGE plpgsql $$
begin
-- TODO: seeding the database should not be a part of migration…

--    if not exists (select id from application limit 1) then

        -- insert empty ARM channels
        insert into channel values ('5dfe7b12-c94a-470d-a2b6-2eae78c5c9f5', 'stable', '#1458d6', '2015-09-19 05:09:34.261241', 'e96281a6-d1af-4bde-9a0a-97b76e56dc57', null, 2);
        insert into channel values ('cf20698b-1f19-43d6-b6f6-d15c796cb217', 'beta', '#fce433', '2015-09-19 05:09:34.264334', 'e96281a6-d1af-4bde-9a0a-97b76e56dc57', null, 2);
        insert into channel values ('def12ce0-3ba4-4649-b290-8843f3b455eb', 'alpha', '#1fa2bb', '2015-09-19 05:09:34.265754', 'e96281a6-d1af-4bde-9a0a-97b76e56dc57', null, 2);
        insert into channel values ('30b6ffa6-e6dc-4a01-bea6-9ce7f1a5bb34', 'edge', '#f4ab3b', '2015-09-19 05:09:34.265754', 'e96281a6-d1af-4bde-9a0a-97b76e56dc57', null, 2);

        -- insert ARM groups
        insert into groups values ('11a585f6-9418-4df0-8863-78b2fd3240f8', 'Stable (ARM)', 'For production clusters (ARM)', false, true, false, false, 'Europe/Berlin', '1 minutes', 999999, '60 minutes', '2015-09-19 05:09:34.269062', 'e96281a6-d1af-4bde-9a0a-97b76e56dc57', '5dfe7b12-c94a-470d-a2b6-2eae78c5c9f5');
        insert into groups values ('d112ec01-ba34-4a9e-9d4b-9814a685f266', 'Beta (ARM)', 'Promoted alpha releases, to catch bugs specific to your configuration (ARM)', false, true, false, false, 'Europe/Berlin', '1 minutes', 999999, '60 minutes', '2015-09-19 05:09:34.273244', 'e96281a6-d1af-4bde-9a0a-97b76e56dc57', 'cf20698b-1f19-43d6-b6f6-d15c796cb217');
        insert into groups values ('e641708d-fb48-4260-8bdf-ba2074a1147a', 'Alpha (ARM)', 'Tracks current development work and is released frequently (ARM)', false, true, false, false, 'Europe/Berlin', '1 minutes', 999999, '60 minutes', '2015-09-19 05:09:34.274911', 'e96281a6-d1af-4bde-9a0a-97b76e56dc57', 'def12ce0-3ba4-4649-b290-8843f3b455eb');
        insert into groups values ('b4b2fa22-c1ea-498c-a8ac-c1dc0b1d7c17', 'Edge (ARM)', 'Experimental features and patches (ARM)', false, true, false, false, 'Europe/Berlin', '1 minutes', 999999, '60 minutes', '2018-04-10 10:25:39.677359+00', 'e96281a6-d1af-4bde-9a0a-97b76e56dc57', '30b6ffa6-e6dc-4a01-bea6-9ce7f1a5bb34');

        -- update descriptions of the AMD64 groups
        update groups set description = 'For production clusters (AMD64)', name = 'Stable (AMD64)' where id = '9a2deb70-37be-4026-853f-bfdd6b347bbe' and description = 'For production clusters';
        update groups set name = 'Stable (AMD64)' where id = '9a2deb70-37be-4026-853f-bfdd6b347bbe' and name = 'Stable';
        update groups set description = 'Promoted alpha releases, to catch bugs specific to your configuration (AMD64)' where id = '3fe10490-dd73-4b49-b72a-28ac19acfcdc' and description = 'Promoted alpha releases, to catch bugs specific to your configuration';
        update groups set name = 'Beta (AMD64)' where id = '3fe10490-dd73-4b49-b72a-28ac19acfcdc' and name = 'Beta';
        update groups set description = 'Tracks current development work and is released frequently (AMD64)' where id = '5b810680-e36a-4879-b98a-4f989e80b899' and description = 'Tracks current development work and is released frequently';
        update groups set name = 'Alpha (AMD64)' where id = '5b810680-e36a-4879-b98a-4f989e80b899' and name = 'Alpha';
        update groups set description = 'Experimental features and patches (AMD64)' where id = '72834a2b-ad86-4d6d-b498-e08a19ebe54e' and description = 'Experimental features and patches';
        update groups set name = 'Edge (AMD64)' where id = '72834a2b-ad86-4d6d-b498-e08a19ebe54e' and name = 'Edge';

--    end if;

end;

$$;

-- +migrate StatementEnd

-- +migrate Down
