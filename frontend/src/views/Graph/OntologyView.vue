<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div class="h-full overflow-y-auto bg-gray-50 p-3 sm:p-4">
        <div class="mx-auto flex max-w-6xl flex-col gap-y-4">
            <div>
                <h1 class="text-base font-semibold leading-6 text-gray-900">
                    {{ t("graph.ontology.title") }}
                </h1>
                <p class="mt-1 text-sm text-gray-500">{{ t("graph.ontology.description") }}</p>
            </div>

            <div v-if="graphTypes.errorMessage.value" class="rounded-md bg-red-50 p-4">
                <div class="flex">
                    <XCircleIcon class="h-5 w-5 shrink-0 text-red-400" aria-hidden="true" />
                    <p class="ml-3 text-sm text-red-700">{{ graphTypes.errorMessage.value }}</p>
                </div>
            </div>

            <section
                class="overflow-hidden rounded-xl bg-white ring-1 ring-gray-200"
                :aria-label="t('graph.ontology.map')"
            >
                <div class="border-b border-gray-100 px-4 py-3">
                    <h2 class="text-sm font-semibold text-gray-900">
                        {{ t("graph.ontology.map") }}
                    </h2>
                </div>
                <div class="relative h-80 bg-gray-50/40">
                    <TypeMapCanvas v-if="loaded" :types="types" />
                </div>
                <p
                    v-if="selfRelationCount > 0"
                    class="border-t border-gray-100 px-4 py-3 text-xs text-gray-500"
                >
                    {{ t("graph.ontology.self_relations", { count: selfRelationCount }) }}
                </p>
            </section>

            <section class="overflow-hidden rounded-xl bg-white ring-1 ring-gray-200">
                <div class="border-b border-gray-100 px-4 py-3">
                    <h2 class="text-sm font-semibold text-gray-900">
                        {{ t("graph.ontology.entity_types") }}
                    </h2>
                </div>
                <div class="overflow-x-auto">
                    <table class="min-w-full divide-y divide-gray-200 text-sm">
                        <thead
                            class="bg-gray-50 text-left text-xs font-semibold uppercase tracking-wide text-gray-500"
                        >
                            <tr>
                                <th scope="col" class="px-4 py-3">
                                    {{ t("graph.ontology.column.type") }}
                                </th>
                                <th scope="col" class="px-4 py-3">
                                    {{ t("graph.ontology.column.source") }}
                                </th>
                                <th scope="col" class="px-4 py-3">
                                    {{ t("graph.ontology.column.key") }}
                                </th>
                                <th scope="col" class="px-4 py-3">
                                    {{ t("graph.ontology.column.name") }}
                                </th>
                            </tr>
                        </thead>
                        <tbody class="divide-y divide-gray-200">
                            <tr v-for="entityType in sortedEntityTypes" :key="entityType.id">
                                <td class="px-4 py-3">
                                    <div class="flex items-center gap-x-2">
                                        <span
                                            class="size-2.5 shrink-0 rounded-full"
                                            :style="{ background: styles[entityType.key]?.color }"
                                        ></span>
                                        <span class="font-medium text-gray-900">{{
                                            entityType.name
                                        }}</span>
                                        <span class="font-mono text-xs text-gray-400">{{
                                            entityType.key
                                        }}</span>
                                    </div>
                                </td>
                                <td class="px-4 py-3">
                                    <span
                                        v-if="entityType.built_in"
                                        class="inline-flex items-center gap-x-1 rounded-md bg-gray-50 px-2 py-1 text-xs font-medium text-gray-600 ring-1 ring-inset ring-gray-500/10"
                                    >
                                        <LockClosedIcon class="size-3" aria-hidden="true" />
                                        {{ t("graph.ontology.built_in") }}
                                    </span>
                                    <span v-else-if="entityType.source_name" class="text-gray-700">
                                        {{
                                            t("graph.ontology.from_dataset", {
                                                dataset: entityType.source_name,
                                            })
                                        }}
                                    </span>
                                    <span v-else class="text-gray-700">{{
                                        t("graph.ontology.from_a_dataset")
                                    }}</span>
                                </td>
                                <td class="px-4 py-3 font-mono text-xs text-gray-600">
                                    {{ entityType.key_property_name || "—" }}
                                </td>
                                <td class="px-4 py-3 font-mono text-xs text-gray-600">
                                    {{ entityType.title_property_name || "—" }}
                                </td>
                            </tr>
                        </tbody>
                    </table>
                </div>
            </section>

            <section class="overflow-hidden rounded-xl bg-white ring-1 ring-gray-200">
                <div class="border-b border-gray-100 px-4 py-3">
                    <h2 class="text-sm font-semibold text-gray-900">
                        {{ t("graph.ontology.relation_types") }}
                    </h2>
                </div>
                <div class="overflow-x-auto">
                    <table class="min-w-full divide-y divide-gray-200 text-sm">
                        <thead
                            class="bg-gray-50 text-left text-xs font-semibold uppercase tracking-wide text-gray-500"
                        >
                            <tr>
                                <th scope="col" class="px-4 py-3">
                                    {{ t("graph.ontology.column.relation_type") }}
                                </th>
                                <th scope="col" class="px-4 py-3">
                                    {{ t("graph.ontology.column.kind") }}
                                </th>
                                <th scope="col" class="px-4 py-3">
                                    {{ t("graph.ontology.column.read_from") }}
                                </th>
                            </tr>
                        </thead>
                        <tbody class="divide-y divide-gray-200">
                            <tr v-for="relation in relations" :key="relation.key">
                                <td class="px-4 py-3">
                                    <p class="text-gray-900">
                                        <span class="font-medium">{{ relation.from }}</span>
                                        {{ " " }}
                                        <span class="text-indigo-600">{{ relation.name }}</span>
                                        {{ " " }}
                                        <span class="font-medium">{{ relation.to }}</span>
                                    </p>
                                    <p class="mt-0.5 text-xs text-gray-500">
                                        {{ relation.to }} {{ relation.reverseName }}
                                        {{ relation.from }}
                                    </p>
                                </td>
                                <td class="px-4 py-3">
                                    <span
                                        :class="[
                                            relation.kindClass,
                                            'inline-flex items-center gap-x-1 rounded-md px-2 py-1 text-xs font-medium ring-1 ring-inset',
                                        ]"
                                    >
                                        <LockClosedIcon
                                            v-if="relation.builtIn"
                                            class="size-3"
                                            aria-hidden="true"
                                        />
                                        {{ relation.kindLabel }}
                                    </span>
                                </td>
                                <td class="px-4 py-3 text-gray-600">
                                    <span
                                        :class="relation.readFromIsField ? 'font-mono text-xs' : ''"
                                    >
                                        {{ relation.readFrom }}
                                    </span>
                                </td>
                            </tr>
                        </tbody>
                    </table>
                </div>
            </section>
        </div>
    </div>
