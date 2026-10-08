<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <TransitionRoot as="template" :show="true">
        <Dialog class="relative z-50" @close="handleCancel">
            <TransitionChild
                as="template"
                enter="ease-out duration-300"
                enter-from="opacity-0"
                enter-to="opacity-100"
                leave="ease-in duration-200"
                leave-from="opacity-100"
                leave-to="opacity-0"
            >
                <div class="fixed inset-0 bg-gray-500/75 transition-opacity" />
            </TransitionChild>

            <div class="fixed inset-0 z-10 w-screen overflow-y-auto">
                <div class="flex min-h-full items-center justify-center p-4 sm:p-6">
                    <TransitionChild
                        as="template"
                        enter="ease-out duration-300"
                        enter-from="opacity-0 translate-y-4 sm:translate-y-0 sm:scale-95"
                        enter-to="opacity-100 translate-y-0 sm:scale-100"
                        leave="ease-in duration-200"
                        leave-from="opacity-100 translate-y-0 sm:scale-100"
                        leave-to="opacity-0 translate-y-4 sm:translate-y-0 sm:scale-95"
                    >
                        <DialogPanel
                            class="relative w-full max-w-3xl transform overflow-hidden rounded-xl bg-white shadow-xl transition-all"
                        >
                            <!-- Header -->
                            <div class="border-b border-gray-200 px-6 py-5">
                                <div class="flex items-center gap-3">
                                    <div
                                        class="flex h-9 w-9 shrink-0 items-center justify-center rounded-lg bg-indigo-100"
                                    >
                                        <VideoCameraIcon class="h-5 w-5 text-indigo-600" />
                                    </div>
                                    <div>
                                        <DialogTitle
                                            as="h3"
                                            class="text-base font-semibold leading-6 text-gray-900"
                                        >
                                            {{
                                                channelName
                                                    ? t("meetings.waiting_room.title_channel", {
                                                          channel: channelName,
                                                      })
                                                    : t("meetings.waiting_room.title")
                                            }}
                                        </DialogTitle>
                                        <p class="text-sm text-gray-500">
                                            {{ t("meetings.waiting_room.subtitle") }}
                                        </p>
                                    </div>
                                </div>
                            </div>

                            <!-- Body -->
                            <div class="grid grid-cols-1 gap-8 p-6 lg:grid-cols-2">
                                <!-- Left: Video preview -->
                                <div class="space-y-4">
                                    <div
                                        class="relative w-full h-[258px] overflow-hidden rounded-xl bg-gray-100 ring-1 ring-gray-200"
                                    >
                                        <video
                                            ref="videoPreview"
                                            autoplay
                                            playsinline
                                            muted
                                            class="w-full h-full object-contain"
                                            :class="{
                                                '-scale-x-100': mirrorVideo,
                                            }"
                                        />
                                        <div
                                            v-if="!cameraEnabled"
                                            class="absolute inset-0 flex flex-col items-center justify-center bg-gray-100"
                                        >
                                            <UserCircleIcon class="mb-2 h-20 w-20 text-gray-300" />
                                            <span class="text-sm text-gray-400">
                                                {{ t("meetings.waiting_room.camera_off") }}
                                            </span>
                                        </div>
                                    </div>

                                    <!-- Mirror toggle -->
                                    <div
                                        class="flex items-center justify-between rounded-lg px-1 py-2"
                                    >
                                        <div class="flex items-center gap-2">
                                            <ArrowsRightLeftIcon class="h-4 w-4 text-gray-400" />
                                            <span class="text-sm text-gray-700">
                                                {{ t("meetings.waiting_room.mirror") }}
                                            </span>
                                        </div>
                                        <Switch
                                            v-model="mirrorVideo"
                                            :class="[
                                                mirrorVideo ? 'bg-indigo-600' : 'bg-gray-200',
                                                'relative inline-flex h-6 w-11 shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors duration-200 ease-in-out focus:outline-none focus:ring-2 focus:ring-indigo-600 focus:ring-offset-2',
                                            ]"
                                        >
                                            <span
                                                aria-hidden="true"
                                                :class="[
                                                    mirrorVideo ? 'translate-x-5' : 'translate-x-0',
                                                    'pointer-events-none inline-block h-5 w-5 transform rounded-full bg-white shadow ring-0 transition duration-200 ease-in-out',
                                                ]"
                                            />
                                        </Switch>
                                    </div>
                                </div>

                                <!-- Right: Settings -->
                                <div class="space-y-6">
                                    <!-- Camera -->
                                    <div class="space-y-3">
                                        <div class="flex items-center justify-between">
                                            <div class="flex items-center gap-2">
                                                <VideoCameraIcon class="h-4 w-4 text-gray-500" />
                                                <label class="text-sm font-medium text-gray-700">
                                                    {{ t("meetings.waiting_room.camera") }}
                                                </label>
                                            </div>
                                            <Switch
                                                v-model="cameraEnabled"
                                                @click="toggleCamera"
                                                :class="[
                                                    cameraEnabled ? 'bg-indigo-600' : 'bg-gray-200',
                                                    'relative inline-flex h-6 w-11 shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors duration-200 ease-in-out focus:outline-none focus:ring-2 focus:ring-indigo-600 focus:ring-offset-2',
                                                ]"
                                            >
                                                <span
                                                    aria-hidden="true"
                                                    :class="[
                                                        cameraEnabled
                                                            ? 'translate-x-5'
                                                            : 'translate-x-0',
                                                        'pointer-events-none inline-block h-5 w-5 transform rounded-full bg-white shadow ring-0 transition duration-200 ease-in-out',
                                                    ]"
                                                />
                                            </Switch>
                                        </div>

                                        <Listbox
                                            v-model="selectedCamera"
                                            :disabled="!cameraEnabled"
                                        >
                                            <div class="relative">
                                                <ListboxButton
                                                    class="relative w-full cursor-default rounded-md bg-white py-2 pl-3 pr-10 text-left text-sm shadow-sm ring-1 ring-inset ring-gray-300 focus:outline-none focus:ring-2 focus:ring-indigo-600 disabled:cursor-not-allowed disabled:bg-gray-50 disabled:text-gray-400 disabled:ring-gray-200"
                                                >
                                                    <span
                                                        class="block truncate text-gray-900"
                                                        :class="{
                                                            'text-gray-400': !cameraEnabled,
                                                        }"
                                                    >
                                                        {{
                                                            getDeviceName(
                                                                selectedCamera,
                                                                cameraDevices,
                                                            ) ||
                                                            t("meetings.waiting_room.select_camera")
                                                        }}
                                                    </span>
                                                    <span
                                                        class="pointer-events-none absolute inset-y-0 right-0 flex items-center pr-2"
                                                    >
                                                        <ChevronUpDownIcon
                                                            class="h-5 w-5 text-gray-400"
                                                            aria-hidden="true"
                                                        />
                                                    </span>
                                                </ListboxButton>

                                                <ListboxOptions
                                                    class="absolute z-10 mt-1 max-h-60 w-full overflow-auto rounded-md bg-white py-1 text-base shadow-lg ring-1 ring-black/5 focus:outline-none sm:text-sm"
                                                >
                                                    <ListboxOption
                                                        v-for="device in cameraDevices"
                                                        :key="device.deviceId"
                                                        v-slot="{ active, selected }"
                                                        as="template"
                                                        :value="device.deviceId"
                                                    >
                                                        <li
                                                            :class="[
                                                                active
                                                                    ? 'bg-indigo-600 text-white'
                                                                    : 'text-gray-900',
                                                                'relative cursor-default select-none py-2 pl-3 pr-9',
                                                            ]"
                                                        >
                                                            <span
                                                                :class="[
                                                                    selected
                                                                        ? 'font-semibold'
                                                                        : 'font-normal',
                                                                    'block truncate',
                                                                ]"
                                                            >
                                                                {{
                                                                    device.label ||
                                                                    t(
                                                                        "meetings.waiting_room.camera_fallback",
                                                                        {
                                                                            id: device.deviceId.slice(
                                                                                0,
                                                                                8,
                                                                            ),
                                                                        },
                                                                    )
                                                                }}
                                                            </span>
                                                            <span
                                                                v-if="selected"
                                                                :class="[
                                                                    active
                                                                        ? 'text-white'
                                                                        : 'text-indigo-600',
                                                                    'absolute inset-y-0 right-0 flex items-center pr-4',
                                                                ]"
                                                            >
                                                                <CheckIcon
                                                                    class="h-5 w-5"
                                                                    aria-hidden="true"
                                                                />
                                                            </span>
                                                        </li>
                                                    </ListboxOption>
                                                </ListboxOptions>
                                            </div>
                                        </Listbox>
                                    </div>

                                    <!-- Microphone -->
                                    <div class="space-y-3">
                                        <div class="flex items-center justify-between">
                                            <div class="flex items-center gap-2">
                                                <MicrophoneIcon class="h-4 w-4 text-gray-500" />
                                                <label class="text-sm font-medium text-gray-700">
                                                    {{ t("meetings.waiting_room.microphone") }}
                                                </label>
                                            </div>
                                            <Switch
                                                v-model="microphoneEnabled"
                                                @click="toggleMicrophone"
                                                :class="[
                                                    microphoneEnabled
                                                        ? 'bg-indigo-600'
                                                        : 'bg-gray-200',
                                                    'relative inline-flex h-6 w-11 shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors duration-200 ease-in-out focus:outline-none focus:ring-2 focus:ring-indigo-600 focus:ring-offset-2',
                                                ]"
                                            >
                                                <span
                                                    aria-hidden="true"
                                                    :class="[
                                                        microphoneEnabled
                                                            ? 'translate-x-5'
                                                            : 'translate-x-0',
                                                        'pointer-events-none inline-block h-5 w-5 transform rounded-full bg-white shadow ring-0 transition duration-200 ease-in-out',
                                                    ]"
                                                />
                                            </Switch>
                                        </div>

                                        <Listbox
                                            v-model="selectedMicrophone"
                                            :disabled="!microphoneEnabled"
                                        >
                                            <div class="relative">
                                                <ListboxButton
                                                    class="relative w-full cursor-default rounded-md bg-white py-2 pl-3 pr-10 text-left text-sm shadow-sm ring-1 ring-inset ring-gray-300 focus:outline-none focus:ring-2 focus:ring-indigo-600 disabled:cursor-not-allowed disabled:bg-gray-50 disabled:text-gray-400 disabled:ring-gray-200"
                                                >
                                                    <span
                                                        class="block truncate text-gray-900"
                                                        :class="{
                                                            'text-gray-400': !microphoneEnabled,
                                                        }"
                                                    >
                                                        {{
                                                            getDeviceName(
                                                                selectedMicrophone,
                                                                microphoneDevices,
                                                            ) ||
                                                            t(
                                                                "meetings.waiting_room.select_microphone",
                                                            )
                                                        }}
                                                    </span>
                                                    <span
                                                        class="pointer-events-none absolute inset-y-0 right-0 flex items-center pr-2"
                                                    >
                                                        <ChevronUpDownIcon
                                                            class="h-5 w-5 text-gray-400"
                                                            aria-hidden="true"
                                                        />
                                                    </span>
                                                </ListboxButton>

                                                <ListboxOptions
                                                    class="absolute z-10 mt-1 max-h-60 w-full overflow-auto rounded-md bg-white py-1 text-base shadow-lg ring-1 ring-black/5 focus:outline-none sm:text-sm"
                                                >
                                                    <ListboxOption
                                                        v-for="device in microphoneDevices"
                                                        :key="device.deviceId"
                                                        v-slot="{ active, selected }"
                                                        as="template"
                                                        :value="device.deviceId"
                                                    >
                                                        <li
                                                            :class="[
                                                                active
                                                                    ? 'bg-indigo-600 text-white'
                                                                    : 'text-gray-900',
                                                                'relative cursor-default select-none py-2 pl-3 pr-9',
                                                            ]"
                                                        >
                                                            <span
                                                                :class="[
                                                                    selected
                                                                        ? 'font-semibold'
                                                                        : 'font-normal',
                                                                    'block truncate',
                                                                ]"
                                                            >
                                                                {{
                                                                    device.label ||
                                                                    t(
                                                                        "meetings.waiting_room.microphone_fallback",
                                                                        {
                                                                            id: device.deviceId.slice(
                                                                                0,
                                                                                8,
                                                                            ),
                                                                        },
                                                                    )
                                                                }}
                                                            </span>
                                                            <span
                                                                v-if="selected"
                                                                :class="[
                                                                    active
                                                                        ? 'text-white'
                                                                        : 'text-indigo-600',
                                                                    'absolute inset-y-0 right-0 flex items-center pr-4',
                                                                ]"
                                                            >
                                                                <CheckIcon
                                                                    class="h-5 w-5"
                                                                    aria-hidden="true"
                                                                />
                                                            </span>
                                                        </li>
                                                    </ListboxOption>
                                                </ListboxOptions>
                                            </div>
                                        </Listbox>
                                    </div>

                                    <!-- Device status -->
                                    <div
                                        class="flex items-center justify-between rounded-lg bg-gray-50 px-3 py-2.5"
                                    >
                                        <div class="flex items-center gap-2">
                                            <div
                                                class="h-2.5 w-2.5 rounded-full"
                                                :class="
                                                    settingsLoaded
                                                        ? 'animate-pulse bg-green-500'
                                                        : 'bg-yellow-400'
                                                "
                                            />
                                            <span class="text-sm text-gray-600">
                                                {{
                                                    settingsLoaded
                                                        ? t("meetings.waiting_room.devices_ready")
                                                        : t("meetings.waiting_room.devices_loading")
                                                }}
                                            </span>
                                        </div>
                                        <button
                                            @click="refreshDevices"
                                            class="rounded-md p-1.5 text-gray-400 transition-colors hover:bg-gray-200 hover:text-gray-600"
                                            :title="t('meetings.waiting_room.refresh_devices')"
                                        >
                                            <ArrowPathIcon class="h-4 w-4" />
                                        </button>
                                    </div>

                                    <!-- Participant count -->
                                    <p class="text-center text-xs text-gray-400">
                                        {{
                                            t("meetings.waiting_room.participants", {
                                                count: participantCount,
                                            })
                                        }}
                                    </p>
                                </div>
                            </div>

                            <!-- Footer -->
                            <div
                                class="flex flex-row-reverse gap-3 border-t border-gray-200 px-6 py-4"
                            >
                                <button
                                    type="button"
                                    @click="joinCall"
                                    :disabled="!settingsLoaded"
                                    class="inline-flex items-center justify-center gap-2 rounded-md bg-indigo-600 px-4 py-2 text-sm font-semibold text-white shadow-sm hover:bg-indigo-500 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600 disabled:cursor-not-allowed disabled:opacity-50"
                                >
                                    <svg
                                        v-if="!settingsLoaded"
                                        class="h-4 w-4 animate-spin"
                                        xmlns="http://www.w3.org/2000/svg"
                                        fill="none"
                                        viewBox="0 0 24 24"
                                    >
                                        <circle
                                            class="opacity-25"
                                            cx="12"
                                            cy="12"
                                            r="10"
                                            stroke="currentColor"
                                            stroke-width="4"
                                        />
                                        <path
                                            class="opacity-75"
                                            fill="currentColor"
                                            d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
                                        />
                                    </svg>
                                    <VideoCameraIcon v-else class="h-4 w-4" aria-hidden="true" />
                                    {{
                                        settingsLoaded
                                            ? t("meetings.waiting_room.join")
                                            : t("meetings.waiting_room.loading")
                                    }}
                                </button>

                                <button
                                    type="button"
                                    @click="handleCancel"
                                    class="inline-flex items-center justify-center rounded-md bg-white px-4 py-2 text-sm font-semibold text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 hover:bg-gray-50"
                                >
                                    {{ t("meetings.waiting_room.cancel") }}
                                </button>
                            </div>
                        </DialogPanel>
                    </TransitionChild>
                </div>
            </div>
        </Dialog>
    </TransitionRoot>
