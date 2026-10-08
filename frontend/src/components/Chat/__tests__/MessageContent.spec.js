// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { describe, it, expect } from "vitest";
import { shallowMount } from "@vue/test-utils";
import { createPinia } from "pinia";
import MessageContent from "@/components/Chat/Messages/MessageContent.vue";
import MentionTag from "@/components/Mentions/MentionTag.vue";
import CodeBlock from "@/components/Chat/Messages/CodeBlock.vue";

// shallowMount stubs <MentionTag>/<CodeBlock>, so the walker's own output (escaped
// text, links, marks, and where components are emitted) is what gets asserted,
// without pulling in the user store, floating-ui, or the network.
function render(text) {
    return shallowMount(MessageContent, {
        props: { text },
        global: { plugins: [createPinia()] },
    });
}

describe("MessageContent XSS", () => {
    it("renders a raw <script> as inert text", () => {
        const w = render("<script>alert(1)</script>");

        expect(w.find("script").exists()).toBe(false);
        expect(w.text()).toContain("<script>alert(1)</script>");
    });

    it("renders an <img onerror> as inert text", () => {
        const w = render("<img src=x onerror=alert(1)>");

        expect(w.find("img").exists()).toBe(false);
    });

    it("keeps highlight content inert (==<img onerror>==)", () => {
        const w = render("==<img src=x onerror=alert(1)>==");

        expect(w.find("mark").exists()).toBe(true);
        expect(w.find("img").exists()).toBe(false);
    });

    it("keeps highlight content inert (==</mark><script>==)", () => {
        const w = render("==</mark><script>==");

        expect(w.find("script").exists()).toBe(false);
    });

    it("drops a javascript: link, leaving literal text", () => {
        const w = render("[x](javascript:alert(1))");

        expect(w.find("a").exists()).toBe(false);
        expect(w.text()).toContain("[x](javascript:alert(1))");
    });
});

describe("MessageContent links", () => {
    it("renders a safe https link with a blank target", () => {
        const w = render("[site](https://example.com)");
        const a = w.find("a");

        expect(a.exists()).toBe(true);
        expect(a.attributes("href")).toBe("https://example.com");
        expect(a.attributes("target")).toBe("_blank");
        expect(a.attributes("rel")).toContain("noopener");
    });
});

describe("MessageContent highlight", () => {
    it("wraps ==text== in a <mark>", () => {
        const w = render("==important==");

        expect(w.find("mark").exists()).toBe(true);
        expect(w.find("mark").text()).toBe("important");
    });
});

describe("MessageContent mentions", () => {
    it("emits a <MentionTag> per @handle", () => {
        const w = render("hey @bob and @all");
        const mentions = w.findAllComponents(MentionTag);

        expect(mentions).toHaveLength(2);
        expect(mentions[0].props("username")).toBe("bob");
        expect(mentions[1].props("username")).toBe("all");
    });

    it("leaves @handles inside inline code literal", () => {
        const w = render("look at `@bob`");

        expect(w.findAllComponents(MentionTag)).toHaveLength(0);
    });

    it("leaves @handles inside a code block literal", () => {
        const w = render("```\n@bob\n```");

        expect(w.findAllComponents(MentionTag)).toHaveLength(0);
        expect(w.findComponent(CodeBlock).exists()).toBe(true);
    });

    it("leaves @handles inside link text literal", () => {
        const w = render("[@bob](https://example.com)");

        expect(w.findAllComponents(MentionTag)).toHaveLength(0);
    });

    it("matches directory style usernames containing dots and hyphens", () => {
        const w = render("ping @jane.doe and @jane-doe and @jane_doe");
        const mentions = w.findAllComponents(MentionTag);

        expect(mentions.map((m) => m.props("username"))).toEqual([
            "jane.doe",
            "jane-doe",
            "jane_doe",
        ]);
    });

    it("emits a <MentionTag> carrying the id for the stored form", () => {
        const id = "0f8fad5bd9cb469fa16570867728950e";
        const w = render(`hey <@${id}> and @bob`);
        const mentions = w.findAllComponents(MentionTag);

        expect(mentions).toHaveLength(2);
        expect(mentions[0].props("userId")).toBe(id);
        expect(mentions[1].props("username")).toBe("bob");
    });

    it("leaves an id form mention inside code literal", () => {
        const id = "0f8fad5bd9cb469fa16570867728950e";
        const w = render(`look at \`<@${id}>\` and [<@${id}>](https://e.com)`);

        expect(w.findAllComponents(MentionTag)).toHaveLength(0);
    });

    it("leaves trailing sentence punctuation out of the handle", () => {
        const w = render("ask @jane.doe. then @bob- or @carol_");
        const mentions = w.findAllComponents(MentionTag);

        expect(mentions.map((m) => m.props("username"))).toEqual(["jane.doe", "bob", "carol"]);
    });
});

describe("MessageContent code blocks", () => {
    it("emits a <CodeBlock> with language and raw code", () => {
        const w = render("```js\nconst x = 1;\n```");
        const cb = w.findComponent(CodeBlock);

        expect(cb.exists()).toBe(true);
        expect(cb.props("lang")).toBe("js");
        expect(cb.props("code")).toContain("const x = 1;");
    });

    it("emits a <CodeBlock> with no language for a plain fence", () => {
        const w = render("```\nplain text\n```");
        const cb = w.findComponent(CodeBlock);

        expect(cb.exists()).toBe(true);
        expect(cb.props("lang")).toBeNull();
    });
});
