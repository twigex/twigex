<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div
        v-if="channelStore.currentChannel != null"
        class="flex items-center justify-between px-4 h-full w-full"
    >
        <!-- Channel identity -->
        <div class="flex items-center gap-x-2 min-w-0">
            <HashtagIcon
                v-if="channel?.type === 'O'"
                class="h-5 w-5 text-gray-400 shrink-0"
                aria-hidden="true"
            />
            <LockClosedIcon
                v-else-if="channel?.type === 'P'"
                class="h-5 w-5 text-gray-400 shrink-0"
                aria-hidden="true"
            />
            <ChatBubbleLeftRightIcon
                v-else
                class="h-5 w-5 text-gray-400 shrink-0"
                aria-hidden="true"
            />

            <h1 class="text-sm font-semibold text-gray-900 truncate">
                {{ channel?.displayname }}
            </h1>

            <span
                v-if="channel?.channel_members?.length && channel?.type !== 'D'"
                class="hidden sm:inline-flex shrink-0 items-center rounded-full bg-gray-100 px-2 py-0.5 text-xs font-medium text-gray-500"
            >
                {{ channel.channel_members.length }}
                {{ channel.channel_members.length === 1 ? "member" : "members" }}
            </span>
        </div>

        <!-- Actions -->
        <div class="flex items-center gap-x-1 shrink-0">
            <button
                v-if="settingsStore.chatSettings.enable_video"
                @click="onVideoClick()"
                type="button"
                :title="videoLabel"
                class="inline-flex items-center gap-x-1.5 rounded-md px-2.5 py-1.5 text-sm font-medium ring-1 ring-inset transition-colors"
                :class="
                    liveHere.length > 0
                        ? 'bg-green-50 text-green-700 ring-green-600/30 hover:bg-green-100'
                        : 'text-gray-700 ring-gray-300 hover:bg-gray-50'
                "
            >
                <span class="relative">
                    <VideoCameraIcon
                        class="h-4 w-4"
                        :class="liveHere.length > 0 ? 'text-green-600' : 'text-indigo-500'"
                        aria-hidden="true"
                    />
                    <span
                        v-if="liveHere.length > 0"
                        class="absolute -right-0.5 -top-0.5 size-1.5 rounded-full bg-green-500 ring-1 ring-white"
                        aria-hidden="true"
                    />
                </span>
                <span class="hidden sm:inline text-xs">{{ videoLabel }}</span>
            </button>

            <button
                @click="
                    ((channelStore.details = !channelStore.details),
                    (channelStore.mobileDetails = !channelStore.mobileDetails))
                "
                type="button"
                class="rounded-md p-1.5 transition-colors"
                :class="
                    channelStore.details
                        ? 'bg-gray-100 text-gray-700'
                        : 'text-gray-400 hover:text-gray-600 hover:bg-gray-100'
                "
            >
                <InformationCircleIcon class="h-5 w-5" aria-hidden="true" />
            </button>
        </div>

        <MeetingPicker
            v-if="showPicker"
            :channel-id="channel.id"
            :active="activeMeetings"
            :scheduled="scheduledMeetings"
            @join="joinMeeting"
            @new="openOptions('instant')"
            @schedule="openOptions('schedule')"
            @edit="openEdit"
            @changed="loadMeetings"
            @close="showPicker = false"
        />

        <MeetingOptionsDialog
            v-if="optionsMode"
            :channel-id="channel.id"
            :mode="optionsMode"
            :meeting="editingMeeting"
            @started="joinMeeting"
            @created="closeOptions"
            @updated="onMeetingUpdated"
            @close="closeOptions"
        />
    </div>
</template>

<script setup>
import { t } from "@/i18n/index.js";
import { ref, computed } from "vue";
import { VideoCameraIcon } from "@heroicons/vue/24/solid";
import {
    InformationCircleIcon,
    HashtagIcon,
    LockClosedIcon,
    ChatBubbleLeftRightIcon,
} from "@heroicons/vue/24/outline";
import { useChannelsStore } from "@/store/channels";
import { useSettingsStore } from "@/store/settings";
import { useAlertStore } from "@/store/alerts";
import { useCallsStore } from "@/store/calls";
import { useUserStore } from "@/store/user";
import { useMeetingsStore } from "@/store/meetings";
import chatService from "@/services/chatService";
import MeetingPicker from "@/components/Chat/Meetings/MeetingPicker.vue";
import MeetingOptionsDialog from "@/components/Chat/Meetings/MeetingOptionsDialog.vue";

const props = defineProps({
    channel: {
        type: [Object, null],
        default: null,
    },
});

const channelStore = useChannelsStore();
const settingsStore = useSettingsStore();
const alertStore = useAlertStore();
const callsStore = useCallsStore();
const userStore = useUserStore();
const meetingsStore = useMeetingsStore();

const isDirect = computed(() => props.channel?.type === "D");

const liveHere = computed(() => {
    if (isDirect.value) return [];

    return meetingsStore.liveMeetings.filter((m) => m.channel_id === props.channel?.id);
});

const videoLabel = computed(() =>
    isDirect.value ? t.value("channels.header.call") : t.value("channels.header.meet"),
);

const calleeId = computed(
    () => props.channel?.channel_members?.find((m) => m.user_id !== userStore.user?.id)?.user_id,
);

const showPicker = ref(false);
const optionsMode = ref(null);
const editingMeeting = ref(null);
const activeMeetings = ref([]);
const scheduledMeetings = ref([]);

async function loadMeetings() {
    const [active, scheduled] = await Promise.all([
        chatService.getMeetings(props.channel.id),
        chatService.getScheduledMeetings(props.channel.id),
    ]);

    activeMeetings.value = active.data || [];
    scheduledMeetings.value = scheduled.data || [];
}

function onVideoClick() {
    if (isDirect.value) {
        startDirectCall();
    } else {
        openVideo();
    }
}

async function openVideo() {
    try {
        await loadMeetings();
        showPicker.value = true;
    } catch {
        alertStore.showError(t.value("meetings.error.load"));
    }
}

async function startDirectCall() {
    // Set outgoing before awaiting so that if the other member is ringing us at
    // the same instant, their incoming call is suppressed rather than flashing a
    // dialog we're about to resolve.
    callsStore.setOutgoing({
        channelId: props.channel.id,
        calleeId: calleeId.value,
        name: props.channel.displayname,
    });
    try {
        const { data } = await chatService.startCall(props.channel.id);

        if (data.status === "connected") {
            callsStore.clearOutgoing();
            callsStore.clearIncoming();
            channelStore.openVideoChat(data.meeting);
        }
    } catch {
        callsStore.clearOutgoing();
        alertStore.showError(t.value("calls.error.start"));
    }
}

function joinMeeting(meeting) {
    showPicker.value = false;
    optionsMode.value = null;
    channelStore.openVideoChat(meeting);
}

function openOptions(mode) {
    showPicker.value = false;
    editingMeeting.value = null;
    optionsMode.value = mode;
}

function openEdit(meeting) {
    showPicker.value = false;
    editingMeeting.value = meeting;
    optionsMode.value = "edit";
}

function closeOptions() {
    optionsMode.value = null;
    editingMeeting.value = null;
}

async function onMeetingUpdated() {
    closeOptions();
    await loadMeetings();
}
</script>
