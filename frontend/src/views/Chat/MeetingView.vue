<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <SidebarLayout :title="t('main.sections.chat')">
        <template #sidebar="{ mini }">
            <ChatNavigation :mini="mini" />
        </template>

        <div class="flex h-full w-full items-center justify-center p-6">
            <div
                class="w-full max-w-md rounded-xl bg-white p-8 text-center shadow-sm ring-1 ring-gray-200"
            >
                <div v-if="state === 'loading'" class="text-sm text-gray-500">
                    {{ t("meetings.page.loading") }}
                </div>

                <template v-else-if="state === 'lobby'">
                    <CalendarIcon class="mx-auto h-10 w-10 text-indigo-500" />
                    <h1 class="mt-4 text-lg font-semibold leading-6 text-gray-900">
                        {{ meeting.title }}
                    </h1>
                    <p class="mt-1 text-sm text-gray-500">
                        {{ getDateAndTime(meeting.scheduled_at) }}
                    </p>
                    <p v-if="meeting.channel_name" class="text-sm text-gray-400">
                        {{
                            t("meetings.nav.in_channel", {
                                channel: meeting.channel_name,
                            })
                        }}
                    </p>
                    <p class="mt-4 text-sm text-gray-500">
                        {{
                            isHost
                                ? t("meetings.page.host_not_started")
                                : t("meetings.page.not_started")
                        }}
                    </p>
                    <div class="mt-6 flex flex-col gap-2">
                        <template v-if="isHost">
                            <button
                                type="button"
                                :disabled="starting"
                                @click="startMeeting"
                                class="inline-flex items-center justify-center rounded-md bg-indigo-600 px-4 py-2 text-sm font-semibold text-white hover:bg-indigo-500 disabled:opacity-50"
                            >
                                {{ t("meetings.page.start") }}
                            </button>
                            <button
                                type="button"
                                @click="showEdit = true"
                                class="inline-flex items-center justify-center rounded-md bg-white px-4 py-2 text-sm font-semibold text-gray-900 ring-1 ring-inset ring-gray-300 hover:bg-gray-50"
                            >
                                {{ t("meetings.page.edit") }}
                            </button>
                        </template>
                        <button
                            type="button"
                            @click="leave"
                            class="text-sm font-semibold text-gray-500 hover:text-gray-700"
                        >
                            {{ t("meetings.page.back") }}
                        </button>
                    </div>
                </template>

                <template v-else-if="state === 'live'">
                    <VideoCameraIcon class="mx-auto h-10 w-10 text-indigo-500" />
                    <h1 class="mt-4 text-lg font-semibold leading-6 text-gray-900">
                        {{ meeting.title }}
                    </h1>
                    <p v-if="meeting.channel_name" class="mt-1 text-sm text-gray-400">
                        {{
                            t("meetings.nav.in_channel", {
                                channel: meeting.channel_name,
                            })
                        }}
                    </p>
                    <p class="mt-4 text-sm text-gray-500">
                        {{ t("meetings.page.live") }}
                    </p>
                    <div class="mt-6 flex flex-col gap-2">
                        <button
                            type="button"
                            @click="join(meeting)"
                            class="inline-flex items-center justify-center rounded-md bg-indigo-600 px-4 py-2 text-sm font-semibold text-white hover:bg-indigo-500"
                        >
                            {{ t("meetings.page.join") }}
                        </button>
                        <button
                            type="button"
                            @click="leave"
                            class="text-sm font-semibold text-gray-500 hover:text-gray-700"
                        >
                            {{ t("meetings.page.back") }}
                        </button>
                    </div>
                </template>

                <template v-else-if="state === 'ended'">
                    <VideoCameraSlashIcon class="mx-auto h-10 w-10 text-gray-400" />
                    <h1 class="mt-4 text-lg font-semibold leading-6 text-gray-900">
                        {{ meeting.title }}
                    </h1>
                    <p class="mt-2 text-sm text-gray-500">
                        {{ t("meetings.page.ended") }}
                    </p>
                    <button
                        type="button"
                        @click="leave"
                        class="mt-6 inline-flex items-center justify-center rounded-md bg-white px-4 py-2 text-sm font-semibold text-gray-900 ring-1 ring-inset ring-gray-300 hover:bg-gray-50"
                    >
                        {{ t("meetings.page.back") }}
                    </button>
                </template>

                <template v-else>
                    <ExclamationTriangleIcon class="mx-auto h-10 w-10 text-gray-400" />
                    <p class="mt-4 text-sm text-gray-500">
                        {{ error || t("meetings.page.unavailable") }}
                    </p>
                    <button
                        type="button"
                        @click="leave"
                        class="mt-6 inline-flex items-center justify-center rounded-md bg-white px-4 py-2 text-sm font-semibold text-gray-900 ring-1 ring-inset ring-gray-300 hover:bg-gray-50"
                    >
                        {{ t("meetings.page.back") }}
                    </button>
                </template>
            </div>

            <MeetingOptionsDialog
                v-if="showEdit && meeting"
                :channel-id="meeting.channel_id"
                mode="edit"
                :meeting="meeting"
                @updated="onEdited"
                @close="showEdit = false"
            />
        </div>

        <ChannelDialogs />
    </SidebarLayout>
