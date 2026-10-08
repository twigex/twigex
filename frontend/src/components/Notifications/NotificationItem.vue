<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <li
        class="flex p-3 my-1 rounded-md mr-2"
        :class="notification.read_at != 0 ? 'hover:bg-gray-50' : 'bg-indigo-50 hover:bg-indigo-100'"
        @click="emits('notificationClicked')"
    >
        <div class="flex-none">
            <UserCircleIcon v-if="actorless" class="h-7 w-7 text-gray-400" />
            <div v-else class="h-7 w-7">
                <UserAvatar :user-id="notification.sender" />
            </div>
        </div>
        <div class="flex flex-col justify-between flex-auto min-w-0 ml-4">
            <div class="text-left">
                <p v-if="!actorless" class="text-sm font-semibold leading-6 text-gray-900">
                    <a class="hover:underline">{{ senderName }}</a>
                </p>
                <p class="mt-1 text-xs leading-5 text-gray-500">
                    {{ message(notification) }}
                </p>
            </div>
        </div>
        <div class="flex shrink-0 items-start gap-x-4">
            <div class="hidden sm:flex sm:flex-col sm:items-end">
                <p class="text-sm leading-6 text-gray-900">
                    {{ useDateOperations().getDateAndTime(notification.created_at) }}
                </p>
            </div>
        </div>
        <div class="flex shrink-0 items-center px-2">
            <Menu as="div" class="relative flex-none">
                <MenuButton
                    ref="reference"
                    @click.stop
                    class="-m-2.5 block p-2.5 text-gray-500 hover:text-gray-900"
                >
                    <span class="sr-only">{{ t("notifications.open_options") }}</span>
                    <EllipsisVerticalIcon class="h-5 w-5" aria-hidden="true" />
                </MenuButton>
                <transition
                    enter-active-class="transition ease-out duration-100"
                    enter-from-class="transform opacity-0 scale-95"
                    enter-to-class="transform opacity-100 scale-100"
                    leave-active-class="transition ease-in duration-75"
                    leave-from-class="transform opacity-100 scale-100"
                    leave-to-class="transform opacity-0 scale-95"
                >
                    <MenuItems
                        ref="floating"
                        :style="floatingStyles"
                        class="absolute right-0 z-10 mt-2 w-32 origin-top-right rounded-md bg-white py-2 shadow-lg ring-1 ring-gray-900/5 focus:outline-none"
                    >
                        <MenuItem v-slot="{ active }">
                            <a
                                href="#"
                                :class="[
                                    active ? 'bg-gray-50' : '',
                                    'block px-3 py-1 text-sm leading-6 text-gray-900 text-left',
                                ]"
                                @click.stop="emits('notificationRead')"
                                >{{ t("notifications.mark_as_read") }}
                            </a>
                        </MenuItem>
                        <MenuItem v-slot="{ active }">
                            <a
                                href="#"
                                :class="[
                                    active ? 'bg-gray-50' : '',
                                    'block px-3 py-1 text-sm leading-6 text-gray-900 text-left',
                                ]"
                                @click.stop="emits('notificationDeleted')"
                                >{{ t("common.button.delete") }}</a
                            >
                        </MenuItem>
                    </MenuItems>
                </transition>
            </Menu>
        </div>
    </li>
</template>

<script setup>
import { ref, computed } from "vue";
import { Menu, MenuButton, MenuItems, MenuItem } from "@headlessui/vue";
import { EllipsisVerticalIcon, UserCircleIcon } from "@heroicons/vue/20/solid";
import { t } from "@/i18n/index";
import UserAvatar from "@/components/UserAvatar.vue";
import { useUser } from "@/composables/useUser";
import useDateOperations from "@/composables/useDateOperations.js";
import constructNotificationMessage, { actorlessNotifications } from "@/utils/notifications.js";
import { useFloating, autoPlacement } from "@floating-ui/vue";
import { autoUpdate } from "@floating-ui/dom";

const props = defineProps({
    notification: {
        type: Object,
        required: true,
    },
});

const emits = defineEmits(["notificationClicked", "notificationRead", "notificationDeleted"]);

const reference = ref(null);
const floating = ref(null);
const { floatingStyles } = useFloating(reference, floating, {
    strategy: "absolute", // keeps it tied to the list item while scrolling
    transform: false,
    middleware: [
        autoPlacement({
            allowedPlacements: ["bottom-end", "bottom-start", "top-end", "top-start"],
        }),
    ],
    whileElementsMounted: autoUpdate,
});

const sender = useUser(() => props.notification.sender);

const senderName = computed(() => {
    if (!sender.value) return "";

    return `${sender.value.name ?? ""} ${sender.value.lastname ?? ""}`.trim();
});

const actorless = computed(() =>
    actorlessNotifications.has(props.notification.details?.notificationType),
);

function message(notification) {
    return constructNotificationMessage(notification);
}
</script>
