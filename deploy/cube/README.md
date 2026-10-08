# Collimato Cube.js deployment

A single, shared, multi-tenant Cube.js backend for Collimato analytics. It
replaces the old model where Twigex provisioned a Cube.js + Cube Store
StatefulSet per workspace via the Kubernetes API.

## How it works

```
Browser → Twigex (session auth + workspace authz + field filtering)
        → signs a short-lived JWT { workspaceId } with the shared secret
        → proxies /cubejs-api/v1/{load,sql,meta} to Cube.js
Cube.js → calls back to Twigex /api/internal/cube/{schema,schema-version,connection}
          (authenticated with CUBE_CALLBACK_SECRET) to resolve, per workspace:
            • the data model files          (repositoryFactory)
            • the model version, for reloads (schemaVersion)
            • the database connection        (driverFactory)
        → Cube Store router + worker for pre-aggregations
```

A new workspace, a new data-model file, or a new database connection is just a
row in Twigex's database — there is nothing to provision or restart.

## Running

1. `cp .env.example .env` and set the three values.
2. `docker compose up -d`.
3. Point Twigex at this instance (see below).

Only Twigex should be able to reach `cube_api:4000`, and only Cube.js should be
able to reach Twigex's `/api/internal/cube` endpoints (they return database
credentials). Keep both on a private network.

## Matching Twigex configuration

Set these on the Twigex side (env vars):

| Twigex env var                          | must equal                                  |
|------------------------------------------|---------------------------------------------|
| `TWIGEX_COLLIMATO_CUBE_API_URL`         | URL of `cube_api`, e.g. `http://cube:4000`  |
| `TWIGEX_COLLIMATO_CUBE_API_SECRET`      | `CUBEJS_API_SECRET` here                     |
| `TWIGEX_COLLIMATO_CUBE_CALLBACK_SECRET` | `CUBE_CALLBACK_SECRET` here                  |

`TWIGEX_COLLIMATO_CUBE_API_URL` being set is also what makes Twigex report
Collimato as available (`GET /api/collimato/status`).

## Pre-aggregations

Cube Store (router + worker) is included because some workspaces use
pre-aggregations. Isolation between tenants is handled by `contextToOrchestratorId`
(keyed on `workspaceId`) in `cube.js`, so a single shared Cube Store serves all
workspaces. If no workspace uses pre-aggregations you can drop the two
`cubestore_*` services and the `CUBEJS_CUBESTORE_HOST` env var.
