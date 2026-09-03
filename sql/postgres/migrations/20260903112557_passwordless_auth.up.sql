-- Regular users authenticate with an email one-time code and have no password.
-- Only admin accounts keep a password.
ALTER TABLE users ALTER COLUMN password DROP NOT NULL;
