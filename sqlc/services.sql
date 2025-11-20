-- name: AddService :exec
insert into services (id, name, logo) values (?, ?, ?);

-- name: GetService :one
select id, logo from services where name = ?;

-- name: CheckExistingRecord :one
select accounts.id from accounts
  where accounts.user_id <> ?
  and accounts.service_id in (
    select services.id from services
      where services.name = ?
  );

-- name: GetServicesList :many
select name, logo from services;

-- name: GetUserServicesList :many
select distinct services.name, services.logo from services
  left join accounts on accounts.service_id = services.id
  where accounts.user_id = ?;

-- name: UpdateService :exec
update services set name = ?, logo = ? where name = sqlc.arg(old_name);

-- name: RemoveService :exec
delete from services where name = ?;
