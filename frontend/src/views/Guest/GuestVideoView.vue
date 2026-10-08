<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div class="fixed inset-0 bg-gray-950">
        <!-- Validating -->
        <div v-if="state === 'loading'" class="flex h-full w-full items-center justify-center">
            <div class="flex items-center gap-3 text-gray-300">
                <svg
                    class="h-5 w-5 animate-spin"
                    xmlns="http://www.w3.org/2000/svg"
                    fill="none"
                    viewBox="0 0 24 24"
                >
                    <circle
                        class="opacity-25"
                        cx="12"
                        cy="12"
                        r="10"
                        stroke="currentColor"
                        stroke-width="4"
                    />
                    <path
                        class="opacity-75"
                        fill="currentColor"
                        d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
                    />
                </svg>
                <span class="text-sm">{{ t("guest_video.validating") }}</span>
            </div>
        </div>

        <!-- Invalid / expired link -->
        <div
            v-else-if="state === 'error'"
            class="flex h-full w-full items-center justify-center p-4"
        >
            <div class="w-full max-w-md rounded-xl bg-white p-8 text-center shadow-xl">
                <div
                    class="mx-auto mb-4 flex h-12 w-12 items-center justify-center rounded-full bg-red-100"
                >
                    <ExclamationTriangleIcon class="h-6 w-6 text-red-600" />
                </div>
                <h2 class="text-lg font-semibold text-gray-900">
                    {{ t("guest_video.invalid_title") }}
                </h2>
                <p class="mt-2 text-sm text-gray-500">{{ errorMessage }}</p>
            </div>
        </div>

        <!-- Name entry -->
        <div
            v-else-if="state === 'name'"
            class="flex h-full w-full items-center justify-center p-4"
        >
            <div class="w-full max-w-md rounded-xl bg-white p-8 shadow-xl">
                <div class="flex items-center gap-3">
                    <div
                        class="flex h-9 w-9 shrink-0 items-center justify-center rounded-lg bg-indigo-100"
                    >
                        <VideoCameraIcon class="h-5 w-5 text-indigo-600" />
                    </div>
                    <div>
                        <h2 class="text-base font-semibold text-gray-900">
                            {{
                                channelName
                                    ? t("guest_video.name_title_channel", {
                                          channel: channelName,
                                      })
                                    : t("guest_video.name_title")
                            }}
                        </h2>
                        <p class="text-sm text-gray-500">
                            {{ t("guest_video.name_subtitle") }}
                        </p>
                    </div>
                </div>

                <form class="mt-6" @submit.prevent="submitName">
                    <label for="guest-name" class="block text-sm font-medium text-gray-700">
                        {{ t("guest_video.name_label") }}
                    </label>
                    <input
                        id="guest-name"
                        v-model="name"
                        type="text"
                        autofocus
                        maxlength="80"
                        :placeholder="t('guest_video.name_placeholder')"
                        class="mt-1 block w-full rounded-md border-0 py-2 px-3 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm"
                    />

                    <div v-if="requiresPassword" class="mt-4">
                        <label for="guest-password" class="block text-sm font-medium text-gray-700">
                            {{ t("guest_video.password_label") }}
                        </label>
                        <input
                            id="guest-password"
                            v-model="password"
                            type="password"
                            :placeholder="t('guest_video.password_placeholder')"
                            class="mt-1 block w-full rounded-md border-0 py-2 px-3 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm"
                        />
                    </div>

                    <p v-if="nameError" class="mt-2 text-xs text-red-600">
                        {{ nameError }}
                    </p>

                    <button
                        type="submit"
                        :disabled="!name.trim() || (requiresPassword && !password)"
                        class="mt-6 inline-flex w-full items-center justify-center gap-2 rounded-md bg-indigo-600 px-4 py-2 text-sm font-semibold text-white shadow-sm hover:bg-indigo-500 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600 disabled:cursor-not-allowed disabled:opacity-50"
                    >
                        {{ t("guest_video.continue") }}
                    </button>
                </form>
            </div>
        </div>

        <!-- Scheduled: meeting hasn't started yet -->
        <div
            v-else-if="state === 'scheduled'"
            class="flex h-full w-full items-center justify-center p-4"
        >
            <div class="w-full max-w-md rounded-xl bg-white p-8 text-center shadow-xl">
                <div
                    class="mx-auto mb-4 flex h-12 w-12 items-center justify-center rounded-full bg-indigo-100"
                >
                    <ClockIcon class="h-6 w-6 text-indigo-600" />
                </div>
                <h2 class="text-lg font-semibold text-gray-900">
                    {{ meetingTitle || t("guest_video.scheduled_title") }}
                </h2>
                <p class="mt-2 text-sm text-gray-500">
                    {{ t("guest_video.scheduled_text") }}
                </p>
                <p v-if="scheduledAt" class="mt-1 text-sm text-gray-700">
                    {{ scheduledLabel() }}
                </p>
                <button
                    type="button"
                    @click="validate"
                    class="mt-6 inline-flex items-center justify-center rounded-md bg-indigo-600 px-4 py-2 text-sm font-semibold text-white shadow-sm hover:bg-indigo-500"
                >
                    {{ t("guest_video.check_again") }}
                </button>
            </div>
        </div>

        <!-- Device setup -->
        <WaitingRoom
            v-else-if="state === 'waiting'"
            :channel-name="meetingTitle || channelName"
            :participants="participants"
            @join="joinCall"
            @cancel="backToName"
        />

        <!-- In call -->
        <VideoView
            v-else-if="state === 'video'"
            :settings="callSettings"
            :video-token="videoToken"
            :host="host"
            :guest-token="token"
            standalone
            @close="endCall"
        />

        <!-- Call ended -->
        <div
            v-else-if="state === 'ended'"
            class="flex h-full w-full items-center justify-center p-4"
        >
            <div class="w-full max-w-md rounded-xl bg-white p-8 text-center shadow-xl">
                <h2 class="text-lg font-semibold text-gray-900">
                    {{ t("guest_video.ended_title") }}
                </h2>
                <p class="mt-2 text-sm text-gray-500">
                    {{ t("guest_video.ended_text") }}
                </p>
                <button
                    type="button"
                    @click="rejoin"
                    class="mt-6 inline-flex items-center justify-center rounded-md bg-indigo-600 px-4 py-2 text-sm font-semibold text-white shadow-sm hover:bg-indigo-500"
                >
                    {{ t("guest_video.rejoin") }}
                </button>
            </div>
        </div>
    </div>
