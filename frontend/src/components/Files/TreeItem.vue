<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <li class="list-none truncate">
        <div
            :data-move-target="isFolder ? model.id : undefined"
            :class="{
                'font-semibold': isFolder,
                'cursor-pointer': true,
                'bg-gray-50 text-indigo-600': route.params.id == model.id,
                'text-gray-700 hover:text-indigo-600 hover:bg-gray-50':
                    !route.params.id != model.id,
            }"
            @click.stop="openNode(model)"
            class="group flex items-center gap-x-2 rounded-md p-2 pl-3 text-xs leading-2 truncate data-[drop-target]:bg-indigo-50 data-[drop-target]:ring-1 data-[drop-target]:ring-inset data-[drop-target]:ring-indigo-500"
        >
            <span @click.stop="load(model)" v-if="isFolder || isDrive">
                <ChevronDownIcon v-if="isOpen" class="h-5 w-5" />
                <ChevronRightIcon v-else class="h-5 w-5" />
            </span>
            <span v-if="isDrive">
                <ServerIcon class="h-5 w-5" />
            </span>
            <span v-else-if="isFolder">
                <FolderOpenIcon v-if="isOpen" class="h-5 w-5" />
                <FolderIcon v-else class="h-5 w-5" />
            </span>
            <span v-else>
                <DocumentIcon class="h-5 w-5" />
            </span>
            {{ model.name }}
        </div>
        <ul v-show="isOpen" v-if="isFolder" class="pl-3">
            <TreeItem
                class="item my-1"
                v-for="(childModel, index) in model.children"
                :model="childModel"
                :key="index"
                :load-children="loadChildren"
                :open="open"
            />
        </ul>
    </li>
</template>

<script setup>
import { ref, computed } from "vue";
import { useRoute } from "vue-router";
import {
    ChevronDownIcon,
    ChevronRightIcon,
    FolderIcon,
    FolderOpenIcon,
    DocumentIcon,
    ServerIcon,
} from "@heroicons/vue/24/outline";

const props = defineProps({
    model: Object,
    loadChildren: Function,
    open: Function,
});

const route = useRoute();
const isOpen = ref(false);

const isFolder = computed(() => {
    return props.model.isFolder || props.model.type == "cloud#drive";
});

const isDrive = computed(() => {
    return props.model.type == "cloud#drive";
});

defineEmits(["load-children", "open"]);

function load(file) {
    props.loadChildren(file);
    isOpen.value = !isOpen.value;
}

function openNode(file) {
    props.open(file);
}
</script>
