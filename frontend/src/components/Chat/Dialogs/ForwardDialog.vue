<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <TransitionRoot as="template" :show="modelValue">
        <Dialog class="relative z-50" @close="emits('update:modelValue', false)">
            <TransitionChild
                as="template"
                enter="ease-out duration-300"
                enter-from="opacity-0"
                enter-to="opacity-100"
                leave="ease-in duration-200"
                leave-from="opacity-100"
                leave-to="opacity-0"
            >
                <div class="fixed inset-0 bg-gray-500/75 transition-opacity" />
            </TransitionChild>

            <div class="fixed inset-0 z-10 w-screen overflow-y-auto">
                <div
                    class="flex min-h-full items-end justify-center p-4 text-center sm:items-center sm:p-0"
                >
                    <TransitionChild
                        as="template"
                        enter="ease-out duration-300"
                        enter-from="opacity-0 translate-y-4 sm:translate-y-0 sm:scale-95"
                        enter-to="opacity-100 translate-y-0 sm:scale-100"
                        leave="ease-in duration-200"
                        leave-from="opacity-100 translate-y-0 sm:scale-100"
                        leave-to="opacity-0 translate-y-4 sm:translate-y-0 sm:scale-95"
                    >
                        <DialogPanel
                            class="relative transform overflow-hidden rounded-lg bg-white px-4 pb-4 pt-5 text-left shadow-xl transition-all sm:my-8 sm:w-full sm:max-w-2xl sm:p-6"
                        >
                            <div class="sm:flex sm:items-start">
                                <div
                                    class="mx-auto flex size-12 shrink-0 items-center justify-center rounded-full bg-indigo-100 sm:mx-0 sm:size-10"
                                >
                                    <ArrowRightIcon
                                        class="size-6 text-indigo-600"
                                        aria-hidden="true"
                                    />
                                </div>
                                <div class="mt-3 text-center sm:ml-4 sm:mt-0 sm:text-left w-full">
                                    <DialogTitle
                                        as="h3"
                                        class="text-base font-semibold text-gray-900"
                                        >{{ t("channels.forward_dialog.title") }}</DialogTitle
                                    >
                                    <div class="mt-2">
                                        <p class="text-sm text-gray-500">
                                            {{ t("channels.forward_dialog.select_channel") }}
                                        </p>

                                        <div class="max-h-56 overflow-y-auto">
                                            <fieldset>
                                                <!-- Channels label -->
                                                <legend
                                                    class="mt-4 text-sm font-semibold text-gray-600"
                                                >
                                                    {{ t("channels.forward_dialog.channels") }}
                                                </legend>
                                                <div
                                                    class="mt-4 divide-y divide-gray-200 border-b border-t border-gray-200"
                                                >
                                                    <label
                                                        v-for="(channel, index) in channels"
                                                        :key="index"
                                                        :for="`channel-${channel.id}`"
                                                        class="relative flex gap-3 py-4 hover:bg-gray-100 px-2 cursor-pointer"
                                                    >
                                                        <div class="flex h-6 shrink-0 items-center">
                                                            <div
                                                                class="group grid size-4 grid-cols-1"
                                                            >
                                                                <input
                                                                    :id="`channel-${channel.id}`"
                                                                    :name="`channel-${channel.id}`"
                                                                    type="checkbox"
                                                                    :value="channel.id"
                                                                    v-model="selectedChannels"
                                                                    class="col-start-1 row-start-1 appearance-none rounded border border-gray-300 bg-white checked:border-indigo-600 checked:bg-indigo-600 indeterminate:border-indigo-600 indeterminate:bg-indigo-600 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600 disabled:border-gray-300 disabled:bg-gray-100 disabled:checked:bg-gray-100 forced-colors:appearance-auto"
                                                                />
                                                                <svg
                                                                    class="pointer-events-none col-start-1 row-start-1 size-3.5 self-center justify-self-center stroke-white group-has-[:disabled]:stroke-gray-950/25"
                                                                    viewBox="0 0 14 14"
                                                                    fill="none"
                                                                >
                                                                    <path
                                                                        class="opacity-0 group-has-[:checked]:opacity-100"
                                                                        d="M3 8L6 11L11 3.5"
                                                                        stroke-width="2"
                                                                        stroke-linecap="round"
                                                                        stroke-linejoin="round"
                                                                    />
                                                                    <path
                                                                        class="opacity-0 group-has-[:indeterminate]:opacity-100"
                                                                        d="M3 7H11"
                                                                        stroke-width="2"
                                                                        stroke-linecap="round"
                                                                        stroke-linejoin="round"
                                                                    />
                                                                </svg>
                                                            </div>
                                                        </div>
                                                        <div class="min-w-0 flex-1 text-sm/6">
                                                            <span
                                                                class="select-none font-medium text-gray-900"
                                                            >
                                                                {{ channel.name }}
                                                            </span>
                                                        </div>
                                                    </label>
                                                </div>
                                            </fieldset>
                                            <fieldset>
                                                <legend
                                                    class="mt-4 text-sm font-semibold text-gray-600"
                                                >
                                                    {{ t("channels.forward_dialog.users") }}
                                                </legend>
                                                <div
                                                    class="mt-4 divide-y divide-gray-200 border-b border-t border-gray-200"
                                                >
                                                    <label
                                                        v-for="(user, index) in users"
                                                        :key="index"
                                                        :for="`channel-${user.id}`"
                                                        class="relative flex gap-3 py-4 hover:bg-gray-100 px-2 cursor-pointer"
                                                    >
                                                        <div class="flex h-6 shrink-0 items-center">
                                                            <div
                                                                class="group grid size-4 grid-cols-1"
                                                            >
                                                                <input
                                                                    :id="`channel-${user.id}`"
                                                                    :name="`channel-${user.id}`"
                                                                    type="checkbox"
                                                                    :value="user.id"
                                                                    v-model="selectedUsers"
                                                                    class="col-start-1 row-start-1 appearance-none rounded border border-gray-300 bg-white checked:border-indigo-600 checked:bg-indigo-600 indeterminate:border-indigo-600 indeterminate:bg-indigo-600 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600 disabled:border-gray-300 disabled:bg-gray-100 disabled:checked:bg-gray-100 forced-colors:appearance-auto"
                                                                />
                                                                <svg
                                                                    class="pointer-events-none col-start-1 row-start-1 size-3.5 self-center justify-self-center stroke-white group-has-[:disabled]:stroke-gray-950/25"
                                                                    viewBox="0 0 14 14"
                                                                    fill="none"
                                                                >
                                                                    <path
                                                                        class="opacity-0 group-has-[:checked]:opacity-100"
                                                                        d="M3 8L6 11L11 3.5"
                                                                        stroke-width="2"
                                                                        stroke-linecap="round"
                                                                        stroke-linejoin="round"
                                                                    />
                                                                    <path
                                                                        class="opacity-0 group-has-[:indeterminate]:opacity-100"
                                                                        d="M3 7H11"
                                                                        stroke-width="2"
                                                                        stroke-linecap="round"
                                                                        stroke-linejoin="round"
                                                                    />
                                                                </svg>
                                                            </div>
                                                        </div>
                                                        <div
                                                            class="min-w-0 text-sm/6 flex flex-row items-center gap-x-1"
                                                        >
                                                            <div class="h-5 w-5 flex-shrink-0">
                                                                <UserAvatar :user="user" />
                                                            </div>

                                                            <span
                                                                class="select-none font-medium text-gray-900 truncate"
                                                            >
                                                                {{
                                                                    user.name + " " + user.lastname
                                                                }}
                                                            </span>
                                                        </div>
                                                    </label>
                                                </div>
                                            </fieldset>
                                        </div>
                                    </div>
                                </div>
                            </div>
                            <div class="mt-5 sm:mt-4 sm:flex sm:flex-row-reverse">
                                <BaseButton
                                    class="mx-3"
                                    :isDisabled="!selectedChannels.length && !selectedUsers.length"
                                    @click="forward"
                                    >{{ t("common.button.forward") }}</BaseButton
                                >

                                <button
                                    type="button"
                                    class="mt-3 inline-flex w-full justify-center rounded-md bg-white px-3 py-2 text-sm font-semibold text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 hover:bg-gray-50 sm:mt-0 sm:w-auto"
                                    @click="close"
                                    ref="cancelButtonRef"
                                >
                                    {{ t("common.button.cancel") }}
                                </button>
                            </div>
                        </DialogPanel>
                    </TransitionChild>
                </div>
            </div>
        </Dialog>
    </TransitionRoot>
