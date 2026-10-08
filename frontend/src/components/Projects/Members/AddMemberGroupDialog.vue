<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <TransitionRoot as="template" :show="modelValue">
        <Dialog class="relative z-50" @close="close">
            <TransitionChild
                as="template"
                enter="ease-out duration-200"
                enter-from="opacity-0"
                enter-to="opacity-100"
                leave="ease-in duration-150"
                leave-from="opacity-100"
                leave-to="opacity-0"
            >
                <div class="fixed inset-0 bg-gray-500 bg-opacity-75 transition-opacity" />
            </TransitionChild>

            <div class="fixed inset-0 z-10 flex items-end justify-center p-4 sm:items-center">
                <TransitionChild
                    as="template"
                    enter="ease-out duration-200"
                    enter-from="opacity-0 scale-95"
                    enter-to="opacity-100 scale-100"
                    leave="ease-in duration-150"
                    leave-from="opacity-100 scale-100"
                    leave-to="opacity-0 scale-95"
                >
                    <DialogPanel
                        class="relative w-full transform rounded-lg bg-white px-4 pb-4 pt-5 text-left shadow-xl transition-all sm:my-8 sm:max-w-lg sm:p-6"
                    >
                        <DialogTitle
                            as="h3"
                            class="text-base font-semibold leading-6 text-gray-900"
                        >
                            {{ t("projects.workspace_members.add_dialog_title") }}
                        </DialogTitle>
                        <p class="mt-1 text-sm text-gray-500">
                            {{ workspace?.title }}
                        </p>

                        <div class="mt-4">
                            <UserGroupPicker
                                v-model:selected-users="selectedUsers"
                                v-model:selected-groups="selectedGroups"
                                :exclude-user-ids="excludeUserIds"
                                :exclude-group-ids="excludeGroupIds"
                                :placeholder="
                                    t('projects.workspace_members.add_dialog_placeholder')
                                "
                            />
                        </div>

                        <div class="mt-5 sm:mt-6 sm:flex sm:flex-row-reverse">
                            <button
                                type="button"
                                class="inline-flex w-full justify-center rounded-md bg-indigo-600 px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-indigo-500 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600 disabled:cursor-not-allowed disabled:opacity-50 sm:ml-3 sm:w-auto"
                                :disabled="!selectedUsers.length && !selectedGroups.length"
                                @click="add"
                            >
                                {{ t("common.button.add") }}
                            </button>
                            <button
                                type="button"
                                class="mt-3 inline-flex w-full justify-center rounded-md bg-white px-3 py-2 text-sm font-semibold text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 hover:bg-gray-50 sm:mt-0 sm:w-auto"
                                @click="close"
                            >
                                {{ t("common.button.cancel") }}
                            </button>
                        </div>
                    </DialogPanel>
                </TransitionChild>
            </div>
        </Dialog>
    </TransitionRoot>
</template>

<script setup>
import { ref, watch } from "vue";
import { Dialog, DialogPanel, DialogTitle, TransitionChild, TransitionRoot } from "@headlessui/vue";
import { t } from "@/i18n/index.js";
import UserGroupPicker from "@/components/UserGroupPicker.vue";
import workspaceService from "@/services/workspaceService";

const props = defineProps({
    modelValue: Boolean,
    workspace: { type: Object, default: null },
    existingMemberIds: { type: Array, default: () => [] },
});

const emit = defineEmits(["add", "add-groups", "update:modelValue"]);

const selectedUsers = ref([]);
const selectedGroups = ref([]);
const excludeUserIds = ref([]);
const excludeGroupIds = ref([]);

watch(
    () => props.modelValue,
    async (open) => {
        if (!open) return;
        selectedUsers.value = [];
        selectedGroups.value = [];
        excludeUserIds.value = props.existingMemberIds;

        excludeGroupIds.value = [];
        if (props.workspace?.id) {
            try {
                const res = await workspaceService.getWorkspaceGroups(props.workspace.id);

                excludeGroupIds.value = (res.data ?? []).map((g) => g.group_id);
            } catch {
                // non-admins may be denied, fall back to empty
            }
        }
    },
);

function add() {
    if (selectedUsers.value.length) {
        emit("add", selectedUsers.value);
    }

    if (selectedGroups.value.length) {
        emit("add-groups", selectedGroups.value);
    }

    close();
}

function close() {
    emit("update:modelValue", false);
}
</script>
