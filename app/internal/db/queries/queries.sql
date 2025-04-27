-- name: GetSnippetByPublicID :one 
-- Retrieves a snippet by its public ID 
SELECT s.id, s.public_id, s.title, s.created_at, s.expires_at, s.password_hash, s.edit_token, s.view_count, s.last_edited_at,
       c.content_type, c.encrypted_content
FROM snippets s
JOIN snippet_contents c ON s.id = c.snippet_id
WHERE s.public_id = $1 
LIMIT 1;


-- name: CreateSnippet :one
-- Creates a new snippet 
WITH new_snippet AS (
    INSERT INTO snippets (
        title, 
        expires_at, 
        password_hash, 
        edit_token
    ) VALUES (
        $1, $2, $3, $4
    )
    RETURNING id, public_id, created_at, edit_token
)
INSERT INTO snippet_contents (
    snippet_id,
    content_type,
    encrypted_content
) 
SELECT 
    id, $5, $6
FROM new_snippet
RETURNING snippet_id,
          (SELECT public_id FROM new_snippet),
          (SELECT created_at FROM new_snippet),
          (SELECT edit_token FROM new_snippet);

-- name: DeleteExpiredSnippets :execrows
-- Deletes all snippets that have expired
DELETE FROM snippets
WHERE expires_at IS NOT NULL AND expires_at < NOW();

-- name: DeleteSnippetById :execrows 
-- Deletes a snippet by id 
DELETE FROM snippets 
WHERE id = $1;

-- name: UpdateSnippet :one
-- Updates an existing snippet by ID
UPDATE snippets
SET 
    title = COALESCE($2, title),
    expires_at = COALESCE($3, expires_at),
    password_hash = COALESCE($4, password_hash),
    last_edited_at = NOW()
WHERE id = $1
RETURNING id, public_id, created_at, last_edited_at;

-- name: UpdateSnippetContent :exec
-- Updates the content of a snippet
UPDATE snippet_contents
SET 
    content_type = COALESCE($2, content_type),
    encrypted_content = COALESCE($3, encrypted_content)
WHERE snippet_id = $1;


-- name: IncrementSnippetViewCount :one
-- Increments the view count for a snippet
UPDATE snippets
SET view_count = view_count + 1
WHERE id = $1
RETURNING view_count;

-- name: ListRecentSnippets :many
-- Lists recently created snippets (for admin purposes)
SELECT s.id, s.public_id, s.title, s.created_at, s.expires_at 
FROM snippets s
ORDER BY s.created_at DESC
LIMIT $1;


