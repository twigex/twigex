# Testing guide — cloud-go

Tests live in this repo in a few kinds, and they're set up very differently. Pick based on what you're testing.

| You're testing... | Use | Needs a DB? |
|---|---|---|
| A SQL query / store function | **Store test** (real MySQL) | Yes (`make test-db`) |
| Business logic, permissions, validation in `app/` | **App test** (fake store) | No |
| A pure helper (no DB, no app) | Plain Go test | No |
| A Vue component | **Frontend test** (Vitest) | No |

As a rule: test the SQL against a real database, and test the decisions with a fake. Do not imitate SQL behavior in a fake, and do not start a database to check an `if`.

---

## 1. Store-layer test (real MySQL)

These run against a throwaway schema. The harness (`store/sqlstore/testhelper_test.go`) gives you `requireDB(t)` (a `*sql.DB`) and `cleanTables(t, ...)`. You run them with:

```bash
export TEST_MYSQL_DSN='admin:password@tcp(127.0.0.1:3306)/?multiStatements=true'
make test-db
```

If `TEST_MYSQL_DSN` is unset, `requireDB` skips the test instead of failing — so these are safe in CI without a DB.

### Boilerplate

```go
package sqlstore

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/twigex/twigex/model"
)

func resetThingTables(t *testing.T) {
	cleanTables(t,
		"child_table",
		"thing",
	)
}

func seedThing(t *testing.T, db *sql.DB) string {
	t.Helper()
	id := model.NewID()
	now := time.Now().Unix()
	_, err := db.Exec(
		`INSERT INTO thing (id, name, created_at, deleted_at)
		 VALUES (?, ?, ?, 0)`,
		id, "thing-"+id[:8], now,
	)
	if err != nil {
		t.Fatalf("seedThing: %v", err)
	}
	return id
}

func TestCreateThing(t *testing.T) {
	db := requireDB(t)
	resetThingTables(t)

	repo := &thingRepository{Db: db}
	ctx := context.Background()

	id := seedThing(t, db)

	got, err := repo.Get(ctx, id)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got == nil {
		t.Fatal("expected a thing, got nil")
	}
	if got.ID != id {
		t.Errorf("want id %s, got %s", id, got.ID)
	}
}
```

### Rules to follow

1. **First two lines, always:** `db := requireDB(t)` then `resetThingTables(t)`. The reset makes tests order-independent — never assume an empty DB.
2. **Build the repo by hand:** `&thingRepository{Db: db}`. You're calling one store struct directly, not the whole app.
3. **Every seed helper gets `t.Helper()`** — then a failure points at your test line, not inside the helper.
4. **Seed with the *applied* schema, not the migration file.** Columns get added/dropped by later migrations. If an insert fails with `Field 'x' doesn't have a default value`, check what columns actually exist:
   ```sql
   SELECT COLUMN_NAME FROM information_schema.columns
   WHERE table_schema = DATABASE() AND table_name = 'thing'
   ORDER BY ORDINAL_POSITION;
   ```
5. **Mind timestamp units.** Most columns are seconds (`time.Now().Unix()`), chat columns are milliseconds (`UnixMilli()`). The wrong unit stores wrong data without failing the test.
6. **`t.Fatalf` in setup, `t.Errorf` in assertions.** If seeding fails the test can't continue (fatal). If one of several checks fails, keep going so you see all of them (error).

### What to test in SQL

The things plain Go can't tell you: `NOT EXISTS` clauses, `UNION` dedupe, `ON DUPLICATE KEY UPDATE`, soft-delete filters, joins skipping deleted rows. If a test would pass without a DB, it doesn't belong here.

---

## 2. App-layer test (fake store, no DB)

For permissions, validation, error codes, ordering of checks. The trick is a **fake store** so there's no database.

### Boilerplate

```go
package app

import (
	"context"
	"net/http"
	"testing"

	"github.com/twigex/twigex/model"
	"github.com/twigex/twigex/store"
)

type fakeThingStore struct {
	store.ThingStore // embed the interface -> all methods exist (nil)
	things           []model.Thing
}

func (f *fakeThingStore) GetByIDs(_ context.Context, _ []string) ([]model.Thing, error) {
	return f.things, nil
}

func thingApp(s *fakeThingStore) *App {
	return &App{
		Store: store.Store{Thing: s},
	}
}

func TestDoThingForbidden(t *testing.T) {
	a := thingApp(&fakeThingStore{})
	user := model.User{ID: "u1", Role: model.SystemUserRoleId}

	appErr := a.DoThing(user, "id1")
	if appErr == nil || appErr.Status != http.StatusForbidden {
		t.Fatalf("expected 403, got %v", appErr)
	}
}
```

### How the fake works

