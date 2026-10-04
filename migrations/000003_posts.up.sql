CREATE TABLE xchat.posts (
    id BIGSERIAL PRIMARY KEY,
    content VARCHAR(280) NOT NULL CHECK (char_length(content) BETWEEN 1 AND 280),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    -- удалили юзера -> удалились его посты
    author_user_id INTEGER NOT NULL REFERENCES xchat.users(id) ON DELETE CASCADE
);

-- индекс под ленту: ORDER BY created_at DESC, id DESC + WHERE (created_at, id) < (...)
-- по нему постгрес идёт сразу с нужного места, не перебирая предыдущие страницы
CREATE INDEX posts_created_at_id_idx ON xchat.posts (created_at DESC, id DESC);

-- постгрес НЕ создаёт индекс на внешний ключ сам. Без него ON DELETE CASCADE
-- при удалении юзера делает seq scan по всем постам
CREATE INDEX posts_author_user_id_idx ON xchat.posts (author_user_id);
