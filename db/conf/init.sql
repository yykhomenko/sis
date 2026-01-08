\echo CREATE SCHEMA sis;
create schema sis;

\echo CREATE TABLE subscribers;
create table sis.subscribers (
	msisdn        bigint primary key,
	created_at    timestamp not null default now(),
	billing_type  smallint  not null,
	language_type smallint  not null,
	operator_type smallint  not null
)
