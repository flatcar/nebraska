-- +migrate Up

ALTER TABLE event ADD COLUMN event_type integer;
ALTER TABLE event ADD COLUMN event_result integer;

UPDATE event e
SET event_type = et.type, 
    event_result = et.result 
FROM event_type et 
WHERE e.event_type_id = et.id;

ALTER TABLE event ALTER COLUMN event_type SET DEFAULT 0;
ALTER TABLE event ALTER COLUMN event_result SET DEFAULT 0;
ALTER TABLE event ALTER COLUMN event_type SET NOT NULL;
ALTER TABLE event ALTER COLUMN event_result SET NOT NULL;

ALTER TABLE event DROP COLUMN event_type_id;
DROP TABLE event_type;


-- +migrate Down

CREATE TABLE event_type (
    id serial primary key,
    type integer not null,
    result integer not null,
    description varchar(100) not null
);

ALTER TABLE event ADD COLUMN event_type_id integer;