</template>

<script setup>
import { ref, onMounted, onUnmounted, watch } from "vue";
import {
    Dialog,
    DialogPanel,
    DialogTitle,
    TransitionChild,
    TransitionRoot,
    Listbox,
    ListboxButton,
    ListboxOptions,
    ListboxOption,
    Switch,
} from "@headlessui/vue";
import {
    VideoCameraIcon,
    MicrophoneIcon,
    UserCircleIcon,
    ArrowsRightLeftIcon,
    ChevronUpDownIcon,
    CheckIcon,
    ArrowPathIcon,
} from "@heroicons/vue/24/outline";
import { t } from "@/i18n/index.js";

const props = defineProps({
    channelName: {
        type: String,
        default: "",
    },
    participants: {
        type: Number,
        default: null,
    },
});

const emit = defineEmits(["join", "cancel"]);
const videoPreview = ref(null);
const mirrorVideo = ref(true);
const cameraEnabled = ref(true);
const microphoneEnabled = ref(true);
const cameraDevices = ref([]);
const microphoneDevices = ref([]);
const selectedCamera = ref("");
const selectedMicrophone = ref("");
const settingsLoaded = ref(false);
const participantCount = ref(0);

let videoStream = null;
let audioStream = null;
let previewRequest = 0;
let closed = false;

