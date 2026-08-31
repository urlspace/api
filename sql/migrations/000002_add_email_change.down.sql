ALTER TABLE users
    DROP COLUMN email_new_code_hash_expires_at,
    DROP COLUMN email_new_code_hash,
    DROP COLUMN email_new;
