<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div class="flex h-full flex-col">
        <div
            v-if="!can('manage_channels')"
            class="flex h-full flex-col items-center justify-center px-6 text-center"
        >
            <LockClosedIcon class="size-12 text-gray-300" />
            <h3 class="mt-2 text-sm font-semibold text-gray-900">
                {{ t("settings.access_restricted.title") }}
            </h3>
            <p class="mt-1 max-w-sm text-sm text-gray-500">
                {{ t("settings.access_restricted.channels") }}
            </p>
        </div>

        <div v-else class="flex-1 overflow-y-auto">
            <div class="mx-auto max-w-6xl space-y-8 px-6 py-6">
                <DataTable
                    v-model:sort="sort"
                    :columns="columns"
                    :items="filteredChannels"
                    :loading="!loaded"
                    clickable
                    @row-click="editChannel"
                >
                    <template #toolbar>
                        <div class="relative w-full max-w-xs">
                            <MagnifyingGlassIcon
                                class="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-gray-400"
                                aria-hidden="true"
                            />
                            <input
                                v-model="searchQuery"
                                type="search"
                                :placeholder="t('common.placeholder.search')"
                                class="block w-full rounded-md border-0 bg-white py-1.5 pl-9 pr-3 text-sm text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600"
                            />
                        </div>
                    </template>

                    <template #name="{ item }">
                        <div class="flex items-center">
                            <span class="relative size-10 shrink-0">
                                <LetterAvatar
                                    :id="item.id"
                                    :name="item.displayname"
                                    class="size-10 rounded-lg text-sm"
                                />
                                <span
                                    v-if="item.type === channelTypes.Private"
                                    class="absolute -bottom-1 -right-1 flex size-5 items-center justify-center rounded-full bg-white"
                                    aria-hidden="true"
                                >
                                    <LockClosedIcon class="size-3.5 text-gray-500" />
                                </span>
                            </span>
                            <div class="ml-4 min-w-0">
                                <div class="flex items-center gap-x-2">
                                    <span class="truncate font-medium text-gray-900">
                                        {{ item.displayname }}
                                    </span>
                                    <span
                                        v-if="item.deleted"
                                        class="inline-flex shrink-0 items-center rounded-md bg-gray-50 px-2 py-0.5 text-xs font-medium text-gray-600 ring-1 ring-inset ring-gray-500/10"
                                    >
                                        {{ t("settings.channels.archived") }}
                                    </span>
                                </div>
                                <div
                                    v-if="item.description"
                                    class="mt-1 max-w-md truncate text-gray-500"
                                >
                                    {{ item.description }}
                                </div>
                            </div>
                        </div>
                    </template>

                    <template #type="{ item }">
                        <span
                            v-if="item.type === channelTypes.Public"
                            class="inline-flex items-center rounded-md bg-green-50 px-2 py-1 text-xs font-medium text-green-700 ring-1 ring-inset ring-green-600/20"
                        >
                            {{ t("settings.channels.type.public") }}
                        </span>
                        <span
                            v-else
                            class="inline-flex items-center rounded-md bg-gray-50 px-2 py-1 text-xs font-medium text-gray-600 ring-1 ring-inset ring-gray-500/10"
                        >
                            {{ t("settings.channels.type.private") }}
                        </span>
                    </template>

                    <template #last_post="{ item }">
                        {{ getDate(item.last_post) }}
                    </template>

                    <template #created="{ item }">
                        {{ getDate(item.created) }}
                    </template>

                    <template #member_count="{ item }">
                        <span class="inline-flex items-center gap-x-1.5">
                            <UsersIcon class="size-4 text-gray-400" aria-hidden="true" />
                            {{ item.member_count }}
                        </span>
                    </template>

                    <template #actions="{ item }">
                        <RouterLink
                            :to="{ name: 'channel', params: { id: item.id } }"
                            class="font-medium text-indigo-600 hover:text-indigo-900"
                            @click.stop
                        >
                            {{ t("common.button.edit") }}
                            <span class="sr-only">, {{ item.displayname }}</span>
                        </RouterLink>
                    </template>
                </DataTable>

                <template v-if="isSystemAdmin">
                    <SectionCard
                        :title="t('settings.channels.link_preview.title')"
                        :description="t('settings.channels.link_preview.description')"
                    >
                        <div class="space-y-4 p-4">
                            <EnvManagedNotice v-if="settings.link_preview_env_locked" />
                            <ToggleSwitch
                                v-model="settings.link_preview_settings.enabled"
                                :label="t('settings.channels.link_preview.enable')"
                                :disabled="settings.link_preview_env_locked"
                            />
                        </div>
                    </SectionCard>

                    <SectionCard
                        :title="t('settings.channels.gif.title')"
                        :description="t('settings.channels.gif.description')"
                    >
                        <div class="space-y-4 p-4">
                            <EnvManagedNotice v-if="settings.gif_env_locked" />
                            <ToggleSwitch
                                v-model="settings.gif_settings.enabled"
                                :label="t('settings.channels.enable_gif_picker')"
                                :disabled="settings.gif_env_locked"
                            />
                            <div class="max-w-sm">
                                <label
                                    for="gif-api-key"
                                    class="block text-sm font-medium text-gray-900"
                                >
                                    {{ t("settings.channels.gif.api_key") }}
                                </label>
                                <input
                                    id="gif-api-key"
                                    v-model="settings.gif_settings.api_key"
                                    type="password"
                                    :disabled="settings.gif_env_locked"
                                    placeholder="••••••"
                                    :class="inputClass"
                                />
                            </div>
                        </div>
                    </SectionCard>

                    <SectionCard
                        :title="t('settings.channels.video.title')"
                        :description="t('settings.channels.video.description')"
                    >
                        <div class="space-y-4 p-4">
                            <EnvManagedNotice v-if="settings.video_env_locked" />
                            <ToggleSwitch
                                v-model="settings.video_settings.enabled"
                                :label="t('settings.channels.enable_video')"
                                :disabled="settings.video_env_locked"
                            />
                            <div class="grid max-w-2xl gap-4 sm:grid-cols-2">
                                <div class="sm:col-span-2">
                                    <label
                                        for="video-host"
                                        class="block text-sm font-medium text-gray-900"
                                    >
                                        {{ t("settings.channels.video.host") }}
                                    </label>
                                    <input
                                        id="video-host"
                                        v-model="settings.video_settings.host"
                                        type="text"
                                        :disabled="settings.video_env_locked"
                                        placeholder="wss://video.example.com"
                                        :class="inputClass"
                                    />
                                </div>
                                <div>
                                    <label
                                        for="video-key"
                                        class="block text-sm font-medium text-gray-900"
                                    >
                                        {{ t("settings.channels.video.api_key") }}
                                    </label>
                                    <input
                                        id="video-key"
                                        v-model="settings.video_settings.key"
                                        type="text"
                                        :disabled="settings.video_env_locked"
                                        placeholder="API key"
                                        :class="inputClass"
                                    />
                                </div>
                                <div>
                                    <label
                                        for="video-secret"
                                        class="block text-sm font-medium text-gray-900"
                                    >
                                        {{ t("settings.channels.video.secret_key") }}
                                    </label>
                                    <input
                                        id="video-secret"
                                        v-model="settings.video_settings.secret"
                                        type="password"
                                        :disabled="settings.video_env_locked"
                                        placeholder="••••••"
                                        :class="inputClass"
                                    />
                                </div>
                            </div>
                        </div>
                    </SectionCard>
                </template>
            </div>
        </div>

        <div
            v-if="isSystemAdmin && can('manage_channels')"
            class="flex shrink-0 items-center justify-end gap-x-3 border-t bg-gray-50 px-6 py-4"
        >
            <BaseButton :is-disabled="hasChanges() || allEnvLocked" @click="onSave">
                {{ t("common.button.save") }}
            </BaseButton>
        </div>
    </div>
