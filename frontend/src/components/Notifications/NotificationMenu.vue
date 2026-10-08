<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <Popover v-if="!isMobile" class="relative flex">
        <PopoverButton
            ref="reference"
            class="inline-flex items-center gap-x-1 text-sm font-semibold leading-6 text-gray-900"
        >
            <button
                type="button"
                class="rounded-full p-1 text-white shadow-sm focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600"
            >
                <BellIcon class="h-6 w-6" aria-hidden="true" />
                <div
                    v-if="unreadCount() > 0"
                    class="absolute inline-flex items-center justify-center w-6 h-6 text-xs font-bold text-white bg-red-500 border-2 border-white rounded-full -top-2 -end-2 dark:border-gray-900"
                >
                    {{ unreadCount() }}
                </div>
            </button>
        </PopoverButton>

        <transition
            enter-active-class="transition ease-out duration-100"
            enter-from-class="transform opacity-0 scale-95"
            enter-to-class="transform opacity-100 scale-100"
            leave-active-class="transition ease-in duration-75"
            leave-from-class="transform opacity-100 scale-100"
            leave-to-class="transform opacity-0 scale-95"
        >
            <PopoverPanel
                v-slot="{ close }"
                ref="floating"
                class="z-10 flex max-h-96"
                :style="floatingStyles"
            >
                <div
                    class="w-screen max-w-md flex-auto rounded-md bg-white text-sm leading-6 shadow-lg ring-1 ring-gray-900/5"
                >
                    <div class="p-4">
                        <div
                            v-if="notificationStore.notifications.length > 0"
                            class="flex flex-row mb-3 text-sm leading-6 justify-between"
                        >
                            <label for="comments" class="font-medium text-gray-500">{{
                                t("notifications.title")
                            }}</label>
                            <button
                                @click="markAllAsRead()"
                                type="button"
                                class="rounded bg-indigo-600 px-2 py-1 text-xs font-semibold text-white shadow-sm hover:bg-indigo-500 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600"
                            >
                                {{ t("notifications.mark_all_read") }}
                            </button>
                        </div>
                        <ul
                            v-if="notificationStore.notifications.length > 0"
                            role="list"
                            class="divide-y divide-gray-100 overflow-y-auto h-80"
                        >
                            <div
                                v-for="notification in notificationStore.notifications"
                                :key="notification.id"
                            >
                                <NotificationItem
                                    :notification="notification"
                                    @notificationClicked="onClicked(notification, close)"
                                    @notificationRead="readNotification(notification)"
                                    @notificationDeleted="deleteNotification(notification)"
                                />
                            </div>
                        </ul>
                        <div v-else class="text-center p-8">
                            <BellIcon class="mx-auto h-12 w-12 text-gray-400" aria-hidden="true" />
                            <h3 class="mt-2 text-sm font-semibold text-gray-900">
                                {{ t("notifications.title") }}
                            </h3>
                            <p class="mt-1 text-sm text-gray-500">
                                {{ t("notifications.empty") }}
                            </p>
                        </div>
                    </div>
                </div>
            </PopoverPanel>
        </transition>
    </Popover>

    <div v-else>
        <button
            @click="open = true"
            type="button"
            class="rounded-full p-1 text-white shadow-sm focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600"
        >
            <BellIcon class="h-6 w-6" aria-hidden="true" />
            <div
                v-if="unreadCount() > 0"
                class="absolute inline-flex items-center justify-center w-6 h-6 text-xs font-bold text-white bg-red-500 border-2 border-white rounded-full top-1 dark:border-gray-900"
            >
                {{ unreadCount() }}
            </div>
        </button>
        <TransitionRoot as="template" :show="open">
            <Dialog class="relative z-10" @close="open = false">
                <TransitionChild
                    as="template"
                    enter="ease-out duration-300"
                    enter-from="opacity-0"
                    enter-to=""
                    leave="ease-in duration-200"
                    leave-from=""
                    leave-to="opacity-0"
                >
                    <div class="fixed inset-0 bg-gray-500/75 transition-opacity"></div>
                </TransitionChild>

                <div class="fixed inset-0 z-10 w-screen overflow-y-auto">
                    <div
                        class="flex min-h-full items-end justify-center p-4 text-center sm:items-center sm:p-0"
                    >
                        <TransitionChild
                            as="template"
                            enter="ease-out duration-300"
                            enter-from="opacity-0 translate-y-4 sm:translate-y-0 sm:scale-95"
                            enter-to=" translate-y-0 sm:scale-100"
                            leave="ease-in duration-200"
                            leave-from=" translate-y-0 sm:scale-100"
                            leave-to="opacity-0 translate-y-4 sm:translate-y-0 sm:scale-95"
                        >
                            <DialogPanel
                                class="relative transform overflow-hidden rounded-lg bg-white px-4 pt-5 pb-4 text-left shadow-xl transition-all sm:my-8 sm:w-full sm:max-w-sm sm:p-6"
                            >
                                <div>
                                    <div class="text-left sm:mt-5">
                                        <div class="flex flex-row justify-between items-center">
                                            <DialogTitle
                                                as="h3"
                                                class="text-base font-semibold text-gray-900"
                                                >{{ t("notifications.title") }}</DialogTitle
                                            >

                                            <button
                                                type="button"
                                                class="rounded-md bg-white text-gray-400 hover:text-gray-500 focus:outline-2 focus:outline-offset-2 focus:outline-indigo-600"
                                                @click="open = false"
                                            >
                                                <span class="sr-only">{{
                                                    t("common.button.close")
                                                }}</span>
                                                <XMarkIcon class="size-6" aria-hidden="true" />
                                            </button>
                                        </div>

                                        <div
                                            class="mt-2 divide-y divide-gray-100 max-h-96 overflow-y-auto"
                                        >
                                            <div
                                                v-for="notification in notificationStore.notifications"
                                                :key="notification.id"
                                            >
                                                <NotificationItem
                                                    :notification="notification"
                                                    @notificationClicked="
                                                        onClicked(notification, close)
                                                    "
                                                    @notificationRead="
                                                        readNotification(notification)
                                                    "
                                                    @notificationDeleted="
                                                        deleteNotification(notification)
                                                    "
                                                />
                                            </div>
                                        </div>
                                    </div>
                                </div>
                            </DialogPanel>
                        </TransitionChild>
                    </div>
                </div>
            </Dialog>
        </TransitionRoot>
    </div>
