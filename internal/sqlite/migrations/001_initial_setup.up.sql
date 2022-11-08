create table balls (
    id integer primary key,
    brand text not null,
    name text not null,
    approval_date text not null,
    image_url text not null
) strict;

create index balls_approval_date_idx on balls(approval_date);
