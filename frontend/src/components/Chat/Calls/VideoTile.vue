<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div
        class="relative h-full w-full object-cover items-center justify-center flex"
        :class="{
            'bg-black': video,
            'bg-gray-900': !video,
        }"
    >
        <video
            v-show="video"
            ref="videoElement"
            :id="props.participant.identity"
            class="absolute h-full w-full object-cover rounded-md transition-all duration-200"
            :class="{
                'mirror-video': shouldMirror,
                'border-transparent': !speaking,
                'border-indigo-500 ring-2 ring-indigo-500/40': speaking,
            }"
        ></video>

        <img
            v-if="!isGuest && !photoFailed"
            v-show="!video"
            class="inline-block rounded-full border-2 transition-all duration-200"
            :class="{
                'border-transparent': !speaking,
                'border-indigo-500 ring-2 ring-indigo-500/40': speaking,
                'w-48 h-48': !props.compact,
                'w-20 h-20': props.compact,
            }"
            :src="photoSrc"
            alt=""
            @error="photoFailed = true"
        />

        <div
            v-if="!isGuest && photoFailed"
            v-show="!video"
            class="inline-flex items-center justify-center rounded-full border-2 text-white transition-all duration-200"
            :class="[
                colorClass,
                {
                    'border-transparent': !speaking,
                    'border-indigo-500 ring-2 ring-indigo-500/40': speaking,
                    'w-48 h-48': !props.compact,
                    'w-20 h-20': props.compact,
                },
            ]"
        >
            <svg class="h-full w-full" viewBox="0 0 10 10" aria-hidden="true">
                <text
                    x="5"
                    y="5"
                    text-anchor="middle"
                    dominant-baseline="central"
                    font-size="6"
                    font-weight="400"
                    fill="currentColor"
                >
                    {{ initial }}
                </text>
            </svg>
        </div>

        <div
            v-show="!video && isGuest"
            class="inline-flex items-center justify-center rounded-full border-2 bg-gray-700 transition-all duration-200"
            :class="{
                'border-transparent': !speaking,
                'border-indigo-500 ring-2 ring-indigo-500/40': speaking,
                'w-48 h-48': !props.compact,
                'w-20 h-20': props.compact,
            }"
        >
            <UserIcon
                :class="props.compact ? 'w-10 h-10' : 'w-24 h-24'"
                class="text-gray-300"
                aria-hidden="true"
            />
        </div>

        <div
            v-if="!props.compact"
            class="absolute bottom-0 inset-x-0 flex items-end justify-between px-2.5 py-2 bg-gradient-to-t from-black/70 to-transparent"
        >
            <div
                class="flex items-center gap-1.5 bg-black/50 backdrop-blur-sm rounded-lg px-2 py-1"
            >
                <WifiIcon
                    class="w-4 h-4"
                    :class="isConnected ? 'text-green-400' : 'text-red-400'"
                />
                <MicrophoneIcon v-if="!micMuted" class="w-4 h-4 text-white" />
                <div v-else class="relative w-4 h-4">
                    <MicrophoneIcon class="w-4 h-4 text-red-400" />
                    <SlashIcon
                        class="w-4 h-4 text-red-400 absolute inset-0 -rotate-[20deg] scale-x-[-1]"
                    />
                </div>
                <VideoCameraIcon v-if="!camMuted" class="w-4 h-4 text-white" />
                <VideoCameraSlashIcon v-if="camMuted" class="w-4 h-4 text-red-400" />
                <ComputerDesktopIcon v-if="isScreenSharing" class="w-4 h-4 text-blue-400" />
                <SpeakerWaveIcon v-if="speaking" class="w-4 h-4 text-indigo-400 animate-pulse" />
            </div>

            <div
                class="flex items-center gap-1 bg-black/50 backdrop-blur-sm rounded-lg px-2.5 py-1 text-white text-sm font-medium"
            >
                <StarIcon
                    v-if="props.isHost"
                    class="h-3.5 w-3.5 shrink-0 text-amber-400"
                    title="Host"
                />
                <span>
                    {{ displayName }}
                </span>
            </div>
        </div>

        <div
            v-if="props.compact"
            class="absolute top-1 right-1 flex gap-1 items-center bg-black/60 backdrop-blur rounded px-1.5 py-1"
        >
            <StarIcon v-if="props.isHost" class="w-3 h-3 text-amber-400" />

            <MicrophoneIcon class="w-3 h-3" :class="micMuted ? 'text-red-400' : 'text-white'" />

            <component
                :is="camMuted ? VideoCameraSlashIcon : VideoCameraIcon"
                class="w-3 h-3"
                :class="camMuted ? 'text-gray-400' : 'text-white'"
            />

            <ComputerDesktopIcon v-if="isScreenSharing" class="w-3 h-3 text-blue-400" />

            <SpeakerWaveIcon v-if="speaking" class="w-3 h-3 text-indigo-400 animate-pulse" />
        </div>
    </div>
