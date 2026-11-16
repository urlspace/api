ALTER TABLE collections
    ADD COLUMN password TEXT,
    ADD CONSTRAINT collections_password_requires_public CHECK (password IS NULL OR public = TRUE);
