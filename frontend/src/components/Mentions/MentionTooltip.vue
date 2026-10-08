<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div
        v-if="open"
        ref="panel"
        class="z-50 rounded-md border bg-white shadow-lg p-3 w-64 max-w-64"
        role="dialog"
    >
        <div class="flex flex-row justify-between items-start">
            <div class="flex">
                <div class="w-10 h-10">
                    <UserAvatar :user="user" status />
                </div>
                <div class="ml-3">
                    <p
                        class="text-sm font-medium text-gray-700 group-hover:text-gray-900 w-36 truncate"
                    >
                        {{ user.name + " " + user.lastname }}
                    </p>
                    <p
                        class="text-xs font-medium text-gray-500 group-hover:text-gray-700 w-36 truncate"
                    >
                        {{ "@" + user.username }}
                    </p>
                    <p class="text-xs font-medium text-gray-500 group-hover:text-gray-700">
                        Last activity:
                        {{ getLastActivity(userStore.getUserStatus(user.id).last_activity) }}
                    </p>
                </div>
            </div>

            <button
                @click="close"
                type="button"
                class="rounded-md bg-white text-gray-400 hover:text-gray-500 focus:outline-2 focus:outline-offset-2 focus:outline-indigo-600"
            >
                <span class="sr-only">Close</span>
                <XMarkIcon class="size-6" aria-hidden="true" />
            </button>
        </div>

        <div class="flex flex-row border-t pt-3 mt-5 justify-between">
            <button
                v-if="user != null && userStore.user.id != user.id"
                @click="onSendMessage"
                type="button"
                class="inline-flex items-center gap-x-1.5 rounded-md bg-indigo-600 px-2.5 py-1.5 text-sm font-semibold text-white shadow-xs hover:bg-indigo-500 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600 dark:bg-indigo-500 dark:shadow-none dark:hover:bg-indigo-400 dark:focus-visible:outline-indigo-500"
            >
                <PaperAirplaneIcon class="-ml-0.5 size-5" aria-hidden="true" />
                Send message
            </button>
            <button
                v-if="user != null && userStore.user.id == user.id"
                @click="
                    router.push({ name: 'profile' });
                    close();
                "
                type="button"
                class="inline-flex items-center gap-x-1.5 rounded-md bg-indigo-600 px-2.5 py-1.5 text-sm font-semibold text-white shadow-xs hover:bg-indigo-500 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600 dark:bg-indigo-500 dark:shadow-none dark:hover:bg-indigo-400 dark:focus-visible:outline-indigo-500"
            >
                <PencilIcon class="-ml-0.5 size-5" aria-hidden="true" />
                Edit profile
            </button>
        </div>
    </div>
</template>

<script setup>
import { ref, watch, onMounted, onBeforeUnmount } from "vue";
import { useRouter } from "vue-router";
import UserAvatar from "@/components/UserAvatar.vue";
import { PaperAirplaneIcon, PencilIcon, XMarkIcon } from "@heroicons/vue/24/outline";
import { useUserStore } from "@/store/user";
import { useChannelsStore } from "@/store/channels";
import { channelTypes } from "@/constants/channels";
import useDateOperations from "@/composables/useDateOperations";
import chatService from "@/services/chatService";

const props = defineProps({
    open: { type: Boolean, default: false },
    user: { type: Object, default: null },
});
const emit = defineEmits(["close"]);

const { getLastActivity } = useDateOperations();

const panel = ref(null);
const userStore = useUserStore();
const channelsStore = useChannelsStore();
const router = useRouter();

function close() {
    emit("close");
}

function onSendMessage() {
    for (let i = 0; i < channelsStore.channels.length; i++) {
        if (channelsStore.channels[i].type === channelTypes.Direct) {
            //If channel has you and user, open it
            const members = channelsStore.channels[i].channel_members;

            let me = members.find((m) => m.user_id === userStore.user.id);
            let him = members.find((m) => m.user_id === props.user.id);

            if (me && him) {
                channelsStore.setCurrentChannel(channelsStore.channels[i]);

                router.push({
                    name: "chat",
                    params: { chatId: channelsStore.channels[i].id },
                });

                close();

                return;
            }
        }
    }

    //Create new DM channel
    chatService
        .createChannel({
            type: "direct",
            name: props.user.name + props.user.email,
            description: "",
            user: props.user.id,
        })
        .then((response) => {
            channelsStore.setChannels([response.data]);
            channelsStore.setCurrentChannel(response.data);

            router.push({
                name: "chat",
                params: { chatId: response.data.id },
            });
        });

    close();
}

function onDocPointer(e) {
    if (!props.open) return;
    const t = e.target;

    if (panel.value && !panel.value.contains(t)) close();
}

function onKey(e) {
    if (!props.open) return;
    if (e.key === "Escape") close();
}

function addGlobalListeners() {
    document.addEventListener("mousedown", onDocPointer, true);
    document.addEventListener("touchstart", onDocPointer, true);
    document.addEventListener("keydown", onKey);
}

function removeGlobalListeners() {
    document.removeEventListener("mousedown", onDocPointer, true);
    document.removeEventListener("touchstart", onDocPointer, true);
    document.removeEventListener("keydown", onKey);
}

watch(
    () => props.open,
    (open) => {
        if (open) addGlobalListeners();
        else removeGlobalListeners();
    },
    { immediate: true },
);

onMounted(() => {
    if (props.open) addGlobalListeners();
});
onBeforeUnmount(removeGlobalListeners);
</script>
