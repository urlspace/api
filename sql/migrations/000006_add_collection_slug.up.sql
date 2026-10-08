ALTER TABLE collections
ADD COLUMN slug TEXT;

-- Existing collections get a random placeholder slug. Users can replace it
-- with a readable one when they next edit the collection.
UPDATE collections
SET slug = uuidv7()::text;

ALTER TABLE collections
ALTER COLUMN slug SET NOT NULL;

ALTER TABLE collections
ADD CONSTRAINT collections_user_id_slug_key UNIQUE (user_id, slug);