</template>

<script setup>
import { ref, computed, onMounted } from "vue";
import { LockClosedIcon, XCircleIcon } from "@heroicons/vue/20/solid";
import { t } from "@/i18n/index.js";
import { typeStyles } from "@/utils/graph/palette.js";
import { useGraphTypes } from "@/composables/graph/useGraphTypes.js";
import TypeMapCanvas from "@/components/Graph/TypeMapCanvas.vue";

const KIND_STYLES = {
    stored: "bg-indigo-50 text-indigo-700 ring-indigo-700/10",
    table: "bg-gray-50 text-gray-600 ring-gray-500/10",
    property: "bg-emerald-50 text-emerald-700 ring-emerald-600/20",
};

// Where Twigex reads each built-in relation from, in words.
const BUILT_IN_SOURCES = {
    chart_uses_dataset_field: "graph.ontology.source.chart_queries",
    dashboard_contains_chart: "graph.ontology.source.dashboard_charts",
    dashboard_uses_dataset_field: "graph.ontology.source.dashboard_filters",
    dataset_field_belongs_to_dataset: "graph.ontology.source.data_model",
};

const graphTypes = useGraphTypes();
const loaded = ref(false);

const types = computed(() => graphTypes.types.value);

const styles = computed(() => typeStyles(types.value.entity_types));

// Your own types first, then the built-in ones.
const sortedEntityTypes = computed(() => {
    const own = types.value.entity_types.filter((entityType) => !entityType.built_in);
    const builtIn = types.value.entity_types.filter((entityType) => entityType.built_in);

    return [...own, ...builtIn];
});

const selfRelationCount = computed(() => {
    return types.value.relation_types.filter(
        (relationType) => relationType.from_type_id === relationType.to_type_id,
    ).length;
});

function typeNameByID(id) {
    const entityType = types.value.entity_types.find((candidate) => candidate.id === id);

    if (!entityType) {
        return "?";
    }

    return entityType.name;
}

function readFrom(relationType) {
    if (relationType.built_in) {
        return t.value(BUILT_IN_SOURCES[relationType.key] ?? "graph.ontology.source.twigex");
    }

    const propertyName = relationType.kind_config?.property_name;

    if (propertyName) {
        return propertyName;
    }

    return t.value("graph.ontology.source.a_field");
}

const relations = computed(() => {
    const own = types.value.relation_types.filter((relationType) => !relationType.built_in);
    const builtIn = types.value.relation_types.filter((relationType) => relationType.built_in);

    return [...own, ...builtIn].map((relationType) => {
        return {
            key: relationType.key,
            from: typeNameByID(relationType.from_type_id),
            to: typeNameByID(relationType.to_type_id),
            name: relationType.name,
            reverseName: relationType.reverse_name,
            builtIn: relationType.built_in,
            kindLabel: t.value(`graph.ontology.kind.${relationType.kind}`),
            kindClass: KIND_STYLES[relationType.kind] ?? KIND_STYLES.table,
            readFrom: readFrom(relationType),
            readFromIsField:
                !relationType.built_in && Boolean(relationType.kind_config?.property_name),
        };
    });
});

onMounted(async () => {
    await graphTypes.ready;

    loaded.value = true;
});
</script>
