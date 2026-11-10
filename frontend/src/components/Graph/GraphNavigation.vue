<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div class="flex h-full flex-col">
        <nav class="flex flex-1 flex-col pt-3" :aria-label="t('graph.title')">
            <ul role="list" class="space-y-1">
                <li>
                    <RouterLink
                        :to="{ name: 'graph-explore' }"
                        :aria-current="isExplore ? 'page' : undefined"
                        :class="[
                            isExplore
                                ? 'bg-gray-50 text-indigo-600'
                                : 'text-gray-700 hover:bg-gray-50 hover:text-indigo-600',
                            'group flex gap-x-3 rounded-md p-2 text-sm font-semibold leading-6',
                        ]"
                    >
                        <ShareIcon
                            :class="[
                                isExplore
                                    ? 'text-indigo-600'
                                    : 'text-gray-400 group-hover:text-indigo-600',
                                'h-6 w-6 shrink-0',
                            ]"
                            aria-hidden="true"
                        />
                        {{ t("graph.navigation.explore") }}
                    </RouterLink>
                </li>
                <li>
                    <RouterLink
                        :to="{ name: 'graph-ontology' }"
                        :aria-current="isOntology ? 'page' : undefined"
                        :class="[
                            isOntology
                                ? 'bg-gray-50 text-indigo-600'
                                : 'text-gray-700 hover:bg-gray-50 hover:text-indigo-600',
                            'group flex gap-x-3 rounded-md p-2 text-sm font-semibold leading-6',
                        ]"
                    >
                        <Squares2X2Icon
                            :class="[
                                isOntology
                                    ? 'text-indigo-600'
                                    : 'text-gray-400 group-hover:text-indigo-600',
                                'h-6 w-6 shrink-0',
                            ]"
                            aria-hidden="true"
                        />
                        {{ t("graph.navigation.ontology") }}
                    </RouterLink>
                </li>
            </ul>
        </nav>

        <RouterLink
            :to="{ name: 'graph-ontology' }"
            class="mt-4 block rounded-lg p-3 text-sm text-gray-500 ring-1 ring-inset ring-gray-200 hover:bg-gray-50"
        >
            <p class="font-semibold text-gray-900">{{ t("graph.navigation.types_title") }}</p>
            <p class="mt-1">
                {{
                    t("graph.navigation.types_count", {
                        types: entityTypeCount,
                        relations: relationTypeCount,
                    })
                }}
            </p>
        </RouterLink>
    </div>
</template>

<script setup>
import { computed } from "vue";
import { RouterLink, useRoute } from "vue-router";
import { ShareIcon, Squares2X2Icon } from "@heroicons/vue/24/outline";
import { t } from "@/i18n/index.js";

defineProps({
    entityTypeCount: {
        type: Number,
        required: true,
    },
    relationTypeCount: {
        type: Number,
        required: true,
    },
});

const route = useRoute();

const isExplore = computed(() => route.name === "graph-explore");
const isOntology = computed(() => {
    return route.name === "graph-ontology" || route.name === "graph-ontology-type";
});
</script>
