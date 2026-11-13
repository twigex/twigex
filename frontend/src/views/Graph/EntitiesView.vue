<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div class="h-full overflow-y-auto bg-gray-50 p-3 sm:p-4">
        <div class="mx-auto flex max-w-6xl flex-col gap-y-4">
            <div>
                <h1 class="text-base font-semibold leading-6 text-gray-900">
                    {{ t("graph.entities.title") }}
                </h1>
                <p class="mt-1 text-sm text-gray-500">{{ t("graph.entities.description") }}</p>
            </div>

            <div v-if="ontology.graphTypes.errorMessage.value" class="rounded-md bg-red-50 p-4">
                <div class="flex">
                    <XCircleIcon class="h-5 w-5 shrink-0 text-red-400" aria-hidden="true" />
                    <p class="ml-3 text-sm text-red-700">
                        {{ ontology.graphTypes.errorMessage.value }}
                    </p>
                </div>
            </div>

            <div
                v-if="loaded && datasetTypes.length === 0"
                class="rounded-xl bg-white p-8 text-center ring-1 ring-gray-200"
            >
                <ListBulletIcon class="mx-auto size-10 text-gray-300" aria-hidden="true" />
                <p class="mt-2 text-sm font-semibold text-gray-900">
                    {{ t("graph.entities.no_types") }}
                </p>
                <p class="mt-1 text-sm text-gray-500">{{ t("graph.entities.no_types_help") }}</p>
                <RouterLink
                    :to="{ name: 'graph-ontology' }"
                    class="mt-4 inline-block text-sm font-semibold text-indigo-600 hover:text-indigo-500"
                >
                    {{ t("graph.entities.open_ontology") }}
                </RouterLink>
            </div>

            <template v-else-if="loaded">
                <nav class="flex flex-wrap gap-2" :aria-label="t('graph.ontology.entity_types')">
                    <RouterLink
                        v-for="entityType in datasetTypes"
                        :key="entityType.key"
                        :to="{ name: 'graph-entities', params: { key: entityType.key } }"
                        :aria-current="entityType.key === currentKey ? 'page' : undefined"
                        :class="[
                            entityType.key === currentKey
                                ? 'bg-white text-gray-900 ring-2 ring-indigo-600'
                                : 'bg-white text-gray-600 ring-1 ring-gray-200 hover:ring-indigo-300',
                            'inline-flex items-center gap-x-2 rounded-lg py-1.5 pl-1.5 pr-3 text-sm font-medium',
                        ]"
                    >
                        <span
                            :class="[
                                ontology.styles.value[entityType.key].tile,
                                'flex size-6 items-center justify-center rounded-md',
                            ]"
                        >
                            <component
                                :is="ontology.iconFor(entityType)"
                                class="size-4"
                                aria-hidden="true"
                            />
                        </span>
                        {{ ontology.pluralName(entityType) }}
                    </RouterLink>
                </nav>

                <p
                    v-if="!currentType"
                    class="rounded-xl bg-white p-8 text-center text-sm text-gray-500 ring-1 ring-gray-200"
                >
                    {{ t("graph.entities.type_not_found") }}
                </p>

                <section v-else class="overflow-hidden rounded-xl bg-white ring-1 ring-gray-200">
                    <div
                        class="flex flex-wrap items-center justify-between gap-3 border-b border-gray-200 px-4 py-3"
                    >
                        <RouterLink
                            :to="{
                                name: 'graph-ontology-type',
                                params: { key: currentType.key },
                            }"
                            class="text-sm font-medium text-indigo-600 hover:text-indigo-500"
                        >
                            {{ t("graph.entities.type_details") }}
                        </RouterLink>
                        <div class="relative w-full sm:w-72">
                            <MagnifyingGlassIcon
                                class="pointer-events-none absolute inset-y-0 left-2.5 my-auto size-4 text-gray-400"
                                aria-hidden="true"
                            />
                            <input
                                v-model="search"
                                type="search"
                                :aria-label="t('graph.entities.search')"
                                :placeholder="t('graph.entities.search')"
                                class="block w-full rounded-md border-0 py-1.5 pl-8 text-sm text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600"
                                @input="scheduleSearch"
                            />
                        </div>
                    </div>

                    <div v-if="list.loading.value" class="flex justify-center py-10">
                        <BaseSpinner :label="t('graph.form.loading')" />
                    </div>

                    <p
                        v-else-if="list.errorMessage.value && list.items.value.length === 0"
                        class="m-4 rounded-md bg-red-50 p-3 text-sm text-red-700"
                    >
                        {{ list.errorMessage.value }}
                    </p>

                    <p
                        v-else-if="list.items.value.length === 0"
                        class="py-10 text-center text-sm text-gray-500"
                    >
                        {{ emptyMessage }}
                    </p>

                    <div v-else class="overflow-x-auto">
                        <table class="min-w-full divide-y divide-gray-200 text-sm">
                            <thead
                                class="bg-gray-50 text-left text-xs font-semibold uppercase tracking-wide text-gray-500"
                            >
                                <tr>
                                    <th scope="col" class="px-4 py-3">
                                        {{ t("graph.entities.column.name") }}
                                    </th>
                                    <th scope="col" class="px-4 py-3">
                                        {{ t("graph.form.key") }}
                                    </th>
                                    <th scope="col" class="px-4 py-3">
                                        <span class="sr-only">{{ t("graph.form.actions") }}</span>
                                    </th>
                                </tr>
                            </thead>
                            <tbody class="divide-y divide-gray-200">
                                <tr v-for="entity in list.items.value" :key="entity.id">
                                    <td class="px-4 py-3 font-medium text-gray-900">
                                        {{ entity.title || entity.id }}
                                    </td>
                                    <td class="px-4 py-3 font-mono text-xs text-gray-500">
                                        {{ entity.id }}
                                    </td>
                                    <td class="whitespace-nowrap px-4 py-3 text-right">
                                        <RouterLink
                                            :to="{
                                                name: 'graph-explore',
                                                query: { type: entity.type, id: entity.id },
                                            }"
                                            class="font-medium text-indigo-600 hover:text-indigo-500"
                                        >
                                            {{ t("graph.entities.show_on_graph") }}
                                        </RouterLink>
                                    </td>
                                </tr>
                            </tbody>
                        </table>
                    </div>

                    <div
                        v-if="!list.loading.value && list.items.value.length > 0"
                        class="flex flex-wrap items-center justify-between gap-3 border-t border-gray-200 px-4 py-3 text-sm text-gray-500"
                    >
                        <span>
                            {{ t("graph.entities.shown", { count: list.items.value.length }) }}
                        </span>
                        <span v-if="list.errorMessage.value" class="text-red-700">
                            {{ list.errorMessage.value }}
                        </span>
                        <BaseButton
                            v-if="list.next.value"
                            size="small"
                            variant="secondary"
                            :is-loading="list.loadingMore.value"
                            @click="list.loadMore()"
                        >
                            {{ t("graph.entities.load_more") }}
                        </BaseButton>
                    </div>
                </section>
            </template>
        </div>
    </div>