const handleCancel = () => {
    cleanupPreview();
    emit("cancel");
};

const getDeviceName = (deviceId, deviceList) => {
    const device = deviceList.find((d) => d.deviceId === deviceId);

    return device?.label || device?.deviceId?.slice(0, 8) || "";
};

const refreshDevices = async () => {
    settingsLoaded.value = false;
    await loadDevices();
};

async function loadDevices() {
    try {
        // Probe only to unlock device labels/permission; release it immediately
        // so it doesn't keep the camera/mic on (startPreview opens the real one).
        const probe = await navigator.mediaDevices.getUserMedia({
            audio: true,
            video: true,
        });

        const devices = await navigator.mediaDevices.enumerateDevices();

        probe.getTracks().forEach((track) => track.stop());

        cameraDevices.value = devices.filter((device) => device.kind === "videoinput");
        microphoneDevices.value = devices.filter((device) => device.kind === "audioinput");

        if (cameraDevices.value.length > 0 && !selectedCamera.value) {
            selectedCamera.value = cameraDevices.value[0].deviceId;
        }

        if (microphoneDevices.value.length > 0 && !selectedMicrophone.value) {
            selectedMicrophone.value = microphoneDevices.value[0].deviceId;
        }

        settingsLoaded.value = true;
        await startPreview();
    } catch (error) {
        console.error("Error loading devices:", error);
        settingsLoaded.value = true;
    }
}