</template>

<script setup>
import { onMounted, onUnmounted, ref } from "vue";
import {
    Popover,
    PopoverButton,
    PopoverPanel,
    Dialog,
    DialogPanel,
    DialogTitle,
    TransitionChild,
    TransitionRoot,
} from "@headlessui/vue";
import { BellIcon, XMarkIcon } from "@heroicons/vue/24/outline";
import { t } from "@/i18n/index";
import { useNotificationsStore } from "@/store/notifications";
import notificationService from "@/services/notificationService";
import { useFloating } from "@floating-ui/vue";
import useNotificationOperations from "@/composables/useNotifications.js";
import NotificationItem from "./NotificationItem.vue";

onMounted(() => {
    //add resize listener
    handleResize();
    window.addEventListener("resize", handleResize);

    notificationService.notifications().then((res) => {
        notificationStore.setNotifications(res.data);
    });
});

onUnmounted(() => {
    window.removeEventListener("resize", handleResize);
});

const { redirect } = useNotificationOperations();

const isMobile = ref(false);
const open = ref(false);
const reference = ref(null);
const floating = ref(null);
const { floatingStyles } = useFloating(reference, floating, {
    placement: "bottom-end",
    transform: false,
    strategy: "fixed",
});

const notificationStore = useNotificationsStore();

function onClicked(notification, close) {
    redirect(notification);

    if (isMobile.value) {
        open.value = false;

        return;
    }

    close();
}

function readNotification(notification) {
    notificationService.read({ read: true, id: [notification.id] }).then(() => {
        notification.read_at = true;
    });
}

function deleteNotification(notification) {
    notificationService.delete(notification.id).then(() => {
        notificationStore.notifications = notificationStore.notifications.filter(
            (item) => item.id !== notification.id,
        );
    });
}

function markAllAsRead() {
    notificationService
        .read({
            read: true,
            id: notificationStore.notifications.map((item) => item.id),
        })
        .then(() => {
            notificationStore.notifications.forEach((item) => (item.read_at = true));
        });
}

function unreadCount() {
    return notificationStore.notifications.filter((item) => item.read_at == 0).length;
}

function handleResize() {
    isMobile.value = window.innerWidth < 640;
    open.value = false;
}
</script>
