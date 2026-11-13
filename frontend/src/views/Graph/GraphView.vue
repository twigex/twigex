<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div class="flex h-full flex-col gap-y-3 overflow-hidden bg-gray-50 p-3 sm:p-4">
        <GraphSearch :entity-types="searchTypes" :styles="styles" @pick="pickStart" />

        <div
            class="flex shrink-0 flex-wrap gap-2"
            role="group"
            :aria-label="t('graph.chips.label')"
        >
            <button
                type="button"
                :aria-pressed="map.shownTypes.value.length === 0"
                :class="chipClass(map.shownTypes.value.length === 0)"
                @click="map.showAllTypes()"
            >
                {{ t("graph.chips.all") }}
            </button>
            <button
                v-for="entityType in entityTypes"
                :key="entityType.key"
                type="button"
                :aria-pressed="map.shownTypes.value.includes(entityType.key)"
                :class="chipClass(map.shownTypes.value.includes(entityType.key))"
                @click="map.toggleType(entityType.key)"
            >
                <span
                    class="size-2.5 rounded-full"
                    :style="{ background: styles[entityType.key]?.color }"
                ></span>
                {{ entityType.name }}
            </button>
        </div>

        <div class="flex min-h-0 flex-1 flex-col gap-4 md:flex-row">
            <section
                class="relative flex min-h-[24rem] min-w-0 flex-1 flex-col overflow-hidden rounded-xl bg-white ring-1 ring-gray-200"
                :aria-label="t('graph.card.label')"
            >
                <div class="flex flex-wrap items-center gap-3 border-b border-gray-100 px-4 py-3">
                    <nav
                        v-if="pathSteps.length > 0"
                        class="flex min-w-0 flex-1"
                        :aria-label="t('graph.path')"
                    >
                        <ol role="list" class="flex flex-wrap items-center gap-x-2 gap-y-1 text-sm">
                            <li
                                v-for="(step, index) in pathSteps"
                                :key="step.key"
                                class="flex items-center gap-x-2"
                            >
                                <ChevronRightIcon
                                    v-if="index > 0"
                                    class="size-4 shrink-0 text-gray-400"
                                    aria-hidden="true"
                                />
                                <button
                                    type="button"
                                    :class="
                                        index === pathSteps.length - 1
                                            ? 'font-semibold text-gray-900'
                                            : 'text-indigo-600 hover:text-indigo-500'
                                    "
                                    @click="inspect(step.key)"
                                >
                                    {{ step.label }}
                                </button>
                            </li>
                        </ol>
                    </nav>
                    <p v-else class="min-w-0 flex-1 truncate text-sm font-semibold text-gray-900">
                        {{ t("graph.title") }}
                    </p>

                    <span v-if="hasStart" class="text-sm text-gray-500">
                        {{
                            t("graph.card.counts", {
                                entities: counts.entities,
                                relations: counts.relations,
                            })
                        }}
                    </span>

                    <span class="isolate inline-flex rounded-md shadow-sm">
                        <button
                            v-for="(option, index) in layoutOptions"
                            :key="option.value"
                            type="button"
                            :class="[
                                map.layout.value === option.value
                                    ? 'bg-indigo-50 text-indigo-700'
                                    : 'bg-white text-gray-700 hover:bg-gray-50',
                                index === 0 ? 'rounded-l-md' : '-ml-px rounded-r-md',
                                'relative inline-flex items-center px-3 py-1.5 text-sm font-semibold ring-1 ring-inset ring-gray-300 focus:z-10',
                            ]"
                            @click="map.setLayout(option.value)"
                        >
                            {{ option.label }}
                        </button>
                    </span>

                    <BaseButton
                        variant="secondary"
                        size="small"
                        :is-disabled="!canExpandSelected"
                        @click="expandSelected"
                    >
                        {{ t("graph.card.expand_selected") }}
                    </BaseButton>
                    <BaseButton
                        variant="secondary"
                        size="small"
                        :prepend-icon="ArrowPathIcon"
                        :is-disabled="!hasStart"
                        @click="startFromRoute"
                    >
                        {{ t("graph.start_over") }}
                    </BaseButton>
                </div>

                <div class="relative flex-1 bg-gray-50/40">
                    <GraphCanvas :map="map" @inspect="inspect" @expand="expand" />

                    <div
                        v-if="!hasStart"
                        class="pointer-events-none absolute inset-0 flex flex-col items-center justify-center text-center"
                    >
                        <ShareIcon class="mx-auto h-12 w-12 text-gray-300" aria-hidden="true" />
                        <h3 class="mt-2 text-sm font-semibold text-gray-900">
                            {{ t("graph.empty.title") }}
                        </h3>
                        <p class="mt-1 text-sm text-gray-500">
                            {{ t("graph.empty.description") }}
                        </p>
                    </div>

                    <div
                        v-if="map.errorMessage.value"
                        class="absolute inset-x-4 top-4 mx-auto max-w-md rounded-md bg-red-50 p-4 shadow-sm"
                    >
                        <div class="flex">
                            <XCircleIcon class="h-5 w-5 shrink-0 text-red-400" aria-hidden="true" />
                            <p class="ml-3 text-sm text-red-700">
                                {{ map.errorMessage.value }}
                            </p>
                        </div>
                    </div>
                </div>

                <div
                    class="flex flex-wrap items-center gap-x-5 gap-y-2 border-t border-gray-100 px-4 py-3 text-xs text-gray-500"
                >
                    <span class="inline-flex items-center gap-x-1.5">
                        <span class="h-px w-7 bg-gray-400"></span>
                        {{ t("graph.legend.relation") }}
                    </span>
                    <span class="inline-flex items-center gap-x-1.5">
                        <span class="size-2.5 rounded-full bg-gray-400"></span>
                        {{ t("graph.legend.group") }}
                    </span>
                    <span class="flex-1"></span>
                    <span>{{ t("graph.legend.hint") }}</span>
                </div>
            </section>

            <GraphDetails
                v-if="panelNode"
                :map="map"
                :node-key="panelKey"
                @close="panelKey = ''"
                @open="openConnection"
                @add="addFromBox"
                @expand="expand"
            />
        </div>
    </div>
