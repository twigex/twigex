<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div class="h-full">
        <div class="flex flex-col h-full">
            <div
                class="sticky top-0 z-40 flex h-14 shrink-0 items-center gap-x-4 bg-gray-900 px-4 shadow-sm sm:gap-x-6 sm:px-2 lg:px-2"
            >
                <button
                    type="button"
                    class="-m-2.5 p-2.5 text-gray-300 hover:text-gray-100 lg:hidden"
                    @click="navigationStore.mobileOpen = !navigationStore.mobileOpen"
                >
                    <Bars3Icon class="h-6 w-6" aria-hidden="true" />
                </button>

                <div
                    class="flex flex-1 justify-between items-center gap-x-4 self-stretch lg:gap-x-6"
                >
                    <div class="flex gap-x-6">
                        <div class="flex h-16 shrink-0 items-center justify-center">
                            <img class="h-8 w-auto" :src="'/logo.png'" alt="Your Company" />
                        </div>
                    </div>

                    <div class="flex flex-row items-center justify-between gap-x-2 lg:gap-x-2">
                        <div class="hidden lg:block" v-for="item in sections" :key="item.name">
                            <li class="cursor-pointer flex">
                                <a
                                    :class="[
                                        isActive(item)
                                            ? 'bg-gray-800 text-white'
                                            : 'text-gray-400 hover:text-white hover:bg-gray-800',
                                        'group flex gap-x-3 rounded-md p-3 text-sm leading-6 font-semibold relative',
                                    ]"
                                    @click="open(item)"
                                >
                                    <component :is="item.icon" class="h-6 w-6" />

                                    <span class="sr-only">{{ item.displayname }}</span>
                                    <div
                                        v-if="item.name == 'chat' && totalUnreadPosts > 0"
                                        class="z-20 absolute flex items-center justify-center whitespace-nowrap rounded-full bg-red-500 px-1.5 text-center text-xs font-medium leading-5 text-white transform"
                                        :class="[
                                            totalUnreadPosts > 99
                                                ? 'top-0 -right-3'
                                                : 'top-0 right-0',
                                        ]"
                                    >
                                        {{ totalUnreadPosts > 99 ? "99+" : totalUnreadPosts }}
                                    </div>
                                </a>
                            </li>
                        </div>

                        <div class="flex">
                            <div
                                class="mx-1 w-px self-stretch my-3 bg-gray-200 dark:bg-gray-700"
                            ></div>

                            <button
                                type="button"
                                class="py-3 px-1 text-gray-400 hover:text-gray-500"
                            >
                                <NotificationMenu />
                            </button>

                            <div
                                class="mx-1 w-px self-stretch my-3 bg-gray-200 dark:bg-gray-700"
                            ></div>
                        </div>

                        <Menu v-if="userStore.user != null" as="div" class="relative">
                            <MenuButton class="-m-1.5 flex items-center p-1.5">
                                <div class="h-8 w-8">
                                    <UserAvatar :user="userStore.user" />
                                </div>
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
                                    class="absolute right-0 z-10 mt-2.5 w-32 origin-top-right rounded-md bg-white py-2 shadow-lg ring-1 ring-gray-900/5 focus:outline-none"
                                >
                                    <MenuItem
                                        class="cursor-pointer"
                                        @click="open(settings)"
                                        v-slot="{ active }"
                                    >
                                        <a
                                            :class="[
                                                active ? 'bg-gray-50' : '',
                                                'block px-3 py-1 text-sm leading-6 text-gray-900',
                                            ]"
                                            >{{ t("main.menu.settings") }}</a
                                        >
                                    </MenuItem>
                                    <MenuItem
                                        @click="logout"
                                        class="cursor-pointer"
                                        v-slot="{ active }"
                                    >
                                        <a
                                            :class="[
                                                active ? 'bg-gray-50' : '',
                                                'block px-3 py-1 text-sm leading-6 text-gray-900',
                                            ]"
                                            >{{ t("main.menu.sign_out") }}</a
                                        >
                                    </MenuItem>
                                    <MenuItem
                                        @click="aboutDialog = true"
                                        class="cursor-pointer"
                                        v-slot="{ active }"
                                    >
                                        <a
                                            :class="[
                                                active ? 'bg-gray-50' : '',
                                                'block px-3 py-1 text-sm leading-6 text-gray-900',
                                            ]"
                                            >{{ t("main.menu.about") }}</a
                                        >
                                    </MenuItem>
                                </MenuItems>
                            </transition>
                        </Menu>
                    </div>
                </div>
            </div>

            <MobileDrawer v-if="!navigationStore.hasSidebar" />

            <main class="xl:pl-0 flex h-[calc(100%-3.5rem)] overflow-x-hidden">
                <div v-if="loaded" class="flex h-full w-full">
                    <RouterView />
                </div>
            </main>

            <CollaboraOffice />
            <EuroOfficeEditor />
            <MediaViewer />
            <AboutDialog v-model="aboutDialog" />
            <VideoCallOverlay />
            <IncomingCallDialog />
            <OutgoingCallDialog />
        </div>
    </div>
</template>

