-- +goose Up
create table if not exists subscribers (
	msisdn        bigint primary key,
	updated_at    timestamp not null default now(),
	billing_type  smallint  not null,
	language_type smallint  not null,
	operator_type smallint  not null
);

-- +goose Down
drop table if exists subscribers;
