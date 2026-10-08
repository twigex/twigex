<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div class="overflow-y-auto">
        <ul role="list" class="space-y-1">
            <li v-for="db in tables" :key="db.name">
                <Disclosure v-slot="{ open }">
                    <DisclosureButton
                        class="flex w-full items-center gap-x-2 rounded-md p-2 text-sm font-semibold text-gray-700 hover:bg-gray-50"
                    >
                        <ChevronRightIcon
                            :class="[
                                open ? 'rotate-90' : '',
                                'h-4 w-4 shrink-0 text-gray-400 transition-transform',
                            ]"
                            aria-hidden="true"
                        />
                        <span class="truncate">{{ db.name }}</span>
                    </DisclosureButton>
                    <DisclosurePanel as="ul" class="ml-4 mt-1 space-y-1">
                        <li v-for="table in db.tables" :key="table.name">
                            <Disclosure v-slot="{ open: tableOpen }">
                                <DisclosureButton
                                    class="flex w-full items-center gap-x-2 rounded-md p-2 text-sm font-semibold text-gray-700 hover:bg-gray-50"
                                >
                                    <ChevronRightIcon
                                        :class="[
                                            tableOpen ? 'rotate-90' : '',
                                            'h-4 w-4 shrink-0 text-gray-400 transition-transform',
                                        ]"
                                        aria-hidden="true"
                                    />
                                    <input
                                        v-model="table.checked"
                                        type="checkbox"
                                        class="h-4 w-4 rounded border-gray-300 text-indigo-600 focus:ring-indigo-600"
                                        @click.stop
                                        @change="selectAll(db, table)"
                                    />
                                    <span class="truncate">{{ table.name }}</span>
                                </DisclosureButton>
                                <DisclosurePanel as="ul" class="ml-8 mt-1 space-y-1">
                                    <li
                                        v-for="column in table.columns"
                                        :key="column.name"
                                        class="flex items-center gap-x-2 rounded-md px-2 py-1.5"
                                    >
                                        <input
                                            v-model="column.checked"
                                            type="checkbox"
                                            class="h-4 w-4 rounded border-gray-300 text-indigo-600 focus:ring-indigo-600"
                                            @change="
                                                column.checked
                                                    ? add(db, table, column)
                                                    : remove(db, table, column)
                                            "
                                        />
                                        <span class="flex-1 truncate text-sm text-gray-600">{{
                                            column.name
                                        }}</span>
                                        <span
                                            class="shrink-0 inline-flex items-center rounded-md bg-gray-100 px-1.5 py-0.5 text-xs font-medium text-gray-500"
                                        >
                                            {{ column.type }}
                                        </span>
                                    </li>
                                </DisclosurePanel>
                            </Disclosure>
                        </li>
                    </DisclosurePanel>
                </Disclosure>
            </li>
        </ul>
    </div>
</template>

<script setup>
import { ref } from "vue";
import { Disclosure, DisclosureButton, DisclosurePanel } from "@headlessui/vue";
import { ChevronRightIcon } from "@heroicons/vue/20/solid";

defineProps({
    tables: {
        type: Array,
        default: () => [],
    },
});

const emit = defineEmits(["add"]);

const selectedTables = ref([]);

function add(db, table, column) {
    if (
        !selectedTables.value.some(
            (item) =>
                item.database === db.database &&
                item.table === table.name &&
                item.column === column.name,
        )
    ) {
        selectedTables.value.push({
            database: db.database,
            table: table.name,
            column: column.name,
            type: column.type,
        });
    }

    emit("add", selectedTables.value);
}

function remove(db, table, column) {
    selectedTables.value = selectedTables.value.filter(
        (item) =>
            item.database !== db.database ||
            item.table !== table.name ||
            item.column !== column.name,
    );
    emit("add", selectedTables.value);
}

function selectAll(db, table) {
    if (table.checked) {
        table.columns.forEach((column) => {
            column.checked = true;
            if (
                !selectedTables.value.some(
                    (item) =>
                        item.database === db.database &&
                        item.table === table.name &&
                        item.column === column.name,
                )
            ) {
                selectedTables.value.push({
                    database: db.database,
                    table: table.name,
                    column: column.name,
                    type: column.type,
                });
            }
        });
    } else {
        table.columns.forEach((column) => {
            column.checked = false;
        });
        selectedTables.value = selectedTables.value.filter(
            (item) => item.database !== db.database || item.table !== table.name,
        );
    }

    emit("add", selectedTables.value);
}
</script>