<script setup>
import { t } from "@/i18n/index.js";
import CollaboraOffice from "@/components/Office/CollaboraOffice.vue";
import EuroOfficeEditor from "@/components/Office/EuroOfficeEditor.vue";
import NotificationMenu from "@/components/Notifications/NotificationMenu.vue";
import AboutDialog from "@/components/AboutDialog.vue";
import MediaViewer from "@/components/MediaViewer.vue";
import UserAvatar from "@/components/UserAvatar.vue";
import VideoCallOverlay from "@/components/Chat/Calls/VideoCallOverlay.vue";
import IncomingCallDialog from "@/components/Chat/Calls/IncomingCallDialog.vue";
import OutgoingCallDialog from "@/components/Chat/Calls/OutgoingCallDialog.vue";
import { ref, onMounted, onUnmounted, watch } from "vue";
import { RouterView, useRouter, useRoute } from "vue-router";
import { useUserStore } from "@/store/user";
import { useSettingsStore } from "@/store/settings";
import { useChannelsStore } from "@/store/channels";
import { useNavigationStore } from "@/store/navigation";
import { usePermissionsStore } from "@/store/permissions";
import { useAlertStore } from "@/store/alerts";
import userService from "@/services/userService";
import settingsService from "@/services/settingsService";
import authService from "@/services/authService";
import chatService from "@/services/chatService";
import collimatoService from "@/services/collimatoService";
import configService from "@/services/configService";
import { close, initialize, reconnectIfDead, sendPresenceEvent } from "@/js/websocket";
import { resetAllStores } from "@/store/reset";
import { Menu, MenuButton, MenuItem, MenuItems } from "@headlessui/vue";
import { Bars3Icon } from "@heroicons/vue/24/outline";
import MobileDrawer from "@/components/Navigation/MobileDrawer.vue";
import { useSections, sectionGuards } from "@/composables/useSections";
import { setLocale } from "@/i18n/index.js";

const settings = {
    name: t.value("main.menu.settings"),
    href: "#",
    path: "profile",
};

const aboutDialog = ref(false);
const router = useRouter();
const route = useRoute();
const userStore = useUserStore();
const settingsStore = useSettingsStore();
const channelsStore = useChannelsStore();
const navigationStore = useNavigationStore();

const loaded = ref(false);
const lastSent = ref(Date.now());
const MIN_DELAY = 6 * 1000;
const interval = ref(null);

const { sections, totalUnreadPosts, isActive, open } = useSections();

watch(totalUnreadPosts, (newVal) => {
    window.electronAPI?.notifyUnread?.(newVal > 0);
});

// Sub-nav links navigate without touching the drawer, so close it on any route change
watch(
    () => route.fullPath,
    () => {
        navigationStore.mobileOpen = false;
    },
);

function onWake() {
    if (document.visibilityState === "visible") {
        reconnectIfDead();
    }
}

onMounted(async () => {
    initialize();

    window.addEventListener("online", reconnectIfDead);
    document.addEventListener("visibilitychange", onWake);

    ["mousemove", "keydown", "scroll", "touchstart"].forEach((event) => {
        window.addEventListener(event, onUserActivity);
    });

    configService.config().then((response) => {
        settingsStore.setConfig(response.data);
    });

    collimatoService
        .status()
        .then((response) => {
            settingsStore.setCollimatoStatus(response.data);
        })
        .catch(() => {});

    settingsService
        .chatSettings()
        .then((res) => {
            settingsStore.setChatSettings(res.data);
        })
        .catch(() => {});

    userService.statuses().then((response) => {
        userStore.setStatuses(response.data);
    });

    chatService.getChannels().then((response) => {
        channelsStore.setChannels(response.data);
    });

    settingsService.metadata().then((res) => {
        settingsStore.setMetadata(res.data);
    });

    try {
        await Promise.all([
            settingsService.getLicense().then((response) => {
                settingsStore.setLicense(response.data);
            }),
            userService.preferences().then((response) => {
                userStore.setPreferences(response.data);

                setLocale(userStore.getLanguage);
            }),
            userService.me().then((response) => {
                userStore.setUser(response.data);
            }),
            userService.getMyPermissions().then((response) => {
                usePermissionsStore().setPermissions(response.data);
            }),
        ]);

        const currentSection = route.matched.find((r) => sectionGuards[r.name]);

        if (
            currentSection &&
            !usePermissionsStore().permissions.includes(sectionGuards[currentSection.name])
        ) {
            const fallback = sections.value[0];

            router.replace({ name: fallback?.path ?? "profile" });
        }
    } catch {
        useAlertStore().showError(t.value("main.error.load_failed"));
    } finally {
        loaded.value = userStore.user !== null;
    }
});

onUnmounted(() => {
    window.removeEventListener("online", reconnectIfDead);
    document.removeEventListener("visibilitychange", onWake);

    ["mousemove", "keydown", "scroll", "touchstart"].forEach((event) => {
        window.removeEventListener(event, onUserActivity);
    });

    clearInterval(interval.value);

    close();
});

function onUserActivity() {
    const now = Date.now();

    if (now - lastSent.value > MIN_DELAY) {
        sendPresenceEvent();

        lastSent.value = now;
    }
}

function logout() {
    close();

    authService.logOut().then(() => {
        resetAllStores();
        router.push({ name: "login" });
    });
}

watch(
    () => channelsStore.getUnreadChannelsCount(userStore.user ? userStore.user.id : 0),
    (val) => {
        if (val > 0) {
            document.title = `(${val}) Twigex`;
        } else {
            document.title = `Twigex`;
        }
    },
);
</script>
