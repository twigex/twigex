<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <aside
        class="flex max-h-full w-full shrink-0 flex-col gap-y-5 overflow-y-auto rounded-xl bg-white p-5 ring-1 ring-gray-200 md:w-72 xl:w-80"
        :aria-label="t('graph.details.label')"
    >
        <div>
            <div class="flex items-start justify-between gap-x-3">
                <span
                    :class="[
                        badgeClass,
                        'inline-flex items-center gap-x-1.5 rounded-md px-2 py-1 text-xs font-semibold ring-1 ring-inset',
                    ]"
                >
                    <span class="size-2 rounded-full" :style="{ background: dotColor }"></span>
                    {{ typeLabel }}
                </span>
                <button
                    type="button"
                    class="rounded-md text-gray-400 hover:text-gray-500"
                    @click="emit('close')"
                >
                    <span class="sr-only">{{ t("graph.details.close") }}</span>
                    <XMarkIcon class="h-5 w-5" aria-hidden="true" />
                </button>
            </div>

            <h2 class="mt-3 break-words text-xl font-bold text-gray-900">{{ node.label }}</h2>
            <p class="mt-1 text-sm text-gray-500">{{ hint }}</p>

            <div v-if="!node.isGroup" class="mt-4 flex gap-x-2">
                <BaseButton
                    size="small"
                    :prepend-icon="ArrowsPointingOutIcon"
                    :is-disabled="node.expanded"
                    @click="emit('expand', nodeKey)"
                >
                    {{ t("graph.details.expand") }}
                </BaseButton>
            </div>
        </div>

        <div v-if="node.isGroup">
            <p class="text-xs font-semibold uppercase tracking-wide text-gray-500">
                {{ t("graph.details.list") }}
            </p>
            <ul role="list" class="mt-1">
                <li
                    v-for="item in page.items"
                    :key="item.id"
                    class="flex items-center gap-x-3 border-t border-gray-100 py-2.5 text-sm"
                >
                    <span
                        class="size-2.5 shrink-0 rounded-full"
                        :style="{ background: map.styleFor(item.type).color }"
                    ></span>
                    <button
                        type="button"
                        class="flex-1 truncate text-left text-indigo-600 hover:text-indigo-500"
                        @click="emit('add', item)"
                    >
                        {{ item.title }}
                    </button>
                </li>
            </ul>
            <BaseButton
                v-if="page.next"
                class="mt-3"
                variant="secondary"
                size="small"
                :is-loading="loadingPage"
                @click="loadPage(page.next)"
            >
                {{ t("graph.details.load_more") }}
            </BaseButton>
        </div>

        <div v-else>
            <p class="text-xs font-semibold uppercase tracking-wide text-gray-500">
                {{ t("graph.details.relations") }}
            </p>

            <p v-if="rows.length === 0" class="mt-2 text-sm text-gray-500">
                {{ t("graph.details.no_connections") }}
            </p>

            <ul v-else role="list" class="mt-1">
                <li
                    v-for="row in rows"
                    :key="row.id"
                    class="flex items-center gap-x-3 border-t border-gray-100 py-2.5 text-sm"
                >
                    <span
                        class="w-28 shrink-0 truncate text-[11px] font-semibold uppercase text-gray-500"
                    >
                        {{ row.label }}
                    </span>
                    <span
                        class="size-2.5 shrink-0 rounded-full"
                        :style="{ background: row.color }"
                    ></span>
                    <button
                        type="button"
                        class="min-w-0 flex-1 truncate text-left text-indigo-600 hover:text-indigo-500"
                        @click="emit('open', row.key)"
                    >
                        {{ row.title }}
                    </button>
                </li>
            </ul>
        </div>
    </aside>
</template>

<script setup>
import { ref, computed, watch } from "vue";
import { XMarkIcon } from "@heroicons/vue/24/outline";
import { ArrowsPointingOutIcon } from "@heroicons/vue/20/solid";
import { t } from "@/i18n/index.js";
import graphService from "@/services/graphService";
import { extractErrorMessage } from "@/utils/errors";
import { GROUP_STYLE } from "@/utils/graph/palette.js";
import BaseButton from "@/components/BaseButton.vue";

const props = defineProps({
    map: {
        type: Object,
        required: true,
    },
    nodeKey: {
        type: String,
        required: true,
    },
});

const emit = defineEmits(["close", "open", "add", "expand"]);

const page = ref({ items: [], next: "" });
const loadingPage = ref(false);

const node = computed(() => props.map.nodeAttributes(props.nodeKey) ?? {});

const typeLabel = computed(() => {
    if (node.value.isGroup) {
        return t.value("graph.details.group");
    }

    return props.map.typeName(node.value.entityType);
});

const badgeClass = computed(() => {
    if (node.value.isGroup) {
        return GROUP_STYLE.badge;
    }

    return props.map.styleFor(node.value.entityType).badge;
});

const dotColor = computed(() => {
    if (node.value.isGroup) {
        return GROUP_STYLE.color;
    }

    return props.map.styleFor(node.value.entityType).color;
});

const hint = computed(() => {
    if (node.value.isGroup) {
        return t.value("graph.details.hint_group");
    }

    if (node.value.expanded) {
        return t.value("graph.details.hint_opened");
    }

    return t.value("graph.details.hint_unopened");
});

// One row per related entity. A group too big to send with the entity gets
// one row with its count, which opens the group's list.
const rows = computed(() => {
    const result = [];

    for (const connection of props.map.connections(props.nodeKey)) {
        const color = props.map.styleFor(connection.entityType).color;

        if (connection.items.length === 0) {
            let count = `${connection.count}`;

            if (connection.capped) {
                count = `${connection.count}+`;
            }

            result.push({
                id: connection.id,
                label: connection.label,
                color: color,
                key: connection.boxKey,
                title: t.value("graph.details.group_row", {
                    count: count,
                    type: props.map.typeName(connection.entityType),
                }),
            });

            continue;
        }

        for (const item of connection.items) {
            result.push({
                id: `${connection.id}:${item.key}`,
                label: connection.label,
                color: color,
                key: item.key,
                title: item.title,
            });
        }
    }

    return result;
});

// A box's list is read 50 at a time, from where the last page ended. A page
// that arrives after another dot was opened is dropped.
let pageRequest = 0;

async function loadPage(after) {
    pageRequest = pageRequest + 1;

    const request = pageRequest;
    const group = node.value.group;
    const parent = props.map.graph.getNodeAttributes(group.parentKey);

    loadingPage.value = true;

    try {
        const response = await graphService.getRelationPage(
            parent.entityType,
            parent.entityId,
            group.relation_type,
            group.direction,
            after,
        );

        if (request !== pageRequest) {
            return;
        }

        const entities = response.data.items.map((item) => item.entity);
        const items = [...page.value.items, ...entities];

        page.value = { items: items, next: response.data.next ?? "" };
    } catch (e) {
        if (request === pageRequest) {
            props.map.errorMessage.value = extractErrorMessage(e);
        }
    } finally {
        if (request === pageRequest) {
            loadingPage.value = false;
        }
    }
}

watch(
    () => props.nodeKey,
    () => {
        pageRequest = pageRequest + 1;
        page.value = { items: [], next: "" };
        loadingPage.value = false;

        if (node.value.isGroup) {
            loadPage("");
        }
    },
    { immediate: true },
);
</script>