</template>

<script setup>
import { t } from "@/i18n/index.js";

import { ref, computed } from "vue";
import BaseButton from "@/components/BaseButton.vue";
import { Dialog, DialogPanel, DialogTitle, TransitionRoot, TransitionChild } from "@headlessui/vue";
import { ArrowRightIcon } from "@heroicons/vue/24/outline";
import { useChannelsStore } from "@/store/channels";
import UserAvatar from "@/components/UserAvatar.vue";
import { useUserStore } from "@/store/user";
import { channelTypes } from "@/constants/channels";

defineProps({
    modelValue: {
        type: Boolean,
        required: true,
    },
});

const emits = defineEmits(["update:modelValue", "forward"]);

const channelStore = useChannelsStore();
const userStore = useUserStore();

const selectedChannels = ref([]);
const selectedUsers = ref([]);

const channels = computed(() => {
    return channelStore.channels.filter(
        (channel) =>
            channel.type != channelTypes.Direct && channel.id != channelStore.currentChannel.id,
    );
});

const users = computed(() => {
    return userStore.users.filter((user) => user.id != userStore.user.id);
});

function forward() {
    if (selectedChannels.value.length || selectedUsers.value.length) {
        emits("forward", {
            channels: selectedChannels.value,
            users: selectedUsers.value,
        });
    }

    close();
}

function close() {
    selectedChannels.value = [];

    emits("update:modelValue", false);
}
</script>
