CREATE TABLE IF NOT EXISTS things(
    uuid text PRIMARY KEY,
    name text,
    value text,
    updated TIMESTAMP,
    created TIMESTAMP
)