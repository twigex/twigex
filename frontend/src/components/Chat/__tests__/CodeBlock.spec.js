// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { describe, it, expect, vi } from "vitest";
import { mount } from "@vue/test-utils";
import CodeBlock from "@/components/Chat/Messages/CodeBlock.vue";

// The copy-button label reads from i18n; stub it so tests do not depend on a
// loaded locale bundle.
vi.mock("@/i18n/index.js", () => ({ t: { value: (key) => key } }));

describe("CodeBlock", () => {
    it("highlights a known language and labels it", () => {
        const w = mount(CodeBlock, {
            props: { code: "const x = 1;", lang: "js" },
        });

        expect(w.find(".code-block").exists()).toBe(true);
        expect(w.find(".code-lang").text()).toBe("js");
        expect(w.find("code.hljs").exists()).toBe(true);
        expect(w.find("button.copy-btn").exists()).toBe(true);
    });

    it("renders plain code with no language label", () => {
        const w = mount(CodeBlock, { props: { code: "hello", lang: null } });

        expect(w.find("code.plain").exists()).toBe(true);
        expect(w.find("code.plain").text()).toBe("hello");
        expect(w.find(".code-lang").exists()).toBe(false);
    });

    it("escapes html in plain code", () => {
        const w = mount(CodeBlock, {
            props: { code: "<img src=x onerror=y>", lang: null },
        });

        expect(w.find("img").exists()).toBe(false);
        expect(w.find("code.plain").text()).toContain("<img");
    });
});
