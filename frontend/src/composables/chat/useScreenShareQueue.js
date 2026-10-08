// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { ref, computed, watch, nextTick, toRaw } from "vue";
import { Track } from "livekit-client";

const SHARE_TILE_ID = "share-tile";
const SHARE_TILE_RETRIES = 20;
const SHARE_TILE_RETRY_MS = 50;

function wait(ms) {
    return new Promise((resolve) => setTimeout(resolve, ms));
}

function screenShareTrack(participant) {
    if (!participant) return null;

    return toRaw(participant).getTrackPublication(Track.Source.ScreenShare)?.track ?? null;
}

export function useScreenShareQueue() {
    const shareQueue = ref([]);
    const sharedView = computed(() => shareQueue.value.length > 0);

    let attachedTrack = null;
    let attachedElement = null;
    let showRequest = 0;

    function isScreenShareEnabled(participant) {
        for (const value of participant.trackPublications.values()) {
            if (value.source === Track.Source.ScreenShare) {
                return true;
            }
        }

        return false;
    }

    function detachShareTrack() {
        if (attachedTrack && attachedElement) {
            attachedTrack.detach(attachedElement);
        }

        attachedTrack = null;
        attachedElement = null;
    }

    async function showShareTrack(track) {
        const request = ++showRequest;

        detachShareTrack();

        if (!track) return;

        await nextTick();

        for (let attempt = 0; attempt < SHARE_TILE_RETRIES; attempt++) {
            if (request !== showRequest) return;

            const element = document.getElementById(SHARE_TILE_ID);

            if (element) {
                track.attach(element);
                attachedTrack = track;
                attachedElement = element;

                return;
            }

            await wait(SHARE_TILE_RETRY_MS);
        }
    }

    watch(() => screenShareTrack(shareQueue.value[0]), showShareTrack);

    function addParticipantToShareQueue(participant) {
        if (shareQueue.value.some((p) => p.identity === participant.identity)) return;

        shareQueue.value.push(participant);
    }

    function removeParticipantFromShareQueue(participant) {
        const index = shareQueue.value.findIndex((p) => p.identity === participant.identity);

        if (index >= 0) {
            shareQueue.value.splice(index, 1);
        }
    }

    return {
        sharedView,
        shareQueue,
        isScreenShareEnabled,
        addParticipantToShareQueue,
        removeParticipantFromShareQueue,
    };
}
