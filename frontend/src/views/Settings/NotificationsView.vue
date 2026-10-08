<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div class="flex flex-col h-full overflow-hidden">
        <div class="flex flex-row justify-center flex-1 overflow-y-auto p-2">
            <div class="w-full md:w-3/4 space-y-0">
                <!-- Tab navigation -->
                <div class="border-b border-gray-200">
                    <nav class="-mb-px flex space-x-8" aria-label="Tabs">
                        <a
                            v-for="tab in tabs"
                            :key="tab.name"
                            @click="updateActiveTab(tab)"
                            :class="[
                                tab.current
                                    ? 'border-indigo-500 text-indigo-600'
                                    : 'border-transparent text-gray-500 hover:border-gray-300 hover:text-gray-700',
                                'cursor-pointer whitespace-nowrap border-b-2 py-4 px-1 text-sm font-medium',
                            ]"
                            :aria-current="tab.current ? 'page' : undefined"
                        >
                            {{ tab.name }}
                        </a>
                    </nav>
                </div>

                <!-- App notifications tab -->
                <div v-if="tabs[0].current" class="mt-6 space-y-8">
                    <div class="pb-6 border-b border-gray-900/10">
                        <h2 class="text-base font-semibold leading-7 text-gray-900">
                            {{ t("settings.notifications.file_notifications.title") }}
                        </h2>
                        <p class="mt-1 text-sm leading-6 text-gray-600">
                            {{ t("settings.notifications.file_notifications.description") }}
                        </p>
                        <div class="mt-4 space-y-4">
                            <div
                                v-for="(item, index) in items.fileNotifications"
                                :key="index"
                                class="relative flex gap-x-3"
                            >
                                <div class="flex h-6 items-center">
                                    <input
                                        v-model="item.app"
                                        :id="`app-file-${index}`"
                                        type="checkbox"
                                        class="h-4 w-4 rounded border-gray-300 text-indigo-600 focus:ring-indigo-600"
                                    />
                                </div>
                                <div class="text-sm leading-6">
                                    <label
                                        :for="`app-file-${index}`"
                                        class="text-gray-700 cursor-pointer"
                                    >
                                        {{ item.text }}
                                    </label>
                                </div>
                            </div>
                        </div>
                    </div>

                    <div class="pb-6 border-b border-gray-900/10">
                        <h2 class="text-base font-semibold leading-7 text-gray-900">
                            {{ t("settings.notifications.project_notifications.title") }}
                        </h2>
                        <p class="mt-1 text-sm leading-6 text-gray-600">
                            {{ t("settings.notifications.project_notifications.description") }}
                        </p>
                        <div class="mt-4 space-y-4">
                            <div
                                v-for="(item, index) in items.projectNotifications"
                                :key="index"
                                class="relative flex gap-x-3"
                            >
                                <div class="flex h-6 items-center">
                                    <input
                                        v-model="item.app"
                                        :id="`app-project-${index}`"
                                        type="checkbox"
                                        class="h-4 w-4 rounded border-gray-300 text-indigo-600 focus:ring-indigo-600"
                                    />
                                </div>
                                <div class="text-sm leading-6">
                                    <label
                                        :for="`app-project-${index}`"
                                        class="text-gray-700 cursor-pointer"
                                    >
                                        {{ item.text }}
                                    </label>
                                </div>
                            </div>
                        </div>
                    </div>

                    <div class="pb-6">
                        <h2 class="text-base font-semibold leading-7 text-gray-900">
                            {{ t("settings.notifications.meeting_notifications.title") }}
                        </h2>
                        <p class="mt-1 text-sm leading-6 text-gray-600">
                            {{ t("settings.notifications.meeting_notifications.description") }}
                        </p>
                        <div class="mt-4 space-y-4">
                            <div
                                v-for="(item, index) in items.meetingNotifications"
                                :key="index"
                                class="relative flex gap-x-3"
                            >
                                <div class="flex h-6 items-center">
                                    <input
                                        v-model="item.app"
                                        :id="`app-meeting-${index}`"
                                        type="checkbox"
                                        class="h-4 w-4 rounded border-gray-300 text-indigo-600 focus:ring-indigo-600"
                                    />
                                </div>
                                <div class="text-sm leading-6">
                                    <label
                                        :for="`app-meeting-${index}`"
                                        class="text-gray-700 cursor-pointer"
                                    >
                                        {{ item.text }}
                                    </label>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>

                <!-- Email notifications tab -->
                <div v-if="tabs[1].current" class="mt-6 space-y-8">
                    <div class="pb-6 border-b border-gray-900/10">
                        <h2 class="text-base font-semibold leading-7 text-gray-900">
                            {{ t("settings.notifications.email_frequency.title") }}
                        </h2>
                        <div class="mt-4 w-72">
                            <Listbox as="div" v-model="selected">
                                <div class="relative">
                                    <ListboxButton
                                        class="grid w-full cursor-default grid-cols-1 rounded-md bg-white py-1.5 pl-3 pr-2 text-left text-gray-900 shadow-sm outline outline-1 -outline-offset-1 outline-gray-300 focus-visible:outline focus-visible:outline-2 focus-visible:-outline-offset-2 focus-visible:outline-indigo-600 sm:text-sm/6"
                                    >
                                        <span class="col-start-1 row-start-1 truncate pr-6">{{
                                            selected ? selected.text : "Select an option"
                                        }}</span>
                                        <ChevronUpDownIcon
                                            class="col-start-1 row-start-1 size-5 self-center justify-self-end text-gray-500 sm:size-4"
                                            aria-hidden="true"
                                        />
                                    </ListboxButton>

                                    <transition
                                        leave-active-class="transition ease-in duration-100"
                                        leave-from-class="opacity-100"
                                        leave-to-class="opacity-0"
                                    >
                                        <ListboxOptions
                                            class="absolute z-10 mt-1 max-h-60 w-full overflow-auto rounded-md bg-white py-1 text-base shadow-lg outline outline-1 outline-black/5 sm:text-sm"
                                        >
                                            <ListboxOption
                                                as="template"
                                                v-for="option in options"
                                                :key="option.key"
                                                :value="option"
                                                v-slot="{ active, selected: isSelected }"
                                            >
                                                <li
                                                    :class="[
                                                        active
                                                            ? 'bg-indigo-600 text-white outline-none'
                                                            : 'text-gray-900',
                                                        'relative cursor-default select-none py-2 pl-3 pr-9',
                                                    ]"
                                                >
                                                    <span
                                                        :class="[
                                                            isSelected
                                                                ? 'font-semibold'
                                                                : 'font-normal',
                                                            'block truncate',
                                                        ]"
                                                        >{{ option.text }}</span
                                                    >
                                                    <span
                                                        v-if="isSelected"
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
                                    </transition>
                                </div>
                            </Listbox>
                        </div>
                    </div>

                    <div class="pb-6 border-b border-gray-900/10">
                        <h2 class="text-base font-semibold leading-7 text-gray-900">
                            {{ t("settings.notifications.file_notifications.title") }}
                        </h2>
                        <p class="mt-1 text-sm leading-6 text-gray-600">
                            {{ t("settings.notifications.file_notifications.description") }}
                        </p>
                        <div class="mt-4 space-y-4">
                            <div
                                v-for="(item, index) in items.fileNotifications"
                                :key="index"
                                class="relative flex gap-x-3"
                            >
                                <div class="flex h-6 items-center">
                                    <input
                                        v-model="item.email"
                                        :id="`email-file-${index}`"
                                        type="checkbox"
                                        class="h-4 w-4 rounded border-gray-300 text-indigo-600 focus:ring-indigo-600"
                                    />
                                </div>
                                <div class="text-sm leading-6">
                                    <label
                                        :for="`email-file-${index}`"
                                        class="text-gray-700 cursor-pointer"
                                    >
                                        {{ item.text }}
                                    </label>
                                </div>
                            </div>
                        </div>
                    </div>

                    <div class="pb-6 border-b border-gray-900/10">
                        <h2 class="text-base font-semibold leading-7 text-gray-900">
                            {{ t("settings.notifications.project_notifications.title") }}
                        </h2>
                        <p class="mt-1 text-sm leading-6 text-gray-600">
                            {{ t("settings.notifications.project_notifications.description") }}
                        </p>
                        <div class="mt-4 space-y-4">
                            <div
                                v-for="(item, index) in items.projectNotifications"
                                :key="index"
                                class="relative flex gap-x-3"
                            >
                                <div class="flex h-6 items-center">
                                    <input
                                        v-model="item.email"
                                        :id="`email-project-${index}`"
                                        type="checkbox"
                                        class="h-4 w-4 rounded border-gray-300 text-indigo-600 focus:ring-indigo-600"
                                    />
                                </div>
                                <div class="text-sm leading-6">
                                    <label
                                        :for="`email-project-${index}`"
                                        class="text-gray-700 cursor-pointer"
                                    >
                                        {{ item.text }}
                                    </label>
                                </div>
                            </div>
                        </div>
                    </div>

                    <div class="pb-6">
                        <h2 class="text-base font-semibold leading-7 text-gray-900">
                            {{ t("settings.notifications.meeting_notifications.title") }}
                        </h2>
                        <p class="mt-1 text-sm leading-6 text-gray-600">
                            {{ t("settings.notifications.meeting_notifications.description") }}
                        </p>
                        <div class="mt-4 space-y-4">
                            <div
                                v-for="(item, index) in items.meetingNotifications"
                                :key="index"
                                class="relative flex gap-x-3"
                            >
                                <div class="flex h-6 items-center">
                                    <input
                                        v-model="item.email"
                                        :id="`email-meeting-${index}`"
                                        type="checkbox"
                                        class="h-4 w-4 rounded border-gray-300 text-indigo-600 focus:ring-indigo-600"
                                    />
                                </div>
                                <div class="text-sm leading-6">
                                    <label
                                        :for="`email-meeting-${index}`"
                                        class="text-gray-700 cursor-pointer"
                                    >
                                        {{ item.text }}
                                    </label>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </div>

        <div class="shrink-0 flex items-center justify-end gap-x-3 border-t bg-gray-50 px-6 py-4">
            <button
                :disabled="!hasChanges"
                @click.prevent="save"
                class="rounded-md bg-indigo-600 px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-indigo-500 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600 disabled:opacity-50 disabled:cursor-not-allowed"
            >
                {{ t("common.button.save") }}
            </button>
        </div>
    </div>
