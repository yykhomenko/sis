-- name: GetSubscriber :one
select
	msisdn,
	updated_at,
	billing_type,
	language_type,
	operator_type
from subscribers
where msisdn = $1;

-- name: UpdateSubscriber :exec
insert into subscribers (
	msisdn,
	billing_type,
	language_type,
	operator_type
)
values (
	$1,
	$2,
	$3,
	$4
)
on conflict (msisdn) do update
set
	updated_at = now(),
	billing_type = excluded.billing_type,
	language_type = excluded.language_type,
	operator_type = excluded.operator_type;
