-- +goose Up
create schema if not exists sis;
create table if not exists sis.subscribers (
	msisdn        bigint primary key,
	billing_type  smallint,
	language_type smallint,
	operator_type smallint,
	change_date   timestamp default now()
);

-- +goose Down
drop table if exists sis.subscribers;
drop schema if exists sis;
