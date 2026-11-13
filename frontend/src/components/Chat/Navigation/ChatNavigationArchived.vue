<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <Disclosure v-slot="{ open }">
        <DisclosureButton
            as="div"
            class="sticky top-0 z-10 flex w-full cursor-pointer items-center gap-x-1 rounded-md bg-white px-2 py-1 text-xs font-semibold leading-6 text-gray-400 hover:bg-gray-50 hover:text-gray-600"
        >
            <ChevronDownIcon
                :class="[
                    'h-3 w-3 shrink-0 transition-transform duration-200',
                    open ? '' : '-rotate-90',
                ]"
                aria-hidden="true"
            />
            <span>{{ t("channels.navigation.archived") }}</span>
        </DisclosureButton>
        <DisclosurePanel>
            <ul role="list" class="mt-2 space-y-1">
                <li
                    v-for="item in archivedChannels"
                    :key="item.id"
                    role="button"
                    tabindex="0"
                    class="cursor-pointer rounded-md focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-inset focus-visible:ring-indigo-500"
                    @click="openChannel(item)"
                    @keydown.enter="openChannel(item)"
                    @keydown.space.prevent="openChannel(item)"
                >
                    <a
                        :class="[
                            item.id == route.params.chatId
                                ? 'bg-gray-50 text-indigo-600'
                                : 'text-gray-400 hover:text-indigo-600 hover:bg-gray-50',
                            'group flex items-center gap-x-3 rounded-md p-2 text-sm leading-6 font-medium',
                        ]"
                    >
                        <ArchiveBoxIcon class="size-5 shrink-0" aria-hidden="true" />
                        <span class="truncate">{{ item.displayname }}</span>
                    </a>
                </li>
            </ul>
        </DisclosurePanel>
    </Disclosure>
</template>

<script setup>
import { Disclosure, DisclosureButton, DisclosurePanel } from "@headlessui/vue";
import { ArchiveBoxIcon } from "@heroicons/vue/24/outline";
import { ChevronDownIcon } from "@heroicons/vue/20/solid";
import { t } from "@/i18n/index.js";
import { useChatNavigation } from "@/composables/chat/useChatNavigation";

const { route, archivedChannels, openChannel } = useChatNavigation();
</script>
