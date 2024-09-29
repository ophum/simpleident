
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

CREATE TABLE `oauth2_codes` (
    id TEXT PRIMARY KEY,
    oauth2_client_id TEXT,
    code TEXT,
    account_id TEXT,
    created_at DATETIME,
    updated_at DATETIME,
    deleted_at DATETIME
);

CREATE TABLE `oauth2_tokens` (
    id TEXT PRIMARY KEY,
    oauth2_client_id TEXT,
    token TEXT,
    account_id TEXT,
    created_at DATETIME,
    updated_at DATETIME,
    deleted_at DATETIME
);

-- +migrate Down
DROP TABLE `oauth2_tokens`;
DROP TABLE `oauth2_codes`;
DROP TABLE `oauth2_client_secrets`;
DROP TABLE `oauth2_clients`;
DROP TABLE `accounts`;
