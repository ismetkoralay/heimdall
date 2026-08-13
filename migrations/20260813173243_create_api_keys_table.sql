-- +goose Up
create table api_keys (
id UUID primary key default gen_random_uuid(),
hashed_key text not null,
name varchar(25) not null,
rpm_limit int not null default 0,
daily_token_quota int not null default 0,
revoked bool not null default false,
created_at timestamp with time zone default current_timestamp
);

-- +goose Down
drop table if exists api_keys;
