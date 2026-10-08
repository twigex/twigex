<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div ref="audioContainer" style="display: none"></div>

    <div v-if="loaded" class="flex flex-col h-full w-full bg-gray-950 relative">
        <VideoCallStage
            :room="room"
            :participants="participants"
            :active-speaker="activeSpeaker"
            :shared-view="sharedView"
            :settings="settings"
            :guest-token="guestToken"
            :participant-is-host="participantIsHost"
            @share="addParticipantToShareQueue"
            @stop-share="removeParticipantFromShareQueue"
        />

        <div aria-live="assertive" class="flex items-center justify-center px-4 py-4">
            <VideoCallControls
                v-if="room"
                :settings="settings"
                :screen-sharing="localScreenSharing"
                :video-on="videoOn"
                :muted="muted"
                :camera-devices="getCameraDevices()"
                :audio-input-devices="getAudioInputDevices()"
                :active-camera-device="activeCameraDevice"
                :active-audio-device="activeAudioDevice"
                :attach-camera-preview="attachCameraPreview"
                :stop-camera-previews="stopCameraPreviews"
                :is-host="isHost"
                :standalone="standalone"
                :is-direct-call="isDirectCall"
                :promotable-members="promotableMembers"
                @share-screen="shareScreen()"
                @toggle-camera="disableCamera()"
                @select-camera="setActiveCameraDevice"
                @mute="muteAudio()"
                @unmute="unmuteAudio()"
                @select-microphone="setActiveAudioDevice"
                @invite="showInviteDialog = true"
                @make-host="makeHost"
                @minimize="minimizeWindow()"
                @leave="onLeaveClick()"
            />
        </div>

        <GuestInviteDialog
            v-if="showInviteDialog && !standalone && meeting"
            :channel-id="meeting.channel_id"
            :meeting-id="meeting.id"
            @close="showInviteDialog = false"
        />

        <VideoCallLeaveDialog
            :open="showLeaveOptions"
            @close="showLeaveOptions = false"
            @end-for-all="endMeetingForAll()"
            @keep-running="leaveAndKeepRunning()"
        />
    </div>
</template>

