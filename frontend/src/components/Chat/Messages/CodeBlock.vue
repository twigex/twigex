<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div class="code-block not-prose">
        <span v-if="validLang" class="code-lang">{{ validLang }}</span>
        <button type="button" class="copy-btn" aria-label="Copy code" @click="copy">
            {{ label }}
        </button>
        <!-- eslint-disable vue/no-v-html -- highlight.js escapes the code it highlights -->
        <pre><code
            v-if="validLang"
            :class="`hljs language-${validLang}`"
            v-html="highlighted"
        ></code><code v-else class="plain">{{ code }}</code></pre>
        <!-- eslint-enable vue/no-v-html -->
    </div>
</template>

<script setup>
import { computed, ref } from "vue";
import hljs from "highlight.js";
import "highlight.js/styles/github-dark.css";
import { t } from "@/i18n/index.js";

const props = defineProps({
    code: { type: String, default: "" },
    lang: { type: String, default: null },
});

const validLang = computed(() => (props.lang && hljs.getLanguage(props.lang) ? props.lang : ""));

// hljs escapes the input, so its output is safe to bind with v-html; the plain
// branch relies on Vue's own text escaping instead.
const highlighted = computed(() =>
    validLang.value ? hljs.highlight(props.code, { language: validLang.value }).value : "",
);

const copied = ref(false);
const label = computed(() =>
    copied.value
        ? t.value("channels.message.copy_code_copied")
        : t.value("channels.message.copy_code"),
);

async function copy() {
    try {
        await navigator.clipboard.writeText(props.code);
        copied.value = true;
        setTimeout(() => (copied.value = false), 1500);
    } catch (err) {
        console.error("Clipboard copy failed", err);
    }
}
</script>

<style>
.code-block {
    position: relative;
    margin: 0.5rem 0;
}

.code-block pre {
    margin: 0;
    padding: 1rem;
    overflow-x: auto;
    border-radius: 0.5rem;
    background-color: #1f2937; /* gray-800 */
    font-size: 0.875rem;
    line-height: 1.6;
    font-family:
        ui-monospace,
        SFMono-Regular,
        SF Mono,
        Menlo,
        Consolas,
        Liberation Mono,
        monospace;
}

.code-block pre code.plain {
    color: #e5e7eb; /* gray-200, matches --tw-prose-pre-code */
    background: transparent;
    padding: 0;
}

.code-block pre code.hljs {
    padding: 0;
    background: transparent;
}

.code-lang {
    position: absolute;
    top: 8px;
    right: 10px;
    font-size: 11px;
    line-height: 20px;
    letter-spacing: 0.03em;
    text-transform: uppercase;
    color: #9ca3af; /* gray-400 */
    pointer-events: none;
    transition: opacity 0.15s ease;
}

/* The Copy button takes the corner on hover, so the label steps aside. */
.code-block:hover .code-lang {
    opacity: 0;
}

.copy-btn {
    position: absolute;
    top: 8px;
    right: 8px;
    cursor: pointer;
    opacity: 0;
    transition: opacity 0.15s ease;
    background: #374151; /* gray-700 */
    border: 1px solid #4b5563; /* gray-600 */
    border-radius: 6px;
    padding: 3px 10px;
    font-size: 12px;
    line-height: 20px;
    color: #d1d5db; /* gray-300 */
}

.copy-btn:hover {
    background: #4b5563; /* gray-600 */
    color: #f9fafb; /* gray-50 */
}

.code-block:hover .copy-btn {
    opacity: 1;
}
</style>
