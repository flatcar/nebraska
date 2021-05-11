-- +migrate Up

alter table groups add column track varchar(256) not null default 'missing' check (track <> '');

-- populate track from UUID
update groups set track = id;

-- overwrite known amd64 tracks
update groups set track = 'stable' where id = '9a2deb70-37be-4026-853f-bfdd6b347bbe';
update groups set track = 'beta' where id = '3fe10490-dd73-4b49-b72a-28ac19acfcdc';
update groups set track = 'alpha' where id = '5b810680-e36a-4879-b98a-4f989e80b899';
update groups set track = 'edge' where id = '72834a2b-ad86-4d6d-b498-e08a19ebe54e';
-- overwrite known arm64 tracks
update groups set track = 'stable' where id = '11a585f6-9418-4df0-8863-78b2fd3240f8';
update groups set track = 'beta' where id = 'd112ec01-ba34-4a9e-9d4b-9814a685f266';
update groups set track = 'alpha' where id = 'e641708d-fb48-4260-8bdf-ba2074a1147a';
update groups set track = 'edge' where id = 'b4b2fa22-c1ea-498c-a8ac-c1dc0b1d7c17';

-- +migrate Down

alter table groups drop column track;

