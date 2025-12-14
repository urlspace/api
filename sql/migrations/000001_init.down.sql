-- Drop triggers first (depend on tables and function)
DROP TRIGGER IF EXISTS UPDATE_USERS_UPDATED_AT ON users ;
DROP TRIGGER IF EXISTS UPDATE_RESOURCES_UPDATED_AT ON resources ;

-- Drop tables (depend on function via triggers)
DROP TABLE IF EXISTS users ;
DROP TABLE IF EXISTS resources ;

-- Drop function last (nothing depends on it anymore)
DROP FUNCTION IF EXISTS update_updated_at_column () ;
