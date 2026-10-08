// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { describe, it, expect, vi, afterEach } from "vitest";
import { useCallMedia } from "@/composables/chat/useCallMedia";

const props = {
    settings: {
        cameraEnabled: true,
        microphoneEnabled: true,
        selectedCamera: null,
        selectedMicrophone: null,
    },
};

function fakeRoom() {
    return {
        switchActiveDevice: vi.fn(),
        localParticipant: {
            getTrackPublication: vi.fn(() => undefined),
            setCameraEnabled: vi.fn(),
            unpublishTrack: vi.fn(),
            publishTrack: vi.fn(),
        },
    };
}

function stubDevices(devices) {
    vi.stubGlobal("navigator", {
        mediaDevices: {
            enumerateDevices: vi.fn().mockResolvedValue(devices),
        },
    });
}

afterEach(() => {
    vi.unstubAllGlobals();
});

describe("useCallMedia enumerateDevices", () => {
    it("leaves the camera off when no camera is left", async () => {
        stubDevices([{ kind: "audioinput", deviceId: "mic" }]);

        const room = fakeRoom();
        const { enumerateDevices, videoOn, activeCameraDevice } = useCallMedia(() => room, props);

        activeCameraDevice.value = { kind: "videoinput", deviceId: "gone" };
        videoOn.value = true;

        await enumerateDevices();

        expect(room.switchActiveDevice).not.toHaveBeenCalled();
        expect(room.localParticipant.setCameraEnabled).not.toHaveBeenCalled();
        expect(videoOn.value).toBe(false);
        expect(activeCameraDevice.value).toBeNull();
    });

    it("keeps the current camera when it is still present", async () => {
        const camera = { kind: "videoinput", deviceId: "cam" };

        stubDevices([camera]);

        const room = fakeRoom();
        const { enumerateDevices, activeCameraDevice } = useCallMedia(() => room, props);

        activeCameraDevice.value = camera;

        await enumerateDevices();

        expect(room.localParticipant.setCameraEnabled).not.toHaveBeenCalled();
        expect(activeCameraDevice.value).toEqual(camera);
    });

    it("does not touch the camera when none was chosen", async () => {
        stubDevices([{ kind: "videoinput", deviceId: "cam" }]);

        const room = fakeRoom();
        const { enumerateDevices, activeCameraDevice } = useCallMedia(() => room, props);

        await enumerateDevices();

        expect(room.switchActiveDevice).not.toHaveBeenCalled();
        expect(room.localParticipant.setCameraEnabled).not.toHaveBeenCalled();
        expect(activeCameraDevice.value).toBeNull();
    });

    it("keeps the camera off when the chosen camera is replaced", async () => {
        const camera = { kind: "videoinput", deviceId: "cam" };

        stubDevices([camera]);

        const room = fakeRoom();
        const { enumerateDevices, videoOn, activeCameraDevice } = useCallMedia(() => room, props);

        activeCameraDevice.value = { kind: "videoinput", deviceId: "gone" };
        videoOn.value = false;

        await enumerateDevices();

        expect(room.switchActiveDevice).not.toHaveBeenCalled();
        expect(room.localParticipant.setCameraEnabled).not.toHaveBeenCalled();
        expect(videoOn.value).toBe(false);
        expect(activeCameraDevice.value).toEqual(camera);
    });
});

function fakeStream() {
    const track = { stop: vi.fn() };

    return {
        track,
        getTracks: () => [track],
    };
}

function stubCamera(getUserMedia) {
    vi.stubGlobal("navigator", {
        mediaDevices: {
            enumerateDevices: vi.fn().mockResolvedValue([]),
            getUserMedia,
        },
    });
}

function videoElement() {
    return { srcObject: null, play: vi.fn(() => Promise.resolve()) };
}

describe("useCallMedia camera previews", () => {
    it("opens one stream when a camera preview is attached twice at once", async () => {
        const stream = fakeStream();
        const getUserMedia = vi.fn().mockResolvedValue(stream);

        stubCamera(getUserMedia);

        const { attachCameraPreview } = useCallMedia(() => fakeRoom(), props);
        const first = videoElement();
        const second = videoElement();

        await Promise.all([attachCameraPreview(first, "cam"), attachCameraPreview(second, "cam")]);

        expect(getUserMedia).toHaveBeenCalledTimes(1);
        expect(first.srcObject).toBe(stream);
        expect(second.srcObject).toBe(stream);
    });

    it("stops a preview that arrives after the previews were stopped", async () => {
        const stream = fakeStream();
        let resolve;

        stubCamera(vi.fn(() => new Promise((r) => (resolve = r))));

        const { attachCameraPreview, stopCameraPreviews } = useCallMedia(() => fakeRoom(), props);
        const el = videoElement();

        const attaching = attachCameraPreview(el, "cam");

        stopCameraPreviews();
        resolve(stream);
        await attaching;
        await Promise.resolve();

        expect(stream.track.stop).toHaveBeenCalled();
        expect(el.srcObject).toBeNull();
    });
});