</template>

<script setup>
import { ref, computed, onMounted, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import { ArrowPathIcon, ChevronRightIcon } from "@heroicons/vue/20/solid";
import { ShareIcon, XCircleIcon } from "@heroicons/vue/24/outline";
import { t } from "@/i18n/index.js";
import { typeStyles } from "@/utils/graph/palette.js";
import { useGraphMap } from "@/composables/graph/useGraphMap.js";
import { useGraphTypes } from "@/composables/graph/useGraphTypes.js";
import BaseButton from "@/components/BaseButton.vue";
import GraphSearch from "@/components/Graph/GraphSearch.vue";
import GraphCanvas from "@/components/Graph/GraphCanvas.vue";
import GraphDetails from "@/components/Graph/GraphDetails.vue";

const route = useRoute();
const router = useRouter();
const map = useGraphMap();

const graphTypes = useGraphTypes();

const entityTypes = computed(() => graphTypes.types.value.entity_types);

const layoutOptions = computed(() => [
    { value: "columns", label: t.value("graph.layout.columns") },
    { value: "free", label: t.value("graph.layout.free") },
]);

const currentType = computed(() => route.query.type ?? "");
const currentId = computed(() => route.query.id ?? "");
const hasStart = computed(() => Boolean(currentType.value && currentId.value));

// Only types that read their entities from a dataset can be searched; the
// built-in ones are reached by following relations.
const datasetTypes = computed(() => {
    return entityTypes.value.filter((entityType) => entityType.source === "dataset");
});

const searchTypes = computed(() => {
    return datasetTypes.value.filter((entityType) => map.isTypeShown(entityType.key));
});

const styles = computed(() => typeStyles(entityTypes.value));

const selectedNode = computed(() => map.nodeAttributes(map.selectedKey.value));

// The panel opens only once a dot is clicked, so the map uses the full width
// until there is something to show. The starting dot is selected from the
// start, for the path, without opening the panel.
const panelKey = ref("");

const panelNode = computed(() => map.nodeAttributes(panelKey.value));

const counts = computed(() => map.counts());

const canExpandSelected = computed(() => {
    const node = selectedNode.value;

    return Boolean(node && !node.isGroup && !node.expanded);
});

const pathSteps = computed(() => {
    return map.currentPath().keys.map((key) => {
        return { key: key, label: map.graph.getNodeAttribute(key, "label") };
    });
});

function chipClass(on) {
    const base =
        "inline-flex h-8 items-center gap-x-1.5 rounded-full px-3 text-sm font-medium ring-1 ring-inset";

    if (on) {
        return `${base} bg-gray-900 text-white ring-gray-900`;
    }

    return `${base} bg-white text-gray-700 ring-gray-300 hover:bg-gray-50`;
}

function pickStart(type, id) {
    router.push({ name: "graph-explore", query: { type: type, id: id } });
}

function startFromRoute() {
    panelKey.value = "";

    map.start(currentType.value, currentId.value);
}

// A click shows a dot in the panel without changing the map; its relations are
// read so the panel can list them.
async function inspect(key) {
    map.select(key);
    panelKey.value = key;

    const node = map.graph.getNodeAttributes(key);

    if (!node.isGroup) {
        await map.loadRelations(key);
    }
}

async function expand(key) {
    await map.expand(key);

    if (!map.graph.hasNode(key)) {
        return;
    }

    map.select(key);
    panelKey.value = key;
}

function expandSelected() {
    expand(map.selectedKey.value);
}

// A name in the panel may not be on the map yet when the selected dot hasn't
// been opened; opening it puts its neighbours there first.
async function openConnection(key) {
    if (!map.graph.hasNode(key)) {
        await map.expand(panelKey.value);
    }

    if (map.graph.hasNode(key)) {
        await inspect(key);
    }
}

function addFromBox(item) {
    const key = map.addFromGroup(panelKey.value, item);

    inspect(key);
}

onMounted(async () => {
    await graphTypes.ready;

    map.setTypes(graphTypes.types.value);

    if (graphTypes.errorMessage.value) {
        map.errorMessage.value = graphTypes.errorMessage.value;
    }

    startFromRoute();
});

watch(
    () => [route.query.type, route.query.id],
    () => startFromRoute(),
);
</script>
