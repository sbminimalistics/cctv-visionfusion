create table
    events (
        id text primary key,
        src text,
        src_type text,
        created_at timestamptz
);