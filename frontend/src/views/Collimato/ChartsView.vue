<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div class="flex h-full flex-col overflow-hidden">
        <!-- Loading -->
        <div v-if="!loaded" class="flex h-full items-center justify-center">
            <div
                class="loader h-8 w-8 rounded-full border-4 border-indigo-500 border-t-transparent animate-spin"
            ></div>
        </div>

        <template v-else>
            <!-- Header -->
            <div class="shrink-0 flex items-center justify-between border-b px-6 py-4">
                <div>
                    <h1 class="text-base font-semibold leading-6 text-gray-900">
                        {{ t("collimato.charts.table.title") }}
                    </h1>
                    <p class="mt-1 text-sm text-gray-500">
                        {{ t("collimato.charts.table.description") }}
                    </p>
                </div>
                <button
                    v-if="collimatoStore.hasPermissionToCreateCharts"
                    type="button"
                    @click="router.push({ name: 'new-chart' })"
                    class="inline-flex items-center gap-x-1.5 rounded-md bg-indigo-600 px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-indigo-500 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600"
                >
                    <PlusIcon class="-ml-0.5 h-4 w-4" aria-hidden="true" />
                    {{ t("collimato.charts.button.new_chart") }}
                </button>
            </div>

            <!-- Empty state (no charts at all) -->
            <div
                v-if="charts.length === 0"
                class="flex flex-1 flex-col items-center justify-center"
            >
                <ChartBarIconOutline class="mx-auto h-16 w-16 text-gray-300" aria-hidden="true" />
                <h3 class="mt-3 text-sm font-semibold text-gray-900">
                    {{ t("collimato.charts.empty.title") }}
                </h3>
                <p class="mt-1 text-sm text-gray-500">
                    {{ t("collimato.charts.empty.description") }}
                </p>
                <button
                    v-if="collimatoStore.hasPermissionToCreateCharts"
                    type="button"
                    @click="router.push({ name: 'new-chart' })"
                    class="mt-6 inline-flex items-center gap-x-1.5 rounded-md bg-indigo-600 px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-indigo-500 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600"
                >
                    <PlusIcon class="-ml-0.5 h-4 w-4" aria-hidden="true" />
                    {{ t("collimato.charts.button.new_chart") }}
                </button>
            </div>

            <template v-else>
                <!-- Search + type filter bar -->
                <div class="shrink-0 flex flex-wrap items-center gap-3 border-b px-6 py-3">
                    <!-- Search input -->
                    <div class="relative min-w-48 flex-1 sm:max-w-xs">
                        <MagnifyingGlassIcon
                            class="pointer-events-none absolute left-2.5 top-1/2 -translate-y-1/2 h-4 w-4 text-gray-400"
                            aria-hidden="true"
                        />
                        <input
                            v-model="searchQuery"
                            type="search"
                            :placeholder="t('collimato.charts.search_placeholder')"
                            class="block w-full rounded-md border-0 py-1.5 pl-8 pr-3 text-sm text-gray-900 ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600"
                        />
                    </div>

                    <!-- Type filter chips -->
                    <div class="flex flex-wrap gap-1.5">
                        <button
                            type="button"
                            @click="selectedType = null"
                            :class="[
                                selectedType === null
                                    ? 'bg-indigo-600 text-white'
                                    : 'bg-white text-gray-600 ring-1 ring-inset ring-gray-300 hover:bg-gray-50',
                                'rounded-full px-3 py-1 text-xs font-medium',
                            ]"
                        >
                            {{ t("collimato.charts.filter.all") }}
                        </button>
                        <button
                            v-for="type in availableTypes"
                            :key="type"
                            type="button"
                            @click="selectedType = selectedType === type ? null : type"
                            :class="[
                                selectedType === type
                                    ? getTypeConfig(type).activePill
                                    : 'bg-white text-gray-600 ring-1 ring-inset ring-gray-300 hover:bg-gray-50',
                                'inline-flex items-center gap-x-1 rounded-full px-3 py-1 text-xs font-medium',
                            ]"
                        >
                            <component
                                :is="getTypeConfig(type).icon"
                                class="h-3 w-3"
                                aria-hidden="true"
                            />
                            {{ getTypeConfig(type).label }}
                        </button>
                    </div>
                </div>

                <!-- Card grid -->
                <div class="flex-1 overflow-y-auto p-6">
                    <!-- No search results -->
                    <div
                        v-if="filteredCharts.length === 0"
                        class="flex h-full flex-col items-center justify-center"
                    >
                        <MagnifyingGlassIcon class="h-10 w-10 text-gray-300" aria-hidden="true" />
                        <p class="mt-3 text-sm font-medium text-gray-900">
                            {{ t("collimato.charts.search_no_results") }}
                        </p>
                        <button
                            type="button"
                            @click="
                                searchQuery = '';
                                selectedType = null;
                            "
                            class="mt-3 text-sm text-indigo-600 hover:text-indigo-500"
                        >
                            {{ t("collimato.charts.search_clear") }}
                        </button>
                    </div>

                    <ul
                        v-else
                        role="list"
                        class="grid grid-cols-[repeat(auto-fill,minmax(18rem,1fr))] gap-4 content-start"
                    >
                        <ItemCard
                            v-for="chart in filteredCharts"
                            :key="chart.id"
                            :title="chart.name"
                            @click="openChart(chart)"
                        >
                            <template #icon>
                                <span
                                    :class="[
                                        getTypeConfig(chart.chart_type).bg,
                                        'flex h-10 w-10 shrink-0 items-center justify-center rounded-lg',
                                    ]"
                                    aria-hidden="true"
                                >
                                    <component
                                        :is="getTypeConfig(chart.chart_type).icon"
                                        class="h-5 w-5 text-white"
                                    />
                                </span>
                            </template>
                            <template #subtitle>
                                <span
                                    :class="[
                                        getTypeConfig(chart.chart_type).badge,
                                        'inline-flex items-center rounded-md px-1.5 py-0.5 text-xs font-medium ring-1 ring-inset',
                                    ]"
                                >
                                    {{ getTypeConfig(chart.chart_type).label }}
                                </span>
                            </template>
                            <template
                                v-if="
                                    collimatoStore.hasPermissionToEditCharts ||
                                    collimatoStore.hasPermissionToDeleteCharts
                                "
                                #actions
                            >
                                <button
                                    v-if="collimatoStore.hasPermissionToEditCharts"
                                    @click="
                                        router.push({
                                            name: 'edit-chart',
                                            params: { id: chart.id },
                                        })
                                    "
                                    type="button"
                                    class="rounded p-1 text-gray-400 hover:bg-gray-100 hover:text-indigo-600"
                                    :title="t('common.button.edit')"
                                >
                                    <PencilIcon class="h-4 w-4" aria-hidden="true" />
                                </button>
                                <button
                                    v-if="collimatoStore.hasPermissionToDeleteCharts"
                                    @click="openDeleteDialog(chart)"
                                    type="button"
                                    class="rounded p-1 text-gray-400 hover:bg-red-50 hover:text-red-600"
                                    :title="t('common.button.delete')"
                                >
                                    <TrashIcon class="h-4 w-4" aria-hidden="true" />
                                </button>
                            </template>
                            <template #footer>
                                <UserAvatarWithText
                                    class="gap-x-2"
                                    :user-id="chart.owner_id"
                                    avatar-class="size-6 shrink-0"
                                    text-class="truncate text-xs text-gray-500"
                                />
                                <span class="shrink-0 text-xs text-gray-400">
                                    {{ getDateAndTime(chart.updated_at) }}
                                </span>
                            </template>
                        </ItemCard>
                    </ul>
                </div>
            </template>
        </template>

        <ChartDeleteDialog
            v-model="deleteDialog"
            @update:modelValue="deleteDialog = $event"
            @delete="deleteChart"
        />
    </div>
