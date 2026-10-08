<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div>
        <GridDialog
            :isOpen="isDialogOpen"
            @close="closeDialog"
            :workspaceId="route.params.id"
            :tableId="props.parentTableID"
            :currentLinkedIDs="selectedItems"
            :singleSelect="props.singleSelect"
            :field="props.field"
            :editItemLink="props.editItemLink"
            :linkedID="props.linkedID"
            :currentLocalTableID="props.currentLocalTableID"
            :linkedItemsForCheckbox="props.linkedItemsForCheckbox"
            :isAtFirstLinkedLevel="props.isAtFirstLinkedLevel"
            :title="props.title"
            @saved="emit('saved', $event)"
        />
    </div>
</template>

<script setup>
import { watch, ref, onMounted, computed } from "vue";
import { useRoute } from "vue-router";
import { useWorkspaceStore } from "@/store/workspaces";
import GridDialog from "./GridViewDialog.vue";

const props = defineProps({
    open: {
        type: Boolean,
        default: false,
    },
    linkedID: {
        type: String,
        default: "no-id",
    },
    editItemLink: {
        type: Object,
        default: () => ({}),
    },
    singleSelect: {
        type: Boolean,
        default: false,
    },
    field: {
        type: String,
        default: "",
    },
    parentTableID: {
        type: String,
        default: "",
    },
    currentLocalTableID: {
        type: String,
        default: "",
    },
    selectedItemsForLinked: {
        type: Array,
        default: () => [],
    },
    linkedItemsForCheckbox: {
        type: Array,
        default: () => [],
    },
    isAtFirstLinkedLevel: {
        type: Boolean,
        default: false,
    },
    title: { type: String, default: "" },
});

const route = useRoute();
const workspaceStore = useWorkspaceStore();
const selectedItems = ref([]);
const isDialogOpen = ref(props.open);
const emit = defineEmits(["close", "create", "updateSelectedItem", "saved"]);

const dialogTableID = computed({
    get() {
        return workspaceStore.getDialogTableID;
    },
    set(value) {
        workspaceStore.setDialogTableID(value);
    },
});

const closeDialog = () => {
    isDialogOpen.value = false;
    emit("close");
};

const fetchTableData = () => {
    dialogTableID.value = props.parentTableID;
    selectedItems.value = props.selectedItemsForLinked || [];
};

onMounted(() => {
    selectedItems.value = [];
    fetchTableData();
});

watch(
    () => props.open,
    (isOpen) => {
        isDialogOpen.value = isOpen;
        if (isOpen) {
            selectedItems.value = [];
            fetchTableData();
        } else {
            selectedItems.value = [];
        }
    },
);

watch(
    () => props.linkedID,
    () => {
        fetchTableData();
    },
);
</script>
