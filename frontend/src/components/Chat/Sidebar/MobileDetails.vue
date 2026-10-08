<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <!-- Menu for mobile or small screens -->
    <TransitionRoot as="template" :show="channelStore.mobileDetails">
        <Dialog
            class="relative z-50 lg:hidden"
            @close="channelStore.mobileDetails = !channelStore.mobileDetails"
        >
            <TransitionChild
                as="template"
                enter="transition-opacity ease-linear duration-300"
                enter-from="opacity-0"
                enter-to=""
                leave="transition-opacity ease-linear duration-300"
                leave-from=""
                leave-to="opacity-0"
            >
                <div class="fixed inset-0 bg-gray-900/80" />
            </TransitionChild>

            <div class="fixed inset-0 flex justify-end">
                <TransitionChild
                    as="template"
                    enter="transition ease-in-out duration-300 transform"
                    enter-from="translate-x-full"
                    enter-to="translate-x-0"
                    leave="transition ease-in-out duration-300 transform"
                    leave-from="translate-x-0"
                    leave-to="translate-x-full"
                >
                    <DialogPanel class="relative ml-16 flex w-full max-w-xs flex-1">
                        <TransitionChild
                            as="template"
                            enter="ease-in-out duration-300"
                            enter-from="opacity-0"
                            enter-to=""
                            leave="ease-in-out duration-300"
                            leave-from=""
                            leave-to="opacity-0"
                        >
                            <div class="absolute top-0 right-full flex w-16 justify-center pt-5">
                                <button
                                    type="button"
                                    class="-m-2.5 p-2.5"
                                    @click="
                                        channelStore.mobileDetails = !channelStore.mobileDetails
                                    "
                                >
                                    <span class="sr-only">Close sidebar</span>
                                    <XMarkIcon class="size-6 text-white" aria-hidden="true" />
                                </button>
                            </div>
                        </TransitionChild>

                        <!-- Sidebar component -->
                        <div
                            class="flex grow flex-col justify-center gap-y-5 overflow-y-auto bg-white pb-2 dark:bg-white dark:ring-1"
                        >
                            <ChatDetailsContent />
                        </div>
                    </DialogPanel>
                </TransitionChild>
            </div>
        </Dialog>
    </TransitionRoot>
</template>

<script setup>
import ChatDetailsContent from "@/components/Chat/Sidebar/ChatDetailsContent.vue";
import { Dialog, DialogPanel, TransitionChild, TransitionRoot } from "@headlessui/vue";
import { XMarkIcon } from "@heroicons/vue/24/outline";
import { useChannelsStore } from "@/store/channels";

const channelStore = useChannelsStore();
</script>
