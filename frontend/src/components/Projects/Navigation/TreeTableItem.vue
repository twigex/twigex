<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div
        :class="[
            isCurrent
                ? 'bg-gray-50 text-indigo-600'
                : 'text-gray-700 hover:bg-gray-50 hover:text-indigo-600',
            dragged?.id === table.id ? 'opacity-50' : '',
            'group flex cursor-pointer items-center gap-x-2 truncate rounded-md p-2 pl-3 text-xs font-medium leading-2',
        ]"
        :title="name"
        @click.stop="open"
        @pointerdown="pressItem($event, table)"
    >
        <span v-if="alignWithFolders" class="h-5 w-5 flex-none" aria-hidden="true" />
        <TableCellsIcon class="h-5 w-5 flex-none" aria-hidden="true" />
        <span class="truncate">{{ name }}</span>
    </div>
</template>

<script setup>
import { computed } from "vue";
import { useRoute, useRouter } from "vue-router";
import { TableCellsIcon } from "@heroicons/vue/24/outline";
import { useMoveProjectItem } from "@/composables/projects/useMoveProjectItem";
import { useWorkspaceStore } from "@/store/workspaces";

const props = defineProps({
    table: { type: Object, required: true },
    // Leaves the width of a folder's chevron before the icon: always inside a
    // folder, so the table reads as in it, and at the top of a workspace only
    // when folders are there to line up with.
    alignWithFolders: { type: Boolean, default: false },
});

const route = useRoute();
const router = useRouter();
const workspaceStore = useWorkspaceStore();
const { dragged, pressItem } = useMoveProjectItem();

const name = computed(() => props.table.display_name?.String || props.table.name);
const isCurrent = computed(() => route.params.tid === props.table.id);

function open() {
    workspaceStore.setSortOptions([{ field: "", direction: "asc" }]);
    router.push(`/projects/${route.params.id}/grid/${props.table.id}/view/${props.table.view_id}`);
}
</script>
