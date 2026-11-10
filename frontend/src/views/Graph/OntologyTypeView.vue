<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div class="h-full overflow-y-auto bg-gray-50 p-3 sm:p-4">
        <div class="mx-auto flex max-w-6xl flex-col gap-y-4">
            <RouterLink
                :to="{ name: 'graph-ontology' }"
                class="inline-flex w-fit items-center gap-x-1 text-sm font-medium text-gray-500 hover:text-gray-700"
            >
                <ChevronLeftIcon class="size-4" aria-hidden="true" />
                {{ t("graph.ontology.title") }}
            </RouterLink>

            <p
                v-if="loaded && !entityType"
                class="rounded-md bg-white p-6 text-sm text-gray-500 ring-1 ring-gray-200"
            >
                {{ t("graph.ontology.type_not_found") }}
            </p>

            <template v-if="entityType">
                <div class="flex items-start gap-x-4">
                    <span
                        :class="[
                            style.tile,
                            'flex size-12 shrink-0 items-center justify-center rounded-xl',
                        ]"
                    >
                        <component
                            :is="ontology.iconFor(entityType)"
                            class="size-7"
                            aria-hidden="true"
                        />
                    </span>
                    <div class="min-w-0 flex-1">
                        <div class="flex flex-wrap items-center gap-2">
                            <h1 class="text-xl font-bold text-gray-900">{{ entityType.name }}</h1>
                            <span
                                :class="[
                                    ontology.statusClass(entityType.status),
                                    'inline-flex items-center rounded-md px-2 py-1 text-xs font-medium ring-1 ring-inset',
                                ]"
                            >
                                {{ ontology.statusLabel(entityType.status) }}
                            </span>
                            <span
                                v-if="entityType.built_in"
                                class="inline-flex items-center gap-x-1 rounded-md bg-gray-50 px-2 py-1 text-xs font-medium text-gray-600 ring-1 ring-inset ring-gray-500/10"
                            >
                                <LockClosedIcon class="size-3" aria-hidden="true" />
                                {{ t("graph.ontology.built_in") }}
                            </span>
                        </div>
                        <p class="mt-1 text-sm text-gray-500">
                            {{ ontology.pluralName(entityType) }} ·
                            <span class="font-mono text-xs">{{ entityType.key }}</span>
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
                                'whitespace-nowrap border-b-2 px-1 py-3 text-sm font-medium',
                            ]"
                            @click="showTab(tab.value)"
                        >
                            {{ tab.label }}
                        </button>
                    </nav>
                </div>

                <section
                    v-if="currentTab === 'overview'"
                    class="rounded-xl bg-white p-5 ring-1 ring-gray-200"
                >
                    <dl class="grid grid-cols-1 gap-x-6 gap-y-5 sm:grid-cols-2">
                        <div
                            v-for="item in overview"
                            :key="item.label"
                            :class="item.wide ? 'sm:col-span-2' : ''"
                        >
                            <dt class="text-xs font-medium text-gray-500">{{ item.label }}</dt>
                            <dd
                                class="mt-1 text-sm text-gray-900"
                                :class="item.mono ? 'font-mono text-xs' : ''"
                            >
                                {{ item.value }}
                            </dd>
                        </div>
                    </dl>
                </section>

                <section
                    v-else-if="currentTab === 'fields'"
                    class="overflow-hidden rounded-xl bg-white ring-1 ring-gray-200"
                >
                    <table class="min-w-full divide-y divide-gray-200 text-sm">
                        <thead
                            class="bg-gray-50 text-left text-xs font-semibold uppercase tracking-wide text-gray-500"
                        >
                            <tr>
                                <th scope="col" class="px-4 py-3">
                                    {{ t("graph.ontology.column.field") }}
                                </th>
                                <th scope="col" class="px-4 py-3">
                                    {{ t("graph.ontology.column.role") }}
                                </th>
                            </tr>
                        </thead>
                        <tbody class="divide-y divide-gray-200">
                            <tr v-for="field in fields" :key="field.role">
                                <td class="px-4 py-3 font-mono text-xs text-gray-700">
                                    {{ field.name }}
                                </td>
                                <td class="px-4 py-3 text-gray-600">{{ field.role }}</td>
                            </tr>
                        </tbody>
                    </table>
                    <p class="border-t border-gray-100 px-4 py-3 text-xs text-gray-500">
                        {{ fieldsNote }}
                    </p>
                </section>

                <section v-else class="overflow-hidden rounded-xl bg-white ring-1 ring-gray-200">
                    <p v-if="relations.length === 0" class="p-6 text-sm text-gray-500">
                        {{ t("graph.ontology.no_relation_types") }}
                    </p>
                    <table v-else class="min-w-full divide-y divide-gray-200 text-sm">
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
                            </tr>
                        </thead>
                        <tbody class="divide-y divide-gray-200">
                            <tr v-for="relation in relations" :key="relation.id">
                                <td class="px-4 py-3 text-gray-900">
                                    <span class="font-medium">{{ entityType.name }}</span>
                                    {{ " " }}
                                    <span class="text-indigo-600">{{ relation.name }}</span>
                                    {{ " " }}
                                    <RouterLink
                                        v-if="relation.other"
                                        :to="{
                                            name: 'graph-ontology-type',
                                            params: { key: relation.other.key },
                                        }"
                                        class="font-medium text-gray-900 underline decoration-gray-300 underline-offset-2 hover:decoration-gray-500"
                                    >
                                        {{ relation.other.name }}
                                    </RouterLink>
                                </td>
                                <td class="px-4 py-3 text-gray-600">
                                    {{ ontology.cardinalityLabel(relation.cardinality) }}
                                </td>
                                <td class="px-4 py-3">
                                    <span
                                        :class="[
                                            ontology.kindClass(relation.relationType),
                                            'inline-flex items-center rounded-md px-2 py-1 text-xs font-medium ring-1 ring-inset',
                                        ]"
                                    >
                                        {{ ontology.kindLabel(relation.relationType) }}
                                    </span>
                                </td>
                            </tr>
                        </tbody>
                    </table>
                </section>
            </template>
        </div>
    </div>