</template>

<script setup>
import { t } from "@/i18n/index";

import { computed, onMounted, reactive, ref } from "vue";
import { Listbox, ListboxButton, ListboxOption, ListboxOptions } from "@headlessui/vue";
import { CheckIcon, ChevronUpDownIcon } from "@heroicons/vue/20/solid";
import { useUserStore } from "@/store/user";
import { useAlertStore } from "@/store/alerts";
import { extractErrorMessage } from "@/utils/errors";
import userServices from "@/services/userService";

onMounted(() => {
    items.fileNotifications = fileNotifications();
    items.projectNotifications = projectNotifications();
    items.meetingNotifications = meetingNotifications();

    items.fileNotifications.forEach((element) => {
        for (let index = 0; index < userStore.preferences.Notifications.length; index++) {
            if (element.email_key == userStore.preferences.Notifications[index].name) {
                element.email =
                    userStore.preferences.Notifications[index].value == "true" ? true : false;
            } else if (element.app_key == userStore.preferences.Notifications[index].name) {
                element.app =
                    userStore.preferences.Notifications[index].value == "true" ? true : false;
            }
        }
    });

    items.projectNotifications.forEach((element) => {
        for (let index = 0; index < userStore.preferences.Notifications.length; index++) {
            if (element.email_key == userStore.preferences.Notifications[index].name) {
                element.email =
                    userStore.preferences.Notifications[index].value == "true" ? true : false;
            } else if (element.app_key == userStore.preferences.Notifications[index].name) {
                element.app =
                    userStore.preferences.Notifications[index].value == "true" ? true : false;
            }
        }
    });

    items.meetingNotifications.forEach((element) => {
        for (let index = 0; index < userStore.preferences.Notifications.length; index++) {
            if (element.email_key == userStore.preferences.Notifications[index].name) {
                element.email =
                    userStore.preferences.Notifications[index].value == "true" ? true : false;
            } else if (element.app_key == userStore.preferences.Notifications[index].name) {
                element.app =
                    userStore.preferences.Notifications[index].value == "true" ? true : false;
            }
        }
    });

    for (let i = 0; i < userStore.preferences.Notifications.length; i++) {
        if (userStore.preferences.Notifications[i].name == "send_email_notifications") {
            selected.value =
                userStore.preferences.Notifications[i].value == "immediately"
                    ? options[0]
                    : options[1];
        }
    }

    snapshotState();
});