</template>

<script setup>
import { ref, computed, onMounted } from "vue";
import { useRouter, useRoute } from "vue-router";
import { t } from "@/i18n/index.js";
import collimatoService from "@/services/collimatoService.js";
import useDateOperations from "@/composables/useDateOperations.js";
import { useAlertStore } from "@/store/alerts";
import { extractErrorMessage } from "@/utils/errors";
import { useUsers } from "@/composables/useUser";
import { useCollimatoStore } from "@/store/collimato";
import ChartDeleteDialog from "@/components/Collimato/Dialogs/ChartDeleteDialog.vue";
import ItemCard from "@/components/ItemCard.vue";
import UserAvatarWithText from "@/components/UserAvatarWithText.vue";
import { PlusIcon, PencilIcon, TrashIcon, MagnifyingGlassIcon } from "@heroicons/vue/20/solid";
import { ChartBarIcon as ChartBarIconOutline } from "@heroicons/vue/24/outline";
import { CHART_TYPE_IDS } from "@/utils/collimato/chartTypes.js";
import {
    presentationFor,
    FALLBACK_PRESENTATION,
} from "@/components/Collimato/Charts/chartTypePresentation.js";

const router = useRouter();
const route = useRoute();
const alertStore = useAlertStore();
const collimatoStore = useCollimatoStore();
const { getDateAndTime } = useDateOperations();

const loaded = ref(false);
const charts = ref([]);

// Trigger a lazy load so owner names resolve in the list.
useUsers(() => charts.value.map((c) => c.owner_id));
const deleteDialog = ref(false);
const chartToDelete = ref(null);
const searchQuery = ref("");
const selectedType = ref(null);

const typeConfigs = computed(() =>
    Object.fromEntries(
        CHART_TYPE_IDS.map((id) => {
            const presentation = presentationFor(id);

            return [
                id,
                {
                    icon: presentation.icon,
                    label: t.value(presentation.shortLabelKey),
                    ...presentation.listStyle,
                },
            ];
        }),
    ),
);

function getTypeConfig(type) {
    return (
        typeConfigs.value[type] ?? {
            icon: FALLBACK_PRESENTATION.icon,
            label: type,
            ...FALLBACK_PRESENTATION.listStyle,
        }
    );
}

const availableTypes = computed(() =>
    Object.keys(typeConfigs.value).filter((type) =>
        charts.value.some((c) => c.chart_type === type),
    ),
);

const filteredCharts = computed(() => {
    const query = searchQuery.value.trim().toLowerCase();

    return charts.value.filter((chart) => {
        const matchesSearch = !query || chart.name.toLowerCase().includes(query);
        const matchesType = !selectedType.value || chart.chart_type === selectedType.value;

        return matchesSearch && matchesType;
    });
});

function openChart(chart) {
    router.push({ name: "edit-chart", params: { id: chart.id } });
}

function openDeleteDialog(item) {
    chartToDelete.value = item;
    deleteDialog.value = true;
}

function deleteChart() {
    collimatoService
        .deleteChart(route.params.workspaceId, chartToDelete.value.id)
        .then(() => {
            charts.value = charts.value.filter((c) => c.id !== chartToDelete.value.id);
            chartToDelete.value = null;
        })
        .catch((error) => {
            alertStore.showError(extractErrorMessage(error));
        });
    deleteDialog.value = false;
}

onMounted(() => {
    if (!collimatoStore.hasPermissionToViewCharts) {
        alertStore.showError(t.value("collimato.charts.error.no_permission_view"));

        return;
    }

    collimatoService.getCharts(route.params.workspaceId).then((response) => {
        charts.value = response.data;
        loaded.value = true;
    });
});
</script>