</template>

<script setup>
import { t } from "@/i18n/index.js";

import { ref, computed, onMounted } from "vue";
import { RouterLink, useRouter } from "vue-router";
import BaseButton from "@/components/BaseButton.vue";
import DataTable from "@/components/DataTable.vue";
import SectionCard from "@/components/SectionCard.vue";
import ToggleSwitch from "@/components/ToggleSwitch.vue";
import EnvManagedNotice from "@/components/Settings/EnvManagedNotice.vue";
import chatService from "@/services/chatService";
import settingsService from "@/services/settingsService";
import { channelTypes } from "@/constants/channels";
import { LockClosedIcon } from "@heroicons/vue/24/outline";
import LetterAvatar from "@/components/LetterAvatar.vue";
import { MagnifyingGlassIcon, UsersIcon } from "@heroicons/vue/20/solid";
import { usePermissions } from "@/composables/usePermissions";
import useDateOperations from "@/composables/useDateOperations.js";
import { useAlertStore } from "@/store/alerts";
import { useUserStore } from "@/store/user";

const router = useRouter();
const { can } = usePermissions();
const { getDate } = useDateOperations();
const userStore = useUserStore();
const isSystemAdmin = computed(() => userStore.user?.role === "system_admin");
const oldSettings = ref({});
const settings = ref({
    video_settings: {
        enabled: false,
        host: "",
        key: "",
        secret: "",
    },

    gif_settings: {
        enabled: false,
        api_key: "",
    },

    link_preview_settings: {
        enabled: false,
    },
});