async function startPreview() {
    cleanupPreview();

    const request = previewRequest;

    if (closed || (!cameraEnabled.value && !microphoneEnabled.value)) {
        return;
    }

    const constraints = {};

    if (cameraEnabled.value && selectedCamera.value) {
        constraints.video = {
            deviceId: { exact: selectedCamera.value },
            width: { ideal: 1280 },
            height: { ideal: 720 },
        };
    }

    if (microphoneEnabled.value && selectedMicrophone.value) {
        constraints.audio = {
            deviceId: { exact: selectedMicrophone.value },
            echoCancellation: true,
            noiseSuppression: true,
            autoGainControl: true,
        };
    }

    try {
        if (Object.keys(constraints).length > 0) {
            const stream = await navigator.mediaDevices.getUserMedia(constraints);

            // A newer preview, a join or a cancel started while this one waited for the camera.
            if (closed || request !== previewRequest) {
                stream.getTracks().forEach((track) => track.stop());

                return;
            }

            const videoTracks = stream.getVideoTracks();
            const audioTracks = stream.getAudioTracks();

            if (videoTracks.length > 0) {
                videoStream = new MediaStream(videoTracks);
                if (videoPreview.value) {
                    videoPreview.value.srcObject = videoStream;
                }
            }

            if (audioTracks.length > 0) {
                audioStream = new MediaStream(audioTracks);
            }
        } else if (videoPreview.value) {
            videoPreview.value.srcObject = null;
        }
    } catch (error) {
        console.error("Error accessing media devices:", error);
        if (videoPreview.value) {
            videoPreview.value.srcObject = null;
        }
    }
}

