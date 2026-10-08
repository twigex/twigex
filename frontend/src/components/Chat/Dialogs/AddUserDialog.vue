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
                <div class="fixed inset-0 bg-gray-900/40 backdrop-blur-sm" />
            </TransitionChild>

            <div class="fixed inset-0 z-10 flex items-center justify-center p-4">
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
                        class="w-full max-w-xl overflow-hidden rounded-xl bg-white text-gray-900 shadow-2xl ring-1 ring-black/5"
                    >
                        <header class="flex items-center justify-between px-5 pt-5">
                            <DialogTitle class="text-lg font-semibold text-gray-900">
                                {{ t("channels.add_users_dialog.title") }}
                            </DialogTitle>

                            <button
                                type="button"
                                class="rounded-md p-2 text-gray-400 hover:bg-gray-100 hover:text-gray-700"
                                @click="close"
                            >
                                ×
                            </button>
                        </header>

                        <div class="px-5 pt-2 text-center text-gray-700">
                            <span class="text-lg"> #{{ channelName }} </span>
                        </div>

                        <div class="mt-2 px-5 text-xs text-gray-500">
                            {{ t("channels.add_users_dialog.description") }}
                        </div>

                        <div class="px-5 pt-4 pb-4">
                            <UserGroupPicker
                                v-model:selected-users="selected"
                                v-model:selected-groups="selectedGroups"
                                :exclude-user-ids="excludeUserIds"
                                :exclude-group-ids="excludeGroupIds"
                                :placeholder="t('channels.add_users_dialog.input')"
                            />
                        </div>

                        <footer class="border-t border-gray-200 bg-white px-5 py-4">
                            <div class="grid grid-cols-2 gap-3">
                                <button
                                    class="rounded-md bg-gray-100 px-3 py-2 text-sm font-semibold hover:bg-gray-200"
                                    @click="close"
                                >
                                    {{ t("common.button.cancel") }}
                                </button>

                                <button
                                    class="rounded-md bg-indigo-600 px-3 py-2 text-sm font-semibold text-white hover:bg-indigo-500 disabled:opacity-50"
                                    :disabled="!selected.length && !selectedGroups.length"
                                    @click="add"
                                >
                                    {{ t("common.button.add") }}
                                </button>
                            </div>
                        </footer>
                    </DialogPanel>
                </TransitionChild>
            </div>
        </Dialog>
    </TransitionRoot>
</template>

<script setup>
import { t } from "@/i18n/index.js";

import { ref, computed, watch } from "vue";
import { Dialog, DialogPanel, DialogTitle, TransitionChild, TransitionRoot } from "@headlessui/vue";
import chatService from "@/services/chatService";
import { useChannelsStore } from "@/store/channels";
import UserGroupPicker from "@/components/UserGroupPicker.vue";

const props = defineProps({
    modelValue: Boolean,
    channel: {
        type: Object,
        default: null,
    },
});

const emit = defineEmits(["add", "add-groups", "update:modelValue"]);

const channelStore = useChannelsStore();

const selected = ref([]);
const selectedGroups = ref([]);
const excludeUserIds = ref([]);
const excludeGroupIds = ref([]);

const targetChannel = computed(() => props.channel ?? channelStore.addUserDialog.channel);

const channelName = computed(() => targetChannel.value?.name ?? "channel");

watch(
    () => props.modelValue,
    async (open) => {
        if (!open) return;

        selected.value = [];
        selectedGroups.value = [];

        // Only exclude *direct* members from the picker. Materialized members
        // (is_direct=false, came in via a group) should still be selectable;
        // picking them promotes their row to direct on the backend while
        // preserving their per-user state.
        excludeUserIds.value = (targetChannel.value?.channel_members ?? [])
            .filter((m) => m.is_direct)
            .map((m) => m.user_id);

        // Fetch the groups already attached to this channel so we can exclude
        // them from the picker.
        excludeGroupIds.value = [];
        const channelId = targetChannel.value?.id;

        if (channelId) {
            try {
                const res = await chatService.getChannelGroups(channelId);

                excludeGroupIds.value = (res.data ?? []).map((g) => g.group_id);
            } catch {
                // Non-admins can't list, fall back to empty. Backend will
                // reject duplicates if they slip through.
            }
        }
    },
);

function add() {
    if (selected.value.length) {
        emit("add", selected.value);
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
