<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div class="flex-1 relative min-h-0">
        <Teleport to="#video-minimized-root" v-if="channelsStore.isMinimized">
            <div class="w-full h-full relative">
                <button
                    @click="channelsStore.isMinimized = false"
                    class="absolute top-1 left-1 z-50 p-1 rounded bg-black/50 hover:bg-black/70 transition-colors"
                >
                    <ArrowTopRightOnSquareIcon class="h-3 w-3 text-white" />
                </button>
                <VideoTile
                    v-if="activeSpeaker"
                    :key="activeSpeaker.identity"
                    :participant="activeSpeaker"
                    :mirror-video="settings.mirrorVideo"
                    :is-host="participantIsHost(activeSpeaker)"
                    :guest-token="guestToken"
                    compact
                />
            </div>
        </Teleport>

        <div
            v-if="!sharedView && participants.length === 1"
            class="flex h-full w-full items-center justify-center p-2"
        >
            <div
                class="overflow-hidden rounded-2xl bg-gray-800 shadow-2xl ring-1 ring-white/10"
                :style="{
                    width: `min(calc(100vw - 16px), calc((100vh - 100px) * 16 / 9))`,
                    aspectRatio: '16 / 9',
                }"
            >
                <VideoTile
                    v-if="room?.localParticipant"
                    :participant="room?.localParticipant"
                    :mirror-video="settings.mirrorVideo"
                    :is-host="participantIsHost(room?.localParticipant)"
                    :guest-token="guestToken"
                    @share="emit('share', $event)"
                    @stop-share="emit('stopShare', $event)"
                />
            </div>
        </div>

        <div
            v-if="!sharedView && participants.length > 1"
            class="w-full h-full flex items-center justify-center bg-gray-950 p-2"
        >
            <div class="flex flex-col items-center gap-2">
                <div
                    v-for="(row, rowIndex) in gridRows"
                    :key="rowIndex"
                    class="flex justify-center gap-2"
                >
                    <div
                        v-for="(participant, colIndex) in row"
                        :key="colIndex"
                        class="overflow-hidden rounded-2xl bg-gray-800 ring-1 ring-white/5"
                        :style="{
                            width: `min(calc((100vw - ${(grid.columns + 1) * 8}px) / ${grid.columns}), calc((100vh - ${(grid.rows + 1) * 8 + 84}px) / ${grid.rows} * 16 / 9))`,
                            aspectRatio: '16 / 9',
                        }"
                    >
                        <VideoTile
                            :participant="participant"
                            :key="participant.identity"
                            :mirror-video="settings.mirrorVideo"
                            :is-host="participantIsHost(participant)"
                            :guest-token="guestToken"
                            @share="emit('share', $event)"
                            @stop-share="emit('stopShare', $event)"
                        />
                    </div>
                </div>
            </div>
        </div>

        <div v-if="sharedView" class="flex h-full w-full gap-2 pl-2">
            <div class="flex flex-1 min-w-0 items-center justify-center">
                <div
                    class="w-full h-full overflow-hidden rounded-xl bg-gray-900 ring-1 ring-white/5"
                >
                    <ShareScreenTile id="share-tile" />
                </div>
            </div>

            <div class="relative flex flex-col items-center w-52 py-2">
                <div
                    ref="sidebarScroll"
                    class="flex flex-col items-center gap-y-2 h-full min-h-0 overflow-y-auto hide-scrollbar w-full"
                >
                    <div
                        v-for="i in participants"
                        :key="i.identity"
                        class="w-full h-32 min-h-32 overflow-hidden rounded-xl bg-gray-800 ring-1 ring-white/5"
                    >
                        <VideoTile
                            :participant="i"
                            :mirror-video="settings.mirrorVideo"
                            :compact="true"
                            :is-host="participantIsHost(i)"
                            :guest-token="guestToken"
                            @share="emit('share', $event)"
                            @stop-share="emit('stopShare', $event)"
                        />
                    </div>
                </div>

                <div
                    v-show="canScrollUp"
                    class="pointer-events-none absolute top-0 inset-x-0 h-12 bg-gradient-to-b from-gray-950/80 to-transparent"
                />
                <div
                    v-show="canScrollDown"
                    class="pointer-events-none absolute bottom-0 inset-x-0 h-12 bg-gradient-to-t from-gray-950/80 to-transparent"
                />

                <button
                    v-show="canScrollUp"
                    type="button"
                    @click="scrollSidebar(-1)"
                    aria-label="Scroll up"
                    class="absolute top-2 left-1/2 -translate-x-1/2 z-10 p-1.5 rounded-full bg-gray-800/90 text-white shadow-lg ring-1 ring-white/10 backdrop-blur hover:bg-gray-700 transition-colors"
                >
                    <ChevronUpIcon class="size-4" aria-hidden="true" />
                </button>
                <button
                    v-show="canScrollDown"
                    type="button"
                    @click="scrollSidebar(1)"
                    aria-label="Scroll down"
                    class="absolute bottom-2 left-1/2 -translate-x-1/2 z-10 p-1.5 rounded-full bg-gray-800/90 text-white shadow-lg ring-1 ring-white/10 backdrop-blur hover:bg-gray-700 transition-colors"
                >
                    <ChevronDownIcon class="size-4" aria-hidden="true" />
                </button>
            </div>
        </div>
    </div>
</template>

<script setup>
import { toRef } from "vue";
import VideoTile from "@/components/Chat/Calls/VideoTile.vue";
import ShareScreenTile from "@/components/Chat/Calls/ShareScreenTile.vue";
import { useChannelsStore } from "@/store/channels";
import { useCallGrid } from "@/composables/chat/useCallGrid";
import { useCallSidebarScroll } from "@/composables/chat/useCallSidebarScroll";
import { ChevronUpIcon, ChevronDownIcon } from "@heroicons/vue/20/solid";
import { ArrowTopRightOnSquareIcon } from "@heroicons/vue/24/outline";

const props = defineProps({
    room: {
        type: Object,
        default: null,
    },
    participants: {
        type: Array,
        default: () => [],
    },
    activeSpeaker: {
        type: Object,
        default: null,
    },
    sharedView: {
        type: Boolean,
        default: false,
    },
    settings: {
        type: Object,
        required: true,
    },
    guestToken: {
        type: String,
        default: "",
    },
    participantIsHost: {
        type: Function,
        required: true,
    },
});

const emit = defineEmits(["share", "stopShare"]);

const channelsStore = useChannelsStore();

const participants = toRef(props, "participants");

const { grid, gridRows } = useCallGrid(participants);

const { sidebarScroll, canScrollUp, canScrollDown, scrollSidebar } = useCallSidebarScroll(
    toRef(props, "sharedView"),
    participants,
);
</script>

<style scoped>
.hide-scrollbar {
    scrollbar-width: none; /* Firefox */
    -ms-overflow-style: none; /* IE/Edge legacy */
}
.hide-scrollbar::-webkit-scrollbar {
    display: none; /* Chrome/Safari */
}
</style>
