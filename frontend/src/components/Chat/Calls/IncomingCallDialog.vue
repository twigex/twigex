<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div
        v-if="calls.incoming"
        class="fixed inset-0 z-[60] flex items-center justify-center bg-black/40"
    >
        <div class="w-80 rounded-2xl bg-white p-6 text-center shadow-2xl ring-1 ring-gray-200">
            <div class="mx-auto h-20 w-20">
                <UserAvatar :user="callerUser" status />
            </div>
            <p class="mt-4 truncate text-lg font-semibold text-gray-900">
                {{ calls.incoming.callerName }}
            </p>
            <p class="text-sm text-gray-500">
                {{ t("calls.incoming.calling") }}
            </p>

            <div class="mt-6 flex items-center justify-center gap-8">
                <button
                    type="button"
                    @click="decline"
                    :title="t('calls.decline')"
                    class="inline-flex h-14 w-14 items-center justify-center rounded-full bg-red-600 text-white shadow-md hover:bg-red-500"
                >
                    <PhoneXMarkIcon class="h-6 w-6" />
                </button>
                <button
                    type="button"
                    @click="accept"
                    :title="t('calls.accept')"
                    class="inline-flex h-14 w-14 items-center justify-center rounded-full bg-green-600 text-white shadow-md hover:bg-green-500"
                >
                    <PhoneIcon class="h-6 w-6" />
                </button>
            </div>
        </div>
    </div>
</template>

<script setup>
import { computed, watch, onUnmounted } from "vue";
import { PhoneIcon, PhoneXMarkIcon } from "@heroicons/vue/24/solid";
import { t } from "@/i18n/index.js";
import UserAvatar from "@/components/UserAvatar.vue";
import chatService from "@/services/chatService";
import { useCallsStore } from "@/store/calls";
import { useChannelsStore } from "@/store/channels";
import { useUser } from "@/composables/useUser";

const calls = useCallsStore();
const channelsStore = useChannelsStore();

const caller = useUser(() => calls.incoming?.callerId);

const callerUser = computed(() => {
    const c = calls.incoming;

    if (!c) return { id: "" };

    return (
        caller.value || {
            id: c.callerId,
            name: c.callerName,
        }
    );
});

let timeout = null;
let ringtone = null;

function stopRingtone() {
    if (ringtone) {
        ringtone.pause();
        ringtone = null;
    }
}

watch(
    () => calls.incoming,
    (incoming) => {
        clearTimeout(timeout);
        stopRingtone();
        if (incoming) {
            ringtone = new Audio("/ringtone.mp3");
            ringtone.loop = true;
            ringtone.play().catch(() => {});
            timeout = setTimeout(() => calls.clearIncoming(), 45000);
        }
    },
);

onUnmounted(() => {
    clearTimeout(timeout);
    stopRingtone();
});

async function accept() {
    const c = calls.incoming;

    if (!c) return;
    calls.clearIncoming();
    try {
        const { data } = await chatService.answerCall(c.channelId);

        if (data.status === "connected") {
            channelsStore.openVideoChat(data.meeting);
        }
    } catch {
        // Caller hung up before we answered; nothing to join.
    }
}

function decline() {
    const c = calls.incoming;

    if (!c) return;
    calls.clearIncoming();
    chatService.declineCall(c.channelId).catch(() => {});
}
</script>
