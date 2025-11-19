-- name: AddUser :exec
insert into users (id, username, password) values (?, ?, ?);

-- name: GetUser :one
select id, password from users where username = ?;

-- name: GetUserByID :one
select username, password from users where id = ?;

-- name: UpdateUser :exec
update users set username = ?, password = ? where id = ?;

-- name: RemoveUser :exec
delete from users where id = ?;