const inputClass =
    "mt-2 block w-full rounded-md border-0 px-3 py-1.5 text-sm text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 disabled:cursor-not-allowed disabled:bg-gray-50 disabled:text-gray-500";

const channels = ref([]);
const loaded = ref(false);
const searchQuery = ref("");
const sort = ref({ key: "name", desc: false });

const columns = computed(() => [
    { key: "name", label: t.value("data_table.name"), sortable: true, sortKey: "displayname" },
    { key: "type", label: t.value("data_table.type"), sortable: true, hiddenBelow: "md" },
    {
        key: "member_count",
        label: t.value("data_table.members"),
        sortable: true,
        hiddenBelow: "sm",
    },
    {
        key: "msg_count",
        label: t.value("data_table.messages"),
        sortable: true,
        hiddenBelow: "lg",
    },
    {
        key: "last_post",
        label: t.value("data_table.last_activity"),
        sortable: true,
        hiddenBelow: "md",
    },
    { key: "created", label: t.value("data_table.created"), sortable: true, hiddenBelow: "lg" },
    { key: "actions", label: t.value("data_table.actions"), srOnly: true, align: "right" },
]);

const filteredChannels = computed(() => {
    const q = searchQuery.value.trim().toLowerCase();

    return channels.value
        .map((channel) => ({ ...channel, member_count: channel.channel_members?.length ?? 0 }))
        .filter((channel) => {
            if (!q) return true;

            return (
                (channel.displayname ?? "").toLowerCase().includes(q) ||
                (channel.description ?? "").toLowerCase().includes(q)
            );
        });
});

function editChannel(channel) {
    router.push({ name: "channel", params: { id: channel.id } });
}

async function onSave() {
    await settingsService
        .updateChatSettings({
            video_settings: settings.value.video_settings,
            gif_settings: settings.value.gif_settings,
            link_preview_settings: settings.value.link_preview_settings,
        })
        .then(() => {
            oldSettings.value = JSON.parse(JSON.stringify(settings.value));
            useAlertStore().showSuccess(t.value("settings.channels.settings_saved"));
        })
        .catch(() => {
            useAlertStore().showError(t.value("settings.channels.settings_save_failed"));
        });
}

const allEnvLocked = computed(
    () =>
        settings.value.video_env_locked &&
        settings.value.gif_env_locked &&
        settings.value.link_preview_env_locked,
);

function hasChanges() {
    return JSON.stringify(oldSettings.value) == JSON.stringify(settings.value);
}

onMounted(async () => {
    try {
        const { data } = await chatService.getAllChannels();

        channels.value = data ?? [];
    } catch {
        useAlertStore().showError(t.value("settings.channels.list_failed"));
    } finally {
        loaded.value = true;
    }

    if (!isSystemAdmin.value) {
        return;
    }

    await settingsService.getChatSettings().then((response) => {
        settings.value = response.data;

        oldSettings.value = JSON.parse(JSON.stringify(settings.value));
    });
});
</script>
