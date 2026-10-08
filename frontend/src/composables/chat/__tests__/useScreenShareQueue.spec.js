// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { describe, it, expect, vi, afterEach } from "vitest";
import { effectScope, nextTick } from "vue";
import { Track } from "livekit-client";
import { useScreenShareQueue } from "@/composables/chat/useScreenShareQueue";

function sharer(identity) {
    const track = {
        attach: vi.fn(),
        detach: vi.fn(),
    };

    return {
        identity,
        track,
        trackPublications: new Map([["s", { source: Track.Source.ScreenShare }]]),
        getTrackPublication: (source) =>
            source === Track.Source.ScreenShare ? { track } : undefined,
    };
}

function shareTile() {
    const el = document.createElement("video");

    el.id = "share-tile";
    document.body.appendChild(el);

    return el;
}

async function flush() {
    await nextTick();
    await nextTick();
}

function setup() {
    const scope = effectScope();
    const queue = scope.run(() => useScreenShareQueue());

    return { scope, ...queue };
}

afterEach(() => {
    document.body.innerHTML = "";
});

describe("useScreenShareQueue", () => {
    it("attaches the first sharer and switches to the next when it stops", async () => {
        const tile = shareTile();
        const { scope, sharedView, addParticipantToShareQueue, removeParticipantFromShareQueue } =
            setup();

        const first = sharer("a");
        const second = sharer("b");

        addParticipantToShareQueue(first);
        addParticipantToShareQueue(second);
        await flush();

        expect(sharedView.value).toBe(true);
        expect(first.track.attach).toHaveBeenCalledWith(tile);
        expect(second.track.attach).not.toHaveBeenCalled();

        removeParticipantFromShareQueue(first);
        await flush();

        expect(first.track.detach).toHaveBeenCalledWith(tile);
        expect(second.track.attach).toHaveBeenCalledWith(tile);

        removeParticipantFromShareQueue(second);
        await flush();

        expect(second.track.detach).toHaveBeenCalledWith(tile);
        expect(sharedView.value).toBe(false);

        scope.stop();
    });

    it("waits for the share tile to render", async () => {
        vi.useFakeTimers();

        const { scope, addParticipantToShareQueue } = setup();
        const first = sharer("a");

        addParticipantToShareQueue(first);
        await flush();

        expect(first.track.attach).not.toHaveBeenCalled();

        const tile = shareTile();

        await vi.advanceTimersByTimeAsync(100);

        expect(first.track.attach).toHaveBeenCalledWith(tile);

        scope.stop();
        vi.useRealTimers();
    });

    it("ignores a second share event from the same participant", () => {
        const { scope, shareQueue, addParticipantToShareQueue } = setup();
        const first = sharer("a");

        addParticipantToShareQueue(first);
        addParticipantToShareQueue(first);

        expect(shareQueue.value).toHaveLength(1);

        scope.stop();
    });

    it("reports false for a participant without a screen share", () => {
        const { scope, isScreenShareEnabled } = setup();

        expect(isScreenShareEnabled({ trackPublications: new Map() })).toBe(false);
        expect(isScreenShareEnabled(sharer("a"))).toBe(true);

        scope.stop();
    });
});
