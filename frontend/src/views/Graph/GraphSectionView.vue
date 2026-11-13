<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <SidebarLayout :title="t('graph.title')">
        <template #sidebar="{ mini }">
            <GraphNavigation
                v-if="!mini"
                :entity-type-count="types.entity_types.length"
                :relation-type-count="types.relation_types.length"
            />
            <div v-else class="flex justify-center pt-2">
                <ShareIcon class="size-5 text-indigo-600" aria-hidden="true" />
            </div>
        </template>

        <RouterView />
    </SidebarLayout>
</template>

<script setup>
import { RouterView } from "vue-router";
import { ShareIcon } from "@heroicons/vue/24/outline";
import { t } from "@/i18n/index.js";
import { provideGraphTypes } from "@/composables/graph/useGraphTypes.js";
import SidebarLayout from "@/components/Navigation/SidebarLayout.vue";
import GraphNavigation from "@/components/Graph/GraphNavigation.vue";

const { types } = provideGraphTypes();
</script>
