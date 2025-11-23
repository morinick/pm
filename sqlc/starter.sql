-- name: AddKeys :exec
insert into ciphers (id, key_value) values (?, ?);

-- name: GetKeys :many
select key_value from ciphers;

-- name: AddAssets :exec
insert into services (id, name, logo) values (?, ?, ?);
