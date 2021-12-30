-- +migrate Up

alter table application add column product_id varchar(155) default null;
alter table application add constraint application_unique_product_id unique(product_id);

-- +migrate Down

alter table application drop constraint application_unique_product_id;
alter table application drop column product_id;
