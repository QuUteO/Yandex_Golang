create schema if not exists shema_users;

create table if not exist schema_name.users
(
    id      bigserial    not null
        constraint users_pk
            primary key,
    name    text,
    surname text,
    email   varchar(254) not null unique
);