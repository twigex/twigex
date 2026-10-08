# Contributing to Twigex

Thanks for taking the time to contribute. Twigex is a self-hosted collaboration
platform: a Go backend, a Vue 3 frontend, and MariaDB/MySQL plus Redis behind it.
Everything in this repository is licensed under
[AGPL-3.0-only](LICENSE), and contributions are accepted under the same licence.

Twigex has **not reached 1.0**. Interfaces, settings and database schema still
change between releases, so please open an issue before starting anything large.

## Table of contents

- [Reporting a security issue](#reporting-a-security-issue)
- [Ground rules](#ground-rules)
- [Reporting bugs](#reporting-bugs)
- [Suggesting a change](#suggesting-a-change)
- [Development setup](#development-setup)
- [Repository layout](#repository-layout)
- [Coding conventions](#coding-conventions)
- [Tests](#tests)
- [Translations](#translations)
- [Database migrations](#database-migrations)
- [Dependencies](#dependencies)
- [Commits](#commits)
- [Opening a pull request](#opening-a-pull-request)
- [Review expectations](#review-expectations)
- [AI-assisted contributions](#ai-assisted-contributions)
- [Licensing of contributions](#licensing-of-contributions)
- [Getting help](#getting-help)

---

## Reporting a security issue

> [!IMPORTANT]
> **Do not open a public issue for a security problem.** See
> [SECURITY.md](SECURITY.md) for how to report one privately.

This includes anything touching authentication, sessions, permissions, public
share links, file access, or secret handling.

## Ground rules

- Be respectful and assume good faith. Technical criticism is welcome, personal
  criticism is not. Everyone taking part is covered by our
  [Code of Conduct](CODE_OF_CONDUCT.md).
- One logical change per issue and per pull request.
- Discuss before building. An issue that gets a "yes, go ahead" saves you from a
  rejected pull request.
- Read [the backend guidelines](docs/development/backend.md) before writing backend code. It is not a
  style suggestion; a change that ignores the layer rules, error handling or
  logging conventions will be sent back for rework.

## Reporting bugs

Open an issue with:

1. **Version:** the version the server reports, or the release tarball name
   (e.g. `twigex-0.15.0-linux-amd64`). For a source build, the commit hash.
2. **Deployment:** single binary, Docker Compose or Helm chart, and which optional services are
   configured (LiveKit, Cube.js, EuroOffice/Collabora).
3. **Environment:** MariaDB or MySQL and its version, Redis version, storage
   backend (local disk or S3).
4. **Steps to reproduce**, what you expected, and what happened instead.
5. **Server logs** around the failure, and the browser console for frontend bugs.

> [!WARNING]
> Scrub logs and screenshots before attaching them. No credentials, tokens,
> connection strings, real email addresses or customer data.

If you can reproduce it on the [demo instance](https://demo.twigex.com/),
say so. That removes the whole configuration surface from the investigation.

## Suggesting a change

Open an issue describing the problem you have, not only the solution you want.
Include who is affected and how you work around it today. For anything that adds
a setting, a table, an API endpoint or a permission, expect a design discussion
first; those are the parts that are expensive to change after release.

## Development setup

### Prerequisites

| Tool                 | Version                                 |
| -------------------- | --------------------------------------- |
| Go                   | 1.26 or newer                           |
| Node.js              | 22.12 or newer (CI builds on 24)        |
| MariaDB (or MySQL 8) | required                                |
| Redis                | required, used for sessions             |

The server will not start without a reachable database and Redis. LiveKit,
Cube.js and EuroOffice/Collabora are only contacted when their feature is used,
so leave them out until you need them.

### Build and run

```sh
make build      # frontend + binary + release archive in build/
make compile    # binary only
./twigex start --port 3000
```

Configuration comes from the environment. Copy the variables documented in the
[README quick start](README.md#option-a-single-binary) into an env file, and set
at least `TWIGEX_DB_*`, `TWIGEX_ADMIN_*`, `TWIGEX_STORAGE_TYPE`,
`TWIGEX_SERVER_SITE_URL` and `TWIGEX_SERVER_ENCRYPT_KEY`. On first boot the
server applies migrations, creates the storage backend and creates the
administrator; each step is skipped once it has nothing to do, so restarts are
safe.

### Frontend development

Run the server on port 3000 and start Vite alongside it. Vite proxies `/api` and
websocket traffic through, so open the URL Vite prints, not port 3000.

```sh
cd frontend && npm install && npm run dev
```

### Notes that save time

- **Use `make tidy`, not `go mod tidy`.** Plain `go mod tidy` fails on the
  build-tagged import in `enterprise/imports.go`; the target handles it.
- **`NOTICE` is generated.** Never edit it by hand.

## Repository layout

```
api/         HTTP handlers: parse, authenticate, call app, respond
app/         business logic, permissions, orchestration
store/       SQL queries only (store/sqlstore holds the implementations)
model/       types, AppError, constants
internal/    packages used only inside this module
db/mysql/    numbered .up.sql migrations, embedded into the binary
i18n/        backend translations (en, lv, kk, pl)
frontend/    Vue 3 + Vite app, with its own i18n under src/i18n
tools/       i18n checker and translation sync

cmd/         CLI commands: start, migrate, install, upgrade, user
config/      configuration loaded from the environment
interfaces/  seams the enterprise module implements (LDAP, OIDC, roles)
enterprise/  build-tag switch between the enterprise module and the stubs
collimato/   Collimato query types
crypto/      encryption and hashing helpers
safehttp/    SSRF-hardened HTTP client for outbound requests
tlog/        structured logger
templates/   email templates and blank office documents
deploy/      deployment extras, such as the shared Cube.js setup
docs/        developer guidelines
```

A package sits at the root when something outside this module imports it, and
under `internal/` when nothing does. There is no third category: put a new
package in `internal/` unless an importer forces otherwise.

The layer rule is strict: **each layer only talks to the layer directly below
it.** The API layer never touches `Store`, the app layer never imports HTTP
types, and the store layer contains no business logic. [the backend guidelines](docs/development/backend.md)
explains where a given branch of logic belongs, including the borderline cases.

## Coding conventions

Read [the backend guidelines](docs/development/backend.md) in full once. The rules most contributions
trip over:

| Rule                                                                                          | Reference                                              |
| --------------------------------------------------------------------------------------------- | ------------------------------------------------------ |
| Return `*model.AppError` from app functions, with an i18n ID from `i18n/en.json`, never a raw English string | [Error handling](docs/development/backend.md#error-handling)  |
| Never re-wrap an `AppError`; pass it through                                                  | [Error handling](docs/development/backend.md#error-handling)          |
| Respond with `respondJSON` / `respondAppError`; never `http.Error` or `json.NewEncoder`       | [API layer](docs/development/backend.md#api-layer)                    |
| Decode bodies with `decodeBody`                                                               | [API layer](docs/development/backend.md#api-layer)                    |
| Permission checks live in the app layer, always                                               | [App layer](docs/development/backend.md#app-layer)                    |
| Log with `tlog.Errorw` / `Warnw` / `Infow`, context fields first, `err` last, never `err.Error()`, never PII | [Logging](docs/development/backend.md#logging)         |
| Distinguish a DB failure (`500`) from a missing row (`404`)                                   | [HTTP status codes](docs/development/backend.md#http-status-codes)    |
| Pass `context.Context` from the request through app (`a.dbCtx`, once) into every store query   | [Store layer](docs/development/backend.md#store-layer)                |
| Build URLs with `url.Values`, never string concatenation                                      | [API layer](docs/development/backend.md#api-layer)                    |
| Record activity with `RecordActivity` and `model.*` constants                                 | [App layer](docs/development/backend.md#app-layer)                    |
| Check the timestamp unit: most columns are seconds, chat columns are milliseconds             | [Timestamp units](docs/development/backend.md#timestamp-units)        |

### Comments

Default to none. Add a comment only when the *why* is non-obvious: a hidden
constraint, an ordering requirement, a workaround for an external bug, a security
rationale. Do not restate the code, do not add section dividers, and do not
comment a line that had no comment before your change. A doc comment on an
exported symbol starts with the identifier name.

### Formatting

```sh
make lint-go                   # gofmt and golangci-lint
cd frontend && npm run format  # prettier; CI checks it with make lint-frontend
cd frontend && npm run lint    # eslint, also run by make lint-frontend
```

### Licence headers

Every new `.go`, `.js`, `.css`, `.vue` and `.html` file starts with these two
lines, in that language's comment syntax. Migrations are left out on purpose.

```go
// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only
```

`make check-headers` fails without them. It only looks for the SPDX line, so
copy both lines as they are.

## Tests

[The testing guide](docs/development/testing.md) has the full detail and
copy-paste boilerplate for both kinds of test. In short: **test SQL against a
real database, test decisions with a fake store.**

| You are testing                                  | Write                                | Needs a DB |
| ------------------------------------------------ | ------------------------------------ | ---------- |
| A query or store function                        | Store test in `store/sqlstore`       | yes        |
| Business logic, permissions, validation, statuses| App test with a fake store           | no         |
| A pure helper                                    | Plain Go test                        | no         |

Store tests need a database. The migrations run on both MariaDB and MySQL 8, so
either will do:

```sh
docker run --rm -d -p 3306:3306 -e MARIADB_ROOT_PASSWORD=root mariadb:11
export TEST_MYSQL_DSN='root:root@tcp(127.0.0.1:3306)/?multiStatements=true'
```

A unique throwaway schema is created and dropped per run.

Before you call a change done, run what CI runs:

```sh
make lint            # gofmt, golangci-lint and prettier
make test-all        # everything, under -race; needs TEST_MYSQL_DSN
make test-frontend   # Vitest
make check-headers   # licence headers on new files
make check-i18n      # the four locales carry the same IDs
make check-tidy      # go.mod and go.sum are tidy
make check-vuln      # no reachable known vulnerability
```

`make test-all` refuses to start without `TEST_MYSQL_DSN`. Without one,
`requireDB(t)` would skip every store test and the run would report green having
tested no SQL at all, which is the failure this guards against. For a quick loop
while working, `go test ./app/...` needs no database.

Fix a bug, add a test that would have caught it. New app-layer logic and new SQL
are expected to arrive with tests.

## Translations

Backend strings live in `i18n/` as an array of `{id, translation}`; frontend
strings live in `frontend/src/i18n/` as flat key/value JSON. Locales are
**en, lv, kk, pl**.

1. Add the new string to `en.json` (backend `i18n/` or `frontend/src/i18n/`).
2. Run `make sync-i18n` to add the key to the other locales.
3. Translate all locales for anything a user reads: chat, notifications,
   emails, error messages. English alone is fine for internal or admin-only
   diagnostics.
4. Run `make check-i18n` and make sure it reports **0 missing**. Every
   `model.NewAppError("x.y", ...)` ID must exist in `i18n/en.json`.

Error IDs are named `<namespace>.<action_or_reason>`, e.g. `file.not_found`,
`user.create_failed`, `upload.invalid_size`.

## Database migrations

- One file per change: `db/mysql/<next-number>_<snake_case_name>.up.sql`. Take
  the next unused number (check `db/mysql/` for the highest). The directory
  is embedded with `go:embed`, so a new file ships with the binary and needs no
  packaging change.
- **Write the `.up.sql` only.** Down migrations are not applied.
- Use syntax both MariaDB and MySQL 8 accept. Watch for the two that differ:
  `ADD COLUMN IF NOT EXISTS` is MariaDB-only, and MySQL rejects a non-null
  `DEFAULT` on a `TEXT` or `JSON` column. `CREATE TABLE IF NOT EXISTS` is fine.
- Don't guard a migration against running twice. Migrations run on every start,
  but an applied one is recorded by version and never opened again.
- Never edit a migration that has shipped. Add a new one. The only exception is
  an edit that cannot change the schema an existing install already has, since
  those installs never read the file again.
- Migrations are a high-risk area: expect a second reviewer (see
  [Review expectations](#review-expectations)).

## Dependencies

Adding a dependency needs a reason in the pull request description. Before you
add one, check that it is compatible with AGPL-3.0-only, that it is maintained,
and that the standard library or an existing dependency cannot do the job. No
third-party source is copied into this repository; dependencies are linked into
the binary and bundled into the frontend assets, and `NOTICE` reproduces their
licences. Run `make tidy` after changing `go.mod`.

## Commits

Write the subject in the imperative mood, capitalised, no trailing period, kept
under ~72 characters, and describing the change rather than the file touched:

```
Show unread file notifications in the sidebar
Keep the kanban card order after a status change
```

Use the body to explain *why*, wrapped at 72 columns, and reference the issue
(`Closes #310`). Keep each commit buildable; squash the "fix typo" and
"address review" commits before the pull request is merged.

## Opening a pull request

The repository is on GitHub at
[github.com/twigex/twigex](https://github.com/twigex/twigex), so changes land as
pull requests. Contributors without write access work from a fork.

1. Branch from `main`. If a maintainer points you at another branch for the
   area you are touching, use that instead.
2. Name the branch after the work in kebab-case, or after the issue it closes:
   `ldap-sync-users`, `310-files-show-unread-file-notifications-badge`.
3. Keep the change focused. Unrelated refactors, reformatting and dependency
   bumps belong in their own pull request.
4. Make sure CI is green. On every pull request it runs the linters, the
   checks, `make test-all` against MariaDB, the frontend tests, and then the
   build. A maintainer may need to approve the run for a first-time
   contributor.
5. Fill in the description:
   - what changed and why, and the issue it closes;
   - how you tested it, including whether store tests were run against a real
     MariaDB;
   - migrations, new settings or new environment variables introduced;
   - breaking changes for existing deployments;
   - screenshots or a short clip for UI changes;
   - an AI-assistance note when it applies (see below).

Self-review checklist:

- [ ] Layer boundaries respected: no `Store` access from `api/`, no HTTP types in `app/`, no logic in `store/`
- [ ] All new errors are `*model.AppError` with an i18n ID, and status codes match the table in the [backend guidelines](docs/development/backend.md#http-status-codes)
- [ ] `context.Context` threaded through every new store call
- [ ] `make lint` clean
- [ ] Tests added, `go test ./app/...` and (if SQL changed) `make test-all` pass
- [ ] `make check-i18n` reports 0 missing, and user-facing strings are translated in all four locales
- [ ] No secrets, tokens or real user data in code, tests, fixtures or logs
- [ ] `NOTICE` untouched by hand; `make tidy` run if `go.mod` changed
- [ ] Substantial AI assistance noted in the description; no AI co-author trailers

## Review expectations

Every change is reviewed, including maintainers' own. Reviewers look at layer
placement, error and log handling, permission checks, SQL correctness, and test
coverage, roughly in that order. Anything touching authentication, sessions,
permissions, cryptography or migrations gets a second pair of eyes and may sit
longer; that is deliberate.

Respond to comments rather than force-pushing silently, and say when you disagree.
A rejected approach is not a rejected contributor.

## AI-assisted contributions

AI tools are permitted. Every line of AI-generated code that enters this
codebase is the full responsibility of the developer who commits it: the AI is a
tool, not a team member, and the rules are the same whichever assistant
produced the code.

### Understand before you commit

You must be able to explain every line you commit: what it does, why it is
correct, and how it fits the surrounding system. If you cannot explain it, do not
submit it. In practice:

- Read the output in full before adding it to the codebase.
- Trace the logic; do not assume the AI applied the right pattern.
- Check it against [the backend](docs/development/backend.md) and
  [frontend](docs/development/frontend.md) guidelines.
- If any part is unclear, have it explained, simplified or rewritten until you
  understand it.
- Accept the change only when you would be comfortable defending it in a review
  as your own.

### Review and disclosure

AI-generated code goes through the same pull request review as any other code,
with the same or closer scrutiny: it can be confidently wrong. The author
understands the full diff before opening the pull request, makes sure it passes
CI, and answers review comments as if they had written it.

Disclose substantial AI assistance in the pull request description, e.g.
"Initial implementation scaffolded with Claude Code; reviewed and adapted to our
error-handling and logging conventions." Autocompleted names and one-liners need
no note.

AI tools are not listed as authors. Many of them add a `Co-authored-by:`
trailer or a "Generated with …" line by default, so please remove those before
committing. If you would like the history to show that a tool helped, you can
add an `Assisted-by: <tool>` trailer instead. The note in the pull request
description is still needed.

Pull requests, issues and comments come from a person, who opens them and
follows the discussion. Ones posted by an autonomous agent with no person behind
them may be closed.

### Areas where AI-generated code needs prior discussion

- Authentication and session handling
- Permission and access-control logic
- Cryptography and secret handling
- Database migrations

These areas carry outsized risk: AI tools can produce subtly incorrect security
code that passes tests and review but fails in production. When in doubt, write
it by hand and have a second developer review it.

### Secrets and personal data

Never paste credentials, tokens, connection strings or personal data into a
prompt. Use placeholders.

### Common mistakes

AI tools tend to repeat the backend mistakes listed in
[the backend guidelines](docs/development/backend.md#common-mistakes), so check
generated code against that list.

### AI output is not a source of truth

AI tools can state incorrect facts with high confidence. Verify explanations of
this codebase, library behaviour or language semantics against the code in the
tree, the upstream docs or a teammate, especially for SQL, concurrency and
anything that talks to an external service.

## Licensing of contributions

By submitting a contribution you agree that it is licensed under
**AGPL-3.0-only**, the same licence as the project, and that you have the right
to submit it.

## Getting help

- Manual: [twigex.com/en/Installation](https://twigex.com/en/Installation)
- Conventions: [the backend guidelines](docs/development/backend.md) and [the testing guide](docs/development/testing.md)
- Bugs and ideas: open an issue
- Security: **security@twigex.com** (never a public issue)
- Enterprise Edition and support: [twigex.com](https://twigex.com/)
