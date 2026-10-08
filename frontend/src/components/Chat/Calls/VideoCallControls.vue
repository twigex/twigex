<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div
        class="flex items-center gap-2 bg-gray-900/90 backdrop-blur-md rounded-2xl px-4 py-3 shadow-2xl ring-1 ring-white/10"
    >
        <button
            @click="emit('shareScreen')"
            type="button"
            title="Share screen"
            class="p-2.5 rounded-xl text-white transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-indigo-500"
            :class="
                screenSharing
                    ? 'bg-indigo-600 hover:bg-indigo-500'
                    : 'bg-white/10 hover:bg-white/20'
            "
        >
            <ComputerDesktopIcon class="h-5 w-5" />
        </button>

        <div class="w-px h-7 bg-white/10" />

        <Listbox as="div" :model-value="activeCameraDevice" by="deviceId" class="relative">
            <div class="flex rounded-xl overflow-hidden">
                <button
                    @click="emit('toggleCamera')"
                    type="button"
                    title="Toggle camera"
                    class="p-2.5 text-white transition-colors"
                    :class="
                        videoOn ? 'bg-white/10 hover:bg-white/20' : 'bg-red-600 hover:bg-red-500'
                    "
                >
                    <VideoCameraIcon v-if="videoOn" class="h-5 w-5" />
                    <VideoCameraSlashIcon v-else class="h-5 w-5" />
                </button>
                <div class="w-px" :class="videoOn ? 'bg-white/20' : 'bg-red-500'" />
                <ListboxButton
                    class="px-2 py-2.5 text-white transition-colors"
                    :class="
                        videoOn ? 'bg-white/10 hover:bg-white/20' : 'bg-red-600 hover:bg-red-500'
                    "
                    title="Select camera"
                >
                    <ChevronUpIcon class="h-3.5 w-3.5" />
                </ListboxButton>
            </div>

            <transition
                leave-active-class="transition ease-in duration-100"
                leave-from-class="opacity-100"
                leave-to-class="opacity-0"
            >
                <ListboxOptions
                    :ref="onCameraMenu"
                    class="absolute right-0 bottom-full mb-2 w-72 rounded-xl bg-gray-900 shadow-2xl ring-1 ring-white/10 overflow-hidden focus:outline-none"
                >
                    <ListboxOption
                        as="template"
                        v-for="option in cameraDevices"
                        :key="option.deviceId"
                        :value="option"
                        @click="emit('selectCamera', option)"
                        v-slot="{ active }"
                    >
                        <li
                            class="cursor-pointer select-none p-1.5 text-sm"
                            :class="active ? 'bg-white/5' : ''"
                        >
                            <div class="relative w-full rounded-lg overflow-hidden">
                                <video
                                    class="w-full h-36 bg-black object-cover"
                                    autoplay
                                    muted
                                    playsinline
                                    :ref="(el) => attachCameraPreview(el, option.deviceId)"
                                />
                                <div
                                    class="absolute bottom-0 inset-x-0 px-3 py-1.5 bg-black/60 backdrop-blur-sm flex items-center justify-between"
                                >
                                    <p class="text-white text-sm font-medium truncate">
                                        {{ option.label }}
                                    </p>
                                    <CheckIcon
                                        v-if="activeCameraDevice?.deviceId === option.deviceId"
                                        class="h-4 w-4 text-indigo-400 shrink-0 ml-2"
                                    />
                                </div>
                            </div>
                        </li>
                    </ListboxOption>

                    <div class="border-t border-white/10 mx-3 my-1" />

                    <div class="px-4 py-3">
                        <label class="flex items-center gap-3 cursor-pointer">
                            <input
                                type="checkbox"
                                class="h-4 w-4 rounded border-gray-600 bg-gray-800 text-indigo-600 focus:ring-indigo-500 focus:ring-offset-gray-900"
                                v-model="settings.mirrorVideo"
                            />
                            <span class="text-sm text-gray-300">Mirror camera</span>
                        </label>
                    </div>
                </ListboxOptions>
            </transition>
        </Listbox>

        <Listbox as="div" :model-value="activeAudioDevice" by="deviceId" class="relative">
            <div class="flex rounded-xl overflow-hidden">
                <button
                    @click="muted ? emit('unmute') : emit('mute')"
                    type="button"
                    title="Toggle microphone"
                    class="p-2.5 text-white transition-colors"
                    :class="
                        !muted ? 'bg-white/10 hover:bg-white/20' : 'bg-red-600 hover:bg-red-500'
                    "
                >
                    <MicrophoneIcon v-if="!muted" class="h-5 w-5" />
                    <div v-else class="relative w-5 h-5">
                        <MicrophoneIcon class="w-5 h-5" />
                        <SlashIcon class="w-5 h-5 absolute inset-0 -rotate-[20deg] scale-x-[-1]" />
                    </div>
                </button>
                <template v-if="audioInputDevices.length > 1">
                    <div class="w-px" :class="!muted ? 'bg-white/20' : 'bg-red-500'" />
                    <ListboxButton
                        class="px-2 py-2.5 text-white transition-colors"
                        :class="
                            !muted ? 'bg-white/10 hover:bg-white/20' : 'bg-red-600 hover:bg-red-500'
                        "
                        title="Select microphone"
                    >
                        <ChevronUpIcon class="h-3.5 w-3.5" />
                    </ListboxButton>
                </template>
            </div>

            <transition
                leave-active-class="transition ease-in duration-100"
                leave-from-class="opacity-100"
                leave-to-class="opacity-0"
            >
                <ListboxOptions
                    class="absolute right-0 bottom-full mb-2 w-72 rounded-xl bg-gray-900 shadow-2xl ring-1 ring-white/10 overflow-hidden focus:outline-none"
                >
                    <ListboxOption
                        as="template"
                        v-for="option in audioInputDevices"
                        :key="option.deviceId"
                        :value="option"
                        @click="emit('selectMicrophone', option)"
                        v-slot="{ active, selected }"
                    >
                        <li
                            class="cursor-pointer select-none px-4 py-3 text-sm transition-colors flex items-center justify-between"
                            :class="active ? 'bg-indigo-600 text-white' : 'text-gray-300'"
                        >
                            <p :class="selected ? 'font-semibold' : 'font-normal'">
                                {{ option.label }}
                            </p>
                            <CheckIcon
                                v-if="activeAudioDevice?.deviceId === option.deviceId"
                                class="h-4 w-4 shrink-0 ml-3"
                                :class="active ? 'text-white' : 'text-indigo-400'"
                            />
                        </li>
                    </ListboxOption>
                </ListboxOptions>
            </transition>
        </Listbox>

        <div class="w-px h-7 bg-white/10" />

        <button
            v-if="isHost && !standalone && !isDirectCall"
            @click="emit('invite')"
            type="button"
            title="Invite people"
            class="p-2.5 rounded-xl bg-white/10 hover:bg-white/20 text-white transition-colors"
        >
            <UserPlusIcon class="h-5 w-5" />
        </button>

        <Menu v-if="isHost && !standalone" as="div" class="relative">
            <MenuButton
                title="Make host"
                class="p-2.5 rounded-xl bg-white/10 hover:bg-white/20 text-white transition-colors"
            >
                <UserGroupIcon class="h-5 w-5" />
            </MenuButton>
            <MenuItems
                class="absolute right-0 bottom-full mb-2 w-56 rounded-xl bg-gray-900 py-1 shadow-2xl ring-1 ring-white/10 focus:outline-none"
            >
                <div class="px-3 py-2 text-xs text-gray-400">
                    {{ t("meetings.host.make_host") }}
                </div>
                <template v-if="promotableMembers.length">
                    <MenuItem v-for="p in promotableMembers" :key="p.identity" v-slot="{ active }">
                        <button
                            @click="emit('makeHost', p)"
                            :class="[
                                active ? 'bg-white/5' : '',
                                'block w-full px-3 py-2 text-left text-sm text-white',
                            ]"
                        >
                            {{ p.name || p.identity }}
                        </button>
                    </MenuItem>
                </template>
                <div v-else class="px-3 py-2 text-sm text-gray-500">
                    {{ t("meetings.host.none") }}
                </div>
            </MenuItems>
        </Menu>

        <button
            v-if="!standalone"
            @click="emit('minimize')"
            type="button"
            title="Minimize"
            class="p-2.5 rounded-xl bg-white/10 hover:bg-white/20 text-white transition-colors"
        >
            <ArrowDownRightIcon class="h-5 w-5" />
        </button>

        <button
            @click="emit('leave')"
            type="button"
            title="Leave call"
            class="p-2.5 rounded-xl bg-red-600 hover:bg-red-500 text-white transition-colors"
        >
            <PhoneXMarkIcon class="h-5 w-5" />
        </button>
    </div>
