// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

// Mirrors internal/catalog/connection_test.go. Both files decide what the same
// four ssl_mode values mean, for two different engines, against the same source
// database. If they disagree the "test connection" button verifies something the
// analytics queries do not.
//
// Run: node --test deploy/cube/cube.test.js
//
// sslOption is module-local because Cube validates its config against a strict
// schema and rejects unknown keys, so cube.js cannot export it. The file is
// evaluated here with one export appended, which keeps a single cube.js and
// avoids matching on its source text.

const test = require('node:test');
const assert = require('node:assert');
const fs = require('node:fs');
const path = require('node:path');

function loadSslOption() {
  const file = path.join(__dirname, 'cube.js');
  const src = `${fs.readFileSync(file, 'utf8')}\nmodule.exports.__sslOption = sslOption;`;
  const mod = { exports: {} };

  new Function('module', 'exports', 'require', 'process', '__dirname', src)(
    mod,
    mod.exports,
    require,
    process,
    __dirname,
  );

  return mod.exports.__sslOption;
}

const sslOption = loadSslOption();
const CA = '-----BEGIN CERTIFICATE-----\nnot-a-real-one\n-----END CERTIFICATE-----';

test('disable negotiates nothing', () => {
  for (const type of ['mysql', 'postgres']) {
    assert.equal(sslOption('disable', CA, type), false);
    assert.equal(sslOption(undefined, CA, type), false);
  }
});

test('require encrypts without checking', () => {
  for (const type of ['mysql', 'postgres']) {
    const o = sslOption('require', CA, type);
    assert.equal(o.rejectUnauthorized, false);
    assert.equal(o.ca, undefined, 'require must not imply a trust anchor');
  }
});

test('verify-ca checks the chain but not the name', () => {
  for (const type of ['mysql', 'postgres']) {
    const o = sslOption('verify-ca', CA, type);
    assert.equal(o.rejectUnauthorized, true);
    assert.equal(o.ca, CA);
    assert.notEqual(o.verifyIdentity, true, `${type}: verify-ca must not check the name`);
  }
});

// pg honours checkServerIdentity; mysql2 overwrites it and reads verifyIdentity
// instead, which defaults to no check. Confirmed against mysql2 3.11.5.
test('verify-ca suppresses the name check in the way each driver reads', () => {
  assert.equal(typeof sslOption('verify-ca', CA, 'postgres').checkServerIdentity, 'function');
  assert.equal(
    sslOption('verify-ca', CA, 'postgres').checkServerIdentity(),
    undefined,
    'postgres: the override must accept any name',
  );
  assert.notEqual(sslOption('verify-ca', CA, 'mysql').verifyIdentity, true);
});

test('verify-full checks the name', () => {
  const pg = sslOption('verify-full', CA, 'postgres');
  assert.equal(pg.rejectUnauthorized, true);
  assert.equal(pg.ca, CA);
  assert.equal(
    pg.checkServerIdentity,
    undefined,
    'postgres: leaving checkServerIdentity unset is what makes node check the name',
  );

  const my = sslOption('verify-full', CA, 'mysql');
  assert.equal(my.rejectUnauthorized, true);
  assert.equal(my.ca, CA);
  assert.equal(
    my.verifyIdentity,
    true,
    'mysql2 ignores checkServerIdentity, so without this verify-full is really verify-ca',
  );
});

// An unmapped driver inheriting the postgres branch is how the mysql gap would
// have been reintroduced for the next engine added.
test('verify-full refuses a driver it has no mapping for', () => {
  assert.throws(
    () => sslOption('verify-full', CA, 'clickhouse'),
    /no verified TLS mapping for driver "clickhouse"/,
  );
});

// The modes that cannot silently weaken stay permissive, so adding an engine does
// not break connections that were never driver specific.
test('the other modes accept any driver', () => {
  assert.equal(sslOption('disable', CA, 'clickhouse'), false);
  assert.equal(sslOption('require', CA, 'clickhouse').rejectUnauthorized, false);

  const ca = sslOption('verify-ca', CA, 'clickhouse');
  assert.equal(ca.rejectUnauthorized, true);
  assert.equal(ca.ca, CA);
});

test('a verifying mode without a certificate still verifies', () => {
  for (const mode of ['verify-ca', 'verify-full']) {
    for (const type of ['mysql', 'postgres']) {
      const o = sslOption(mode, '', type);
      assert.equal(o.rejectUnauthorized, true, `${type} ${mode}`);
      assert.equal(o.ca, undefined, 'an empty certificate must not become an empty trust anchor');
    }
  }
});
