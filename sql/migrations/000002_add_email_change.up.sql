ALTER TABLE users
    ADD COLUMN email_new TEXT,
    ADD COLUMN email_new_code_hash TEXT,
    ADD COLUMN email_new_code_hash_expires_at TIMESTAMPTZ;
