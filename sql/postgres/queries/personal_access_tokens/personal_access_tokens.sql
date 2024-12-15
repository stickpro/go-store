-- name: GetByToken :one
SELECT * FROM personal_access_tokens WHERE (expires_at > now() OR expires_at IS NULL)  AND token=$1 LIMIT 1;