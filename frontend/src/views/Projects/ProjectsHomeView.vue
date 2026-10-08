<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div class="flex h-full flex-col bg-white">
        <div class="flex-none border-b border-gray-200 px-4 py-4 sm:px-6">
            <div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
                <div class="min-w-0">
                    <h1 class="flex items-center gap-x-2 text-base font-semibold text-gray-900">
                        {{ t("projects.home.title") }}
                        <span
                            class="rounded-full bg-gray-100 px-2 py-0.5 text-xs font-medium text-gray-600"
                        >
                            {{
                                shown.length !== workspaces.length
                                    ? `${shown.length} / ${workspaces.length}`
                                    : workspaces.length
                            }}
                        </span>
                    </h1>
                    <p class="mt-0.5 text-sm text-gray-500">{{ t("projects.home.description") }}</p>
                </div>

                <div class="flex items-center gap-x-2">
                    <div class="relative">
                        <MagnifyingGlassIcon
                            class="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-gray-400"
                            aria-hidden="true"
                        />
                        <input
                            v-model="search"
                            type="search"
                            :placeholder="t('projects.home.search')"
                            class="block w-full rounded-md border-0 py-1.5 pl-9 pr-3 text-sm text-gray-900 ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:w-52"
                        />
                    </div>

                    <Listbox v-model="sortBy" as="div" class="relative flex-none">
                        <ListboxButton
                            class="relative w-40 cursor-default rounded-md bg-white py-1.5 pl-3 pr-9 text-left text-sm text-gray-900 ring-1 ring-inset ring-gray-300 focus:outline-none focus:ring-2 focus:ring-indigo-600"
                        >
                            <span class="block truncate">{{ sortLabel(sortBy) }}</span>
                            <span
                                class="pointer-events-none absolute inset-y-0 right-0 flex items-center pr-2"
                            >
                                <ChevronUpDownIcon
                                    class="h-4 w-4 text-gray-400"
                                    aria-hidden="true"
                                />
                            </span>
                        </ListboxButton>
                        <ListboxOptions
                            class="absolute right-0 z-10 mt-1 w-44 overflow-auto rounded-md bg-white py-1 text-sm shadow-lg ring-1 ring-black/5 focus:outline-none"
                        >
                            <ListboxOption
                                v-for="option in sortOptions"
                                :key="option"
                                v-slot="{ active, selected }"
                                :value="option"
                                as="template"
                            >
                                <li
                                    :class="[
                                        active ? 'bg-indigo-600 text-white' : 'text-gray-900',
                                        'relative cursor-default select-none py-2 pl-3 pr-9',
                                    ]"
                                >
                                    <span
                                        :class="[
                                            selected ? 'font-semibold' : 'font-normal',
                                            'block truncate',
                                        ]"
                                    >
                                        {{ sortLabel(option) }}
                                    </span>
                                    <span
                                        v-if="selected"
                                        :class="[
                                            active ? 'text-white' : 'text-indigo-600',
                                            'absolute inset-y-0 right-0 flex items-center pr-3',
                                        ]"
                                    >
                                        <CheckIcon class="h-4 w-4" aria-hidden="true" />
                                    </span>
                                </li>
                            </ListboxOption>
                        </ListboxOptions>
                    </Listbox>

                    <button
                        type="button"
                        class="inline-flex flex-none items-center gap-x-1.5 rounded-md bg-indigo-600 px-3 py-1.5 text-sm font-semibold text-white shadow-sm hover:bg-indigo-500 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600"
                        @click="workspaceStore.setWorkspaceCreateDialogOpen(true)"
                    >
                        <PlusIcon class="-ml-0.5 h-4 w-4" aria-hidden="true" />
                        <span class="hidden sm:inline">{{ t("projects.home.new_workspace") }}</span>
                    </button>
                </div>
            </div>
        </div>

        <div class="min-h-0 flex-1 overflow-y-auto px-4 py-6 sm:px-6">
            <ul
                v-if="shown.length"
                role="list"
                class="grid grid-cols-[repeat(auto-fill,minmax(18rem,1fr))] gap-4"
            >
                <ItemCard
                    v-for="ws in shown"
                    :key="ws.id"
                    :title="ws.title"
                    :description="ws.description"
                    :description-fallback="t('projects.home.no_description')"
                    @click="openWorkspace(ws)"
                    @contextmenu.stop.prevent="toggleMenu($event, ws)"
                >
                    <template #icon>
                        <LetterAvatar
                            :id="ws.id"
                            :name="ws.title"
                            class="h-10 w-10 rounded-lg text-sm"
                        />
                    </template>
                    <template v-if="ws.user_roles?.length" #subtitle>
                        <p class="truncate text-xs text-gray-500">
                            {{ ws.user_roles.map((r) => roleNameLabel(r, "projects")).join(", ") }}
                        </p>
                    </template>
                    <template v-if="canManage(ws)" #actions>
                        <button
                            type="button"
                            class="rounded p-1 text-gray-400 hover:bg-gray-100 hover:text-indigo-600"
                            @click="toggleMenu($event, ws)"
                        >
                            <span class="sr-only">{{ t("projects.folder_view.actions") }}</span>
                            <EllipsisHorizontalIcon class="h-5 w-5" aria-hidden="true" />
                        </button>
                    </template>

                    <div class="mt-3 flex items-center gap-x-4 px-4 text-xs text-gray-500">
                        <span class="inline-flex items-center gap-x-1">
                            <TableCellsIcon class="h-4 w-4 text-gray-400" aria-hidden="true" />
                            {{ t("projects.home.tables", { count: ws.table_count || 0 }) }}
                        </span>
                        <span class="inline-flex items-center gap-x-1">
                            <UsersIcon class="h-4 w-4 text-gray-400" aria-hidden="true" />
                            {{ t("projects.home.members", { count: ws.members?.length || 0 }) }}
                        </span>
                        <span v-if="ws.group_count" class="inline-flex items-center gap-x-1">
                            <UserGroupIcon class="h-4 w-4 text-gray-400" aria-hidden="true" />
                            {{ t("projects.home.groups", { count: ws.group_count }) }}
                        </span>
                    </div>

                    <template #footer>
                        <div class="flex -space-x-1.5">
                            <span
                                v-for="member in (ws.members || []).slice(0, 4)"
                                :key="member.user_id"
                                class="inline-block h-6 w-6 rounded-full ring-2 ring-white"
                            >
                                <UserAvatar :user="memberUser(member)" />
                            </span>
                        </div>
                        <span class="text-xs text-gray-400">{{ getDate(ws.created_at) }}</span>
                    </template>
                </ItemCard>
            </ul>

            <div v-else class="flex h-full flex-col items-center justify-center text-center">
                <MagnifyingGlassIcon class="h-10 w-10 text-gray-300" aria-hidden="true" />
                <p class="mt-3 text-sm font-medium text-gray-500">
                    {{ t("projects.home.no_results") }}
                </p>
                <button
                    type="button"
                    class="mt-2 text-sm text-indigo-600 hover:text-indigo-500"
                    @click="search = ''"
                >
                    {{ t("projects.home.clear_search") }}
                </button>
            </div>
        </div>

        <NavigationItemMenu
            :visible="menuVisible"
            :anchor="menuAnchor"
            :item="menuItem"
            @close="
                menuVisible = false;
                menuItem = null;
            "
        />
    </div>
