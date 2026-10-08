<!-- Base your branch on `main`. Keep the change focused: unrelated
refactors, reformatting and dependency bumps belong in their own pull
request. See CONTRIBUTING.md for the full guidance. -->

## What changed and why

<!-- Describe the change and the problem it solves. Reference the issue it
closes with `Closes #N`. -->

## How it was tested

<!-- Include whether store tests were run against a real MariaDB/MySQL 8. -->

## Migrations, settings, environment variables

<!-- List any new migrations, settings or environment variables. Remove this
section if there are none. -->

## Breaking changes

<!-- Anything an existing deployment must do after upgrading. Remove this
section if there are none. -->

## Screenshots

<!-- For UI changes: a screenshot or a short clip. Remove this section
otherwise. -->

## AI assistance

<!-- Disclose substantial AI assistance, e.g. "Initial implementation
scaffolded with Claude Code; reviewed and adapted to the error-handling and
logging conventions." Autocompleted names and one-liners need no note. -->

---

Self-review checklist:

- [ ] Layer boundaries respected: no `Store` access from `api/`, no HTTP types in `app/`, no logic in `store/`
- [ ] All new errors are `*model.AppError` with an i18n ID, and status codes match the table in the [backend guidelines](https://github.com/twigex/twigex/blob/main/docs/development/backend.md#http-status-codes)
- [ ] `context.Context` threaded through every new store call
- [ ] `make lint` clean
- [ ] Tests added, `go test ./app/...` and (if SQL changed) `make test-all` pass
- [ ] `make check-i18n` reports 0 missing, and user-facing strings are translated in all four locales
- [ ] No secrets, tokens or real user data in code, tests, fixtures or logs
- [ ] `NOTICE` untouched by hand; `make tidy` run if `go.mod` changed
- [ ] Substantial AI assistance noted in the description; no AI co-author trailers
