<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <TransitionRoot as="template" :show="true">
        <Dialog class="relative z-50" @close="emit('close')">
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
                <div class="flex min-h-full items-center justify-center p-4">
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
                            class="relative w-full max-w-lg transform overflow-hidden rounded-xl bg-white shadow-xl transition-all"
                        >
                            <div
                                class="flex items-start justify-between gap-3 border-b border-gray-200 px-6 py-5"
                            >
                                <div>
                                    <DialogTitle
                                        as="h3"
                                        class="text-base font-semibold leading-6 text-gray-900"
                                    >
                                        {{ t("meetings.picker.title") }}
                                    </DialogTitle>
                                    <p class="text-sm text-gray-500">
                                        {{ t("meetings.picker.subtitle") }}
                                    </p>
                                </div>
                                <button
                                    type="button"
                                    :title="t('common.button.close')"
                                    @click="emit('close')"
                                    class="-m-1.5 shrink-0 rounded-md p-1.5 text-gray-400 transition-colors hover:bg-gray-100 hover:text-gray-600"
                                >
                                    <XMarkIcon class="h-5 w-5" />
                                </button>
                            </div>

                            <div class="max-h-96 overflow-y-auto px-6 py-4">
                                <div v-if="active.length">
                                    <h4
                                        class="text-xs font-medium uppercase tracking-wide text-gray-400"
                                    >
                                        {{ t("meetings.picker.active") }}
                                    </h4>
                                    <ul class="mt-2 space-y-2">
                                        <li
                                            v-for="m in active"
                                            :key="m.id"
                                            class="flex items-center justify-between gap-3 rounded-lg ring-1 ring-gray-200 px-3 py-2"
                                        >
                                            <div class="min-w-0">
                                                <p
                                                    class="truncate text-sm font-medium text-gray-900"
                                                >
                                                    {{ m.title }}
                                                </p>
                                                <p class="text-xs text-gray-500">
                                                    {{
                                                        t("meetings.picker.in_call", {
                                                            count: m.participants,
                                                        })
                                                    }}
                                                </p>
                                            </div>
                                            <button
                                                type="button"
                                                @click="emit('join', m)"
                                                class="shrink-0 rounded-md bg-indigo-600 px-3 py-1.5 text-sm font-semibold text-white hover:bg-indigo-500"
                                            >
                                                {{ t("meetings.picker.join") }}
                                            </button>
                                        </li>
                                    </ul>
                                </div>

                                <div v-if="scheduled.length" class="mt-4">
                                    <h4
                                        class="text-xs font-medium uppercase tracking-wide text-gray-400"
                                    >
                                        {{ t("meetings.picker.scheduled") }}
                                    </h4>
                                    <ul class="mt-2 space-y-2">
                                        <li
                                            v-for="m in scheduled"
                                            :key="m.id"
                                            class="flex items-center justify-between gap-3 rounded-lg ring-1 ring-gray-200 px-3 py-2"
                                        >
                                            <div class="min-w-0">
                                                <p
                                                    class="truncate text-sm font-medium text-gray-900"
                                                >
                                                    {{ m.title }}
                                                </p>
                                                <p class="text-xs text-gray-500">
                                                    {{ formatTime(m.scheduled_at) }}
                                                </p>
                                            </div>
                                            <div
                                                v-if="isHost(m)"
                                                class="flex shrink-0 items-center gap-2"
                                            >
                                                <template v-if="confirmingId === m.id">
                                                    <span class="text-xs text-gray-500">
                                                        {{ t("meetings.picker.cancel_confirm") }}
                                                    </span>
                                                    <button
                                                        type="button"
                                                        :disabled="cancelling === m.id"
                                                        @click="cancelMeeting(m)"
                                                        class="rounded-md bg-red-600 px-2.5 py-1.5 text-xs font-semibold text-white hover:bg-red-500 disabled:opacity-50"
                                                    >
                                                        {{ t("meetings.picker.cancel_yes") }}
                                                    </button>
                                                    <button
                                                        type="button"
                                                        @click="confirmingId = ''"
                                                        class="rounded-md bg-white px-2.5 py-1.5 text-xs font-semibold text-gray-700 ring-1 ring-inset ring-gray-300 hover:bg-gray-50"
                                                    >
                                                        {{ t("meetings.picker.cancel_no") }}
                                                    </button>
                                                </template>
                                                <template v-else>
                                                    <button
                                                        type="button"
                                                        :disabled="starting === m.id"
                                                        @click="startMeeting(m)"
                                                        class="rounded-md bg-indigo-600 px-3 py-1.5 text-sm font-semibold text-white hover:bg-indigo-500 disabled:opacity-50"
                                                    >
                                                        {{ t("meetings.picker.start") }}
                                                    </button>
                                                    <button
                                                        type="button"
                                                        :title="t('meetings.picker.edit')"
                                                        @click="emit('edit', m)"
                                                        class="rounded-md p-1.5 text-gray-400 transition-colors hover:bg-gray-100 hover:text-gray-600"
                                                    >
                                                        <PencilSquareIcon class="h-4 w-4" />
                                                    </button>
                                                    <button
                                                        type="button"
                                                        :title="t('meetings.picker.cancel')"
                                                        @click="confirmingId = m.id"
                                                        class="rounded-md p-1.5 text-gray-400 transition-colors hover:bg-red-50 hover:text-red-600"
                                                    >
                                                        <TrashIcon class="h-4 w-4" />
                                                    </button>
                                                </template>
                                            </div>
                                            <span v-else class="shrink-0 text-xs text-gray-400">
                                                {{ t("meetings.picker.waiting_host") }}
                                            </span>
                                        </li>
                                    </ul>
                                </div>

                                <p
                                    v-if="!active.length && !scheduled.length"
                                    class="text-sm text-gray-400"
                                >
                                    {{ t("meetings.picker.none") }}
                                </p>
                            </div>

                            <div
                                class="flex flex-row-reverse gap-3 border-t border-gray-200 px-6 py-4"
                            >
                                <button
                                    type="button"
                                    @click="emit('new')"
                                    class="inline-flex items-center gap-2 rounded-md bg-indigo-600 px-4 py-2 text-sm font-semibold text-white hover:bg-indigo-500"
                                >
                                    <VideoCameraIcon class="h-4 w-4" />
                                    {{ t("meetings.picker.start_new") }}
                                </button>
                                <button
                                    type="button"
                                    @click="emit('schedule')"
                                    class="inline-flex items-center gap-2 rounded-md bg-white px-4 py-2 text-sm font-semibold text-gray-900 ring-1 ring-inset ring-gray-300 hover:bg-gray-50"
                                >
                                    <CalendarIcon class="h-4 w-4" />
                                    {{ t("meetings.picker.schedule") }}
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
import { ref } from "vue";
import { Dialog, DialogPanel, DialogTitle, TransitionChild, TransitionRoot } from "@headlessui/vue";
import {
    VideoCameraIcon,
    CalendarIcon,
    XMarkIcon,
    TrashIcon,
    PencilSquareIcon,
} from "@heroicons/vue/24/outline";
import { t } from "@/i18n";
import chatService from "@/services/chatService";
import { useUserStore } from "@/store/user";
import { useAlertStore } from "@/store/alerts";

const props = defineProps({
    channelId: {
        type: String,
        required: true,
    },
    active: {
        type: Array,
        default: () => [],
    },
    scheduled: {
        type: Array,
        default: () => [],
    },
});

const emit = defineEmits(["join", "new", "schedule", "edit", "changed", "close"]);

const userStore = useUserStore();
const alertStore = useAlertStore();

const starting = ref("");
const confirmingId = ref("");
const cancelling = ref("");

function isHost(meeting) {
    return meeting.host_id === userStore.user?.id;
}

function formatTime(unixSeconds) {
    return new Date(unixSeconds * 1000).toLocaleString();
}

async function startMeeting(meeting) {
    starting.value = meeting.id;
    try {
        const { data } = await chatService.startMeeting(props.channelId, meeting.id);

        emit("join", data);
    } catch {
        alertStore.showError(t.value("meetings.error.start"));
    } finally {
        starting.value = "";
    }
}

async function cancelMeeting(meeting) {
    cancelling.value = meeting.id;
    try {
        await chatService.cancelMeeting(props.channelId, meeting.id);
        confirmingId.value = "";
        emit("changed");
    } catch {
        alertStore.showError(t.value("meetings.error.cancel"));
    } finally {
        cancelling.value = "";
    }
}
</script>