</template>

<script setup>
import { ref, computed, watch, onMounted, onBeforeUnmount } from "vue";
import { RouterLink, useRoute, useRouter } from "vue-router";
import { MagnifyingGlassIcon, XCircleIcon } from "@heroicons/vue/20/solid";
import { ListBulletIcon } from "@heroicons/vue/24/outline";
import { t } from "@/i18n/index.js";
import { useOntology } from "@/composables/graph/useOntology.js";
import { useEntityList } from "@/composables/graph/useEntityList.js";
import BaseButton from "@/components/BaseButton.vue";
import BaseSpinner from "@/components/BaseSpinner.vue";

const SEARCH_DELAY = 400;

const route = useRoute();
const router = useRouter();
const ontology = useOntology();
const list = useEntityList();

const loaded = ref(false);
const search = ref("");
const searched = ref("");

let searchTimer = null;

// Only types read from a dataset have entities to list; the built-in ones
// are Twigex's own charts, dashboards, datasets and fields.
const datasetTypes = computed(() => {
    return ontology.sortedEntityTypes.value.filter((entityType) => {
        return entityType.source === "dataset";
    });
});

const currentKey = computed(() => route.params.key ?? "");

const currentType = computed(() => {
    return datasetTypes.value.find((entityType) => entityType.key === currentKey.value);
});

const emptyMessage = computed(() => {
    if (searched.value) {
        return t.value("graph.entities.none_found");
    }

    return t.value("graph.entities.none");
});

function startList() {
    clearTimeout(searchTimer);

    if (!currentType.value) {
        list.cancel();

        return;
    }

    searched.value = search.value.trim();

    list.start(currentType.value.key, searched.value);
}

function scheduleSearch() {
    clearTimeout(searchTimer);

    searchTimer = setTimeout(startList, SEARCH_DELAY);
}

// Without a type in the address the first one opens, so the section never
// starts empty.
function openFirstType() {
    if (currentKey.value || datasetTypes.value.length === 0) {
        return false;
    }

    router.replace({ name: "graph-entities", params: { key: datasetTypes.value[0].key } });

    return true;
}

watch(currentKey, () => {
    if (!loaded.value || openFirstType()) {
        return;
    }

    search.value = "";

    startList();
});

onMounted(async () => {
    await ontology.graphTypes.ready;

    loaded.value = true;

    if (openFirstType()) {
        return;
    }

    startList();
});

onBeforeUnmount(() => {
    clearTimeout(searchTimer);
    list.cancel();
});
</script>
