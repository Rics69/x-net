-- добавляем nullable -> заполняем -> делаем NOT NULL:
-- сразу NOT NULL без DEFAULT упадёт, если в таблице уже есть строки
ALTER TABLE xchat.users
    ADD COLUMN username VARCHAR(32),
    ADD COLUMN password_hash TEXT,
    ADD COLUMN created_at TIMESTAMPTZ NOT NULL DEFAULT now();

-- у уже существующих (тестовых) юзеров логина/пароля нет. Пустой hash
-- не пройдёт bcrypt.CompareHashAndPassword, так что войти под ними нельзя
UPDATE xchat.users
SET
    username = 'user_' || id,
    password_hash = ''
WHERE username IS NULL;

ALTER TABLE xchat.users
    ALTER COLUMN username SET NOT NULL,
    ALTER COLUMN password_hash SET NOT NULL,
    ADD CONSTRAINT users_username_format CHECK (username ~ '^[a-zA-Z0-9_]{3,32}$');

-- уникальность без учёта регистра: Ivan и ivan - один и тот же username.
-- По этому же индексу идёт поиск при логине (WHERE lower(username) = lower($1))
CREATE UNIQUE INDEX users_username_lower_uidx ON xchat.users (lower(username));
