// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { describe, it, expect } from "vitest";
import {
    SSL_MODE_DISABLE,
    SSL_MODE_REQUIRE,
    SSL_MODE_VERIFY_CA,
    SSL_MODE_VERIFY_FULL,
    CA_CERT_KEY,
    securityPayload,
    securityOf,
} from "@/utils/collimato/connectionSecurity";

const cert = "-----BEGIN CERTIFICATE-----\nabc\n-----END CERTIFICATE-----";

describe("securityPayload", () => {
    it("sends no config when encryption is off", () => {
        expect(securityPayload(SSL_MODE_DISABLE, "")).toEqual({
            ssl_mode: SSL_MODE_DISABLE,
            config: null,
        });
    });

    it("drops a certificate the chosen mode would never check", () => {
        const payload = securityPayload(SSL_MODE_REQUIRE, cert);

        expect(payload.config).toBeNull();
    });

    it("keeps the certificate for the verifying modes", () => {
        for (const mode of [SSL_MODE_VERIFY_CA, SSL_MODE_VERIFY_FULL]) {
            expect(securityPayload(mode, cert).config).toEqual({
                [CA_CERT_KEY]: cert,
            });
        }
    });

    it("treats a blank certificate as none given", () => {
        expect(securityPayload(SSL_MODE_VERIFY_FULL, "   \n ").config).toBeNull();
    });

    it("trims what the user pasted", () => {
        const payload = securityPayload(SSL_MODE_VERIFY_CA, `\n ${cert} \n`);

        expect(payload.config[CA_CERT_KEY]).toBe(cert);
    });
});

describe("securityOf", () => {
    it("reads a stored connection back", () => {
        expect(
            securityOf({
                ssl_mode: SSL_MODE_VERIFY_FULL,
                config: { [CA_CERT_KEY]: cert },
            }),
        ).toEqual({ sslMode: SSL_MODE_VERIFY_FULL, caCert: cert });
    });

    it("falls back to off for a connection saved before encryption existed", () => {
        expect(securityOf({}).sslMode).toBe(SSL_MODE_DISABLE);
    });

    it("survives a null config", () => {
        expect(securityOf({ ssl_mode: SSL_MODE_REQUIRE, config: null })).toEqual({
            sslMode: SSL_MODE_REQUIRE,
            caCert: "",
        });
    });

    it("round trips through the payload", () => {
        const payload = securityPayload(SSL_MODE_VERIFY_CA, cert);

        expect(securityOf(payload)).toEqual({
            sslMode: SSL_MODE_VERIFY_CA,
            caCert: cert,
        });
    });
});