const userStore = useUserStore();
const items = reactive({
    file: fileNotifications(),
});
const options = [
    {
        text: t.value("settings.notifications.email_frequency.immediately"),
        key: "immediately",
    },
    {
        text: t.value("settings.notifications.email_frequency.never"),
        key: "never",
    },
];
const selected = ref("");
const savedState = ref(null);

const hasChanges = computed(() => {
    if (!savedState.value) return false;

    return (
        JSON.stringify(
            items.fileNotifications?.map((i) => ({
                app: i.app,
                email: i.email,
            })),
        ) !== savedState.value.fileNotifications ||
        JSON.stringify(
            items.projectNotifications?.map((i) => ({
                app: i.app,
                email: i.email,
            })),
        ) !== savedState.value.projectNotifications ||
        JSON.stringify(
            items.meetingNotifications?.map((i) => ({
                app: i.app,
                email: i.email,
            })),
        ) !== savedState.value.meetingNotifications ||
        selected.value?.key !== savedState.value.selectedKey
    );
});

const tabs = ref([
    {
        name: t.value("settings.notifications.tabs.app_notifications"),
        current: true,
    },
    {
        name: t.value("settings.notifications.tabs.email_notifications"),
        current: false,
    },
]);

function snapshotState() {
    savedState.value = {
        fileNotifications: JSON.stringify(
            items.fileNotifications?.map((i) => ({
                app: i.app,
                email: i.email,
            })),
        ),
        projectNotifications: JSON.stringify(
            items.projectNotifications?.map((i) => ({
                app: i.app,
                email: i.email,
            })),
        ),
        meetingNotifications: JSON.stringify(
            items.meetingNotifications?.map((i) => ({
                app: i.app,
                email: i.email,
            })),
        ),
        selectedKey: selected.value?.key,
    };
}

