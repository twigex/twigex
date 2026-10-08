// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { describe, it, expect } from "vitest";
import { mount } from "@vue/test-utils";
import ConnectionSecurity from "@/components/Collimato/ConnectionSecurity.vue";
import {
    SSL_MODE_DISABLE,
    SSL_MODE_REQUIRE,
    SSL_MODE_VERIFY_CA,
    SSL_MODE_VERIFY_FULL,
} from "@/utils/collimato/connectionSecurity";

function form(sslMode, caCert = "") {
    return mount(ConnectionSecurity, { props: { sslMode, caCert } });
}

const textarea = (wrapper) => wrapper.find("textarea");

describe("what is on screen", () => {
    it("shows only the checkbox when encryption is off", () => {
        const wrapper = form(SSL_MODE_DISABLE);

        expect(wrapper.findComponent({ name: "BaseSelect" }).exists()).toBe(false);
        expect(textarea(wrapper).exists()).toBe(false);
    });

    it("offers a verification level once encryption is on", () => {
        const wrapper = form(SSL_MODE_REQUIRE);

        expect(wrapper.findComponent({ name: "BaseSelect" }).exists()).toBe(true);
    });

    it("hides the certificate field when nothing is verified", () => {
        expect(textarea(form(SSL_MODE_REQUIRE)).exists()).toBe(false);
    });

    it("asks for a certificate for both verifying modes", () => {
        for (const mode of [SSL_MODE_VERIFY_CA, SSL_MODE_VERIFY_FULL]) {
            expect(textarea(form(mode)).exists()).toBe(true);
        }
    });

    it("shows a certificate the connection already has", () => {
        const wrapper = form(SSL_MODE_VERIFY_FULL, "a certificate");

        expect(textarea(wrapper).element.value).toBe("a certificate");
    });
});

describe("what it emits", () => {
    it("turning encryption on asks for the weakest mode that encrypts", async () => {
        const wrapper = form(SSL_MODE_DISABLE);

        await wrapper.find("input[type=checkbox]").setValue(true);

        expect(wrapper.emitted("update:sslMode").at(-1)).toEqual([SSL_MODE_REQUIRE]);
    });

    it("turning encryption off asks for disable, whatever was verified", async () => {
        const wrapper = form(SSL_MODE_VERIFY_FULL);

        await wrapper.find("input[type=checkbox]").setValue(false);

        expect(wrapper.emitted("update:sslMode").at(-1)).toEqual([SSL_MODE_DISABLE]);
    });

    it("choosing a level passes that level up", async () => {
        const wrapper = form(SSL_MODE_REQUIRE);

        wrapper
            .findComponent({ name: "BaseSelect" })
            .vm.$emit("update:modelValue", SSL_MODE_VERIFY_CA);
        await wrapper.vm.$nextTick();

        expect(wrapper.emitted("update:sslMode").at(-1)).toEqual([SSL_MODE_VERIFY_CA]);
    });

    it("typing a certificate passes it up", async () => {
        const wrapper = form(SSL_MODE_VERIFY_CA);

        await textarea(wrapper).setValue("pasted pem");

        expect(wrapper.emitted("update:caCert").at(-1)).toEqual(["pasted pem"]);
    });

    it("keeps the typed certificate when the level changes between verifying modes", async () => {
        const wrapper = form(SSL_MODE_VERIFY_CA, "kept");

        await wrapper.setProps({ sslMode: SSL_MODE_VERIFY_FULL });

        expect(textarea(wrapper).element.value).toBe("kept");
    });
});
