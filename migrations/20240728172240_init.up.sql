create table users
(
    id       uuid    not null
        constraint users_pk
        primary key,
    email    varchar not null,
    name     varchar,
    birthday date
);

create unique index users_email_uindex
    on users (lower(email));
