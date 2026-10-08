<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<script>
import { computed, h } from "vue";
import MentionTag from "@/components/Mentions/MentionTag.vue";
import CodeBlock from "@/components/Chat/Messages/CodeBlock.vue";
import useChatOperations from "@/composables/chat/useChatOperations";
import { splitMentionTokens } from "@/utils/mentions";

const VOID_TAGS = new Set([
    "area",
    "base",
    "br",
    "col",
    "embed",
    "hr",
    "img",
    "input",
    "link",
    "meta",
    "param",
    "source",
    "track",
    "wbr",
]);

const SAFE_URL_PATTERN = /^(https?:\/\/|mailto:)/i;

// renderMarkdown already sanitized this HTML with DOMPurify, so copying the
// attributes onto VNodes is safe. The href/src and on* checks below are a
// second layer of defense: even if that sanitizer config later changes, a bad
// URL or inline event handler still cannot get through here.
function propsFromElement(el) {
    const props = {};

    for (const attr of el.attributes) {
        const name = attr.name;

        if (name.startsWith("on")) continue;
        const value = attr.value;

        if ((name === "href" || name === "src") && !SAFE_URL_PATTERN.test(value.trim())) {
            continue;
        }

        props[name] = value;
    }

    return props;
}

function splitMentions(text) {
    return splitMentionTokens(text).map((token) => {
        if (token.username !== undefined) {
            return h(MentionTag, { username: token.username });
        }

        if (token.userId !== undefined) {
            return h(MentionTag, { userId: token.userId });
        }

        return token.text;
    });
}

function childVNodes(el, insideCodeOrLink) {
    const out = [];

    el.childNodes.forEach((node) => {
        const vnode = nodeToVNode(node, insideCodeOrLink);

        if (Array.isArray(vnode)) out.push(...vnode);
        else if (vnode !== null) out.push(vnode);
    });

    return out;
}

function nodeToVNode(node, insideCodeOrLink) {
    if (node.nodeType === Node.TEXT_NODE) {
        // Mentions inside code or links stay literal, matching the old DOM walk.
        return insideCodeOrLink ? node.nodeValue : splitMentions(node.nodeValue);
    }

    if (node.nodeType !== Node.ELEMENT_NODE) return null;

    const el = node;
    const tag = el.tagName.toLowerCase();

    if (tag === "pre" && el.classList.contains("md-code-block")) {
        return h(CodeBlock, {
            code: el.textContent ?? "",
            lang: el.getAttribute("data-lang") || null,
        });
    }

    const nextInside = insideCodeOrLink || tag === "a" || tag === "code" || tag === "pre";

    if (VOID_TAGS.has(tag)) return h(tag, propsFromElement(el));

    return h(tag, propsFromElement(el), childVNodes(el, nextInside));
}

export default {
    name: "MessageContent",
    props: {
        text: { type: String, default: "" },
    },
    setup(props) {
        const { renderMarkdown } = useChatOperations();

        // Depends only on the text, so marked + DOMPurify + this walk run once per
        // message body; a user loading later re-renders only its <MentionTag> child.
        const tree = computed(() => {
            const html = renderMarkdown(props.text ?? "");
            const doc = new DOMParser().parseFromString(html, "text/html");

            return childVNodes(doc.body, false);
        });

        return () => tree.value;
    },
};
</script>
