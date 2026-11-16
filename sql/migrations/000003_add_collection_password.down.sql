ALTER TABLE collections
    DROP CONSTRAINT collections_password_requires_public,
    DROP COLUMN password;
