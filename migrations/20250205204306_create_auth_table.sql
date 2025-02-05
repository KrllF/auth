-- +goose Up
create table users (
    id serial primary key,
    name text not null,
    email text not null,
    password text not null,
    role text not null default 'user',
    created_at timestamp not null default now(),
    updated_at timestamp
);

-- +goose Down
drop table users; -- Удаляем таблицу users
drop type user_role;