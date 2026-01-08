-- +migrate Up

ALTER TABLE instance ADD COLUMN oem VARCHAR(256) DEFAULT '';
ALTER TABLE instance ADD COLUMN oem_version VARCHAR(256) DEFAULT '';

-- +migrate Down

ALTER TABLE instance DROP COLUMN oem_version;
ALTER TABLE instance DROP COLUMN oem;