function updateActiveTab(tab) {
    tabs.value.forEach((t) => {
        t.current = false;
    });
    tab.current = true;
}

function save() {
    let notifications = userStore.preferences.Notifications;

    items.fileNotifications.forEach((element) => {
        for (let index = 0; index < notifications.length; index++) {
            if (element.email_key == notifications[index].name) {
                notifications[index].value = element.email == true ? "true" : "false";
            } else if (element.app_key == notifications[index].name) {
                notifications[index].value = element.app == true ? "true" : "false";
            }
        }
    });

    items.projectNotifications.forEach((element) => {
        for (let index = 0; index < notifications.length; index++) {
            if (element.email_key == notifications[index].name) {
                notifications[index].value = element.email == true ? "true" : "false";
            } else if (element.app_key == notifications[index].name) {
                notifications[index].value = element.app == true ? "true" : "false";
            }
        }
    });

    items.meetingNotifications.forEach((element) => {
        for (let index = 0; index < notifications.length; index++) {
            if (element.email_key == notifications[index].name) {
                notifications[index].value = element.email == true ? "true" : "false";
            } else if (element.app_key == notifications[index].name) {
                notifications[index].value = element.app == true ? "true" : "false";
            }
        }
    });

    for (let i = 0; i < notifications.length; i++) {
        if (notifications[i].name == "send_email_notifications") {
            notifications[i].value = selected.value.key;
        }
    }

    userServices
        .updatePreferences({ preferences: notifications })
        .then(() => {
            snapshotState();
            useAlertStore().showSuccess(t.value("common.success.settings_updated"));
        })
        .catch((error) => {
            useAlertStore().showError(extractErrorMessage(error));
        });
}

