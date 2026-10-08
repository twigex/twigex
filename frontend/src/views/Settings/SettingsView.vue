<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <SidebarLayout :title="t('main.sections.settings')">
        <template #sidebar="{ mini }">
            <SettingsNavigation :mini="mini" />
        </template>

        <div class="mx-auto lg:px-0 h-full">
            <div class="mx-auto w-full grow lg:flex lg:inset-y-0 h-full">
                <div class="flex-1 xl:flex lg:inset-y-0 h-full">
                    <div class="flex flex-col h-full w-full overflow-hidden">
                        <!-- Header hidden for views that render their own -->
                        <div
                            v-if="!selfHeadedRoutes.includes(route.name)"
                            class="shrink-0 flex items-center justify-between border-b px-6"
                            :class="headerConfig.description ? 'py-4' : 'h-14'"
                        >
                            <div>
                                <h1 class="text-base font-semibold text-gray-900">
                                    {{ headerConfig.title }}
                                </h1>
                                <p
                                    v-if="headerConfig.description"
                                    class="mt-0.5 text-xs text-gray-500"
                                >
                                    {{ headerConfig.description }}
                                </p>
                            </div>
                            <button
                                v-if="headerConfig.action"
                                type="button"
                                @click="router.push({ name: headerConfig.action.route })"
                                class="inline-flex items-center gap-x-1.5 rounded-md bg-indigo-600 px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-indigo-500 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600"
                            >
                                <PlusIcon class="-ml-0.5 h-4 w-4" aria-hidden="true" />
                                {{ headerConfig.action.label }}
                            </button>
                        </div>

                        <div class="flex-1 min-h-0 min-w-0 overflow-y-auto">
                            <RouterView />
                        </div>
                    </div>
                </div>
            </div>
        </div>
    </SidebarLayout>
</template>

<script setup>
import { computed } from "vue";
import { RouterView, useRoute, useRouter } from "vue-router";
import SidebarLayout from "@/components/Navigation/SidebarLayout.vue";
import SettingsNavigation from "@/components/Settings/SettingsNavigation.vue";
import { PlusIcon } from "@heroicons/vue/20/solid";
import { t } from "@/i18n/index";

const route = useRoute();
const router = useRouter();

const selfHeadedRoutes = [
    "new-role",
    "edit-role",
    "new-group",
    "edit-group",
    "channel",
    "collimato-settings-workspace",
    "projects-settings-workspace",
];

const headerConfig = computed(() => {
    const tr = t.value;
    const headers = {
        profile: {
            title: tr("settings.headers.profile.title"),
            description: tr("settings.headers.profile.description"),
        },
        password: {
            title: tr("settings.headers.password.title"),
            description: tr("settings.headers.password.description"),
        },
        notifications: {
            title: tr("settings.headers.notifications.title"),
            description: tr("settings.headers.notifications.description"),
        },
        "user-security": {
            title: tr("settings.headers.user_security.title"),
            description: tr("settings.headers.user_security.description"),
        },
        "channel-list": {
            title: tr("settings.headers.channel_list.title"),
            description: tr("settings.headers.channel_list.description"),
        },
        "projects-settings-list": {
            title: tr("settings.headers.projects.title"),
            description: tr("settings.headers.projects.description"),
        },
        "collimato-settings-list": {
            title: tr("settings.headers.collimato_workspaces.title"),
            description: tr("settings.headers.collimato_workspaces.description"),
        },
        users: {
            title: tr("settings.headers.users.title"),
            description: tr("settings.headers.users.description"),
        },
        groups: {
            title: tr("settings.headers.groups.title"),
            description: tr("settings.headers.groups.description"),
            action: {
                label: tr("settings.headers.groups.action"),
                route: "new-group",
            },
        },
        license: {
            title: tr("settings.headers.license.title"),
            description: tr("settings.headers.license.description"),
        },
        language: {
            title: tr("settings.headers.language.title"),
            description: tr("settings.headers.language.description"),
        },
        metadata: {
            title: tr("settings.headers.metadata.title"),
            description: tr("settings.headers.metadata.description"),
        },
        smtp: {
            title: tr("settings.headers.smtp.title"),
            description: tr("settings.headers.smtp.description"),
        },
        security: {
            title: tr("settings.headers.security.title"),
            description: tr("settings.headers.security.description"),
        },
        "office-settings": {
            title: tr("settings.headers.office.title"),
            description: tr("settings.headers.office.description"),
        },
        "storage-settings": {
            title: tr("settings.headers.storage.title"),
            description: tr("settings.headers.storage.description"),
        },
        "edit-user": {
            title: tr("settings.headers.edit_user.title"),
            description: tr("settings.headers.edit_user.description"),
        },
        "new-user": {
            title: tr("settings.headers.new_user.title"),
            description: tr("settings.headers.new_user.description"),
        },
        authorization: {
            title: tr("settings.headers.authorization.title"),
            description: tr("settings.headers.authorization.description"),
        },
        roles: {
            title: tr("settings.headers.roles.title"),
            description: tr("settings.headers.roles.description"),
            action: {
                label: tr("settings.headers.roles.action"),
                route: "new-role",
            },
        },
    };

    return headers[route.name] ?? { title: tr("settings.headers.default.title") };
});
</script>