</template>

<script setup>
import { onMounted, onBeforeUnmount, watch, ref, computed } from "vue";

import { Track, ParticipantEvent, LocalParticipant, ConnectionQuality } from "livekit-client";
import { useUserStore } from "@/store/user";
import { useUser } from "@/composables/useUser";
import { colorClassFor, initialsFor } from "@/utils/avatar";

import {
    MicrophoneIcon,
    SlashIcon,
    VideoCameraIcon,
    VideoCameraSlashIcon,
    ComputerDesktopIcon,
    WifiIcon,
    SpeakerWaveIcon,
} from "@heroicons/vue/24/outline";
import { StarIcon, UserIcon } from "@heroicons/vue/24/solid";

const props = defineProps({
    participant: Object,
    video: Boolean,

    mirrorVideo: {
        type: Boolean,
        default: false,
    },

    compact: {
        type: Boolean,
        default: false,
    },

    isHost: {
        type: Boolean,
        default: false,
    },

    // Set only for a guest-link join, where no user lookup may be attempted: a
    // 401 from one resets the stores and redirects out of the call.
    guestToken: {
        type: String,
        default: "",
    },
});

const emits = defineEmits(["stop-share", "share"]);

const userStore = useUserStore();

// Guests have no user record or photo, so show a placeholder icon instead of
// requesting a photo for an identity no user endpoint can resolve.
const isGuest = computed(() => (props.participant?.identity ?? "").startsWith("guest-"));

const user = useUser(() => (isGuest.value || props.guestToken ? null : props.participant.identity));

const photoFailed = ref(false);

const colorClass = computed(() => colorClassFor(props.participant.identity));

const photoSrc = computed(() => {
    const identity = props.participant.identity;

    if (props.guestToken) {
        return `/api/guest/video/${encodeURIComponent(props.guestToken)}/photo/${encodeURIComponent(identity)}`;
    }

    return userStore.getPhotoSrc(identity, user.value?.photo);
});

watch(photoSrc, () => {
    photoFailed.value = false;
});

// Falls back to the LiveKit token name so guests (no user record) and org users
// not in the local store still render instead of crashing on a missing user.
const displayName = computed(() => {
    if (user.value) {
        return (
            `${user.value.name ?? ""} ${user.value.lastname ?? ""}`.trim() ||
            props.participant.name ||
            ""
        );
    }

    return props.participant.name || "";
});

const initial = computed(() => initialsFor(displayName.value));

const speaking = ref(false);
const video = ref(props.participant.isCameraEnabled);
const videoElement = ref(null);

const isConnected = ref(true);
const micMuted = ref(false);
const camMuted = ref(false);
const isScreenSharing = ref(false);
const tileListeners = [];

const shouldMirror = computed(() => {
    return (
        props.mirrorVideo === true &&
        props.participant instanceof LocalParticipant &&
        video.value === true &&
        isScreenSharing.value === false
    );
});

function trackListener(event, handler) {
    tileListeners.push({ event, handler });

    return handler;
}