</template>

<script setup>
import { ref, computed, onMounted } from "vue";
import { RouterLink, useRoute, useRouter } from "vue-router";
import { ChevronLeftIcon, LockClosedIcon } from "@heroicons/vue/20/solid";
import { t } from "@/i18n/index.js";
import { useOntology } from "@/composables/graph/useOntology.js";
import { relationsOf } from "@/utils/graph/relations.js";

const route = useRoute();
const router = useRouter();
const ontology = useOntology();

const loaded = ref(false);

const entityType = computed(() => {
    return ontology.types.value.entity_types.find(
        (candidate) => candidate.key === route.params.key,
    );
});

const style = computed(() => ontology.styles.value[entityType.value.key]);

const currentTab = computed(() => {
    const tab = route.query.tab;

    if (tab === "fields" || tab === "relations") {
        return tab;
    }

    return "overview";
});

const tabs = computed(() => [
    { value: "overview", label: t.value("graph.ontology.tab.overview") },
    { value: "fields", label: t.value("graph.ontology.tab.fields") },
    { value: "relations", label: t.value("graph.ontology.tab.relations") },
]);

function formatDate(seconds) {
    if (!seconds) {
        return "—";
    }

    return new Date(seconds * 1000).toLocaleString();
}

const overview = computed(() => {
    const type = entityType.value;

    return [
        { label: t.value("graph.ontology.overview.key"), value: type.key, mono: true },
        { label: t.value("graph.ontology.column.source"), value: ontology.sourceLabel(type) },
        {
            label: t.value("graph.ontology.overview.status"),
            value: ontology.statusLabel(type.status),
        },
        { label: t.value("graph.ontology.overview.plural"), value: ontology.pluralName(type) },
        {
            label: t.value("graph.ontology.overview.description"),
            value: type.description || t.value("graph.ontology.no_description"),
            wide: true,
        },
        { label: t.value("graph.ontology.overview.created"), value: formatDate(type.created_at) },
        { label: t.value("graph.ontology.overview.updated"), value: formatDate(type.updated_at) },
    ];
});

// A type has two fields of its own: the key and the name.
const fields = computed(() => {
    const type = entityType.value;

    return [
        { name: type.key_property_name || "—", role: t.value("graph.ontology.column.key") },
        { name: type.title_property_name || "—", role: t.value("graph.ontology.column.name") },
    ];
});

const fieldsNote = computed(() => {
    if (entityType.value.built_in) {
        return t.value("graph.ontology.fields_built_in");
    }

    return t.value("graph.ontology.fields_later");
});

const relations = computed(() => {
    return relationsOf(
        entityType.value,
        ontology.types.value.relation_types,
        ontology.typeByID.value,
    );
});

function showTab(tab) {
    router.replace({ query: { ...route.query, tab: tab } });
}

onMounted(async () => {
    await ontology.graphTypes.ready;

    loaded.value = true;
});
</script>
