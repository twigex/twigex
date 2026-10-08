// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { ref, computed } from "vue";
import DOMPurify from "dompurify";
import en from "./en.json";

export const currentLocale = ref("en");

const messages = {
    en,
};

/**
 * Vite-safe dynamic locale imports
 * All JSON files in this folder are included in the build
 */
const localeModules = import.meta.glob("./*.json");

/**
 * Global translate function
 * Fallback:
 * 1. current locale
 * 2. English (default)
 * 3. the key itself
 */
/**
 * Pick the plural form for `count` from a "|"-separated, category-labelled string,
 * e.g. "one {count} item | other {count} items". Uses Intl.PluralRules so each
 * locale resolves its own categories (ru: one/few/many/other, lv: zero/one/other).
 */
function selectPluralForm(text, count, locale) {
    const forms = {};

    for (const part of text.split("|")) {
        const match = part.trim().match(/^(zero|one|two|few|many|other)\s+([\s\S]*)$/);

        if (match) forms[match[1]] = match[2];
    }

    if (!Object.keys(forms).length) return text;

    const category = new Intl.PluralRules(locale).select(count);

    return forms[category] ?? forms.other ?? text;
}

export const t = computed(() => {
    const lang = messages[currentLocale.value] || {};
    const enLang = messages["en"] || {};

    return (key, vars = {}) => {
        let text = lang[key] || enLang[key] || key;

        if (typeof vars.count === "number" && text.includes("|")) {
            text = selectPluralForm(text, vars.count, currentLocale.value);
        }

        // Apply replacements
        for (const [name, value] of Object.entries(vars)) {
            text = text.replaceAll(`{${name}}`, value);
        }

        return text;
    };
});

/**
 * Escape variables for safe HTML interpolation
 */
function escapeHtml(str) {
    return String(str)
        .replace(/&/g, "&amp;")
        .replace(/</g, "&lt;")
        .replace(/>/g, "&gt;")
        .replace(/"/g, "&quot;")
        .replace(/'/g, "&#039;");
}

/**
 * --- HTML TRANSLATION ---
 * Allows controlled HTML (<strong>, <b>, <span>…)
 * Example:
 * tHtml("files.delete_dialog.confirm_delete", { text: "File A" })
 *
 * Use in template with v-html:
 * <p v-html="tHtml('key', {text: 'value'})" />
 */
export function tHtml(key, vars = {}) {
    let template = t.value(key) || key;

    // Insert escaped variables
    for (const [name, value] of Object.entries(vars)) {
        template = template.replaceAll(`{${name}}`, escapeHtml(value));
    }

    // Sanitize output and allow useful formatting tags
    return DOMPurify.sanitize(template, {
        ALLOWED_TAGS: ["strong", "b", "i", "em", "u", "span", "br"],
        ALLOWED_ATTR: ["class"],
    });
}

/**
 * Load & set new locale
 * - Works in dev AND production
 */
export async function setLocale(locale) {
    if (!messages[locale]) {
        const loader = localeModules[`./${locale}.json`];

        if (!loader) {
            console.error(`Locale file for "${locale}" not found.`);

            return;
        }

        const data = await loader();

        messages[locale] = data.default;
    }

    currentLocale.value = locale;
}
