# Tags Feature Spec

## Overview

Tags allow users to categorize and filter their resources. Each tag belongs to a single user — if two users both create a tag named "cars", they are treated as two independent tags. Resources are never mixed across users.

## Tag Rules

- Tag names are strings containing only lowercase letters, digits, hyphens, and underscores (`^[a-z0-9_-]+$`)
- If a user submits a tag with uppercase letters, the server converts it to lowercase
- Tags are unique per user (enforced by a `UNIQUE (user_id, name)` constraint)
- Tags are created implicitly when a resource is created or updated with a tag name that does not yet exist for that user — no explicit "create tag" endpoint is needed
- Deleting a tag removes the association with resources but does not delete the resources themselves
- Tags can be renamed via a dedicated endpoint
- A resource can have at most 20 tags (enforced at the application level)

## Database Schema

### `tags` table

```sql
CREATE TABLE tags (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (user_id, name),
    CHECK (name ~ '^[a-z0-9_-]+$')
);

CREATE INDEX ON tags (user_id);

CREATE TRIGGER update_tags_updated_at
BEFORE UPDATE ON tags
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();
```

### `resource_tags` junction table

```sql
CREATE TABLE resource_tags (
    resource_id UUID NOT NULL REFERENCES resources (id) ON DELETE CASCADE,
    tag_id UUID NOT NULL REFERENCES tags (id) ON DELETE CASCADE,
    PRIMARY KEY (resource_id, tag_id)
);

CREATE INDEX ON resource_tags (tag_id);
```

- Composite primary key on `(resource_id, tag_id)` provides an implicit index on `resource_id`
- Additional index on `tag_id` covers queries going from tag to resources
- `ON DELETE CASCADE` on both foreign keys:
  - Deleting a resource removes its tag associations
  - Deleting a tag removes its resource associations (but not the resources)
- No `updated_at` on the junction table — it is a pure link with no mutable state

## API Endpoints

### New tag endpoints

| Method | Path | Description |
|---|---|---|
| `GET` | `/v1/tags` | List all tags for the authenticated user |
| `PUT` | `/v1/tags/{id}` | Rename a tag |
| `DELETE` | `/v1/tags/{id}` | Delete a tag (keeps resources) |

### Modified resource endpoints

Resource create (`POST /v1/resources`) and update (`PUT /v1/resources/{id}`) payloads gain an optional `tags` field:

```json
{
  "title": "Learn Go",
  "description": "A tutorial on Go",
  "url": "https://example.com",
  "tags": ["go", "tutorial"]
}
```

- Tags are passed as an array of strings (not IDs)
- On update, the tags array is a full replacement (not additive)
- Server lowercases all tag names, validates the format, and upserts tags that do not yet exist for the user

Resource list (`GET /v1/resources`) and get (`GET /v1/resources/{id}`) responses include tags inline.

### Querying resources by tags

```
GET /v1/resources?tags=go,tutorial&tag_mode=all
```

- `tags` — comma-separated list of tag names to filter by
- `tag_mode` — `all` (default) or `any`
  - `all` (intersection): resource must have every listed tag
  - `any` (union): resource must have at least one of the listed tags

#### Intersection query pattern

```sql
SELECT r.* FROM resources r
JOIN resource_tags rt ON r.id = rt.resource_id
JOIN tags t ON rt.tag_id = t.id
WHERE t.user_id = $1 AND t.name = ANY($2)
GROUP BY r.id
HAVING COUNT(DISTINCT t.id) = $3
```

Union is the same query without the `HAVING` clause.

### Querying resources by title search + tags

```
GET /v1/resources?search=tutorial&tags=go
```

- `search` — filters resources where `title ILIKE '%search%'`
- Can be combined with `tags` and `tag_mode`
- Future: may expand search to include `description`

## Implementation Order

1. Database migration — `tags` and `resource_tags` tables
2. SQL queries and sqlc code generation
3. Store layer methods
4. Tag handlers (`GET /v1/tags`, `PUT /v1/tags/{id}`, `DELETE /v1/tags/{id}`)
5. Update resource create/update handlers to accept and process tags
6. Update resource list/get handlers to return tags and support `?tags=` and `?tag_mode=` query parameters
7. Tag name validator
8. Tests
