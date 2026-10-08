<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <DataTable
        :columns="allColumns"
        :items="members"
        item-key="user_id"
        :sort="sort"
        :loading="loading"
        :itemsPerPage="itemsPerPage"
        :totalItems="totalItems"
        :currentPage="currentPage"
        @update:sort="emit('update:sort', $event)"
        @update:currentPage="emit('update:currentPage', $event)"
    >
        <template v-for="name in forwardedSlots" #[name]="scope">
            <slot :name="name" v-bind="scope" />
        </template>

        <template #name="{ item }">
            <div class="flex items-center">
                <div class="size-9 shrink-0">
                    <UserAvatar :user="item.user_info" :user-id="item.user_id" :status="status" />
                </div>
                <div class="ml-4 min-w-0">
                    <div class="flex items-center gap-x-2">
                        <span class="truncate font-medium text-gray-900">
                            {{ fullName(item.user_info) }}
                        </span>
                        <slot name="badges" :item="item" />
                    </div>
                    <div class="mt-1 truncate text-gray-500">{{ item.user_info?.email }}</div>
                </div>
            </div>
        </template>

        <template v-if="$slots.actions" #actions="{ item }">
            <div class="flex items-center justify-end gap-x-4 font-medium">
                <slot name="actions" :item="item" />
            </div>
        </template>
    </DataTable>
</template>

<script setup>
import { t } from "@/i18n/index";

import { computed, useSlots } from "vue";
import DataTable from "@/components/DataTable.vue";
import UserAvatar from "@/components/UserAvatar.vue";

const props = defineProps({
    members: {
        type: Array,
        required: true,
    },
    columns: {
        type: Array,
        default: () => [],
    },
    status: {
        type: Boolean,
        default: false,
    },
    sort: {
        type: Object,
        default: null,
    },
    loading: {
        type: Boolean,
        default: false,
    },
    itemsPerPage: {
        type: Number,
        default: 20,
    },
    totalItems: {
        type: Number,
        default: null,
    },
    currentPage: {
        type: Number,
        default: null,
    },
});

const emit = defineEmits(["update:sort", "update:currentPage"]);

const slots = useSlots();

const OWN_SLOTS = ["name", "badges", "actions"];

const forwardedSlots = computed(() => Object.keys(slots).filter((n) => !OWN_SLOTS.includes(n)));

const allColumns = computed(() => {
    const columns = [
        {
            key: "name",
            label: t.value("members_table.member"),
            sortable: true,
            sortKey: "user_info.name",
        },
        ...props.columns,
    ];

    if (slots.actions) {
        columns.push({
            key: "actions",
            label: t.value("members_table.actions"),
            srOnly: true,
            align: "right",
        });
    }

    return columns;
});

function fullName(user) {
    return `${user?.name ?? ""} ${user?.lastname ?? ""}`.trim() || user?.email;
}
</script>
