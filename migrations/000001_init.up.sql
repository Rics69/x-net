CREATE SCHEMA xchat;

CREATE TABLE xchat.users (
    id SERIAL PRIMARY KEY,
    version BIGINT NOT NULL DEFAULT 1,
    full_name VARCHAR(100) NOT NULL CHECK (char_length(full_name) BETWEEN 3 AND 100),
    phone_number VARCHAR(15) CHECK (phone_number ~ '^\+[0-9]{9,14}$')
);
