// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { describe, it, expect, beforeEach, vi } from "vitest";
import { setActivePinia, createPinia } from "pinia";
import { useCallsStore } from "@/store/calls";
import { handleEvent } from "@/js/websocket";

vi.mock("@/services/userService", () => ({
    default: {
        byUsernames: () => Promise.resolve({ data: [] }),
        byIds: () => Promise.resolve({ data: [] }),
        statuses: () => Promise.resolve({ data: [] }),
    },
}));

vi.mock("@/services/chatService", () => ({ default: {} }));
vi.mock("@/services/notificationService", () => ({ default: {} }));

let calls;

beforeEach(() => {
    setActivePinia(createPinia());
    calls = useCallsStore();
});

function ringing(channelId = "channel-1") {
    calls.setIncoming({
        channelId,
        callerId: "user-caller",
        callerName: "Caller",
    });
}

// The ringtone follows calls.incoming, so clearing it is what silences a tab.
describe("call_handled", () => {
    it("stops the ring when the call was answered in another tab", () => {
        ringing("channel-1");

        handleEvent({
            event: "call_handled",
            app: "chat",
            data: { channel_id: "channel-1" },
        });

        expect(calls.incoming).toBeNull();
    });

    it("leaves a ring for a different channel alone", () => {
        ringing("channel-1");

        handleEvent({
            event: "call_handled",
            app: "chat",
            data: { channel_id: "channel-2" },
        });

        expect(calls.incoming).not.toBeNull();
    });

    it("is a no-op in the tab that already acted", () => {
        handleEvent({
            event: "call_handled",
            app: "chat",
            data: { channel_id: "channel-1" },
        });

        expect(calls.incoming).toBeNull();
        expect(calls.outgoing).toBeNull();
    });
});

describe("call_connected", () => {
    // channelsStore is only wired up by initialize(), so a tab that wrongly
    // tried to join the room would throw here.
    it("is ignored by a tab that did not place the call", () => {
        expect(() =>
            handleEvent({
                event: "call_connected",
                app: "chat",
                data: {
                    channel_id: "channel-1",
                    meeting_id: "meeting-1",
                    title: "Call",
                },
            }),
        ).not.toThrow();
    });

    it("is ignored by a tab calling on a different channel", () => {
        calls.setOutgoing({ channelId: "channel-2", calleeId: "user-callee" });

        expect(() =>
            handleEvent({
                event: "call_connected",
                app: "chat",
                data: { channel_id: "channel-1", meeting_id: "meeting-1" },
            }),
        ).not.toThrow();
        expect(calls.outgoing).not.toBeNull();
    });
});
