// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

export const SSL_MODE_DISABLE = "disable";
export const SSL_MODE_REQUIRE = "require";
export const SSL_MODE_VERIFY_CA = "verify-ca";
export const SSL_MODE_VERIFY_FULL = "verify-full";

export const CA_CERT_KEY = "ca_cert";

export function securityPayload(sslMode, caCert) {
    const carriesCert = sslMode === SSL_MODE_VERIFY_CA || sslMode === SSL_MODE_VERIFY_FULL;
    const cert = caCert?.trim() ?? "";

    return {
        ssl_mode: sslMode,
        config: carriesCert && cert ? { [CA_CERT_KEY]: cert } : null,
    };
}

export function securityOf(connection) {
    return {
        sslMode: connection?.ssl_mode ?? SSL_MODE_DISABLE,
        caCert: connection?.config?.[CA_CERT_KEY] ?? "",
    };
}