function fileNotifications() {
    var items = [
        {
            text: t.value("settings.notifications.file_notification.created"),
            email: "",
            app: "",
            email_key: "notify_email_file_or_folder_created",
            app_key: "notify_app_file_or_folder_created",
        },
        {
            text: t.value("settings.notifications.file_notification.changed"),
            email: "",
            app: "",
            email_key: "notify_email_file_or_folder_renamed",
            app_key: "notify_app_file_or_folder_renamed",
        },
        {
            text: t.value("settings.notifications.file_notification.deleted"),
            email: "",
            app: "",
            email_key: "notify_email_file_or_folder_deleted",
            app_key: "notify_app_file_or_folder_deleted",
        },
        {
            text: t.value("settings.notifications.file_notification.restored"),
            email: "",
            app: "",
            email_key: "notify_email_file_or_folder_restored",
            app_key: "notify_app_file_or_folder_restored",
        },
        {
            text: t.value("settings.notifications.file_notification.shared"),
            email: "",
            app: "",
            email_key: "notify_email_file_or_folder_shared",
            app_key: "notify_app_file_or_folder_shared",
        },
        {
            text: t.value("settings.notifications.file_notification.downloaded"),
            email: "",
            app: "",
            email_key: "notify_email_file_or_folder_downloaded",
            app_key: "notify_app_file_or_folder_downloaded",
        },
        {
            text: t.value("settings.notifications.file_notification.public_downloaded"),
            email: "",
            app: "",
            email_key: "notify_email_file_or_folder_public_downloaded",
            app_key: "notify_app_file_or_folder_public_downloaded",
        },
    ];

    return items;
}

function projectNotifications() {
    var items = [
        {
            text: t.value("settings.notifications.project_notification.invited"),
            email: "",
            app: "",
            email_key: "notify_email_project_invited",
            app_key: "notify_app_project_invited",
        },
        {
            text: t.value("settings.notifications.project_notification.project_deleted"),
            email: "",
            app: "",
            email_key: "notify_email_project_deleted",
            app_key: "notify_app_project_deleted",
        },
        {
            text: t.value("settings.notifications.project_notification.task_deleted"),
            email: "",
            app: "",
            email_key: "notify_email_task_deleted",
            app_key: "notify_app_task_deleted",
        },
        {
            text: t.value("settings.notifications.project_notification.task_created"),
            email: "",
            app: "",
            email_key: "notify_email_task_created",
            app_key: "notify_app_task_created",
        },
        {
            text: t.value("settings.notifications.project_notification.task_assigned"),
            email: "",
            app: "",
            email_key: "notify_email_task_assigned",
            app_key: "notify_app_task_assigned",
        },
        {
            text: t.value("settings.notifications.project_notification.assigned_status_changed"),
            email: "",
            app: "",
            email_key: "notify_email_assigned_task_status_changed",
            app_key: "notify_app_assigned_task_status_changed",
        },
        {
            text: t.value("settings.notifications.project_notification.mention"),
            email: "",
            app: "",
            email_key: "notify_email_task_comment",
            app_key: "notify_app_task_comment",
        },
    ];

    return items;
}

function meetingNotifications() {
    var items = [
        {
            text: t.value("settings.notifications.meeting_notification.scheduled"),
            email: "",
            app: "",
            email_key: "notify_email_meeting_scheduled",
            app_key: "notify_app_meeting_scheduled",
        },
        {
            text: t.value("settings.notifications.meeting_notification.updated"),
            email: "",
            app: "",
            email_key: "notify_email_meeting_updated",
            app_key: "notify_app_meeting_updated",
        },
        {
            text: t.value("settings.notifications.meeting_notification.cancelled"),
            email: "",
            app: "",
            email_key: "notify_email_meeting_cancelled",
            app_key: "notify_app_meeting_cancelled",
        },
        {
            text: t.value("settings.notifications.meeting_notification.mention"),
            email: "",
            app: "",
            email_key: "notify_email_channel_mention",
            app_key: "notify_app_channel_mention",
        },
    ];

    return items;
}
</script>
