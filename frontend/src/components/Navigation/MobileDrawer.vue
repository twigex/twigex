<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <TransitionRoot as="template" :show="navigationStore.mobileOpen">
        <Dialog class="relative z-50 lg:hidden" @close="navigationStore.mobileOpen = false">
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

            <div class="fixed inset-0 flex">
                <TransitionChild
                    as="template"
                    enter="transition ease-in-out duration-300 transform"
                    enter-from="-translate-x-full"
                    enter-to="translate-x-0"
                    leave="transition ease-in-out duration-300 transform"
                    leave-from="translate-x-0"
                    leave-to="-translate-x-full"
                >
                    <DialogPanel class="relative flex w-full max-w-xs flex-1">
                        <div
                            class="flex grow flex-col gap-y-5 overflow-hidden bg-white px-2 pb-2 dark:bg-white dark:ring-1"
                        >
                            <div class="flex h-16 shrink-0 items-center justify-between">
                                <div class="flex h-16 shrink-0 items-center justify-center">
                                    <img class="h-8 w-auto" :src="'/logo.png'" alt="Your Company" />
                                </div>
                                <button
                                    type="button"
                                    class="-m-2.5 p-2.5"
                                    @click="navigationStore.mobileOpen = false"
                                >
                                    <span class="sr-only">Close sidebar</span>
                                    <XMarkIcon class="size-6 text-gray-500" aria-hidden="true" />
                                </button>
                            </div>

                            <div class="flex shrink-0">
                                <div v-for="item in sections" :key="item.name">
                                    <li class="cursor-pointer flex">
                                        <a
                                            :class="[
                                                isActive(item)
                                                    ? 'bg-indigo-600 hover:bg-indigo-500 text-white'
                                                    : 'text-gray-400 hover:text-white hover:bg-indigo-600',
                                                'group flex gap-x-3 rounded-md p-3 text-sm leading-6 font-semibold relative',
                                            ]"
                                            @click="
                                                (open(item), (navigationStore.mobileOpen = false))
                                            "
                                        >
                                            <component :is="item.icon" class="h-6 w-6" />

                                            <span class="sr-only">{{ item.displayname }}</span>
                                            <div
                                                v-if="item.name == 'chat' && totalUnreadPosts > 0"
                                                class="absolute inline-flex items-center justify-center w-5 h-5 text-xs font-bold text-white bg-red-500 rounded-full top-0 right-0 transform"
                                            >
                                                {{ totalUnreadPosts }}
                                            </div>
                                        </a>
                                    </li>
                                </div>
                            </div>

                            <div v-if="$slots.default" class="flex-1 min-h-0">
                                <slot />
                            </div>
                        </div>
                    </DialogPanel>
                </TransitionChild>
            </div>
        </Dialog>
    </TransitionRoot>
</template>

<script setup>
import { Dialog, DialogPanel, TransitionChild, TransitionRoot } from "@headlessui/vue";
import { XMarkIcon } from "@heroicons/vue/24/outline";
import { useNavigationStore } from "@/store/navigation";
import { useSections } from "@/composables/useSections";

const navigationStore = useNavigationStore();
const { sections, totalUnreadPosts, isActive, open } = useSections();
</script>
