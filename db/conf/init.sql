-- Keep in sync with sql/schema (goose migrations).
\echo CREATE TABLE subscribers;
create table if not exists subscribers (
	msisdn        bigint primary key,
	updated_at    timestamp not null default now(),
	billing_type  smallint  not null,
	language_type smallint  not null,
	operator_type smallint  not null
);