onMounted(() => {
    setup();

    props.participant
        .on(
            ParticipantEvent.LocalTrackPublished,
            trackListener(ParticipantEvent.LocalTrackPublished, (trackPub) => {
                if (!videoElement.value || !trackPub.track) return;

                if (trackPub.kind === Track.Kind.Video && trackPub.source === Track.Source.Camera) {
                    trackPub.track.attach(videoElement.value);
                    video.value = true;
                    camMuted.value = false;
                }

                if (
                    trackPub.kind === Track.Kind.Video &&
                    trackPub.source === Track.Source.ScreenShare
                ) {
                    isScreenSharing.value = true;
                    emits("share", props.participant);
                }

                if (
                    trackPub.kind === Track.Kind.Audio &&
                    trackPub.source === Track.Source.Microphone
                ) {
                    micMuted.value = trackPub.isMuted === true;
                }
            }),
        )
        .on(
            ParticipantEvent.LocalTrackUnpublished,
            trackListener(ParticipantEvent.LocalTrackUnpublished, (trackPub) => {
                if (
                    trackPub.kind === Track.Kind.Video &&
                    trackPub.source === Track.Source.ScreenShare
                ) {
                    isScreenSharing.value = false;
                    emits("stop-share", props.participant);
                }
            }),
        )
        .on(
            ParticipantEvent.TrackPublished,
            trackListener(ParticipantEvent.TrackPublished, (trackPub) => {
                if (trackPub.kind === Track.Kind.Video) {
                    video.value = true;
                } else if (
                    trackPub.kind === Track.Kind.Audio &&
                    trackPub.source === Track.Source.Microphone
                ) {
                    micMuted.value = trackPub.isMuted === true;
                }
            }),
        )
        .on(
            ParticipantEvent.TrackUnpublished,
            trackListener(ParticipantEvent.TrackUnpublished, (trackPub) => {
                if (trackPub.kind === Track.Kind.Video) {
                    if (trackPub.source === Track.Source.ScreenShare) {
                        isScreenSharing.value = false;
                        emits("stop-share", props.participant);
                    } else {
                        video.value = false;
                    }
                }
            }),
        )
        .on(
            ParticipantEvent.TrackSubscribed,
            trackListener(ParticipantEvent.TrackSubscribed, (trackPub, remoteTrackPub) => {
                if (remoteTrackPub.kind === Track.Kind.Video) {
                    if (remoteTrackPub.source === Track.Source.ScreenShare) {
                        isScreenSharing.value = true;
                        emits("share", props.participant);
                    } else {
                        remoteTrackPub.track.attach(videoElement.value);
                    }
                }
            }),
        )
        .on(
            ParticipantEvent.TrackUnsubscribed,
            trackListener(ParticipantEvent.TrackUnsubscribed, (trackPub) => {
                if (trackPub.track && trackPub.track.kind === Track.Kind.Video) {
                    if (trackPub.track.source === Track.Source.ScreenShare) {
                        isScreenSharing.value = false;
                        emits("stop-share", props.participant);
                    } else {
                        trackPub.track.detach(videoElement.value);
                    }
                }
            }),
        )
        .on(
            ParticipantEvent.IsSpeakingChanged,
            trackListener(ParticipantEvent.IsSpeakingChanged, (isSpeaking) => {
                speaking.value = isSpeaking;
            }),
        )
        .on(
            ParticipantEvent.TrackMuted,
            trackListener(ParticipantEvent.TrackMuted, (trackPub) => {
                if (trackPub.kind === Track.Kind.Video) {
                    camMuted.value = true;
                    video.value = false;
                }

                if (trackPub.kind === Track.Kind.Audio) {
                    micMuted.value = true;
                }
            }),
        )
        .on(
            ParticipantEvent.TrackUnmuted,
            trackListener(ParticipantEvent.TrackUnmuted, (trackPub) => {
                if (trackPub.kind === Track.Kind.Video) {
                    camMuted.value = false;
                    video.value = true;
                }

                if (trackPub.kind === Track.Kind.Audio) {
                    micMuted.value = false;
                }
            }),
        )
        .on(
            ParticipantEvent.ConnectionQualityChanged,
            trackListener(ParticipantEvent.ConnectionQualityChanged, (quality) => {
                isConnected.value = quality !== ConnectionQuality.Lost;
            }),
        );
});

function removeTileListeners() {
    tileListeners.forEach(({ event, handler }) => {
        props.participant.off(event, handler);
    });

    tileListeners.length = 0;
}

onBeforeUnmount(() => {
    removeTileListeners();
});

function setup() {
    if (!videoElement.value) return;

    const element = videoElement.value;

    props.participant.getTrackPublications()?.forEach((pub) => {
        if (pub.track) {
            pub.track.detach(element);
        }
    });

    const videoTrack = props.participant.getTrackPublication(Track.Source.Camera);

    if (videoTrack && videoTrack.track) {
        videoTrack.track.attach(element);
    }

    const micPub = props.participant.getTrackPublication(Track.Source.Microphone);

    micMuted.value = micPub ? micPub.isMuted === true : true;
}
</script>

<style scoped>
.mirror-video {
    transform: scaleX(-1);
}
</style>
