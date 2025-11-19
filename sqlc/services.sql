-- name: AddService :exec
insert into services (id, name, logo) values (?, ?, ?);

-- name: GetService :one
select id, logo from services where name = ?;

-- name: CheckExistingRecord :one
select creds.id from creds
  where creds.user_id <> ?
  and creds.service_id in (
    select services.id from services
      where services.name = ?
  );

-- name: GetServicesList :many
select name, logo from services;

-- name: GetUserServicesList :many
select distinct services.name, services.logo from services
  left join creds on creds.service_id = services.id
  where creds.user_id = ?;

-- name: UpdateService :exec
update services set name = ?, logo = ? where name = sqlc.arg(old_name);

-- name: RemoveService :exec
delete from services where name = ?;