function stopAllStreams() {
    if (videoStream) {
        videoStream.getTracks().forEach((track) => {
            track.stop();
        });
        videoStream = null;
    }

    if (audioStream) {
        audioStream.getTracks().forEach((track) => {
            track.stop();
        });
        audioStream = null;
    }

    if (videoPreview.value?.srcObject) {
        const stream = videoPreview.value.srcObject;

        stream.getTracks().forEach((track) => {
            track.stop();
        });

        videoPreview.value.srcObject = null;
    }
}

function cleanupPreview() {
    previewRequest++;
    stopAllStreams();
}

async function toggleCamera() {
    cameraEnabled.value = !cameraEnabled.value;
    await startPreview();
}

async function toggleMicrophone() {
    microphoneEnabled.value = !microphoneEnabled.value;
    await startPreview();
}

function joinCall() {
    if (!settingsLoaded.value) return;

    closed = true;
    cleanupPreview();

    emit("join", {
        cameraEnabled: cameraEnabled.value,
        microphoneEnabled: microphoneEnabled.value,
        selectedCamera: selectedCamera.value,
        selectedMicrophone: selectedMicrophone.value,
        mirrorVideo: mirrorVideo.value,
    });
}

watch([selectedCamera, selectedMicrophone], async () => {
    await startPreview();
});

onMounted(async () => {
    participantCount.value = props.participants ?? 0;
    await loadDevices();

    navigator.mediaDevices.addEventListener("devicechange", loadDevices);
});

onUnmounted(() => {
    closed = true;
    cleanupPreview();

    navigator.mediaDevices.removeEventListener("devicechange", loadDevices);
});
</script>
