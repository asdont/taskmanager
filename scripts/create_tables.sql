create table auth
(
    user_id  serial not null
        constraint auth__pk
            primary key,
    username text                                             not null,
    password text                                             not null
);

create unique index auth__username__uindex
    on auth (username);

create index auth__password__index
    on auth (password);

create table task
(
    task_id   serial not null
        constraint user__pk
            primary key,
    user_id   integer                                          not null
        constraint task__user_id__fk
            references auth
            on update cascade on delete cascade,
    status    boolean                                          not null,
    title     text                                             not null,
    created   timestamp                                        not null,
    updated   timestamp                                        not null,
    completed timestamp
);

create index task__user_id__index
    on task (user_id);

create index task__status__index
    on task (status);

