<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div v-if="open" class="flex flex-col w-64 shrink-0 border-r overflow-hidden">
        <div class="shrink-0 flex items-center justify-between px-3 py-2 border-b">
            <span class="text-xs font-semibold uppercase tracking-wide text-gray-400">
                {{ t("collimato.dashboard.filters.title") }}
            </span>
            <div class="flex items-center">
                <button
                    v-if="collimatoStore.hasPermissionToCreateDashboardFilters"
                    type="button"
                    @click="emit('update:dialogOpen', true)"
                    class="rounded p-1 text-gray-400 hover:bg-gray-100 hover:text-gray-600"
                    :title="t('collimato.dashboard.filters.add')"
                >
                    <PlusIcon class="h-4 w-4" aria-hidden="true" />
                </button>
                <button
                    type="button"
                    class="rounded p-1 text-gray-400 hover:bg-gray-100 hover:text-gray-600"
                    @click="emit('toggleSidebar')"
                >
                    <Bars3Icon class="h-4 w-4" aria-hidden="true" />
                </button>
            </div>
        </div>

        <CrossFilterDialog
            :model-value="dialogOpen"
            :data-models="dataModels"
            :charts="charts"
            :edit-filter="filterToEdit"
            @update:model-value="emit('update:dialogOpen', $event)"
            @add-filter="emit('addFilter', $event)"
            @update-filter="emit('updateFilter', $event)"
            @close="emit('closeDialog')"
        />

        <div class="flex-1 overflow-y-auto divide-y divide-gray-100">
            <div
                v-if="filters.length === 0"
                class="flex flex-col items-center justify-center py-10 px-4 text-center"
            >
                <FunnelIcon class="h-8 w-8 text-gray-300" aria-hidden="true" />
                <p class="mt-2 text-xs text-gray-400">
                    {{ t("collimato.dashboard.filters.empty") }}
                </p>
            </div>

            <div v-for="filter in filters" :key="filter.id">
                <div class="flex items-center gap-x-0.5 pr-1">
                    <button
                        type="button"
                        @click="emit('toggleFilter', filter.id)"
                        class="flex-1 flex items-center gap-x-2 px-3 py-2.5 text-left min-w-0 hover:bg-gray-50"
                    >
                        <span class="flex-1 truncate text-sm font-medium text-gray-900">{{
                            filter.name
                        }}</span>
                        <div class="flex items-center gap-x-1 shrink-0">
                            <span
                                v-if="(localSelections[filter.id] ?? []).length > 0"
                                class="inline-flex items-center rounded-full bg-indigo-100 px-1.5 py-0.5 text-xs font-semibold text-indigo-700"
                            >
                                {{ (localSelections[filter.id] ?? []).length }}
                            </span>
                            <ChevronDownIcon
                                class="h-4 w-4 text-gray-400 transition-transform duration-150"
                                :class="openFilters[filter.id] ? 'rotate-180' : ''"
                                aria-hidden="true"
                            />
                        </div>
                    </button>
                    <Menu
                        v-if="
                            collimatoStore.hasPermissionToEditDashboardFilters ||
                            collimatoStore.hasPermissionToDeleteDashboardFilters
                        "
                        as="div"
                        class="relative shrink-0"
                    >
                        <MenuButton
                            class="rounded p-1 text-gray-400 hover:bg-gray-100 hover:text-gray-600"
                        >
                            <EllipsisVerticalIcon class="h-4 w-4" aria-hidden="true" />
                        </MenuButton>
                        <transition
                            enter-active-class="transition ease-out duration-100"
                            enter-from-class="transform opacity-0 scale-95"
                            enter-to-class="transform opacity-100 scale-100"
                            leave-active-class="transition ease-in duration-75"
                            leave-from-class="transform opacity-100 scale-100"
                            leave-to-class="transform opacity-0 scale-95"
                        >
                            <MenuItems
                                class="absolute right-0 z-10 mt-1 w-32 origin-top-right rounded-md bg-white py-1 shadow-lg ring-1 ring-black/5 focus:outline-none"
                            >
                                <MenuItem
                                    v-if="collimatoStore.hasPermissionToEditDashboardFilters"
                                    v-slot="{ active }"
                                >
                                    <button
                                        type="button"
                                        @click="emit('editFilter', filter)"
                                        :class="[
                                            active ? 'bg-gray-50' : '',
                                            'flex w-full items-center px-3 py-1.5 text-sm text-gray-700',
                                        ]"
                                    >
                                        {{ t("common.button.edit") }}
                                    </button>
                                </MenuItem>
                                <MenuItem
                                    v-if="collimatoStore.hasPermissionToDeleteDashboardFilters"
                                    v-slot="{ active }"
                                >
                                    <button
                                        type="button"
                                        @click="emit('deleteFilter', filter)"
                                        :class="[
                                            active ? 'bg-gray-50' : '',
                                            'flex w-full items-center px-3 py-1.5 text-sm text-red-600',
                                        ]"
                                    >
                                        {{ t("common.button.delete") }}
                                    </button>
                                </MenuItem>
                            </MenuItems>
                        </transition>
                    </Menu>
                </div>

                <div v-if="openFilters[filter.id]" class="px-3 pb-3 bg-gray-50">
                    <div class="relative mt-2">
                        <MagnifyingGlassIcon
                            class="pointer-events-none absolute left-2 top-1/2 -translate-y-1/2 h-3.5 w-3.5 text-gray-400"
                            aria-hidden="true"
                        />
                        <input
                            v-model="filterSearch[filter.id]"
                            type="search"
                            :placeholder="t('collimato.dashboard.filters.search_placeholder')"
                            class="block w-full rounded border-0 py-1 pl-6 pr-2 text-xs text-gray-900 ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600"
                            @click.stop
                        />
                    </div>

                    <div v-if="!filterValues[filter.id]" class="mt-2 py-3 text-center">
                        <span class="text-xs text-gray-400">{{
                            t("collimato.dashboard.filters.loading")
                        }}</span>
                    </div>
                    <div
                        v-else-if="getFilteredValues(filter).length === 0"
                        class="mt-2 py-3 text-center"
                    >
                        <span class="text-xs text-gray-400">{{
                            t("collimato.dashboard.filters.no_values")
                        }}</span>
                    </div>
                    <div v-else class="mt-1.5 max-h-44 overflow-y-auto space-y-0.5">
                        <label
                            v-for="val in getFilteredValues(filter)"
                            :key="val"
                            class="flex items-center gap-x-2 rounded px-1 py-1 hover:bg-white cursor-pointer"
                        >
                            <input
                                type="checkbox"
                                :checked="(localSelections[filter.id] ?? []).includes(String(val))"
                                @change="emit('toggleSelection', filter.id, String(val))"
                                class="h-3.5 w-3.5 shrink-0 rounded border-gray-300 text-indigo-600 focus:ring-indigo-600"
                                @click.stop
                            />
                            <span class="text-xs text-gray-700 truncate">{{ val }}</span>
                        </label>
                    </div>

                    <button
                        v-if="(localSelections[filter.id] ?? []).length > 0"
                        type="button"
                        @click="emit('clearFilter', filter.id)"
                        class="mt-2 text-xs text-indigo-600 hover:text-indigo-500"
                    >
                        {{ t("collimato.dashboard.filters.clear") }}
                    </button>
                </div>
            </div>
        </div>

        <div v-if="filters.length > 0" class="shrink-0 border-t px-3 py-3">
            <button
                type="button"
                @click="emit('apply')"
                class="w-full rounded-md bg-indigo-600 px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-indigo-500 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600"
            >
                {{ t("collimato.dashboard.filters.apply") }}
            </button>
        </div>
    </div>

    <div v-else class="flex flex-col items-center shrink-0 w-10 border-r pt-2">
        <button
            type="button"
            class="rounded p-1 text-gray-400 hover:bg-gray-100 hover:text-gray-600"
            @click="emit('toggleSidebar')"
        >
            <Bars3Icon class="h-4 w-4" aria-hidden="true" />
        </button>
    </div>