<script setup>
import { ref, shallowRef, onMounted, onUnmounted, computed, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import VideoCallStage from "@/components/Chat/Calls/VideoCallStage.vue";
import VideoCallControls from "@/components/Chat/Calls/VideoCallControls.vue";
import VideoCallLeaveDialog from "@/components/Chat/Calls/VideoCallLeaveDialog.vue";
import GuestInviteDialog from "@/components/Chat/Meetings/GuestInviteDialog.vue";
import { useSettingsStore } from "@/store/settings";
import chatService from "@/services/chatService";
import { useChannelsStore } from "@/store/channels";
import { useUserStore } from "@/store/user";
import { useAlertStore } from "@/store/alerts";
import { useCallsStore } from "@/store/calls";
import { useCallMedia } from "@/composables/chat/useCallMedia";
import { useScreenShareQueue } from "@/composables/chat/useScreenShareQueue";
import { t } from "@/i18n";
import { Room, RoomEvent, VideoPresets, ScreenSharePresets, Track } from "livekit-client";

const props = defineProps({
    settings: {
        type: Object,
        default: () => ({
            cameraEnabled: true,
            microphoneEnabled: true,
            selectedCamera: null,
            selectedMicrophone: null,
            mirrorVideo: true,
        }),
    },
    videoToken: {
        type: String,
        default: "",
    },
    host: {
        type: String,
        default: "",
    },
    standalone: {
        type: Boolean,
        default: false,
    },
    guestToken: {
        type: String,
        default: "",
    },
    meeting: {
        type: Object,
        default: null,
    },
});

const emits = defineEmits(["close"]);

const room = shallowRef(null);
const channelsStore = useChannelsStore();
const settingsStore = useSettingsStore();
const userStore = useUserStore();
const alertStore = useAlertStore();
const callsStore = useCallsStore();
const route = useRoute();
const router = useRouter();
const token = ref("");
const meeting = ref(props.meeting);
const showLeaveOptions = ref(false);
let endingMeeting = false;
let loadTimer = null;
const participants = ref([]);
const activeSpeaker = ref(null);

// Stop the caller's ringback once someone else joins the room.
watch(
    () => participants.value.length,
    (count) => {
        if (count > 1) callsStore.clearOutgoing();
    },
);

const {
    activeAudioDevice,
    activeCameraDevice,
    muted,
    videoOn,
    initDevices,
    getAudioInputDevices,
    getCameraDevices,
    attachCameraPreview,
    stopCameraPreviews,
    enumerateDevices,
    setActiveAudioDevice,
    setActiveCameraDevice,
    muteAudio,
    unmuteAudio,
    shareScreen,
    disableCamera,
} = useCallMedia(() => room.value, props);

const {
    sharedView,
    shareQueue,
    isScreenShareEnabled,
    addParticipantToShareQueue,
    removeParticipantFromShareQueue,
} = useScreenShareQueue();

const localScreenSharing = computed(() => {
    const identity = room.value?.localParticipant?.identity;

    return !!identity && shareQueue.value.some((p) => p.identity === identity);
});

const loaded = ref(false);
const audioContainer = ref(null);
const audioTracks = new Map();
const showInviteDialog = ref(false);

const isHost = computed(() => !!meeting.value && meeting.value.host_id === userStore.user?.id);

function participantIsHost(p) {
    return !!meeting.value && p?.identity === meeting.value.host_id;
}

// A 1:1 direct-message call has no "leave vs end" choice and ends as soon as either side leaves.
const isDirectCall = computed(() => {
    const id = meeting.value?.channel_id;

    return !!id && channelsStore.channels.find((c) => c.id === id)?.type === "D";
});

async function endDirectCall() {
    // A 1:1 call ends for both sides the moment either leaves, so end it
    // server-side whether or not we are the host. Leave locally first, or
    // deleting the room disconnects us before cleanup can release the camera.
    const m = meeting.value;

    leaveRoom();
    if (props.standalone || !m) {
        return;
    }

    try {
        await chatService.endCall(m.channel_id);
    } catch {
        // Best effort: the peer's client and the empty-room sweep still end it.
    }
}

const promotableMembers = computed(() =>
    participants.value.filter(
        (p) => p.identity !== userStore.user?.id && !p.identity.startsWith("guest-"),
    ),
);

async function load() {
    room.value = new Room({
        adaptiveStream: true,
        dynacast: true,
        videoCaptureDefaults: {
            resolution: VideoPresets.h1080.resolution,
        },
        audioCaptureDefaults: {
            autoGainControl: false,
            noiseSuppression: true,
            echoCancellation: true,
            voiceIsolation: true,
            channelCount: 1,
            sampleRate: 48000,
        },
        publishDefaults: {
            screenShareEncoding: ScreenSharePresets.h1080fps30.encoding,
        },
    });

    const joining = room.value;

    room.value.on(RoomEvent.TrackSubscribed, (track) => {
        if (track.kind === Track.Kind.Audio) {
            if (!audioContainer.value) return;

            if (audioTracks.has(track.sid)) return;

            const audioEl = track.attach();

            audioEl.autoplay = true;
            audioEl.playsInline = true;

            audioContainer.value.appendChild(audioEl);
            audioTracks.set(track.sid, audioEl);
        }
    });

    room.value.on(RoomEvent.TrackUnsubscribed, (track) => {
        if (track.kind !== Track.Kind.Audio) return;

        const existing = audioTracks.get(track.sid);

        if (existing) {
            track.detach(existing);
            existing.remove();
            audioTracks.delete(track.sid);
        }
    });

    room.value.on(RoomEvent.ParticipantConnected, (participant) => {
        addParticipant(participant);
        if (isHost.value) announcePresence();
    });

    room.value.on(RoomEvent.ParticipantDisconnected, (participant) => {
        removeParticipant(participant);

        if (activeSpeaker.value?.identity === participant.identity) {
            activeSpeaker.value =
                room.value.localParticipant ??
                [...room.value.remoteParticipants.values()][0] ??
                null;
        }

        // In a 1:1 call, end it for me once the other side leaves.
        if (isDirectCall.value && room.value.remoteParticipants.size === 0) {
            endDirectCall();
        }

        if (isHost.value) announcePresence();
    });

    room.value.on(RoomEvent.ActiveSpeakersChanged, (speakers) => {
        if (speakers.length > 0) {
            activeSpeaker.value = speakers[0];
        }
    });

    room.value.on(RoomEvent.DataReceived, (payload, _participant, _kind, topic) => {
        if (topic && topic !== "meeting") return;
        let msg;

        try {
            msg = JSON.parse(new TextDecoder().decode(payload));
        } catch {
            return;
        }

        if (msg.type === "meeting_host_changed" && meeting.value) {
            meeting.value = { ...meeting.value, host_id: msg.host_id };
        }
    });

    // Fired when the host ends the meeting (the room is deleted server-side).
    // Intentional leaves remove listeners first, so this only catches remote ends.
    room.value.on(RoomEvent.Disconnected, () => {
        stopLocalTracks();
        if (!props.standalone && !endingMeeting) {
            alertStore.showError(t.value("meetings.ended_notice"));
            channelsStore.closeVideoChat();
        }

        emits("close");
    });

    room.value.on(RoomEvent.MediaDevicesChanged, () => {
        enumerateDevices();
    });

    room.value.on(RoomEvent.MediaDevicesError, (error) => {
        console.error("media devices error", error);
    });

    try {
        const host = props.host || settingsStore.chatSettings.host;

        await room.value.connect(host, token.value);
    } catch (error) {
        console.error("Connection error:", error);

        return;
    }

    if (leftWhileJoining(joining)) return;

    announcePresence();

    const devices = await navigator.mediaDevices.enumerateDevices();

    const hasCamera = devices.some((d) => d.kind === "videoinput");

    if (props.settings.cameraEnabled && hasCamera) {
        try {
            if (props.settings.selectedCamera) {
                await room.value.switchActiveDevice("videoinput", props.settings.selectedCamera);
            }

            await room.value.localParticipant.setCameraEnabled(true);
        } catch (e) {
            console.warn("Camera unavailable, joining without video", e);
        }
    }

    if (leftWhileJoining(joining)) return;

    const hasMic = devices.some((d) => d.kind === "audioinput");

    /*
    if (hasMic) {
        try {
            if (props.settings.selectedMicrophone) {
                await room.value.switchActiveDevice(
                    "audioinput",
                    props.settings.selectedMicrophone,
                );
            }
            if (props.settings.microphoneEnabled) {
                await room.value.localParticipant.setMicrophoneEnabled(true);
                muted.value = false;
            } else {
                muted.value = true;
            }
        } catch (e) {
            console.warn("Microphone unavailable, joining muted", e);
            await room.value.localParticipant.setMicrophoneEnabled(false);
            muted.value = true;
        }
    }
    */

    if (props.settings.microphoneEnabled && hasMic) {
        try {
            if (props.settings.selectedMicrophone) {
                await room.value.switchActiveDevice(
                    "audioinput",
                    props.settings.selectedMicrophone,
                );
            }

            await room.value.localParticipant.setMicrophoneEnabled(true);
            muted.value = false;
        } catch (e) {
            console.warn("Microphone unavailable, joining muted", e);
            muted.value = true;
        }
    }

    if (leftWhileJoining(joining)) return;

    addParticipant(room.value.localParticipant);
    activeSpeaker.value = room.value.localParticipant;

    room.value.remoteParticipants.forEach((p) => {
        if (isScreenShareEnabled(p)) {
            addParticipantToShareQueue(p);
        }

        addParticipant(p);
    });

    loaded.value = true;
}

function minimizeWindow() {
    channelsStore.isMinimized = true;
}

function addParticipant(participant) {
    let index = participants.value.findIndex((p) => p.identity === participant.identity);

    if (index >= 0) {
        participants.value[index] = participant;

        return;
    }

    participants.value.push(participant);
}

function removeParticipant(participant) {
    let index = participants.value.findIndex((p) => p.identity === participant.identity);

    if (index >= 0) {
        participants.value.splice(index, 1);
    }
}

// Stop every local track so the device (and the browser's in-use indicator) is released.
function stopLocalTracks() {
    if (!room.value || !room.value.localParticipant) return;
    room.value.localParticipant.trackPublications.forEach((pub) => {
        pub.track?.stop();
    });
}

// Leaving clears the room while load() is still awaiting, so tracks it turned on after that stay live.
function leftWhileJoining(joining) {
    if (room.value === joining) return false;

    joining.localParticipant.trackPublications.forEach((pub) => {
        pub.track?.stop();
    });

    return true;
}

function cleanupRoom() {
    stopCameraPreviews();

    if (!room.value || !room.value.localParticipant) return;

    stopLocalTracks();

    audioTracks.forEach((el) => {
        el.remove();
    });
    audioTracks.clear();

    room.value.removeAllListeners();
    room.value.disconnect();
    room.value = null;
}

function leaveRoom() {
    announcePresence();
    cleanupRoom();

    emits("close");

    if (props.standalone) {
        return;
    }

    channelsStore.closeVideoChat();

    router.push({
        name: "chat",
        params: { id: route.params.chatId },
    });
}

function onLeaveClick() {
    if (isDirectCall.value) {
        endDirectCall();

        return;
    }

    if (isHost.value && !props.standalone) {
        showLeaveOptions.value = true;

        return;
    }

    leaveRoom();
}

async function endMeetingForAll() {
    endingMeeting = true;
    showLeaveOptions.value = false;

    // Leave locally before ending server-side; ending first deletes the room and disconnects us before cleanup can release the camera.
    const m = meeting.value;

    leaveRoom();

    if (m) {
        try {
            await chatService.endMeeting(m.channel_id, m.id);
        } catch {
            alertStore.showError(t.value("meetings.error.end"));
        }
    }
}

async function leaveAndKeepRunning() {
    if (meeting.value) {
        try {
            await chatService.setMeetingHost(meeting.value.channel_id, meeting.value.id, "");
        } catch {
            // non-fatal: meeting continues, host auto-promotion just didn't apply
        }
    }

    showLeaveOptions.value = false;
    leaveRoom();
}

function announcePresence() {
    if (!meeting.value) return;
    chatService.announceMeetingPresence(meeting.value.channel_id, meeting.value.id).catch(() => {});
}

async function makeHost(participant) {
    if (!meeting.value) return;
    try {
        const { data } = await chatService.setMeetingHost(
            meeting.value.channel_id,
            meeting.value.id,
            participant.identity,
        );

        meeting.value = data;
        alertStore.showSuccess(t.value("meetings.host.transferred"));
    } catch {
        alertStore.showError(t.value("meetings.error.host"));
    }
}

onMounted(async () => {
    await initDevices();

    if (props.videoToken) {
        token.value = props.videoToken;
        loadTimer = setTimeout(() => {
            load();
        }, 1000);

        return;
    }

    if (!props.meeting) {
        return;
    }

    chatService
        .getMeetingVideoToken(props.meeting.channel_id, props.meeting.id)
        .then((response) => {
            token.value = response.data;
            loadTimer = setTimeout(() => {
                load();
            }, 1000);
        })
        .catch(() => {
            alertStore.showError(t.value("meetings.error.join"));
            emits("close");
            channelsStore.closeVideoChat();
        });
});

onUnmounted(() => {
    clearTimeout(loadTimer);
    cleanupRoom();
});
</script>
