<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <Combobox v-model="chosen" as="div" class="relative w-full shrink-0" nullable>
        <label for="graph-search" class="sr-only">{{ t("graph.search.label") }}</label>
        <div class="relative">
            <MagnifyingGlassIcon
                class="pointer-events-none absolute left-3 top-1/2 size-5 -translate-y-1/2 text-gray-400"
                aria-hidden="true"
            />
            <ComboboxInput
                id="graph-search"
                ref="input"
                class="block h-10 w-full rounded-lg border-0 bg-white pl-10 pr-10 text-sm text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600"
                :placeholder="t('graph.search.placeholder')"
                :display-value="typedText"
                autocomplete="off"
                @change="onType($event.target.value)"
            />
            <kbd
                class="pointer-events-none absolute right-3 top-1/2 -translate-y-1/2 rounded border border-gray-300 px-1.5 font-sans text-xs text-gray-500"
            >
                /
            </kbd>
        </div>

        <ComboboxOptions
            class="absolute z-20 mt-2 max-h-96 w-full overflow-y-auto rounded-lg bg-white py-2 text-sm shadow-lg ring-1 ring-black/5 focus:outline-none"
        >
            <li v-if="entityTypes.length === 0" class="px-4 py-2 text-gray-500">
                {{ t("graph.search.no_types") }}
            </li>
            <li v-else-if="searching" class="flex items-center gap-x-3 px-4 py-3 text-gray-500">
                <BaseSpinner
                    size="sm"
                    class="text-indigo-600"
                    :label="t('graph.search.searching')"
                />
                {{ t("graph.search.searching") }}
            </li>
            <li v-else-if="failed" class="px-4 py-2 text-red-600">
                {{ t("graph.search.failed") }}
            </li>
            <li v-else-if="totalResults === 0" class="px-4 py-2 text-gray-500">
                {{ t("graph.search.nothing_found") }}
            </li>

            <template v-for="section in shownSections" :key="section.type">
                <li
                    v-if="section.items.length > 0"
                    class="px-4 pb-1 pt-2 text-xs font-semibold uppercase tracking-wide text-gray-500"
                >
                    {{ section.title }}
                </li>
                <ComboboxOption
                    v-for="item in section.items"
                    :key="`${section.type}:${item.id}`"
                    v-slot="{ active }"
                    :value="choiceValue(section.type, item.id)"
                    as="template"
                >
                    <li
                        :class="[
                            active ? 'bg-indigo-600 text-white' : 'text-gray-900',
                            'flex cursor-pointer select-none items-center gap-x-3 px-4 py-2',
                        ]"
                    >
                        <span
                            class="size-2.5 shrink-0 rounded-full"
                            :style="{ background: section.color }"
                        ></span>
                        <span class="truncate">{{ item.title }}</span>
                    </li>
                </ComboboxOption>
            </template>
        </ComboboxOptions>
    </Combobox>
</template>

<script setup>
import { ref, computed, watch, onMounted, onBeforeUnmount } from "vue";
import { Combobox, ComboboxInput, ComboboxOptions, ComboboxOption } from "@headlessui/vue";
import { MagnifyingGlassIcon } from "@heroicons/vue/20/solid";
import { t } from "@/i18n/index.js";
import graphService from "@/services/graphService";
import BaseSpinner from "@/components/BaseSpinner.vue";

const props = defineProps({
    entityTypes: {
        type: Array,
        required: true,
    },
    styles: {
        type: Object,
        required: true,
    },
});

const emit = defineEmits(["pick"]);

const input = ref(null);
const search = ref("");
const results = ref({});
const searching = ref(false);
const failed = ref(false);
const chosen = ref(null);

// How long to wait after the last keystroke before searching, so a word
// being typed costs one search, not one per letter.
const SEARCH_DELAY = 400;

let searchTimer = null;
let running = null;
let lastSearched = null;

const sections = computed(() => {
    return props.entityTypes.map((entityType) => {
        return {
            type: entityType.key,
            title: entityType.name,
            color: props.styles[entityType.key]?.color,
            items: results.value[entityType.key] ?? [],
        };
    });
});

// While a search runs, the spinner replaces the old results.
const shownSections = computed(() => {
    if (searching.value) {
        return [];
    }

    return sections.value;
});

const totalResults = computed(() => {
    return sections.value.reduce((total, section) => total + section.items.length, 0);
});

// An option's value must stay the same from one redraw to the next: Headless
// UI tracks options by their value, and a new object each time makes it
// redraw without end, which freezes typing.
function choiceValue(type, id) {
    return `${type}:${id}`;
}

watch(chosen, (choice) => pick(choice));

function pick(choice) {
    if (!choice) {
        return;
    }

    const separator = choice.indexOf(":");
    const type = choice.slice(0, separator);
    const id = choice.slice(separator + 1);

    emit("pick", type, id);

    // Cleared, so picking the same result again still counts.
    chosen.value = null;
}

// Cancels the requests of a search still running. The server sees the
// cancelled request too and stops asking Cube.
function cancelRunning() {
    if (running) {
        running.abort();
    }

    running = null;
}

// Each type is searched on its own, as the user; a type the user may not
// read just comes back empty.
async function runSearch() {
    const text = search.value.trim();
    const types = props.entityTypes.map((entityType) => entityType.key).join(",");
    const searchKey = `${types}|${text}`;

    if (searchKey === lastSearched) {
        searching.value = false;

        return;
    }

    cancelRunning();

    const controller = new AbortController();

    running = controller;

    let anyFailed = false;

    const answers = await Promise.all(
        props.entityTypes.map(async (entityType) => {
            try {
                const response = await graphService.searchEntities(
                    entityType.key,
                    text,
                    controller.signal,
                );

                return [entityType.key, response.data.items];
            } catch {
                anyFailed = true;

                return [entityType.key, []];
            }
        }),
    );

    if (controller.signal.aborted) {
        return;
    }

    // A type the user may not read answers with an empty list, not an error,
    // so a failure here is a real problem and must not look like "nothing
    // found".
    failed.value = anyFailed;
    running = null;
    lastSearched = searchKey;
    results.value = Object.fromEntries(answers);
    searching.value = false;
}

// Each keystroke restarts the wait and cancels a search still running, so
// only the last pause in typing reaches the server.
function scheduleSearch() {
    clearTimeout(searchTimer);
    cancelRunning();

    searching.value = true;
    searchTimer = setTimeout(runSearch, SEARCH_DELAY);
}

// The box always shows what the user typed. Without this the combobox puts
// its own text back in the box whenever the results change.
function typedText() {
    return search.value;
}

function onType(text) {
    search.value = text;

    scheduleSearch();
}

// "/" jumps to the search box from anywhere on the page, unless the user is
// already typing somewhere.
function onKeydown(event) {
    const target = event.target;
    const typing =
        target.tagName === "INPUT" || target.tagName === "TEXTAREA" || target.isContentEditable;

    if (event.key !== "/" || typing) {
        return;
    }

    event.preventDefault();
    input.value.$el.focus();
}

watch(() => props.entityTypes, scheduleSearch, { immediate: true });

onMounted(() => {
    window.addEventListener("keydown", onKeydown);
});

onBeforeUnmount(() => {
    clearTimeout(searchTimer);
    cancelRunning();
    window.removeEventListener("keydown", onKeydown);
});
</script>
