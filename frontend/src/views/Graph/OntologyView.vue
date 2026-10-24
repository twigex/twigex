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

            <div v-if="ontology.graphTypes.errorMessage.value" class="rounded-md bg-red-50 p-4">
                <div class="flex">
                    <XCircleIcon class="h-5 w-5 shrink-0 text-red-400" aria-hidden="true" />
                    <p class="ml-3 text-sm text-red-700">
                        {{ ontology.graphTypes.errorMessage.value }}
                    </p>
                </div>
            </div>

            <div class="border-b border-gray-200">
                <nav class="-mb-px flex gap-x-6" :aria-label="t('graph.ontology.tabs')">
                    <button
                        v-for="tab in tabs"
                        :key="tab.value"
                        type="button"
                        :aria-current="currentTab === tab.value ? 'page' : undefined"
                        :class="[
                            currentTab === tab.value
                                ? 'border-indigo-500 text-indigo-600'
                                : 'border-transparent text-gray-500 hover:border-gray-300 hover:text-gray-700',
                            'flex items-center gap-x-2 whitespace-nowrap border-b-2 px-1 py-3 text-sm font-medium',
                        ]"
                        @click="showTab(tab.value)"
                    >
                        {{ tab.label }}
                        <span
                            v-if="tab.count !== undefined"
                            :class="[
                                currentTab === tab.value
                                    ? 'bg-indigo-100 text-indigo-600'
                                    : 'bg-gray-100 text-gray-900',
                                'rounded-full px-2.5 py-0.5 text-xs font-medium',
                            ]"
                        >
                            {{ tab.count }}
                        </span>
                    </button>
                </nav>
            </div>

            <ul
                v-if="currentTab === 'types'"
                role="list"
                class="grid grid-cols-1 gap-4 sm:grid-cols-2 xl:grid-cols-3"
            >
                <li v-for="entityType in ontology.sortedEntityTypes.value" :key="entityType.id">
                    <RouterLink
                        :to="{ name: 'graph-ontology-type', params: { key: entityType.key } }"
                        class="flex h-full flex-col rounded-xl bg-white p-4 ring-1 ring-gray-200 hover:ring-indigo-300"
                    >
                        <div class="flex items-start gap-x-3">
                            <span
                                :class="[
                                    ontology.styles.value[entityType.key].tile,
                                    'flex size-10 shrink-0 items-center justify-center rounded-lg',
                                ]"
                            >
                                <component
                                    :is="ontology.iconFor(entityType)"
                                    class="size-6"
                                    aria-hidden="true"
                                />
                            </span>
                            <div class="min-w-0 flex-1">
                                <p class="truncate text-sm font-semibold text-gray-900">
                                    {{ ontology.pluralName(entityType) }}
                                </p>
                                <p class="truncate font-mono text-xs text-gray-400">
                                    {{ entityType.key }}
                                </p>
                            </div>
                            <span
                                :class="[
                                    ontology.statusClass(entityType.status),
                                    'inline-flex shrink-0 items-center rounded-md px-2 py-1 text-xs font-medium ring-1 ring-inset',
                                ]"
                            >
                                {{ ontology.statusLabel(entityType.status) }}
                            </span>
                        </div>
                        <p class="mt-3 line-clamp-2 flex-1 text-sm text-gray-500">
                            {{ entityType.description || t("graph.ontology.no_description") }}
                        </p>
                        <div
                            class="mt-4 flex items-center justify-between gap-x-3 border-t border-gray-100 pt-3 text-xs text-gray-500"
                        >
                            <span class="inline-flex min-w-0 items-center gap-x-1 truncate">
                                <LockClosedIcon
                                    v-if="entityType.built_in"
                                    class="size-3 shrink-0"
                                    aria-hidden="true"
                                />
                                {{ ontology.sourceLabel(entityType) }}
                            </span>
                            <span class="shrink-0">
                                {{
                                    t("graph.ontology.relation_type_count", {
                                        count: ontology.relationTypeCount(entityType),
                                    })
                                }}
                            </span>
                        </div>
                    </RouterLink>
                </li>
            </ul>

            <section
                v-else-if="currentTab === 'relations'"
                class="overflow-hidden rounded-xl bg-white ring-1 ring-gray-200"
            >
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
                                    {{ t("graph.ontology.column.cardinality") }}
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
                            <tr
                                v-for="relationType in ontology.sortedRelationTypes.value"
                                :key="relationType.key"
                            >
                                <td class="px-4 py-3">
                                    <p class="text-gray-900">
                                        <span class="font-medium">{{
                                            ontology.typeName(relationType.from_type_id)
                                        }}</span>
                                        {{ " " }}
                                        <span class="text-indigo-600">{{ relationType.name }}</span>
                                        {{ " " }}
                                        <span class="font-medium">{{
                                            ontology.typeName(relationType.to_type_id)
                                        }}</span>
                                    </p>
                                    <p class="mt-0.5 text-xs text-gray-500">
                                        {{ ontology.typeName(relationType.to_type_id) }}
                                        {{ relationType.reverse_name }}
                                        {{ ontology.typeName(relationType.from_type_id) }}
                                    </p>
                                </td>
                                <td class="px-4 py-3 text-gray-600">
                                    {{ ontology.cardinalityLabel(relationType.cardinality) }}
                                </td>
                                <td class="px-4 py-3">
                                    <span
                                        :class="[
                                            ontology.kindClass(relationType),
                                            'inline-flex items-center gap-x-1 rounded-md px-2 py-1 text-xs font-medium ring-1 ring-inset',
                                        ]"
                                    >
                                        <LockClosedIcon
                                            v-if="relationType.built_in"
                                            class="size-3"
                                            aria-hidden="true"
                                        />
                                        {{ ontology.kindLabel(relationType) }}
                                    </span>
                                </td>
                                <td class="px-4 py-3 text-gray-600">
                                    <span
                                        :class="
                                            ontology.readsFromField(relationType)
                                                ? 'font-mono text-xs'
                                                : ''
                                        "
                                    >
                                        {{ ontology.readFrom(relationType) }}
                                    </span>
                                </td>
                            </tr>
                        </tbody>
                    </table>
                </div>
            </section>

            <section v-else class="overflow-hidden rounded-xl bg-white ring-1 ring-gray-200">
                <div class="relative h-[28rem] bg-gray-50/40">
                    <TypeMapCanvas v-if="loaded" :types="ontology.types.value" />
                </div>
                <p
                    v-if="selfRelationCount > 0"
                    class="border-t border-gray-100 px-4 py-3 text-xs text-gray-500"
                >
                    {{ t("graph.ontology.self_relations", { count: selfRelationCount }) }}
                </p>
            </section>
        </div>
    </div>
