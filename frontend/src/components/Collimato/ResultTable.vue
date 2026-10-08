<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div class="flex flex-col h-full w-full min-w-0">
        <!-- Header slot -->
        <div class="sm:flex sm:items-center">
            <div class="sm:flex-auto">
                <slot name="header" />
            </div>
        </div>

        <!-- Single scroll port: clamps the wide table's width to this box so it
             cannot expand the mobile layout viewport; sticky thead anchors here -->
        <div class="mt-2 flex-1 min-h-0 overflow-auto">
            <table class="min-w-full divide-y divide-gray-300">
                <thead class="bg-gray-100 sticky top-0 z-10">
                    <tr>
                        <th
                            v-for="header in headers"
                            :key="header.key"
                            scope="col"
                            class="px-3 text-left text-sm font-semibold text-gray-900 bg-gray-100 whitespace-nowrap"
                            :class="dense ? 'py-2' : 'py-3.5'"
                        >
                            {{ header.label }}
                        </th>
                    </tr>
                </thead>
                <tbody class="divide-y divide-gray-200">
                    <tr
                        v-for="item in itemList"
                        :key="item.id"
                        class="hover:bg-gray-50 hover:cursor-pointer"
                        @click="onRowClick(item)"
                    >
                        <td
                            v-for="header in headers"
                            :key="header.key"
                            class="whitespace-nowrap px-3 text-sm text-gray-500"
                            :class="dense ? 'py-2' : 'py-4'"
                        >
                            <slot :name="header.key" :item="item">
                                {{ getNestedProperty(item, header.key) }}
                            </slot>
                        </td>
                    </tr>
                </tbody>
            </table>
        </div>

        <!-- Pagination -->
        <BasePagination
            class="mt-2 px-1"
            :currentPage="page"
            :totalItems="totalCount"
            :itemsPerPage="itemsPerPage"
            @update:currentPage="updateCurrentPage"
        />
    </div>
</template>

<script setup>
import { computed, ref } from "vue";
import BasePagination from "@/components/BasePagination.vue";

const props = defineProps({
    headers: {
        type: Array,
        default: () => [],
    },
    items: {
        type: Array,
        required: true,
    },
    dense: {
        type: Boolean,
        default: false,
    },
    itemsPerPage: {
        type: Number,
        default: 20,
    },
    // Set to page server side: items is then the page, not the whole list.
    totalItems: {
        type: Number,
        default: null,
    },
    // Set to let the parent own the page, so it can reset to 1 on a new query.
    currentPage: {
        type: Number,
        default: null,
    },
});

const emits = defineEmits(["row-click", "update:currentPage"]);

const serverPaged = computed(() => props.totalItems !== null);

const totalCount = computed(() => (serverPaged.value ? props.totalItems : props.items.length));

const internalPage = ref(1);
const page = computed(() => props.currentPage ?? internalPage.value);

const itemList = computed(() => {
    if (serverPaged.value) return props.items;

    return props.items.slice(
        (page.value - 1) * props.itemsPerPage,
        page.value * props.itemsPerPage,
    );
});

const updateCurrentPage = (newPage) => {
    internalPage.value = newPage;
    emits("update:currentPage", newPage);
};

const onRowClick = (item) => {
    emits("row-click", item);
};

function getNestedProperty(obj, path) {
    if (path == undefined) {
        return "";
    }

    return path.split(".").reduce((o, p) => (o ? o[p] : undefined), obj);
}
</script>