</template>

<script setup>
import { t } from "@/i18n";
import {
    Listbox,
    ListboxButton,
    ListboxOption,
    ListboxOptions,
    Menu,
    MenuButton,
    MenuItems,
    MenuItem,
} from "@headlessui/vue";
import { CheckIcon, ChevronUpIcon } from "@heroicons/vue/20/solid";
import {
    PhoneXMarkIcon,
    MicrophoneIcon,
    ComputerDesktopIcon,
    VideoCameraIcon,
    VideoCameraSlashIcon,
    ArrowDownRightIcon,
    SlashIcon,
    UserPlusIcon,
    UserGroupIcon,
} from "@heroicons/vue/24/outline";

const props = defineProps({
    settings: {
        type: Object,
        required: true,
    },
    screenSharing: {
        type: Boolean,
        default: false,
    },
    videoOn: {
        type: Boolean,
        default: false,
    },
    muted: {
        type: Boolean,
        default: false,
    },
    cameraDevices: {
        type: Array,
        default: () => [],
    },
    audioInputDevices: {
        type: Array,
        default: () => [],
    },
    activeCameraDevice: {
        type: Object,
        default: null,
    },
    activeAudioDevice: {
        type: Object,
        default: null,
    },
    attachCameraPreview: {
        type: Function,
        required: true,
    },
    stopCameraPreviews: {
        type: Function,
        required: true,
    },
    isHost: {
        type: Boolean,
        default: false,
    },
    standalone: {
        type: Boolean,
        default: false,
    },
    isDirectCall: {
        type: Boolean,
        default: false,
    },
    promotableMembers: {
        type: Array,
        default: () => [],
    },
});

const emit = defineEmits([
    "shareScreen",
    "toggleCamera",
    "selectCamera",
    "mute",
    "unmute",
    "selectMicrophone",
    "invite",
    "makeHost",
    "minimize",
    "leave",
]);

function onCameraMenu(menu) {
    if (!menu) props.stopCameraPreviews();
}
</script>