</template>

<script setup>
import { ref, computed, watch, onMounted, onUnmounted } from "vue";
import { useRoute, useRouter } from "vue-router";
import SidebarLayout from "@/components/Navigation/SidebarLayout.vue";
import ChatNavigation from "@/components/Chat/Navigation/ChatNavigation.vue";
import ChannelDialogs from "@/components/Chat/Dialogs/ChannelDialogs.vue";
import {
    CalendarIcon,
    VideoCameraIcon,
    VideoCameraSlashIcon,
    ExclamationTriangleIcon,
} from "@heroicons/vue/24/outline";
import { t } from "@/i18n/index.js";
import chatService from "@/services/chatService";
import { useChannelsStore } from "@/store/channels";
import { useMeetingsStore } from "@/store/meetings";
import { useUserStore } from "@/store/user";
import { useAlertStore } from "@/store/alerts";
import useDateOperations from "@/composables/useDateOperations";
import MeetingOptionsDialog from "@/components/Chat/Meetings/MeetingOptionsDialog.vue";

const route = useRoute();
const router = useRouter();
const channelsStore = useChannelsStore();
const meetingsStore = useMeetingsStore();
const userStore = useUserStore();
const alertStore = useAlertStore();
const { getDateAndTime } = useDateOperations();

const state = ref("loading");
const meeting = ref(null);
const error = ref("");
const starting = ref(false);
const showEdit = ref(false);

const isHost = computed(() => meeting.value?.host_id === userStore.user?.id);

// Poll while scheduled so the page flips to the call the moment the host starts it.
let poll = null;

function stopPoll() {
    if (poll) {
        clearInterval(poll);
        poll = null;
    }
}

// Hand the live call to the floating overlay so it keeps running and can be minimized while browsing.
function join(m) {
    stopPoll();
    channelsStore.openVideoChat(m);
    router.replace({ name: "chat" });
}

function leave() {
    stopPoll();
    router.push({ name: "chat" });
}

async function load() {
    try {
        const { data } = await chatService.getMeeting(route.params.id);

        meeting.value = data;

        if (data.status === "active") {
            state.value = "live";
        } else if (data.status === "ended") {
            state.value = "ended";
            stopPoll();
        } else {
            state.value = "lobby";
        }
    } catch (e) {
        error.value = e?.response?.data?.message || t.value("meetings.page.unavailable");
        state.value = "error";
        stopPoll();
    }
}

async function startMeeting() {
    if (!meeting.value) return;
    starting.value = true;
    try {
        const { data } = await chatService.startMeeting(meeting.value.channel_id, meeting.value.id);

        join(data);
    } catch {
        alertStore.showError(t.value("meetings.error.start"));
    } finally {
        starting.value = false;
    }
}

function onEdited() {
    showEdit.value = false;
    load();
}

// Surface the join button the moment the store learns the meeting went live, without waiting for the poll.
watch(
    () => meetingsStore.myMeetings,
    (list) => {
        if (state.value !== "lobby") return;
        const live = list.find((m) => m.id === route.params.id && m.status === "active");

        if (live) {
            meeting.value = live;
            state.value = "live";
        }
    },
);

// The router reuses this instance when navigating between meetings, so onMounted won't refire.
watch(
    () => route.params.id,
    (id) => {
        if (!id) return;
        state.value = "loading";
        meeting.value = null;
        load();
    },
);

onMounted(() => {
    load();
    poll = setInterval(load, 12000);
});

onUnmounted(stopPoll);
</script>
