
-- +migrate Up
CREATE TABLE `accounts` (
    id TEXT PRIMARY KEY,
    username TEXT,
    password TEXT,
    created_at DATETIME,
    updated_at DATETIME,
    deleted_at DATETIME
);

CREATE TABLE `oauth2_clients` (
    id TEXT PRIMARY KEY,
    name TEXT,
    description TEXT,
    callback_url TEXT,
    created_at DATETIME,
    updated_at DATETIME,
    deleted_at DATETIME
);

CREATE TABLE `oauth2_client_secrets` (
    id TEXT PRIMARY KEY,
    -- TODO foreign key
    oauth2_client_id TEXT,
    secret TEXT,
    created_at DATETIME,
    updated_at DATETIME,
    deleted_at DATETIME
);

-- +migrate Down
DROP TABLE `oauth2_client_secrets`;
DROP TABLE `oauth2_clients`;
DROP TABLE `accounts`;
