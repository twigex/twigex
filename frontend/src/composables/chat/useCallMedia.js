// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { ref } from "vue";
import { VideoPresets, createLocalAudioTrack, createLocalVideoTrack, Track } from "livekit-client";

export function useCallMedia(getRoom, props) {
    const devices = ref([]);
    const activeAudioDevice = ref(null);
    const muted = ref(!props.settings.microphoneEnabled);
    const videoOn = ref(props.settings.cameraEnabled);
    const activeCameraDevice = ref(null);
    const previewStreams = new Map();

    async function initDevices() {
        try {
            devices.value = await navigator.mediaDevices.enumerateDevices();
        } catch {
            devices.value = [];
        }

        const mics = devices.value.filter((d) => d.kind === "audioinput");
        const cams = devices.value.filter((d) => d.kind === "videoinput");

        if (props.settings.cameraEnabled) {
            if (props.settings.selectedCamera) {
                activeCameraDevice.value = cams.find(
                    (d) => d.deviceId === props.settings.selectedCamera,
                );
            }
        }

        if (props.settings.microphoneEnabled) {
            if (props.settings.selectedMicrophone) {
                activeAudioDevice.value = mics.find(
                    (d) => d.deviceId === props.settings.selectedMicrophone,
                );
            }
        }
    }

    function getAudioInputDevices() {
        const microphones = devices.value.filter((device) => device.kind === "audioinput");

        return microphones;
    }

    function getCameraDevices() {
        const cameras = devices.value.filter((device) => device.kind === "videoinput");

        return cameras;
    }

    async function attachCameraPreview(el, deviceId) {
        if (!el) return;

        // Keyed by the pending request, so renders that ask again share one stream instead of
        // each opening the camera and leaving all but the last running.
        let request = previewStreams.get(deviceId);

        if (!request) {
            request = navigator.mediaDevices
                .getUserMedia({
                    video: { deviceId: { exact: deviceId } },
                    audio: false,
                })
                .catch((e) => {
                    console.warn("Camera preview failed", e);

                    return null;
                });

            previewStreams.set(deviceId, request);
        }

        const stream = await request;

        if (!stream || previewStreams.get(deviceId) !== request) return;

        if (el.srcObject !== stream) {
            el.srcObject = stream;
            el.play?.().catch(() => {});
        }
    }

    function stopCameraPreviews() {
        previewStreams.forEach((request) => {
            request.then((stream) => stream?.getTracks().forEach((t) => t.stop()));
        });
        previewStreams.clear();
    }

    async function enumerateDevices() {
        devices.value = await navigator.mediaDevices.enumerateDevices();

        if (!activeCameraDevice.value) {
            return;
        }

        const currentCamera = devices.value.find(
            (device) => device.deviceId === activeCameraDevice.value.deviceId,
        );

        if (currentCamera) {
            return;
        }

        const firstCamera = devices.value.find((device) => device.kind === "videoinput");

        if (!firstCamera) {
            activeCameraDevice.value = null;
            videoOn.value = false;

            return;
        }

        if (!videoOn.value) {
            activeCameraDevice.value = firstCamera;

            return;
        }

        await setActiveCameraDevice(firstCamera);
        await getRoom().localParticipant.setCameraEnabled(true);
    }

    async function setActiveAudioDevice(device) {
        activeAudioDevice.value = device;

        for (const pub of getRoom().localParticipant.getTrackPublications()) {
            if (pub.source === "microphone" && pub.track) {
                if (pub.track) {
                    await getRoom().localParticipant.unpublishTrack(pub.track);
                }
            }
        }

        await getRoom().switchActiveDevice("audioinput", device.deviceId);

        const newAudioTrack = await createLocalAudioTrack({
            deviceId: device.deviceId,
        });

        await getRoom().localParticipant.publishTrack(newAudioTrack);

        muted.value = false;
    }

    function localVideoElements() {
        const identity = getRoom().localParticipant.identity;

        return [...document.querySelectorAll("video")].filter((el) => el.id === identity);
    }

    async function setActiveCameraDevice(device) {
        activeCameraDevice.value = device;

        const existingCameraTrack = getRoom().localParticipant.getTrackPublication(
            Track.Source.Camera,
        )?.track;

        if (existingCameraTrack) {
            existingCameraTrack.stop();
            await getRoom().localParticipant.unpublishTrack(existingCameraTrack);
        }

        try {
            await getRoom().switchActiveDevice("videoinput", device.deviceId);

            const newVideoTrack = await createLocalVideoTrack({
                deviceId: device.deviceId,
                resolution: VideoPresets.h1080.resolution,
            });

            await getRoom().localParticipant.publishTrack(newVideoTrack);

            localVideoElements().forEach((videoElement) => {
                newVideoTrack.attach(videoElement);
            });
        } catch (error) {
            console.error("Error switching camera device:", error);
        }
    }

    function muteAudio() {
        getRoom().localParticipant.audioTrackPublications.forEach((track) => {
            track.track.mute();
        });

        muted.value = true;
    }

    async function unmuteAudio() {
        const publications = getRoom().localParticipant.audioTrackPublications;

        if (publications.size > 0) {
            publications.forEach((pub) => {
                pub.track?.unmute();
            });

            muted.value = false;

            return;
        }

        try {
            const newAudioTrack = await createLocalAudioTrack({
                deviceId: props.settings.selectedMicrophone || undefined,
            });

            await getRoom().localParticipant.publishTrack(newAudioTrack);

            muted.value = false;
        } catch (err) {
            console.error("Error enabling microphone track:", err);
            muted.value = true;
        }
    }

    async function shareScreen() {
        const localParticipant = getRoom().localParticipant;

        if (localParticipant.isScreenShareEnabled) {
            for (const publication of localParticipant.trackPublications.values()) {
                if (
                    publication.source === "screen_share" ||
                    publication.source === "screen_share_audio"
                ) {
                    publication.track.stop();
                    localParticipant.unpublishTrack(publication.track);
                }
            }

            return;
        }

        try {
            const tracks = await localParticipant.createScreenTracks({
                audio: true,
            });

            for (const track of tracks) {
                await localParticipant.publishTrack(track);
            }
        } catch (err) {
            console.error("Screen sharing failed:", err);
        }
    }

    async function disableCamera() {
        const devices = await navigator.mediaDevices.enumerateDevices();
        const hasCamera = devices.some((d) => d.kind === "videoinput");

        if (!hasCamera) {
            console.warn("No camera devices found");
            videoOn.value = false;

            return;
        }

        if (getRoom().localParticipant.isCameraEnabled) {
            await getRoom().localParticipant.setCameraEnabled(false);
            videoOn.value = false;

            return;
        }

        //get video track! If undefined create new
        const publications = getRoom().localParticipant.getTrackPublications();
        const hasActiveVideo = [...publications].some((pub) => pub.kind === "video" && pub.track);

        if (!hasActiveVideo) {
            try {
                const newVideoTrack = await createLocalVideoTrack();

                await getRoom().localParticipant.publishTrack(newVideoTrack, {
                    source: Track.Source.Camera,
                });

                // ensure camera is "enabled" on the participant
                await getRoom().localParticipant.setCameraEnabled(false);
                await getRoom().localParticipant.setCameraEnabled(true);

                videoOn.value = true;
            } catch (err) {
                console.error("Error enabling camera track:", err);
            }

            return;
        }

        await getRoom().localParticipant.setCameraEnabled(true);
        videoOn.value = true;
    }

    return {
        devices,
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
    };
}