</template>

<script setup>
import { computed, ref, shallowRef } from "vue";
import { anchorFromEvent } from "@/composables/useAnchoredPopup";
import { useRouter } from "vue-router";
import { Listbox, ListboxButton, ListboxOption, ListboxOptions } from "@headlessui/vue";
import { CheckIcon, ChevronUpDownIcon } from "@heroicons/vue/20/solid";
import {
    EllipsisHorizontalIcon,
    MagnifyingGlassIcon,
    PlusIcon,
    TableCellsIcon,
    UserGroupIcon,
    UsersIcon,
} from "@heroicons/vue/24/outline";
import { t } from "@/i18n/index.js";
import NavigationItemMenu from "@/components/Projects/Menu/NavigationItemMenu.vue";
import UserAvatar from "@/components/UserAvatar.vue";
import useDateOperations from "@/composables/useDateOperations.js";
import { useUserStore } from "@/store/user";
import { useWorkspaceStore } from "@/store/workspaces";
import LetterAvatar from "@/components/LetterAvatar.vue";
import ItemCard from "@/components/ItemCard.vue";
import { roleNameLabel } from "@/utils/roleLabels";

const router = useRouter();
const workspaceStore = useWorkspaceStore();
const userStore = useUserStore();
const { getDate } = useDateOperations();

const sortOptions = ["name", "newest"];
const search = ref("");
const sortBy = ref("name");

const menuVisible = ref(false);
const menuItem = ref(null);
const menuAnchor = shallowRef(null);

const workspaces = computed(() => workspaceStore.getWorkspaces || []);

const shown = computed(() => {
    const query = search.value.trim().toLowerCase();
    const list = workspaces.value.filter(
        (ws) =>
            !query ||
            ws.title?.toLowerCase().includes(query) ||
            ws.description?.toLowerCase().includes(query),
    );

    return sortBy.value === "newest"
        ? [...list].sort((a, b) => (b.created_at || 0) - (a.created_at || 0))
        : [...list].sort((a, b) => (a.title || "").localeCompare(b.title || ""));
});

function sortLabel(option) {
    return t.value(option === "newest" ? "projects.home.sort_newest" : "projects.home.sort_name");
}

function memberUser(member) {
    return {
        ...(member.user_info || {}),
        ...(userStore.usersMap[member.user_id] || {}),
        id: member.user_id,
    };
}

function canManage(ws) {
    if (userStore.user?.role === "system_admin") return true;
    const perms = ws.user_permissions || [];

    return perms.includes("update_workspace") || ws.can_delete;
}

function toggleMenu(event, ws) {
    if (!canManage(ws)) return;
    if (menuVisible.value && menuItem.value?.id === ws.id) {
        menuVisible.value = false;
        menuItem.value = null;

        return;
    }

    menuItem.value = ws;
    menuAnchor.value = anchorFromEvent(event);
    menuVisible.value = true;
}

function openWorkspace(ws) {
    router.push({ name: "project-folder", params: { id: ws.id, fid: ws.id } });
}
</script>
