-- name: AddCredsRecord :exec
insert into creds (id, user_id, service_id, name, secret, payload) values (?, ?, ?, ?, ?, ?);

-- name: GetUserCredsInService :many
select creds.id, creds.name, creds.secret, creds.payload from creds
  left join services on services.id = creds.service_id
  where creds.user_id = ? and services.name = ?;

-- name: GetServiceID :one
select id from services where name = ?;

-- name: GetCredsRecordID :one
select id from creds where name = ? and service_id = ? and user_id = ?;

-- name: UpdateCredsRecord :exec
update creds set name = ?, secret = ?, payload = ? where user_id = ? and name = sqlc.arg(old_name);

-- name: RemoveCredsRecord :exec
delete from creds
  where creds.user_id = ? and
  creds.name = ? and
  creds.service_id in (
    select services.id from services
      where services.name = sqlc.arg(service_name)
  );

-- name: RemoveAllCredsInService :exec
delete from creds
  where id in (
    select creds.id from creds
      left join services on services.id = creds.service_id
      where creds.user_id = ? and services.name = ?
  );
