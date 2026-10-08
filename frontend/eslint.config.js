// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import js from "@eslint/js";
import pluginVue from "eslint-plugin-vue";
import prettier from "eslint-config-prettier";
import globals from "globals";

export default [
    {
        ignores: ["dist/**", "public/**", "node_modules/**"],
    },
    js.configs.recommended,
    ...pluginVue.configs["flat/recommended"],
    {
        languageOptions: {
            ecmaVersion: "latest",
            sourceType: "module",
            globals: {
                ...globals.browser,
            },
        },
    },
    {
        files: ["*.config.js"],
        languageOptions: {
            globals: {
                ...globals.node,
            },
        },
    },
    prettier,
    {
        rules: {
            "padding-line-between-statements": [
                "error",
                { blankLine: "always", prev: "block-like", next: "*" },
                { blankLine: "always", prev: ["const", "let", "var"], next: "*" },
                { blankLine: "any", prev: ["const", "let", "var"], next: ["const", "let", "var"] },
                { blankLine: "always", prev: "*", next: "return" },
            ],

            "vue/require-default-prop": "off",
            "vue/attributes-order": "off",
            "vue/attribute-hyphenation": "off",
            "vue/v-on-event-hyphenation": "off",
            "vue/v-slot-style": "off",

            "no-unused-vars": [
                "error",
                {
                    ignoreRestSiblings: true,
                    argsIgnorePattern: "^_",
                    varsIgnorePattern: "^_",
                    caughtErrorsIgnorePattern: "^_",
                },
            ],

            "no-empty": ["error", { allowEmptyCatch: true }],

            "vue/no-undef-properties": "error",
            "vue/no-import-compiler-macros": "error",

            "no-restricted-syntax": [
                "error",
                {
                    selector: 'CallExpression[callee.type="Identifier"][callee.name="t"]',
                    message:
                        "t from @/i18n is a computed ref: call t.value(...) in script code. Name an injected translate function translate.",
                },
            ],

            // Editing fields of an object prop in place is an accepted pattern here;
            // replacing the prop itself is not.
            "vue/no-mutating-props": ["error", { shallowOnly: true }],
        },
    },
    {
        files: ["**/__tests__/**", "**/*.spec.js"],
        rules: {
            "vue/one-component-per-file": "off",
        },
    },
];
