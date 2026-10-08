// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

// Cube.js configuration for the shared, multi-tenant Collimato analytics backend.
//
// A single Cube.js deployment serves every Collimato workspace. It holds no
// per-workspace files or database credentials of its own. Instead it resolves
// everything at query time by calling back into Twigex's internal Cube API,
// keyed on the workspaceId carried in the JWT that Twigex signs and sends.
//
// Required environment variables:
//   CUBEJS_API_SECRET     shared secret that verifies the JWT from Twigex.
//                         Must equal Twigex TWIGEX_COLLIMATO_CUBE_API_SECRET.
//   TWIGEX_INTERNAL_URL  base URL of Twigex's internal Cube API, e.g.
//                         http://twigex:3000/api/internal/cube (private network).
//   CUBE_CALLBACK_SECRET  shared secret sent back to Twigex on every callback.
//                         Must equal Twigex TWIGEX_COLLIMATO_CUBE_CALLBACK_SECRET.
//
// Optional:
//   CUBE_DB_CONNECT_TIMEOUT_MS  how long to wait for a source database to accept
//                               a connection. Default 15000.

const TWIGEX_URL = process.env.TWIGEX_INTERNAL_URL;
const CALLBACK_SECRET = process.env.CUBE_CALLBACK_SECRET;
const CONNECT_TIMEOUT_MS = Number(process.env.CUBE_DB_CONNECT_TIMEOUT_MS) || 15000;

// callTwigex performs an authenticated GET against a Twigex internal endpoint.
async function callTwigex(pathname, params) {
  const url = new URL(TWIGEX_URL + pathname);
  for (const [key, value] of Object.entries(params)) {
    if (value !== undefined && value !== null) {
      url.searchParams.set(key, value);
    }
  }

  const res = await fetch(url, {
    headers: { 'X-Cube-Callback-Secret': CALLBACK_SECRET },
  });
  if (!res.ok) {
    throw new Error(`Twigex ${pathname} responded ${res.status}`);
  }
  return res.json();
}

// Node tls options. pg hands these to tls.connect unchanged, so the identity
// check is replaced with a no-op for verify-ca. mysql2 does not: it reads only
// rejectUnauthorized and verifyIdentity off this object and overwrites
// checkServerIdentity itself, defaulting to no hostname check. verify-full has to
// ask for that check explicitly there, or it silently degrades to verify-ca.
const sslOption = (mode, ca, type) => {
  const trust = ca ? { ca } : {};

  switch (mode) {
    case 'require':
      return { rejectUnauthorized: false };
    case 'verify-ca':
      return { ...trust, rejectUnauthorized: true, checkServerIdentity: () => undefined };
    case 'verify-full':
      // The other modes are driver independent: an unmapped driver can only end
      // up stricter than asked, which fails visibly. This one can end up quieter
      // than asked, so a driver we have not checked is refused rather than
      // guessed at. Adding one means reading its source to find whether it hands
      // this object to tls.connect or reads its own keys off it.
      if (type === 'mysql') {
        return { ...trust, rejectUnauthorized: true, verifyIdentity: true };
      }
      if (type === 'postgres') {
        return { ...trust, rejectUnauthorized: true };
      }
      throw new Error(`sslOption: no verified TLS mapping for driver "${type}"`);
    default:
      return false;
  }
};

module.exports = {
  // Isolate each workspace's compiled data model and its orchestrator (query
  // cache + pre-aggregations), so tenants never share compiled schema or state.
  contextToAppId: ({ securityContext }) => `ws_${securityContext.workspaceId}`,
  // connVersion busts the orchestrator (connection pool) when a connection is
  // edited, so driverFactory re-runs with the new credentials, no restart.
  contextToOrchestratorId: ({ securityContext }) =>
    `ws_${securityContext.workspaceId}_c${securityContext.connVersion}`,

  // Serve the workspace's cube/view definitions from Twigex (backed by the DB).
  repositoryFactory: ({ securityContext }) => ({
    dataSchemaFiles: async () => {
      const files = await callTwigex('/schema', {
        workspace: securityContext.workspaceId,
      });
      return files.map((f) => ({ fileName: f.fileName, content: f.content }));
    },
  }),

  // Poll for changes: when a workspace edits a file in the Twigex GUI this value
  // changes and Cube hot-reloads the model, no restart, no redeploy.
  schemaVersion: async ({ securityContext }) => {
    const { version } = await callTwigex('/schema-version', {
      workspace: securityContext.workspaceId,
    });
    return String(version);
  },

  // Resolve the database driver for a workspace + dataSource at query time. An
  // empty/"default" dataSource selects the workspace's default connection.
  // Returning a DriverConfig ({ type, ... }) replaces the removed `dbType`
  // option (Cube v1.7.0+).
  driverFactory: async ({ securityContext, dataSource }) => {
    const conn = await callTwigex('/connection', {
      workspace: securityContext.workspaceId,
      dataSource,
    });
    return {
      type: conn.type,
      host: conn.host,
      port: Number(conn.port),
      database: conn.database,
      user: conn.user,
      password: conn.password,
      ssl: sslOption(conn.ssl_mode, conn.ssl_ca, conn.type),
      // Bounded so an unreachable DB surfaces as "connect ETIMEDOUT" instead of
      // hanging into a 504. Generous by default: a server doing a reverse DNS
      // lookup on the client IP can take seconds to answer a working connection,
      // and a timeout here is indistinguishable from unreachable.
      connectTimeout: CONNECT_TIMEOUT_MS,
    };
  },
};