</template>

<script setup>
import { t } from "@/i18n/index.js";
import CrossFilterDialog from "@/components/Collimato/Dialogs/CrossFilterDialog.vue";
import { useCollimatoStore } from "@/store/collimato";
import { Menu, MenuButton, MenuItem, MenuItems } from "@headlessui/vue";
import { EllipsisVerticalIcon, PlusIcon, MagnifyingGlassIcon } from "@heroicons/vue/20/solid";
import { Bars3Icon, ChevronDownIcon, FunnelIcon } from "@heroicons/vue/24/outline";

const props = defineProps({
    open: {
        type: Boolean,
        default: true,
    },
    filters: {
        type: Array,
        default: () => [],
    },
    filterValues: {
        type: Object,
        default: () => ({}),
    },
    localSelections: {
        type: Object,
        default: () => ({}),
    },
    openFilters: {
        type: Object,
        default: () => ({}),
    },
    filterSearch: {
        type: Object,
        default: () => ({}),
    },
    dialogOpen: {
        type: Boolean,
        default: false,
    },
    filterToEdit: {
        type: Object,
        default: null,
    },
    dataModels: {
        type: Array,
        default: () => [],
    },
    charts: {
        type: Array,
        default: () => [],
    },
});

const emit = defineEmits([
    "update:dialogOpen",
    "toggleSidebar",
    "addFilter",
    "updateFilter",
    "closeDialog",
    "toggleFilter",
    "editFilter",
    "deleteFilter",
    "toggleSelection",
    "clearFilter",
    "apply",
]);

const collimatoStore = useCollimatoStore();

function getFilteredValues(filter) {
    const all = props.filterValues[filter.id] ?? [];
    const search = (props.filterSearch[filter.id] ?? "").toLowerCase().trim();

    if (!search) return all;

    return all.filter((v) => String(v).toLowerCase().includes(search));
}
</script>
