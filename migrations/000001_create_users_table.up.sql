CREATE TABLE users (
    id            uuid        PRIMARY KEY,
    name          text        NOT NULL,
    email         text        NOT NULL,
    password_hash text        NOT NULL,
    role          text        NOT NULL,
    is_active     boolean     NOT NULL DEFAULT true,
    created_at    timestamptz NOT NULL,
    updated_at    timestamptz NOT NULL,

    CONSTRAINT users_name_not_blank CHECK (length(btrim(name)) > 0),
    CONSTRAINT users_role_valid     CHECK (role IN ('owner', 'clerk'))
);

CREATE UNIQUE INDEX users_email_unique_idx ON users (lower(email));
