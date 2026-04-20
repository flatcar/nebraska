-- +migrate Up

ALTER TABLE instance ADD COLUMN oem VARCHAR(256) NOT NULL DEFAULT '';
ALTER TABLE instance ADD COLUMN aleph_version VARCHAR(256) NOT NULL DEFAULT '';

-- +migrate Down

ALTER TABLE instance DROP COLUMN aleph_version;
ALTER TABLE instance DROP COLUMN oem;
