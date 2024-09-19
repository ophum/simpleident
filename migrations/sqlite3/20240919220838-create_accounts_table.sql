
-- +migrate Up
CREATE TABLE `accounts` (
    id TEXT PRIMARY KEY,
    username TEXT,
    password TEXT,
    created_at DATETIME,
    updated_at DATETIME,
    deleted_at DATETIME
);

-- +migrate Down
DROP TABLE `accounts`;
