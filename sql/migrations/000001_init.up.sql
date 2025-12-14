-- auto update updated_at field
CREATE OR REPLACE FUNCTION update_updated_at_column ()
RETURNS TRIGGER
AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$
LANGUAGE PLPGSQL ;

-- resources
CREATE TABLE resources (
id uuid PRIMARY KEY DEFAULT gen_random_uuid (),
title text NOT NULL,
description text NOT NULL,
url text NOT NULL,
favourite boolean NOT NULL DEFAULT FALSE,
read_later boolean NOT NULL DEFAULT FALSE,
created_at timestamptz NOT NULL DEFAULT now (),
updated_at timestamptz NOT NULL DEFAULT now ()
) ;

CREATE TRIGGER UPDATE_RESOURCES_UPDATED_AT
BEFORE UPDATE ON resources
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column () ;

-- users
CREATE TABLE users (
id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
email TEXT NOT NULL UNIQUE,
email_verified BOOLEAN NOT NULL DEFAULT FALSE,
email_verification_token TEXT,
email_verification_token_expires_at TIMESTAMPTZ,
password TEXT NOT NULL,
username TEXT NOT NULL UNIQUE,
created_at TIMESTAMPTZ NOT NULL DEFAULT now (),
updated_at TIMESTAMPTZ NOT NULL DEFAULT now ()
) ;


CREATE TRIGGER UPDATE_USERS_UPDATED_AT
BEFORE UPDATE ON users
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column () ;