- **Embed the store interface** (`store.ThingStore` with no field name). Go then treats your struct as a full `ThingStore` even with zero methods written — so it compiles and you can drop it into `store.Store{Thing: s}`.
- The embedded interface is **nil**. So any method you *didn't* override **panics if called**. That's a feature: it tells you a code path reached a method you forgot to fake. Add that method when it happens.
- **Override only the methods the path under test actually calls.** A permission test typically needs two or three of some forty methods.
- **Make the fake behave like the real query when it matters.** A blind `return fixed` is fine for "give me X", but if the same method is called twice for different reasons, respect the arguments. `GetRolesByName` is one example: it serves both the permission check and the validation of requested roles, so a fixed return value makes one of them wrong. The fake has to filter by the names it was asked for.

### Two things specific to this codebase

**License gate is first.** Many app functions start with:
```go
if a.Server.License == nil || !*a.Server.License.Features.Collimato { ... 402 }
```
That's a plain struct read — no signing, no crypto. To get past it, hand-build the struct:
```go
Server: Server{License: &model.License{Features: &model.Features{Collimato: boolPtr(true)}}}
```
Leave `License` nil on purpose to test the 402 case (even a system admin gets 402 — the license sits above the admin shortcut).

**System admin short-circuits permissions.** `user.Role == model.SystemAdminRoleId` usually returns `true` before any store call. Use that to slide past the permission check when you want to test the *validation* that comes after (empty input, missing resource, etc.). Use a non-admin with no roles to test the 403 itself.

### Assert on status, not message

Error IDs are i18n keys; the status code is the contract:
```go
if appErr == nil || appErr.Status != http.StatusBadRequest {
	t.Fatalf("expected 400, got %v", appErr)
}
```

---

## 3. Naming and structure

- File: `thing_test.go` next to `thing.go`, same package.
- Test name says the scenario: `TestAddGroupsToWorkspaceUnknownRole`, `TestGetUserWorkspacesIncludesGroupAttached`. You should know what broke from the name alone.
- One behavior per test. Don't chain unrelated assertions.
- Group helpers (`seedX`, `resetXTables`, fakes) at the top of the file, tests below.

### Table-driven when cases are uniform

When you're checking the same thing with different inputs, a table beats copy-paste:

```go
func TestThingStatus(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want int
	}{
		{"empty", "", http.StatusBadRequest},
		{"missing", "ghost", http.StatusNotFound},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			a := thingApp(&fakeThingStore{})
			appErr := a.DoThing(admin(), tc.in)
			if appErr == nil || appErr.Status != tc.want {
				t.Fatalf("want %d, got %v", tc.want, appErr)
			}
		})
	}
}
```

Don't force it when each case needs different setup — separate functions read better then.

---

## 4. Frontend test (Vitest, no DB)

Most components need no test. Write one when the component holds logic worth
pinning down: sanitization, a state machine, a computed value, or rendering
that branches on props. A component that renders a title, a slot and a button
does not get one.

Assert what the component promises its callers, never what it looks like.

| Assert | Don't assert |
|---|---|
| Emitted events and their payload | A CSS class this repo chose |
| Sanitized or escaped output | Spacing, size, layout |
| Which branch rendered | An element that exists only to carry styling |
| A value the component computed | |

For a class or colour, the question is: **is the value computed or configured,
or is it a constant we typed into the stylesheet?** `text-gray-400` on a
placeholder is ours, so asserting it just restates the source and breaks on the
next restyle. A mention colour picked per mention kind is a decision the
component made, so assert that.

Classes that name a state — `disabled`, `active`, `selected` — are behaviour,
not appearance. Those are fine.

```bash
cd frontend && npm test          # or: make test-frontend
```

---

## 5. Checklist before a change is finished

```bash
go build ./...
make lint-go                      # gofmt and golangci-lint
go test ./app/...                 # fast, no DB
make test-db                      # store tests, needs TEST_MYSQL_DSN
```

- Store test starts with `requireDB(t)` + a reset?
- Seeds match the real applied schema (not the migration file)?
- Fake overrides only what's used, and respects args where a method is called more than once?
- Asserting on `appErr.Status`, not the message?
- New `model.NewAppError("x.y", ...)` id added to `i18n/en.json`? (`make check-i18n` passes)
- Frontend test asserts behaviour, not a class this repo chose?

---

## Worked examples in the tree

Real tests to copy from:

- Store, cascade/bulk SQL: `store/sqlstore/channel_groups_materialize_test.go`, `store/sqlstore/collimato_groups_test.go`
- App, permission/validation gates with a fake store: `app/collimato_groups_test.go`
- Harness (seed/reset/`requireDB`): `store/sqlstore/testhelper_test.go`
- Frontend, a component worth testing: `frontend/src/components/Chat/__tests__/MessageContent.spec.js`
