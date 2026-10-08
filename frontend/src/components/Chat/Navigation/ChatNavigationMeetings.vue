<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <Disclosure v-slot="{ open }" default-open>
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
            <span>{{ t("meetings.nav.title") }}</span>
            <span
                v-if="!open && meetingsStore.liveMeetings.length > 0"
                class="ml-auto flex items-center justify-center whitespace-nowrap rounded-full bg-green-500 px-1.5 text-center text-xs font-medium leading-5 text-white"
                aria-hidden="true"
            >
                {{ meetingsStore.liveMeetings.length }}
            </span>
        </DisclosureButton>
        <transition
            enter-active-class="transition ease-out duration-100"
            enter-from-class="opacity-0"
            enter-to-class="opacity-100"
            leave-active-class="transition ease-in duration-75"
            leave-from-class="opacity-100"
            leave-to-class="opacity-0"
        >
            <DisclosurePanel>
                <ul role="list" class="mt-2 space-y-1">
                    <li
                        v-for="m in meetingsStore.liveMeetings"
                        :key="m.id"
                        role="button"
                        tabindex="0"
                        @click="openMeetingPage(m)"
                        @keydown.enter="openMeetingPage(m)"
                        @keydown.space.prevent="openMeetingPage(m)"
                        :class="[
                            'flex cursor-pointer items-center justify-between gap-x-2 rounded-md px-2 py-1.5 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-inset focus-visible:ring-indigo-500',
                            rowActive(m) ? 'bg-indigo-50' : 'hover:bg-gray-50',
                        ]"
                    >
                        <div class="min-w-0">
                            <div
                                class="flex items-center gap-x-1.5 truncate text-sm font-semibold text-gray-700"
                            >
                                <span
                                    class="h-2 w-2 shrink-0 rounded-full bg-green-500"
                                    aria-hidden="true"
                                />
                                <span class="truncate">{{ m.title }}</span>
                            </div>
                            <div class="truncate text-xs text-gray-400">
                                {{
                                    t("meetings.picker.in_call", {
                                        count: m.participants,
                                    })
                                }}<template v-if="m.channel_name">
                                    ·
                                    {{
                                        t("meetings.nav.in_channel", {
                                            channel: m.channel_name,
                                        })
                                    }}</template
                                >
                            </div>
                        </div>
                        <button
                            type="button"
                            @click.stop="joinMeeting(m)"
                            class="shrink-0 rounded-md bg-indigo-600 px-2.5 py-1 text-xs font-semibold text-white hover:bg-indigo-500"
                        >
                            {{ t("meetings.picker.join") }}
                        </button>
                    </li>

                    <li
                        v-for="m in meetingsStore.scheduledMeetings"
                        :key="m.id"
                        role="button"
                        tabindex="0"
                        @click="openMeetingPage(m)"
                        @keydown.enter="openMeetingPage(m)"
                        @keydown.space.prevent="openMeetingPage(m)"
                        :class="[
                            'cursor-pointer rounded-md px-2 py-1.5 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-inset focus-visible:ring-indigo-500',
                            rowActive(m) ? 'bg-indigo-50' : 'hover:bg-gray-50',
                        ]"
                    >
                        <div class="truncate text-sm font-semibold text-gray-700">
                            {{ m.title }}
                        </div>
                        <div class="truncate text-xs text-gray-400">
                            {{ getDateAndTime(m.scheduled_at)
                            }}<template v-if="m.channel_name">
                                ·
                                {{
                                    t("meetings.nav.in_channel", {
                                        channel: m.channel_name,
                                    })
                                }}</template
                            >
                        </div>
                    </li>
                </ul>
            </DisclosurePanel>
        </transition>
    </Disclosure>
</template>

<script setup>
import { t } from "@/i18n/index.js";

import { Disclosure, DisclosureButton, DisclosurePanel } from "@headlessui/vue";
import { ChevronDownIcon } from "@heroicons/vue/20/solid";
import { useChannelsStore } from "@/store/channels";
import { useMeetingsStore } from "@/store/meetings";
import useDateOperations from "@/composables/useDateOperations";
import { useChatNavigation } from "@/composables/chat/useChatNavigation";

const channelsStore = useChannelsStore();
const meetingsStore = useMeetingsStore();
const { getDateAndTime } = useDateOperations();
const { rowActive, openMeetingPage } = useChatNavigation();

function joinMeeting(meeting) {
    channelsStore.openVideoChat(meeting);
}
</script>
