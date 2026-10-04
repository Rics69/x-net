DROP INDEX xchat.users_username_lower_uidx;

ALTER TABLE xchat.users
    DROP CONSTRAINT users_username_format,
    DROP COLUMN username,
    DROP COLUMN password_hash,
    DROP COLUMN created_at;
