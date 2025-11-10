\echo CREATE SCHEMA sis;
create schema sis;

\echo CREATE TABLE info;
create table sis.info (
	msisdn        bigint primary key,
	billing_type  smallint,
	language_type smallint,
	operator_type smallint,
	change_date   timestamp default now()
)
