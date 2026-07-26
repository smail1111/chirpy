-- +goose Up
create table chirps (
    id text primary key,
    created_at timestamp not null,
    updated_at timestamp not null,
    body text not null,
    user_id text not null references users on delete cascade,
    foreign key (user_id) references users(id)
);

-- +goose Down
drop table chirps;