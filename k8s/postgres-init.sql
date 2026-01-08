create schema if not exists sis;
create table if not exists sis.info (
  msisdn        bigint primary key,
  billing_type  smallint,
  language_type smallint,
  operator_type smallint,
  change_date   timestamp default now()
);

-- Ensure unqualified queries target sis schema.
alter role sis set search_path to sis,public;
