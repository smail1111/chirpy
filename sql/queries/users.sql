-- name: CreateUser :one
insert into users (id, created_at, updated_at, email, hashed_password)
values (
    gen_random_uuid(),
    now(),
    now(),
    $1,
    $2
)
returning *;

-- name: Reset :exec
delete from users;

-- name: GetUserByID :one
select * from users where id = $1;

-- name: GetUserByEmail :one
select * from users where email = $1;

-- name: UpdateUser :one
update users
set email = $1, hashed_password = $2
where id = $3
returning *;

-- name: UpgradeUser :one
update users
set is_chirpy_red = true
where id = $1
returning *;