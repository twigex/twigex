<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <FormDialog
        :open="open"
        :title="t('projects.move_item.title', { name: itemName })"
        :confirm-label="t('projects.move_item.move_here')"
        :confirm-disabled="!current || isCurrentLocation"
        :loading="moving"
        @confirm="confirm"
        @close="emit('close')"
    >
        <nav class="flex min-w-0 items-center gap-x-1 text-sm" aria-label="Breadcrumb">
            <template v-for="(node, index) in path" :key="node.id">
                <ChevronRightIcon
                    v-if="index > 0"
                    class="h-4 w-4 flex-none text-gray-400"
                    aria-hidden="true"
                />
                <button
                    type="button"
                    :class="[
                        index === path.length - 1
                            ? 'font-semibold text-gray-900'
                            : 'text-gray-500 hover:text-gray-700',
                        'max-w-[10rem] truncate',
                    ]"
                    @click="currentId = node.id"
                >
                    {{ node.name }}
                </button>
            </template>
        </nav>

        <div class="mt-3 h-64 overflow-y-auto rounded-md ring-1 ring-inset ring-gray-200">
            <ul v-if="folders.length" role="list" class="divide-y divide-gray-100">
                <li v-for="folder in folders" :key="folder.id">
                    <button
                        type="button"
                        class="flex w-full items-center gap-x-3 px-3 py-2.5 text-left text-sm text-gray-900 hover:bg-gray-50"
                        @click="currentId = folder.id"
                    >
                        <FolderIcon class="h-5 w-5 flex-none text-indigo-600" aria-hidden="true" />
                        <span class="min-w-0 flex-auto truncate">{{ folder.name }}</span>
                        <ChevronRightIcon
                            class="h-4 w-4 flex-none text-gray-400"
                            aria-hidden="true"
                        />
                    </button>
                </li>
            </ul>
            <div v-else class="flex h-full flex-col items-center justify-center text-center">
                <FolderOpenIcon class="h-10 w-10 text-gray-300" aria-hidden="true" />
                <p class="mt-2 text-sm text-gray-500">
                    {{ t("projects.move_item.no_folders") }}
                </p>
            </div>
        </div>

        <p class="mt-2 text-xs text-gray-500">
            {{
                isCurrentLocation
                    ? t("projects.move_item.current_location")
                    : t("projects.move_item.hint")
            }}
        </p>
    </FormDialog>
</template>

<script setup>
import { computed, ref, watch } from "vue";
import { ChevronRightIcon } from "@heroicons/vue/20/solid";
import { FolderIcon, FolderOpenIcon } from "@heroicons/vue/24/outline";
import { t } from "@/i18n/index.js";
import FormDialog from "@/components/FormDialog.vue";
import { useMoveProjectItem } from "@/composables/projects/useMoveProjectItem";
import { useWorkspaceStore } from "@/store/workspaces";
import { findTreeNode, findTreeParent, isTableNode } from "@/utils/projects/tree";

const props = defineProps({
    open: { type: Boolean, default: false },
    item: { type: Object, default: null },
});

const emit = defineEmits(["close"]);

const workspaceStore = useWorkspaceStore();
const { moveItem } = useMoveProjectItem();

const currentId = ref(null);
const moving = ref(false);

const roots = computed(() => workspaceStore.getWorkspaceFolders);
const parent = computed(() => (props.item ? findTreeParent(roots.value, props.item.id) : null));
const current = computed(() => findTreeNode(roots.value, currentId.value));

const itemName = computed(() => props.item?.display_name?.String || props.item?.name || "");
const isCurrentLocation = computed(() => !!current.value && current.value.id === parent.value?.id);

const folders = computed(() =>
    (current.value?.children || []).filter(
        (node) => !isTableNode(node) && node.id !== props.item?.id,
    ),
);

const path = computed(() => {
    const nodes = [];

    for (let node = current.value; node; node = findTreeParent(roots.value, node.id)) {
        nodes.unshift(node);
    }

    return nodes;
});

watch(
    () => props.open,
    (open) => {
        if (open) currentId.value = parent.value?.id || roots.value[0]?.id || null;
    },
    { immediate: true },
);

async function confirm() {
    if (!current.value || isCurrentLocation.value) return;
    moving.value = true;
    const moved = await moveItem(props.item, current.value);

    moving.value = false;
    if (moved) emit("close");
}
</script>
