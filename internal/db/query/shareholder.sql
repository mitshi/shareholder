-- name: GetShareholder :one
SELECT * FROM shareholder
WHERE id = $1 LIMIT 1;

-- name: CreateShareholder :one
INSERT INTO shareholder (
    name, email, mobile, folio_number, certificate_number, pan_number, agree_terms
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
)
RETURNING *;

-- name: UpdateEmailVerifyShareholder :one
UPDATE shareholder SET email_verified = $2
WHERE id = $1
RETURNING *;

-- name: UpdateMobileVerifyShareholder :one
UPDATE shareholder SET mobile_verified = $2
WHERE id = $1
RETURNING *;