</template>

<script setup>
import { ref, computed, onMounted } from "vue";
import { RouterLink, useRoute, useRouter } from "vue-router";
import { LockClosedIcon, XCircleIcon } from "@heroicons/vue/20/solid";
import { t } from "@/i18n/index.js";
import { useOntology } from "@/composables/graph/useOntology.js";
import TypeMapCanvas from "@/components/Graph/TypeMapCanvas.vue";

const route = useRoute();
const router = useRouter();
const ontology = useOntology();

const loaded = ref(false);

// The open tab is kept in the address, so a reload or a shared link opens it.
const currentTab = computed(() => {
    const tab = route.query.tab;

    if (tab === "relations" || tab === "map") {
        return tab;
    }

    return "types";
});

const tabs = computed(() => [
    {
        value: "types",
        label: t.value("graph.ontology.entity_types"),
        count: ontology.types.value.entity_types.length,
    },
    {
        value: "relations",
        label: t.value("graph.ontology.relation_types"),
        count: ontology.types.value.relation_types.length,
    },
    { value: "map", label: t.value("graph.ontology.map") },
]);

const selfRelationCount = computed(() => {
    return ontology.types.value.relation_types.filter(
        (relationType) => relationType.from_type_id === relationType.to_type_id,
    ).length;
});

function showTab(tab) {
    router.replace({ query: { ...route.query, tab: tab } });
}

onMounted(async () => {
    await ontology.graphTypes.ready;

    loaded.value = true;
});
</script>
