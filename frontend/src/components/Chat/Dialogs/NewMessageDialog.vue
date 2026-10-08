<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <TransitionRoot as="template" :show="modelValue">
        <Dialog class="relative z-50" @close="close">
            <!-- Backdrop -->
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
                        <!-- Header -->
                        <header class="flex items-center justify-between px-5 pt-5">
                            <DialogTitle class="text-lg font-semibold text-gray-900">
                                {{ t("channels.new_message_dialog.title") }}
                            </DialogTitle>
                            <button
                                type="button"
                                class="rounded-md p-2 text-gray-400 hover:bg-gray-100 hover:text-gray-700"
                                @click="close"
                            >
                                <XMarkIcon class="h-5 w-5" aria-hidden="true" />
                            </button>
                        </header>

                        <!-- Search / selected pill -->
                        <div class="px-5 pt-4">
                            <div
                                class="rounded-md border border-gray-200 bg-white px-3 py-1.5 flex items-center gap-2 min-h-[2.5rem]"
                            >
                                <!-- Selected user pill -->
                                <span
                                    v-if="selected"
                                    class="inline-flex items-center gap-1.5 rounded-md bg-gray-100 px-2 py-1.5 text-xs font-medium text-gray-700 shrink-0"
                                >
                                    <div class="h-4 w-4">
                                        <UserAvatar :user="selected" />
                                    </div>
                                    {{ selected.name }} {{ selected.lastname }}
                                    <button
                                        class="ml-0.5 text-gray-400 hover:text-gray-700 leading-none"
                                        @click.stop="clearSelection"
                                    >
                                        ×
                                    </button>
                                </span>

                                <!-- Search input, hidden once a user is selected -->
                                <input
                                    v-if="!selected"
                                    ref="searchInput"
                                    class="flex-1 min-w-0 bg-transparent text-sm text-gray-900 placeholder-gray-400 border-0 outline-none focus:outline-none focus:ring-0"
                                    spellcheck="false"
                                    :placeholder="t('channels.new_message_dialog.select_user')"
                                    :value="search.query.value"
                                    @input="onInput"
                                />
                            </div>
                        </div>

                        <!-- User list -->
                        <div class="mt-3 h-64 overflow-y-auto px-5 pb-4">
                            <div
                                class="text-xs font-semibold uppercase tracking-wider text-gray-400 mb-2"
                            >
                                {{ t("common.label.users") }}
                            </div>

                            <div
                                v-if="filteredUsers.length === 0"
                                class="py-8 text-center text-sm text-gray-500"
                            >
                                {{ t("common.label.no_users") }}
                            </div>

                            <div class="space-y-0.5">
                                <div
                                    v-for="user in filteredUsers"
                                    :key="user.id"
                                    class="flex cursor-pointer items-center gap-3 rounded-md px-3 py-2 transition-colors"
                                    :class="
                                        isSelected(user)
                                            ? 'bg-indigo-50 hover:bg-indigo-50'
                                            : 'hover:bg-gray-100'
                                    "
                                    @click="selectUser(user)"
                                >
                                    <!-- Radio indicator -->
                                    <span
                                        class="flex h-[18px] w-[18px] shrink-0 items-center justify-center rounded-full border-2 transition-colors"
                                        :class="
                                            isSelected(user)
                                                ? 'border-indigo-600 bg-indigo-600'
                                                : 'border-gray-300'
                                        "
                                    >
                                        <span
                                            v-if="isSelected(user)"
                                            class="h-2 w-2 rounded-full bg-white"
                                        />
                                    </span>

                                    <div class="h-8 w-8 shrink-0">
                                        <UserAvatar :user="user" />
                                    </div>

                                    <div class="min-w-0 flex-1">
                                        <div class="truncate text-sm font-medium text-gray-900">
                                            {{ user.name }} {{ user.lastname }}
                                        </div>
                                        <div class="truncate text-xs text-gray-500">
                                            {{ user.email }}
                                        </div>
                                    </div>
                                </div>
                            </div>
                        </div>

                        <!-- Footer -->
                        <footer class="border-t border-gray-200 bg-white px-5 py-4">
                            <div class="grid grid-cols-2 gap-3">
                                <button
                                    class="rounded-md bg-gray-100 px-3 py-2 text-sm font-semibold text-gray-700 hover:bg-gray-200 transition-colors"
                                    @click="close"
                                >
                                    {{ t("common.button.cancel") }}
                                </button>
                                <button
                                    class="rounded-md bg-indigo-600 px-3 py-2 text-sm font-semibold text-white hover:bg-indigo-500 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
                                    :disabled="!selected"
                                    @click="createDMChannel"
                                >
                                    {{ t("common.button.create") }}
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
import { ref, computed, watch, nextTick } from "vue";
import { Dialog, DialogPanel, DialogTitle, TransitionChild, TransitionRoot } from "@headlessui/vue";
import { XMarkIcon } from "@heroicons/vue/24/outline";
import UserAvatar from "@/components/UserAvatar.vue";
import { useUserStore } from "@/store/user";
import { useUserSearch } from "@/composables/useUserSearch";

const props = defineProps({
    modelValue: {
        type: Boolean,
        required: true,
    },
});

const emit = defineEmits(["create", "update:modelValue"]);

const USER_LIMIT = 50;

const userStore = useUserStore();
const search = useUserSearch(USER_LIMIT);
const selected = ref(null);
const searchInput = ref(null);

watch(
    () => props.modelValue,
    (open) => {
        if (open) {
            selected.value = null;
            search.reset();
            search.search("");
            nextTick(() => searchInput.value?.focus());
        }
    },
);

const filteredUsers = computed(() =>
    search.results.value.filter((u) => u.id !== userStore.user.id),
);

function onInput(e) {
    search.search(e.target.value);
}

function isSelected(user) {
    return selected.value?.id === user.id;
}

function selectUser(user) {
    selected.value = isSelected(user) ? null : user;
    search.search("");
}

function clearSelection() {
    selected.value = null;
    search.search("");
    nextTick(() => searchInput.value?.focus());
}

function createDMChannel() {
    if (!selected.value) return;
    emit("create", {
        type: "direct",
        name: selected.value.name + `${selected.value.email}`,
        description: "",
        user: selected.value.id,
    });
    close();
}

function close() {
    selected.value = null;
    search.reset();
    emit("update:modelValue", false);
}
</script>
