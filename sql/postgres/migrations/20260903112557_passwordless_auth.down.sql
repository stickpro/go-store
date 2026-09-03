UPDATE users SET password = '' WHERE password IS NULL;
ALTER TABLE users ALTER COLUMN password SET NOT NULL;
