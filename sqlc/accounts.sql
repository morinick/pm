-- name: AddAccount :exec
insert into accounts (id, user_id, service_id, name, secret, payload) values (?, ?, ?, ?, ?, ?);

-- name: GetUserAccountsInService :many
select accounts.id, accounts.name, accounts.secret, accounts.payload from accounts
  left join services on services.id = accounts.service_id
  where accounts.user_id = ? and services.name = ?;

-- name: GetServiceID :one
select id from services where name = ?;

-- name: GetAccountID :one
select id from accounts where name = ? and service_id = ? and user_id = ?;

-- name: UpdateAccount :exec
update accounts set name = ?, secret = ?, payload = ? where user_id = ? and name = sqlc.arg(old_name);

-- name: RemoveAccount :exec
delete from accounts
  where accounts.user_id = ? and
  accounts.name = ? and
  accounts.service_id in (
    select services.id from services
      where services.name = sqlc.arg(service_name)
  );

-- name: RemoveAllAccountsInService :exec
delete from accounts
  where id in (
    select accounts.id from accounts
      left join services on services.id = accounts.service_id
      where accounts.user_id = ? and services.name = ?
  );
