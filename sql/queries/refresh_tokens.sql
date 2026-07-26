-- name: CreateRefreshToken :one
insert into refresh_tokens (token, created_at, updated_at, user_id, expires_at)
values (
    $1,
    now(),
    now(),
    $2,
    now() + interval '60 days'
)
returning *;

-- name: GetRefreshToken :one
select * from refresh_tokens where token = $1;

-- name: RevokeRefreshToken :exec
update refresh_tokens
set revoked_at = now(), updated_at = now()
where token = $1;

-- name: GetUserFromRefreshToken :one
select * from users
join refresh_tokens on refresh_tokens.user_id = users.id
where refresh_tokens.token = $1
limit 1;