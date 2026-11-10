ALTER TABLE tokens
    ADD CONSTRAINT tokens_user_id_fkey
    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE;
