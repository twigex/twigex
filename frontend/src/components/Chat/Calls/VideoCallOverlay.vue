<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div v-if="channelStore.videoIsOpen">
        <WaitingRoom
            v-if="showWaitingRoom"
            :channel-name="channelStore.currentMeeting?.title ?? ''"
            :participants="channelStore.currentMeeting?.participants ?? 0"
            @join="joinVideoCall"
            @cancel="closeWaitingRoom"
        />

        <div v-else>
            <div
                v-show="!channelStore.isMinimized"
                class="fixed top-0 left-0 p-0 w-full h-full flex flex-col justify-start bg-black bg-opacity-100 z-50"
            >
                <VideoView
                    v-if="roomConnected"
                    :key="videoViewKey"
                    :settings="waitingRoomSettings"
                    :meeting="channelStore.currentMeeting"
                    @close="closeVideoCall"
                />
            </div>

            <div
                v-show="channelStore.isMinimized"
                id="video-minimized-root"
                class="fixed w-40 h-28 bg-gray-800 text-white rounded-md shadow-lg cursor-move z-50 overflow-hidden"
                :style="{ top: position.y + 'px', left: position.x + 'px' }"
                @mousedown="startDrag"
            ></div>
        </div>
    </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted, watch } from "vue";
import { useChannelsStore } from "@/store/channels";
import { useCallsStore } from "@/store/calls";
import VideoView from "@/views/Chat/VideoView.vue";
import WaitingRoom from "@/components/Chat/Calls/WaitingRoom.vue";

const channelStore = useChannelsStore();
const callsStore = useCallsStore();
const showWaitingRoom = ref(true);
const roomConnected = ref(false);
const videoViewKey = ref(0);
const waitingRoomSettings = ref(null);

// Position for the draggable square
const position = ref({
    x: window.innerWidth - 120,
    y: window.innerHeight - 120,
});

const BOX_WIDTH = 160; // w-40 = 10rem = 160px
const BOX_HEIGHT = 112; // h-28 = 7rem = 112px
const PADDING = 12; // 10-14px safe padding

let isDragging = false;
let offsetX = 0;
let offsetY = 0;

const joinVideoCall = (settings) => {
    waitingRoomSettings.value = settings;
    showWaitingRoom.value = false;
    roomConnected.value = true;
};

// Direct (1:1) calls skip the device lobby and join straight in, Discord-style.
function isDirectMeeting() {
    const id = channelStore.currentMeeting?.channel_id;

    return !!id && channelStore.channels.find((c) => c.id === id)?.type === "D";
}

watch(
    () => channelStore.videoIsOpen,
    (open) => {
        if (open && showWaitingRoom.value && isDirectMeeting()) {
            joinVideoCall({ cameraEnabled: true, microphoneEnabled: true });
        }
    },
);

const closeWaitingRoom = () => {
    channelStore.closeVideoChat();
    callsStore.clearOutgoing();

    showWaitingRoom.value = true;
    roomConnected.value = false;
    waitingRoomSettings.value = null;
};

const closeVideoCall = () => {
    //reset everythign
    callsStore.clearOutgoing();
    waitingRoomSettings.value = null;
    showWaitingRoom.value = true;
    roomConnected.value = false;
    videoViewKey.value++; // Force re-render
};

const clampPosition = () => {
    const maxX = window.innerWidth - BOX_WIDTH - PADDING;
    const maxY = window.innerHeight - BOX_HEIGHT - PADDING;

    position.value.x = Math.min(Math.max(PADDING, position.value.x), maxX);
    position.value.y = Math.min(Math.max(PADDING, position.value.y), maxY);
};

const startDrag = (e) => {
    isDragging = true;
    offsetX = e.clientX - position.value.x;
    offsetY = e.clientY - position.value.y;

    document.addEventListener("mousemove", onDrag);
    document.addEventListener("mouseup", stopDrag);
};

const onDrag = (e) => {
    if (!isDragging) return;

    const newX = e.clientX - offsetX;
    const newY = e.clientY - offsetY;

    position.value.x = newX;
    position.value.y = newY;

    clampPosition();
};

const stopDrag = () => {
    isDragging = false;
    document.removeEventListener("mousemove", onDrag);
    document.removeEventListener("mouseup", stopDrag);
};

const onResize = () => {
    clampPosition();
};

onMounted(() => {
    channelStore.isMinimized = false;

    window.addEventListener("resize", onResize);

    clampPosition();
});

onUnmounted(() => {
    document.removeEventListener("mousemove", onDrag);
    document.removeEventListener("mouseup", stopDrag);

    window.removeEventListener("resize", onResize);
});

watch(
    () => channelStore.isMinimized,
    (minimized) => {
        if (minimized) {
            clampPosition();
        }
    },
);
</script>
