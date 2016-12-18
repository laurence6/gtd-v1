create table users (
    id text primary key,
    password text not null,

    constraint not_empty_user_id check (id!=''));

create table task (
    user_id text not null references users(id) on delete cascade on update cascade,
    id bigint primary key,
    subject text not null,
    due bigint not null,
    priority smallint not null,
    reminder bigint not null,
    next bigint not null,
    note text not null,
    parent_task_id bigint references task(id) on delete cascade on update cascade,

    constraint not_empty_task_subject check (subject!=''));

create table tag (
    user_id text not null references users(id) on delete cascade on update cascade,
    task_id bigint not null references task(id) on delete cascade on update cascade,
    name text not null,

    unique(user_id, task_id, name),
    constraint not_empty_tag_name check (name!=''));

create table contact_info (
    user_id text not null references users(id) on delete cascade on update cascade,
    info text not null,

    unique(user_id, info),
    constraint not_empty_info check (info!=''));

"task_parent_task_id_idx" btree (parent_task_id)
"task_user_id" btree (user_id)

create view task_tag as select task_id, array_agg(name) as tags from tag group by tag.task_id;
create view task_sub_task_id as select task.parent_task_id as task_id, array_agg(task.id) as sub_task_ids from task where task.parent_task_id is not null group by task.parent_task_id;
