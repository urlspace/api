# Plan: Session-based login (authSignin) handler

## Context

The API has signup, email verification, and resend verification handlers. The next step is login. The user wants session-based auth with HttpOnly cookies. API tokens for third-party/plugin use will be a separate feature later.

Everything is stored in a single `tokens` table — a browser session and a user-managed API token are both tokens. The `type` column distinguishes the origin, and a nullable `description` column allows tokens to be labelled (e.g. `"Token to interact with browser extension"`). The same auth middleware can later validate both (cookie or Authorization header).

**Key decisions:**
- Cookie-only — no token in response body
- Require verified email to log in
- Sliding expiration: 30 days of inactivity → token expires. Refresh is throttled (future middleware concern, not part of this handler)
- Generic error message ("invalid email or password") for both wrong email and wrong password
- `type TEXT NOT NULL DEFAULT 'session' CHECK (type IN ('session', 'token'))` — extensible without a PostgreSQL ENUM; `'session'` = created by sign-in, `'token'` = created manually via UI
- `description TEXT` — nullable; sign-in creates tokens without one, manual token creation will supply a string
- Index on `user_id` for listing/invalidating a user's tokens

---

## Steps

### Step 1 — `sql/migrations/000001_init.up.sql` (modify)
Append the tokens table and its trigger at the end of the file, after the resources section.

```sql
-- tokens
CREATE TABLE tokens (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type TEXT NOT NULL DEFAULT 'session' CHECK (type IN ('session', 'token')),
    description TEXT,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX ON tokens (user_id);

CREATE TRIGGER update_tokens_updated_at
BEFORE UPDATE ON tokens
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();
```

### Step 2 — `sql/migrations/000001_init.down.sql` (modify)
Add drops for tokens **before** the existing drops (tokens references users):

```sql
DROP TRIGGER IF EXISTS update_tokens_updated_at ON tokens;
DROP TABLE IF EXISTS tokens;
```

Insert these two lines at the top, before the existing trigger/table drops.

### Step 3 — `sql/queries/tokens.sql` (create)
New file following the formatting style in `sql/queries/users.sql`:

```sql
-- name: CreateToken :one
INSERT INTO tokens (user_id, type, description, expires_at)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetTokenById :one
SELECT * FROM tokens
WHERE id = $1
LIMIT 1;

-- name: UpdateTokenExpiresAt :one
UPDATE tokens
SET expires_at = $2
WHERE id = $1
RETURNING *;

-- name: DeleteToken :exec
DELETE FROM tokens
WHERE id = $1;
```

### Step 4 — `sqlc.yml` (modify)
Add overrides for tokens timestamp columns (non-pointer `time.Time`, matching resources/users pattern):

```yaml
- column: "tokens.expires_at"
  go_type:
    type: "time.Time"
- column: "tokens.created_at"
  go_type:
    type: "time.Time"
- column: "tokens.updated_at"
  go_type:
    type: "time.Time"
```

After this step, run `sqlc generate` to produce `internal/db/tokens.sql.go`, update `internal/db/models.go` with the `Token` struct, and update `internal/db/querier.go` with token methods.

### Step 5 — `internal/store/store_tokens.go` (create)
Following the exact pattern of `internal/store/store_users.go`:

- `TokenCreateParams` struct: `UserID uuid.UUID`, `Type string` (`"session"` or `"token"`), `Description *string` (pointer, nullable), `ExpiresAt time.Time`
- `TokenUpdateExpiresAtParams` struct: `ID uuid.UUID`, `ExpiresAt time.Time`
- `TokenStore` interface with `Create`, `GetById`, `UpdateExpiresAt`, `Delete`
- `tokenStore` concrete type backed by `db.Querier`
- `NewTokenStore(queries db.Querier) TokenStore` constructor

### Step 6 — `internal/store/store.go` (modify)
- Add `Tokens TokenStore` field to `Store` struct (alphabetically: after Resources, before Users)
- Wire `NewTokenStore(queries)` in `New()`

### Step 7 — `internal/handlers/constants.go` (modify)
Add alongside the existing `TokenExpiryDuration`:

```go
const SessionExpiryDuration = 30 * 24 * time.Hour
const SessionCookieName = "session_id"
const TokenTypeSession = "session"
const TokenTypeManaged = "token"
```

Note: `SessionExpiryDuration` and `SessionCookieName` refer to the browser session concept (the cookie); `TokenTypeSession`/`TokenTypeManaged` are the values stored in the `tokens.type` column.

### Step 8 — `internal/handlers/authSignin.go` (create)
Handler following the established pattern (decode → normalize → validate → business logic → response):

- `AuthSigninBody` with `Email` and `Password` fields
- `Normalize()` — lowercase + trim email, leave password untouched
- `Validate()` — reuse `validator.Email()` and `validator.Password()`
- `AuthSigninResponse` with `Status string` and `Data string`
- Handler flow:
  1. Decode + `DisallowUnknownFields`
  2. `Normalize()` + `Validate()`
  3. `s.Users.GetByEmail()` → on any error, return 401 `"invalid email or password"`
  4. `utils.PasswordValidate()` → on failure, return 401 `"invalid email or password"`
  5. Check `u.EmailVerified` → if false, return 403 `"email not verified"`
  6. Read `r.Header.Get("User-Agent")` into a local variable; pass it as `Description` (as a `*string`) to `s.Tokens.Create()` with `Type = TokenTypeSession`, `ExpiresAt = time.Now().Add(SessionExpiryDuration)`
  7. `http.SetCookie()` — `HttpOnly: true`, `Secure: true`, `SameSite: http.SameSiteLaxMode`, `Path: "/"`, `Name: SessionCookieName`, `Value: token.ID.String()`, `Expires: token.ExpiresAt`
  8. Respond 200 `{"status":"ok","data":"ok"}`

### Step 9 — `internal/handlers/authSignin_test.go` (create)
Integration tests matching the pattern in `authSignup_test.go`.

Test users seeded via `s.Users.Create()` with pre-hashed passwords using `utils.PasswordHash()`.

- `"fails on incorrect body"` → 400
- `"fails on unexpected field in body"` → 400
- `"fails on invalid request body"` (empty email) → 400
- `"fails on non-existent email"` → 401, `"invalid email or password"`
- `"fails on wrong password"` → 401, `"invalid email or password"`
- `"fails on unverified email"` → 403, `"email not verified"`
- `"success"` → 200, check response body + `Set-Cookie` header (name=`session_id`, HttpOnly, Secure, valid UUID value, Expires ~30 days)

### Step 10 — `internal/handlers/helpers_test.go` (modify)
Update the TRUNCATE statement to include tokens:

```go
pool.Exec("TRUNCATE users, resources, tokens RESTART IDENTITY CASCADE")
```

### Step 11 — `internal/server/server.go` (modify)
Uncomment and wire the signin route:

```go
mux.HandleFunc("POST /auth/signin", handlers.AuthSignin(s))
```

---

## Files auto-generated (by sqlc generate, not manually edited)
- `internal/db/models.go` — gains `Token` struct
- `internal/db/querier.go` — gains token methods
- `internal/db/tokens.sql.go` — new file

---

## Verification

After implementation, the user should run:
1. Re-run the migration (drop + up) on dev and test databases since the init migration was modified
2. `sqlc generate`
3. `go build ./...`
4. `make test`
5. Manual: `curl -i -X POST http://localhost:8080/v1/auth/signin -d '{"email":"...","password":"..."}'` and confirm `Set-Cookie: session_id=...` in response headers