</template>

<script setup>
import { ref, onMounted } from "vue";
import { useRoute } from "vue-router";
import { VideoCameraIcon, ExclamationTriangleIcon, ClockIcon } from "@heroicons/vue/24/outline";
import { t } from "@/i18n";
import guestVideoService from "@/services/guestVideoService";
import WaitingRoom from "@/components/Chat/Calls/WaitingRoom.vue";
import VideoView from "@/views/Chat/VideoView.vue";

const route = useRoute();

const state = ref("loading");
const errorMessage = ref("");
const channelName = ref("");
const meetingTitle = ref("");
const scheduledAt = ref(null);
const participants = ref(0);
const host = ref("");
const name = ref("");
const password = ref("");
const requiresPassword = ref(false);
const nameError = ref("");
const videoToken = ref("");
const callSettings = ref(null);

const token = route.params.token;

function scheduledLabel() {
    if (!scheduledAt.value) return "";

    return new Date(scheduledAt.value * 1000).toLocaleString();
}

async function validate() {
    try {
        const { data } = await guestVideoService.validate(token);

        channelName.value = data.channel_name;
        meetingTitle.value = data.meeting_title;
        scheduledAt.value = data.scheduled_at;
        participants.value = data.participants;
        host.value = data.host;
        requiresPassword.value = data.requires_password;
        state.value = data.status === "active" ? "name" : "scheduled";
    } catch (error) {
        errorMessage.value = error?.response?.data?.error || t.value("guest_video.invalid_text");
        state.value = "error";
    }
}

function submitName() {
    if (!name.value.trim()) return;
    if (requiresPassword.value && !password.value) return;
    nameError.value = "";
    state.value = "waiting";
}

function backToName() {
    state.value = "name";
}

async function joinCall(settings) {
    try {
        const { data } = await guestVideoService.getToken(token, name.value, password.value);

        videoToken.value = data.token;
        host.value = data.host || host.value;
        callSettings.value = settings;
        state.value = "video";
    } catch (error) {
        // Meeting may not have started yet, send the guest back to the lobby.
        if (error?.response?.status === 409) {
            state.value = "scheduled";

            return;
        }

        // Wrong password, back to the entry step with a message.
        if (error?.response?.status === 403) {
            nameError.value = error?.response?.data?.error || t.value("guest_video.password_error");
            state.value = "name";

            return;
        }

        errorMessage.value = error?.response?.data?.error || t.value("guest_video.join_error");
        state.value = "error";
    }
}

function endCall() {
    videoToken.value = "";
    callSettings.value = null;
    state.value = "ended";
}

function rejoin() {
    state.value = "name";
}

onMounted(validate);
</script>
