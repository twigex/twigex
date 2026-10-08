# Backend Developer Guidelines

## Table of Contents

1. [Architecture Overview](#architecture-overview)
2. [Layer Responsibilities](#layer-responsibilities)
3. [Error Handling](#error-handling)
4. [Logging](#logging)
5. [HTTP Status Codes](#http-status-codes)
6. [API Layer](#api-layer)
7. [App Layer](#app-layer)
8. [Store Layer](#store-layer)
9. [Timestamp Units](#timestamp-units)
10. [i18n Error IDs](#i18n-error-ids)
11. [AI Use Policy](#ai-use-policy)
12. [Comments](#comments)
13. [Dead Code](#dead-code)
14. [Common Mistakes](#common-mistakes)
15. [Quick Reference](#quick-reference)

---

## Architecture Overview

The backend is split into three distinct layers. Each layer has a single responsibility and must not cross into another layer's concerns.

```
HTTP Request
     |
 API Layer       — parse request, authenticate, call app, respond
     |
 App Layer       — business logic, permissions, orchestration
     |
 Store Layer     — database queries, no logic
     |
  Database
```

The golden rule: each layer only talks to the layer directly below it.

- API layer calls App layer only — never Store directly
- App layer calls Store layer only — never HTTP types
- Store layer executes queries only — no business logic, no logging

---

## Layer Responsibilities

### API Layer (`api/`)

- Parse and validate the HTTP request
- Authenticate the current user via `GetCurrentUser`
- Call the appropriate App layer function
- Respond using `respondJSON` or `respondAppError`
- Never access `a.app.Store` directly
- Never contain business logic or permission checks

### App Layer (`app/`)

- Own all business logic
- Own all permission checks — never rely on the API layer to check permissions
- Call Store layer functions
- Log errors with context using `tlog.Errorw`
- Return `*model.AppError` for all errors
- Never import or reference HTTP types (`http.Request`, `http.ResponseWriter`)
- Exception: when streaming data, accept `io.Writer` instead of `http.ResponseWriter` — see [Streaming](#streaming)

### Store Layer (`store/`)

- Execute database queries only
- Return raw data or errors — no wrapping, no logging
- Never contain business logic
- Never call App layer functions

---

## Error Handling

### AppError

All errors in the App layer must be returned as `*model.AppError`:

```go
return nil, model.NewAppError("error.id", http.StatusInternalServerError)
```

The first argument is an i18n ID that maps to a human-readable translation in `i18n/en.json`. Never put raw English strings here.

### Error propagation rules

| Situation                                                          | Action                                                 |
| ------------------------------------------------------------------ | ------------------------------------------------------ |
| App layer receives `error` from Store                              | Log it, wrap in `model.NewAppError`, return            |
| App layer receives `*model.AppError` from another app function     | Pass through directly, do not re-wrap                  |
| API layer receives `*model.AppError` from App                      | Pass to `respondAppError`, do not re-wrap              |
| API layer receives `error` from stdlib (cookie, JSON decode, etc.) | Wrap in `model.NewAppError`, pass to `respondAppError` |

### What not to do

```go
// Never put raw error messages in AppError
return nil, model.NewAppError(err.Error(), http.StatusInternalServerError)

// Never put English strings in AppError
return nil, model.NewAppError("Failed to create user", http.StatusInternalServerError)

// Never re-wrap an AppError
if appErr != nil {
    return nil, model.NewAppError(appErr.Message, appErr.Status)
}

// Never use http.Error in handlers
http.Error(w, "Bad request", http.StatusBadRequest)
```

### What to do

```go
// Use i18n IDs
return nil, model.NewAppError("user.create_failed", http.StatusInternalServerError)

// Pass AppError through directly
if appErr != nil {
    return nil, appErr
}

// Use respondAppError in handlers
respondAppError(w, r, appErr)
```

### Private helper functions

Private helpers should not log — their callers log. Return the error:

```go
// Private helper — no logging
func (a *App) copyThumbnailImageToDisk(ctx context.Context, f model.File, targetDir model.File) error {
    object, err := a.FileStorageObjects[f.Storage].ReadFile(ctx, f.ID)
    if err != nil {
        return err
    }
    defer object.Close()

    return a.writeThumbnail(ctx, targetDir, object)
}

// Public app function — logs before returning
func (a *App) MoveFile(...) (*model.FileMove, *model.AppError) {
    if err := a.copyThumbnailImageToDisk(ctx, file, targetDir); err != nil {
        tlog.Errorw("Failed to copy thumbnail",
            "file_id", file.ID,
            "error", err,
        )
        return nil, model.NewAppError("file.move_failed", http.StatusInternalServerError)
    }
}
```

---

## Logging

Use `tlog.Errorw`, `tlog.Warnw`, and `tlog.Infow` — never the non-structured `tlog.Error`.

### Format

```go
tlog.Errorw("What went wrong, past tense",
    "context_field", value,
    "error", err, // always last, always pass err directly — never err.Error()
)
```

### Level decision

| Situation                             | Level    |
| ------------------------------------- | -------- |
| DB or store call failed               | `Errorw` |
| Resource not found (valid input)      | No log   |
| Client sent bad or unverifiable input | `Warnw`  |
| Permission denied (expected)          | No log   |
| Session expired                       | No log   |
| Best-effort cleanup after delete      | `Warnw`  |

### Rules

- Message: capital first letter, past tense, no punctuation — `"Failed to retrieve user"`
- Always include relevant context fields before `"error"` — `"user_id"`, `"file_id"`, `"workspace_id"` etc.
- Never log PII — no emails, names, passwords
- Never use `err.Error()` as the error value — pass `err` directly
- `Warnw` does not include an `"error"` field unless there is an underlying error
- `Infow` never includes an `"error"` field
- Every `Errorw` must be paired with a returned `AppError` — never log and continue silently

---

## HTTP Status Codes

| Code  | When to use                                         |
| ----- | --------------------------------------------------- |
| `200` | Success                                             |
| `400` | Client sent invalid input                           |
| `401` | Not authenticated                                   |
| `402` | Feature requires paid plan                          |
| `403` | Authenticated but no permission                     |
| `404` | Valid input, resource does not exist                |
| `409` | Conflict                                            |
| `500` | Server, DB, filesystem, or external service failure |
| `507` | Storage quota exceeded                              |

Common mistakes to avoid:

- `404` on a DB failure — the resource might exist, the query just failed. Use `500`.
- `400` on a DB failure — that is not a client error. Use `500`.
- `403` when the user is not logged in at all. Use `401`.

---

## API Layer

### Handler template

```go
func (a *API) exampleHandler(w http.ResponseWriter, r *http.Request) {
    // 1. Extract route params
    resourceID := mux.Vars(r)["id"]

    // 2. Decode body (if needed)
    var s model.SomeRequest
    if !decodeBody(w, r, &s) {
        return
    }

    // 3. Get current user
    user, appErr := a.app.GetCurrentUser(r)
    if appErr != nil {
        respondAppError(w, r, appErr)
        return
    }

    // 4. Call app layer
    result, appErr := a.app.DoSomething(*user, resourceID, s)
    if appErr != nil {
        respondAppError(w, r, appErr)
        return
    }

    // 5. Respond
    respondJSON(w, http.StatusOK, result)
}
```

### Response helpers

Always use these — never `json.NewEncoder`, never `http.Error`:

```go
respondJSON(w, http.StatusOK, data)   // success with body
respondAppError(w, r, appErr)         // error response
w.WriteHeader(http.StatusOK)          // success with no body
```

### Request decoding

Always use `decodeBody` — never `io.ReadAll` + `json.Unmarshal`:

```go
var s model.SomeRequest
if !decodeBody(w, r, &s) {
    return  // decodeBody already responded with an error
}
```

### Building URLs and query strings

Always use `url.Values` — never string concatenation. String concatenation allows user input to inject additional query parameters:

```go
// Safe
params := url.Values{}
params.Set("q", userInput)
params.Set("key", apiKey)
url := "https://api.example.com/search?" + params.Encode()

// Unsafe — user input can inject parameters
url := "https://api.example.com/search?q=" + userInput + "&key=" + apiKey
```

---

## App Layer

### Function template

```go
func (a *App) DoSomething(ctx context.Context, user model.User, id string) (*model.Result, *model.AppError) {
    // 1. Permission check first
    if !a.HasPermission(user, id) {
        return nil, model.NewAppError("resource.forbidden", http.StatusForbidden)
    }

    // 2. Wrap context with query timeout cap
    ctx, cancel := a.dbCtx(ctx)
    defer cancel()

    // 3. Business logic
    data, err := a.Store.Resource.Get(ctx, id)
    if err != nil {
        tlog.Errorw("Failed to retrieve resource",
            "resource_id", id,
            "user_id", user.ID,
            "error", err,
        )
        return nil, model.NewAppError("resource.retrieval_failed", http.StatusInternalServerError)
    }

    if data == nil {
        return nil, model.NewAppError("resource.not_found", http.StatusNotFound)
    }

    return data, nil
}
```

### Database context

Always wrap the incoming context with `dbCtx` before passing it to the store. This caps every query at `QueryTimeout` while still inheriting `r.Context()` cancellation — if the client disconnects before the timeout, the query cancels immediately.

```go
// dbCtx is a helper on App that caps the context at the configured query timeout
func (a *App) dbCtx(ctx context.Context) (context.Context, context.CancelFunc) {
    return context.WithTimeout(ctx, time.Duration(*a.settings.SqlSettings.QueryTimeout)*time.Second)
}
```

This gives you both behaviours automatically:

- Client disconnects early → `r.Context()` cancels the query immediately
- Query runs too long → `QueryTimeout` cancels it at the cap

For background jobs that have no `r.Context()`, pass `context.Background()` and `dbCtx` will still apply the timeout:

```go
// Background job — no request context, timeout still applied
ctx, cancel := a.dbCtx(context.Background())
defer cancel()
data, err := a.Store.X.Get(ctx, id)
```

Never call `a.dbCtx` more than once per app function — wrap once at the top and reuse the same `ctx` for all store calls within that function.

### Permission checks

Permissions always live in the App layer — never the API layer. This ensures that any caller (HTTP handler, background job, migration, test) is subject to the same rules.

```go
// Permission check in app layer — correct
func (a *App) DeleteFile(user model.User, id string) *model.AppError {
    if !a.isOwner(user, id) {
        return model.NewAppError("file.forbidden", http.StatusForbidden)
    }
    // ...
}

// Permission check in API layer — wrong
func (a *API) deleteFile(w http.ResponseWriter, r *http.Request) {
    if user.Role != "admin" {
        respondAppError(w, r, model.NewAppError("file.forbidden", http.StatusForbidden))
        return
    }
}
```

### DB error vs not found

Always distinguish between a DB failure and a missing resource:

```go
data, err := a.Store.Resource.Get(ctx, id)
if err != nil {
    // DB failed — we don't know if it exists
    tlog.Errorw("Failed to retrieve resource", "id", id, "error", err)
    return nil, model.NewAppError("resource.retrieval_failed", http.StatusInternalServerError)
}

if data == nil {
    // DB succeeded, resource does not exist
    return nil, model.NewAppError("resource.not_found", http.StatusNotFound)
}
```

### Streaming

When the App layer needs to stream data (e.g. writing a ZIP file), accept `io.Writer` instead of `http.ResponseWriter`. The API layer passes `http.ResponseWriter` as the `io.Writer` since it satisfies the interface.

```go
// App layer — accepts io.Writer, no HTTP dependency
func (a *App) CreateZip(user model.User, fileID string, w io.Writer) (*string, *model.AppError) {
    zw := zip.NewWriter(w)
    defer zw.Close()
    // ...
}

// API layer — passes http.ResponseWriter as io.Writer
func (a *API) downloadZip(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/zip")
    _, appErr := a.app.CreateZip(*user, fileID, w)
    if appErr != nil {
        respondAppError(w, r, appErr)
        return
    }
}
```

### Recording Activities

Always use `RecordActivity` — never call `a.Store.Activity.Create` directly. Use `model.App*` constants for context and `model.Activity*` constants for actions.

```go
// Correct
a.RecordActivity(user.ID, model.AppFiles, model.ActivityFileCreate, file.Parent, file.ID, map[string]any{
    "name": file.DisplayName,
    "type": file.Type,
})

// Wrong — never call store directly, never use raw strings
j, _ := json.Marshal(m)
a.Store.Activity.Create(user.ID, "files", "file_create", file.Parent, file.ID, string(j))
```

If the response has already been written (e.g. after `ServeFile`), record in a goroutine:

```go
go a.app.RecordActivity(user.ID, model.AppFiles, model.ActivityFileView, file.Parent, file.ID, map[string]any{
    "name": file.DisplayName,
    "type": file.Type,
})
```

Available context constants — `model.AppFiles`, `model.AppProjects`, `model.AppChat`, `model.AppCollimato`.

Available action constants — `model.ActivityFileCreate`, `model.ActivityFileView`, `model.ActivityFileUpload`, `model.ActivityFileDelete`, `model.ActivityFileRename`, `model.ActivityFileRestore`, `model.ActivityFileShare`, `model.ActivityFileMeta`, `model.ActivityFileMove`.

---

## Store Layer

### Naming

A store method is read as `Store.Channels.Get`, so the interface already says
what the method acts on. Do not say it again in the method name.

```go
// Correct
Store.Channels.Get(id)
Store.Channels.GetMembersByIDs(ids)

// Wrong
Store.Channels.GetChannelByID(id)
Store.Channels.FindChannelMembersByIDs(ids)
```

The exception is a store holding more than one kind of row. `CollimatoStore`
keeps connections, charts, dashboards, workspaces, roles and files, so
`GetWorkspaceRoles` keeps its noun: there it is what tells the methods apart,
not a repetition of the receiver.

One verb per kind of write, so the name says which statement runs:

| Verb     | Means                                                     |
| -------- | --------------------------------------------------------- |
| `Create` | makes a new row                                           |
| `Update` | changes rows that exist                                   |
| `Delete` | removes the row itself                                    |
| `Add`    | puts an existing thing into a collection; creates nothing |
| `Remove` | takes it back out; deletes nothing                        |

The group store shows all four:

```go
Store.Groups.Create(...)          // INSERT INTO user_groups   — the group now exists
Store.Groups.AddMembers(id, ids)  // INSERT INTO group_members — existing users join it
Store.Groups.RemoveMember(id, u)  // DELETE FROM group_members — the user still exists
Store.Groups.SoftDelete(id)       // the group is gone
```

Not `Save`, `Set`, `Insert`, `Edit`, `Change` or `Upsert`: each of those has
been used for two different statements in this repository, and a reader cannot
tell from the call which one they get.

Reads are `Get`, never `Find`. `Get` on its own for one row by id, `GetAll` for
the whole set, `GetBy<Field>` to look one up another way, `GetAllFor<Owner>`
for a subset. `Search` is for a typeahead or a `LIKE` query, `Count` for a
count.

A name must not hide a write. `ClaimNextPending` selects `FOR UPDATE SKIP
LOCKED` and marks the row running, so it cannot be called `GetNextPending`: a
caller would reasonably run that twice.

If two methods differ by a filter, the name carries the filter. `GetByIDs` and
`GetActiveByIDs` read the same table; only the second excludes deactivated
users, and a caller has to be able to see that at the call site.

### Returning raw data

Store functions return raw data and errors only. No logging, no AppError, no business logic:

```go
// Correct store function
func (s *FileStore) Get(ctx context.Context, id string) (*model.File, error) {
    var file model.File
    err := s.db.QueryRowContext(ctx,
        "SELECT id, display_name, parent FROM files WHERE id = ?", id,
    ).Scan(&file.ID, &file.DisplayName, &file.Parent)
    if errors.Is(err, sql.ErrNoRows) {
        return nil, nil // not found — return nil, nil
    }

    return &file, err
}

// Wrong — do not do this in the store layer
func (s *FileStore) Get(id string) (*model.File, *model.AppError) {
    tlog.Errorw("Failed to find file", ...)
}
```

### Returning collections

Return a slice, never a pointer to one. A slice is already a reference; the
pointer adds a dereference at every use and a nil case that means nothing a
caller can act on.

```go
// Correct
func (s *UserStore) GetByIDs(ctx context.Context, ids []string) ([]model.User, error)

// Wrong
func (s *UserStore) GetByIDs(ctx context.Context, ids []string) (*[]model.User, error)
```

Build the slice with `make`, so an empty result is an empty slice rather than a
nil one. Callers can then range over it and take its length without checking
first, and `len(x) == 0` is the only test any of them needs.

```go
users := make([]model.User, 0)
for rows.Next() {
    // ...
}
return users, nil
```

A single row stays a pointer: `*model.User` with `nil, nil` for not found,
which is a distinction a caller does act on.

### Context and query timeouts

Store functions must accept a `context.Context` as their first argument and pass it to every query. Never use `context.Background()` inside a store function.

Context flows from the HTTP request down through every layer — never created inside the store:

```
r.Context() → App layer → Store layer → db.QueryContext
```

```go
// API layer — passes r.Context() to app
func (a *API) getFile(w http.ResponseWriter, r *http.Request) {
    result, appErr := a.app.GetFile(r.Context(), id)
}

// App layer — passes ctx down to store
func (a *App) GetFile(ctx context.Context, id string) (*model.File, *model.AppError) {
    file, err := a.Store.File.Get(ctx, id)
}

// Store layer — passes ctx to query
func (s *FileStore) Get(ctx context.Context, id string) (*model.File, error) {
    err := s.db.QueryRowContext(ctx,
        "SELECT id, display_name, parent FROM files WHERE id = ?", id,
    ).Scan(&file.ID, &file.DisplayName, &file.Parent)
}
```

This gives you request cancellation for free — if the client disconnects, the query cancels automatically. It also propagates tracing spans and HTTP handler deadlines down to the database.

For background jobs or migrations that have no incoming request context, use the configured query timeout instead:

```go
ctx, cancel := context.WithTimeout(context.Background(),
    time.Duration(*settings.QueryTimeout) * time.Second,
)
defer cancel()
```

### Migration rule

When touching any store function for any reason — bug fix, new feature, refactor — migrate it to accept `context.Context` and use `QueryRowContext` / `QueryContext` / `ExecContext`. Do not leave it on the old pattern. Do not migrate files you are not already changing.

### Allowed conditional logic

Store functions may contain minimal conditional logic, but only when it is mechanical or query-shaping — never when it encodes a business or domain decision.

**Translating driver-level sentinels:**

```go
// Allowed — translating sql.ErrNoRows into Go-idiomatic nil, nil
func (s *FileStore) Get(ctx context.Context, id string) (*model.File, error) {
    var file model.File
    err := s.db.QueryRowContext(ctx,
        "SELECT id, display_name, parent FROM files WHERE id = ?", id,
    ).Scan(&file.ID, &file.DisplayName, &file.Parent)
    if errors.Is(err, sql.ErrNoRows) {
        return nil, nil
    }

    return &file, err
}
```

**Avoiding a malformed query:**

```go
// Allowed — an empty IN clause would be invalid SQL
func (s *FileStore) GetByIDs(ctx context.Context, ids []string) ([]model.File, error) {
    files := make([]model.File, 0)
    if len(ids) == 0 {
        return files, nil
    }

    // build and execute query, appending to files
}
```

**Optional filters for dynamic queries:**

```go
// Allowed — conditionally appending WHERE clauses based on which filters are set
func (s *FileStore) GetAll(ctx context.Context, filter model.FileFilter) ([]model.File, error) {
    query := "SELECT id, display_name, parent FROM files WHERE 1 = 1"
    args := make([]any, 0)

    if filter.ParentID != "" {
        query += " AND parent = ?"
        args = append(args, filter.ParentID)
    }

    if filter.CreatedAfter != 0 {
        query += " AND created_at > ?"
        args = append(args, filter.CreatedAfter)
    }

    // execute query with args
}
```

### Forbidden conditional logic

Any branch whose reason can be described in product or domain terms belongs in the App layer, not the Store.

```go
// Wrong — permission decision disguised as a query condition
func (s *FileStore) Get(userRole string, id string) (*model.File, error) {
    if userRole != "admin" {
        return nil, errors.New("forbidden")
    }
    // ...
}

// Wrong — business rule encoded in the store
func (s *FileStore) Delete(id string) error {
    file, err := s.Get(id)
    if file.IsLocked {
        return errors.New("file is locked")
    }
    // ...
}

// Wrong — orchestration across multiple store calls
func (s *FileStore) CreateWithParent(file model.File) error {
    parent, err := s.Get(file.ParentID)
    if parent == nil {
        s.CreateDirectory(file.ParentID)
    }
    // ...
}
```

### The mental test

Before adding a conditional to a Store function, ask: _"why does this branch exist?"_

| Answer                                                           | Layer |
| ---------------------------------------------------------------- | ----- |
| Because the SQL driver returns a sentinel error for missing rows | Store |
| Because an empty slice would produce invalid SQL                 | Store |
| Because the query needs a WHERE clause only when a filter is set | Store |
| Because locked files cannot be deleted                           | App   |
| Because only admins can see this resource                        | App   |
| Because creating a file requires a parent to exist first         | App   |

If a product manager or designer could explain why the branch exists — it belongs in the App layer.

---

## Timestamp Units

Most timestamp columns store Unix **seconds**. Chat-related columns
(`posts.*`, `channel_members.last_viewed_at`, `channel_members.updated_at`,
`channel_members.date_joined`, etc.) store **milliseconds** because
sub-second ordering matters. Always check the model field's `//` comment
before writing — the unit isn't visible from the column type (`BIGINT`).

Writing `time.Now().Unix()` into a milliseconds column (or vice versa)
produces silently wrong data: the value is off by 1000× and tests that
don't compare timestamps will happily pass.

---

## i18n Error IDs

All `model.NewAppError` calls must use an ID from `i18n/en.json`. Never use raw English strings.

### Naming convention

```
<namespace>.<action_or_reason>

file.not_found
user.create_failed
workspace.forbidden
upload.invalid_size
auth.invalid_credentials
```

### Adding new IDs

When you need a new error ID:

1. Add it to `i18n/en.json` with a user-friendly translation
2. Run `make sync-i18n` to add the ID to the other locales
3. Use the ID in `model.NewAppError`
4. Run `make check-i18n`, which fails if any ID is missing a translation

```json
{
  "id": "resource.create_failed",
  "translation": "Failed to create resource. Please try again."
}
```

Before adding an ID, search `i18n/en.json` for one that already says the same
thing; reuse beats a near-duplicate.

---

## AI Use Policy

The AI use policy lives in [CONTRIBUTING.md § AI-assisted contributions](../../CONTRIBUTING.md#ai-assisted-contributions).
Everything in this guide applies to AI-generated code exactly as it does to code
written by hand.

---

## Comments

Default to writing no comments. Well-named functions, types, and variables are self-documenting, and the layer split already explains the structure.

Only add a comment when the _why_ is non-obvious: a hidden constraint, a subtle invariant, an ordering or timing requirement, a workaround for an external bug, or an access/security rationale that would genuinely surprise a reader.

```go
// Stop local tracks before disconnecting, or the camera stays on.
stopLocalTracks()
```

**Format rules:**

- Prefer one short line of plain prose; a brief multi-line block is fine when a single line genuinely can't explain the _why_.
- Don't restate what the code or signature already says (`// GetUser returns a user` adds nothing).
- When a comment explains a workaround, link the upstream issue, so a reader can tell when the workaround can go.
- Give a TODO an owner or an issue link (`// TODO(username):` or `// TODO(#123):`), so it can be chased rather than inherited.
- Write comments in English, so every contributor can read them.

**The test:** if removing the comment wouldn't confuse a future reader, delete it. If it describes _what_ the code does rather than _why_, delete it.

---

## Dead Code

Delete code you are replacing rather than commenting it out. Git remembers it,
and in the file nobody can tell whether it is a draft, a rollback, or something
switched off on purpose, so it is never safe for the next person to remove.

The same goes for a function nothing calls, a branch whose body is empty, and a
stub that returns nothing useful. Each one reads as working code until somebody
traces it.

---

## Common Mistakes

| Mistake                                            | Fix                                       |
| -------------------------------------------------- | ----------------------------------------- |
| Accesses Store from the API layer                  | Move logic to App layer                   |
| Uses `http.Error` or `json.NewEncoder` in handlers | Use `respondJSON` / `respondAppError`     |
| Puts English strings in `AppError`                 | Use a key from `i18n/en.json`             |
| Checks permissions in the API layer                | Move check to App layer function          |
| Logs an error then continues without returning     | Log, then return the `AppError`           |
| Re-wraps a `*model.AppError` in a new `AppError`   | Pass `AppError` through directly          |
| Uses `err.Error()` as the log error value          | Pass `err` directly                       |
| Concatenates user input into URLs                  | Use `url.Values`                          |
| Calls `a.Store.Activity.Create` directly           | Use `RecordActivity` with model constants |

---

## Quick Reference

### Correct patterns

```go
// Decode request body
var s model.Request
if !decodeBody(w, r, &s) {
    return
}

// Respond with data
respondJSON(w, http.StatusOK, result)

// Respond with error
respondAppError(w, r, appErr)

// Log an error
tlog.Errorw("Failed to retrieve file",
    "file_id", id,
    "error", err,
)

// Return AppError
return nil, model.NewAppError("file.not_found", http.StatusNotFound)

// Pass AppError through
if appErr != nil {
    return nil, appErr
}

// DB error vs not found
data, err := a.Store.X.Get(ctx, id)
if err != nil {
    // 500 — query failed
}
if data == nil {
    // 404 — does not exist
}

// Pass r.Context() through all layers, capped with dbCtx in app
func (a *API) handler(w http.ResponseWriter, r *http.Request) {
    result, appErr := a.app.DoSomething(r.Context(), id)
}
func (a *App) DoSomething(ctx context.Context, id string) (*model.Result, *model.AppError) {
    ctx, cancel := a.dbCtx(ctx)
    defer cancel()
    data, err := a.Store.X.Get(ctx, id)
}
func (s *Store) Get(ctx context.Context, id string) (*model.X, error) {
    s.db.QueryRowContext(ctx, "SELECT ...")
}
```

### Forbidden patterns

```go
// Raw strings in AppError
model.NewAppError("Failed to do thing", ...)
model.NewAppError(err.Error(), ...)

// Accessing store from API layer
a.app.Store.User.Get(id)

// http.Error in handlers
http.Error(w, "Bad request", http.StatusBadRequest)

// json.NewEncoder in handlers
json.NewEncoder(w).Encode(data)

// Logging and continuing without returning
tlog.Errorw("Something failed", "error", err)
// ... continuing as if nothing happened

// Permission checks in API layer
if user.Role != model.SystemAdminRoleId { ... }

// User input concatenated into URLs
"https://api.example.com?q=" + userInput

// Committing commented-out code, or a function nothing calls

// Store function without context
func (s *FileStore) Get(id string) (*model.File, error)

// Using context.Background() inside a store function
s.db.QueryRowContext(context.Background(), ...)

// Using query methods without context
s.db.QueryRow(...)
s.db.Exec(...)
s.db.Query(...)

// Calling dbCtx more than once in the same app function
ctx1, cancel1 := a.dbCtx(ctx)
defer cancel1()
ctx2, cancel2 := a.dbCtx(ctx) // wrong — wrap once at the top, reuse ctx
defer cancel2()
```
