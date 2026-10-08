<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div
        v-if="calls.outgoing"
        class="fixed inset-0 z-[60] flex items-center justify-center bg-black/40"
    >
        <div class="w-80 rounded-2xl bg-white p-6 text-center shadow-2xl ring-1 ring-gray-200">
            <div class="mx-auto h-20 w-20">
                <UserAvatar :user="calleeUser" status />
            </div>
            <p class="mt-4 text-xs font-semibold uppercase tracking-wide text-indigo-500">
                {{ t("calls.outgoing.calling") }}
            </p>
            <p class="mt-1 truncate text-lg font-semibold text-gray-900">
                {{ calleeName }}
            </p>

            <div class="mt-6 flex items-center justify-center">
                <button
                    type="button"
                    @click="cancel"
                    :title="t('calls.cancel')"
                    class="inline-flex h-14 w-14 items-center justify-center rounded-full bg-red-600 text-white shadow-md hover:bg-red-500"
                >
                    <PhoneXMarkIcon class="h-6 w-6" />
                </button>
            </div>
        </div>
    </div>
</template>

<script setup>
import { computed, watch, onUnmounted } from "vue";
import { PhoneXMarkIcon } from "@heroicons/vue/24/solid";
import { t } from "@/i18n/index.js";
import UserAvatar from "@/components/UserAvatar.vue";
import chatService from "@/services/chatService";
import { useCallsStore } from "@/store/calls";
import { useUser } from "@/composables/useUser";

const calls = useCallsStore();

const callee = useUser(() => calls.outgoing?.calleeId);

const calleeUser = computed(() => {
    const c = calls.outgoing;

    if (!c) return { id: "" };

    return (
        callee.value || {
            id: c.calleeId,
            name: c.name,
        }
    );
});

const calleeName = computed(() => {
    const u = callee.value;

    if (u) return `${u.name} ${u.lastname ?? ""}`.trim();

    return calls.outgoing?.name ?? "";
});

let timeout = null;
let ringback = null;

function stopRingback() {
    if (ringback) {
        ringback.pause();
        ringback = null;
    }

    clearTimeout(timeout);
}

watch(
    () => calls.outgoing,
    (outgoing) => {
        stopRingback();
        if (outgoing) {
            ringback = new Audio("/calling.mp3");
            ringback.loop = true;
            ringback.play().catch(() => {});
            timeout = setTimeout(cancel, 45000);
        }
    },
    { immediate: true },
);

onUnmounted(stopRingback);

function cancel() {
    const c = calls.outgoing;

    if (!c) return;
    calls.clearOutgoing();
    chatService.cancelCall(c.channelId).catch(() => {});
}
</script>
