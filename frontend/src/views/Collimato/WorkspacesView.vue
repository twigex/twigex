<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div class="flex flex-col h-full">
        <!-- Loading -->
        <div v-if="!loaded" class="flex h-full items-center justify-center">
            <div
                class="h-8 w-8 rounded-full border-4 border-indigo-500 border-t-transparent animate-spin"
            ></div>
        </div>

        <template v-else>
            <div
                v-if="hasLapsedWorkspaces"
                class="flex shrink-0 items-start gap-x-2 border-b border-amber-200 bg-amber-50 px-6 py-3"
            >
                <LockClosedIcon class="mt-0.5 h-4 w-4 shrink-0 text-amber-700" aria-hidden="true" />
                <div class="min-w-0">
                    <p class="text-sm font-medium text-amber-800">
                        {{ t("collimato.lapsed.title") }}
                    </p>
                    <p class="mt-0.5 text-xs text-amber-700">
                        {{ t("collimato.lapsed.description") }}
                    </p>
                </div>
            </div>

            <!-- Header -->
            <div class="shrink-0 border-b px-6 py-4">
                <div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
                    <div>
                        <h1 class="flex items-center gap-x-2 text-base font-semibold text-gray-900">
                            {{ t("collimato.your_workspaces.title") }}
                            <span
                                v-if="workspaces.length > 0"
                                class="rounded-full bg-gray-100 px-2 py-0.5 text-xs font-medium text-gray-600"
                            >
                                {{ filteredWorkspaces.length
                                }}{{
                                    filteredWorkspaces.length !== workspaces.length
                                        ? ` / ${workspaces.length}`
                                        : ""
                                }}
                            </span>
                        </h1>
                        <p class="mt-0.5 text-sm text-gray-500">
                            {{ t("collimato.your_workspaces.description") }}
                        </p>
                    </div>
                    <div class="flex items-center gap-x-2">
                        <!-- Search -->
                        <div class="relative">
                            <MagnifyingGlassIcon
                                class="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-gray-400"
                                aria-hidden="true"
                            />
                            <input
                                v-model="searchQuery"
                                type="search"
                                :placeholder="t('collimato.your_workspaces.search_placeholder')"
                                class="block w-52 rounded-md border-0 py-1.5 pl-9 pr-3 text-sm text-gray-900 ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600"
                            />
                        </div>
                        <!-- Sort -->
                        <Listbox v-model="sortBy" as="div" class="relative">
                            <ListboxButton
                                class="relative w-36 cursor-default rounded-md bg-white py-1.5 pl-3 pr-9 text-left text-sm text-gray-900 ring-1 ring-inset ring-gray-300 focus:outline-none focus:ring-2 focus:ring-indigo-600"
                            >
                                <span class="block truncate">{{ sortBy.label }}</span>
                                <span
                                    class="pointer-events-none absolute inset-y-0 right-0 flex items-center pr-2"
                                >
                                    <ChevronUpDownIcon
                                        class="h-4 w-4 text-gray-400"
                                        aria-hidden="true"
                                    />
                                </span>
                            </ListboxButton>
                            <transition
                                leave-active-class="transition ease-in duration-100"
                                leave-from-class="opacity-100"
                                leave-to-class="opacity-0"
                            >
                                <ListboxOptions
                                    class="absolute right-0 z-10 mt-1 w-40 overflow-auto rounded-md bg-white py-1 text-sm shadow-lg ring-1 ring-black/5 focus:outline-none"
                                >
                                    <ListboxOption
                                        v-for="option in sortOptions"
                                        :key="option.value"
                                        :value="option"
                                        as="template"
                                        v-slot="{ active, selected }"
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
                                                    selected ? 'font-semibold' : 'font-normal',
                                                    'block truncate',
                                                ]"
                                                >{{ option.label }}</span
                                            >
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
                            </transition>
                        </Listbox>
                        <!-- New workspace -->
                        <button
                            v-if="canProvisionWorkspaces && canCreateWorkspace"
                            type="button"
                            @click="router.push({ name: 'workspace-setup' })"
                            class="inline-flex items-center gap-x-1.5 rounded-md bg-indigo-600 px-3 py-1.5 text-sm font-semibold text-white shadow-sm hover:bg-indigo-500 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600"
                        >
                            <PlusIcon class="-ml-0.5 h-4 w-4" aria-hidden="true" />
                            {{ t("collimato.button.new_workspace") }}
                        </button>
                        <span
                            v-else-if="canProvisionWorkspaces"
                            class="inline-flex items-center gap-x-1 rounded-md bg-amber-50 px-2 py-1 text-xs font-medium text-amber-800 ring-1 ring-inset ring-amber-600/20"
                        >
                            <LockClosedIcon class="h-3 w-3" aria-hidden="true" />
                            {{ t("settings.license.business_plan") }}
                        </span>
                    </div>
                </div>

                <!-- Ownership filter tabs -->
                <div v-if="workspaces.length > 0" class="mt-3 flex gap-x-1">
                    <button
                        v-for="tab in ownershipTabs"
                        :key="tab.value"
                        type="button"
                        @click="ownerFilter = tab.value"
                        class="inline-flex items-center gap-x-1 rounded-full px-3 py-1 text-xs font-medium transition-colors"
                        :class="
                            ownerFilter === tab.value
                                ? 'bg-indigo-600 text-white'
                                : 'bg-gray-100 text-gray-600 hover:bg-gray-200'
                        "
                    >
                        {{ tab.label }}
                        <span
                            v-if="tab.count !== null"
                            class="rounded-full px-1.5 py-0.5 text-xs font-semibold"
                            :class="
                                ownerFilter === tab.value
                                    ? 'bg-indigo-500 text-white'
                                    : 'bg-gray-200 text-gray-600'
                            "
                        >
                            {{ tab.count }}
                        </span>
                    </button>
                </div>
            </div>

            <!-- Content area -->
            <div class="flex-1 overflow-y-auto px-6 py-4">
                <!-- Empty state: nothing to show, and nothing this user can do about it -->
                <div
                    v-if="workspaces.length === 0 && !canProvisionWorkspaces"
                    class="flex h-full flex-col items-center justify-center text-center"
                >
                    <UserGroupIcon class="mx-auto h-16 w-16 text-gray-300" aria-hidden="true" />
                    <h2 class="mt-4 text-lg font-semibold text-gray-900">
                        {{ t("collimato.not_invited.title") }}
                    </h2>
                    <p class="mt-2 max-w-sm text-sm text-gray-500">
                        {{ t("collimato.not_invited.description") }}
                    </p>
                </div>

                <!-- Empty state: no workspaces at all -->
                <div
                    v-else-if="workspaces.length === 0"
                    class="flex h-full flex-col items-center justify-center text-center"
                >
                    <CubeIcon class="mx-auto h-16 w-16 text-indigo-400" aria-hidden="true" />
                    <h2 class="mt-4 text-lg font-semibold text-gray-900">
                        {{ t("collimato.create_first_workspace.title") }}
                    </h2>
                    <p class="mt-2 max-w-sm text-sm text-gray-500">
                        {{ t("collimato.create_first_workspace.description") }}
                    </p>
                    <button
                        v-if="canCreateWorkspace"
                        type="button"
                        @click="router.push({ name: 'workspace-setup' })"
                        class="mt-6 inline-flex items-center gap-x-1.5 rounded-md bg-indigo-600 px-4 py-2 text-sm font-semibold text-white shadow-sm hover:bg-indigo-500 disabled:bg-indigo-300"
                    >
                        <PlusIcon class="-ml-0.5 h-4 w-4" aria-hidden="true" />
                        {{ t("collimato.button.create_workspace") }}
                    </button>
                </div>

                <!-- Empty state: search / filter has no results -->
                <div
                    v-else-if="filteredWorkspaces.length === 0"
                    class="flex h-full flex-col items-center justify-center text-center"
                >
                    <MagnifyingGlassIcon class="h-10 w-10 text-gray-300" aria-hidden="true" />
                    <p class="mt-3 text-sm font-medium text-gray-500">
                        {{ t("collimato.your_workspaces.no_results") }}
                    </p>
                    <button
                        type="button"
                        @click="
                            searchQuery = '';
                            ownerFilter = 'all';
                        "
                        class="mt-2 text-sm text-indigo-600 hover:text-indigo-500"
                    >
                        {{ t("collimato.your_workspaces.clear_filters") }}
                    </button>
                </div>

                <!-- Card grid -->
                <ul
                    v-else
                    role="list"
                    class="grid grid-cols-[repeat(auto-fill,minmax(18rem,1fr))] gap-4 content-start"
                >
                    <ItemCard
                        v-for="ws in filteredWorkspaces"
                        :key="ws.id"
                        :title="ws.name"
                        :description="ws.description"
                        :description-fallback="t('common.label.no_description')"
                        @click="openWorkspace(ws)"
                    >
                        <template #icon>
                            <LetterAvatar
                                :id="ws.id"
                                :name="ws.name"
                                class="h-10 w-10 rounded-lg text-sm"
                            />
                        </template>
                        <template #subtitle>
                            <div class="flex flex-wrap items-center gap-1.5">
                                <span
                                    class="inline-flex rounded-full px-2 py-0.5 text-xs font-medium capitalize"
                                    :class="statusConfig(ws.status).badge"
                                >
                                    {{ ws.status }}
                                </span>
                                <span
                                    v-if="
                                        ws.status !== 'draft' &&
                                        ws.created_by &&
                                        ws.created_by !== userStore.user?.id
                                    "
                                    class="inline-flex items-center gap-x-1 rounded-full bg-sky-50 px-2 py-0.5 text-xs font-medium text-sky-700"
                                >
                                    <UsersIcon class="h-3 w-3" aria-hidden="true" />
                                    {{ t("collimato.your_workspaces.shared") }}
                                </span>
                            </div>
                        </template>
                        <template v-if="ws.can_delete" #actions>
                            <button
                                type="button"
                                class="rounded p-1 text-gray-400 hover:bg-red-50 hover:text-red-600"
                                :title="t('common.button.delete')"
                                @click="openDeleteDialog(ws)"
                            >
                                <TrashIcon class="h-5 w-5" aria-hidden="true" />
                            </button>
                        </template>
                        <template #footer>
                            <div class="flex items-center">
                                <div class="flex -space-x-1.5">
                                    <span
                                        v-for="memberId in ws.member_ids || []"
                                        :key="memberId"
                                        class="inline-block h-6 w-6 rounded-full ring-2 ring-white"
                                    >
                                        <UserAvatar :user-id="memberId" />
                                    </span>
                                </div>
                                <span
                                    v-if="(ws.member_count || 0) > (ws.member_ids?.length || 0)"
                                    class="ml-1.5 text-xs text-gray-500"
                                >
                                    +{{ ws.member_count - ws.member_ids.length }}
                                </span>
                            </div>
                            <span
                                v-if="ws.status === 'draft'"
                                class="inline-flex items-center gap-x-1 text-xs font-medium text-indigo-600"
                            >
                                {{ t("collimato.your_workspaces.continue_setup") }}
                                <ArrowRightIcon class="h-3 w-3" aria-hidden="true" />
                            </span>
                            <span v-else class="text-xs text-gray-400">{{
                                getDateAndTime(ws.created_at)
                            }}</span>
                        </template>
                    </ItemCard>
                </ul>
            </div>
        </template>

        <DeleteDialog
            v-model="deleteDialog"
            :title="t('collimato.delete_dialog.title')"
            :message="t('collimato.delete_dialog.confirm')"
            @delete="deleteWorkspace"
        />
    </div>
</template>

<script setup>
import { t } from "@/i18n/index.js";
import { ref, computed, onMounted } from "vue";
import { useRouter } from "vue-router";
import DeleteDialog from "@/components/Collimato/Dialogs/DeleteDialog.vue";
import ItemCard from "@/components/ItemCard.vue";
import LetterAvatar from "@/components/LetterAvatar.vue";
import UserAvatar from "@/components/UserAvatar.vue";
import collimatoService from "@/services/collimatoService";
import useDateOperations from "@/composables/useDateOperations.js";
import { useAlertStore } from "@/store/alerts";
import { extractErrorMessage } from "@/utils/errors";
import { useSettingsStore } from "@/store/settings";
import { usePermissionsStore } from "@/store/permissions";
import { useUserStore } from "@/store/user";
import { Listbox, ListboxButton, ListboxOption, ListboxOptions } from "@headlessui/vue";
import { CubeIcon, UserGroupIcon, LockClosedIcon } from "@heroicons/vue/24/outline";
import {
    TrashIcon,
    PlusIcon,
    MagnifyingGlassIcon,
    UsersIcon,
    CheckIcon,
    ChevronUpDownIcon,
    ArrowRightIcon,
} from "@heroicons/vue/20/solid";

const router = useRouter();
const alertStore = useAlertStore();
const userStore = useUserStore();
const settingsStore = useSettingsStore();
const { getDateAndTime } = useDateOperations();

const loaded = ref(false);
const workspaces = ref([]);
const deleteDialog = ref(false);
const itemToDelete = ref(null);
const searchQuery = ref("");
const ownerFilter = ref("all");

const sortOptions = [
    { value: "newest", label: t.value("collimato.your_workspaces.sort.newest") },
    { value: "oldest", label: t.value("collimato.your_workspaces.sort.oldest") },
    { value: "name_asc", label: t.value("collimato.your_workspaces.sort.name_asc") },
    { value: "name_desc", label: t.value("collimato.your_workspaces.sort.name_desc") },
];
const sortBy = ref(sortOptions[0]);

const hasUnlimitedWorkspaces = computed(() =>
    settingsStore.getLicenseFeature("collimato_unlimited_workspaces"),
);
// Unlicensed instances get a single workspace; the count is already loaded here.
const canCreateWorkspace = computed(
    () => hasUnlimitedWorkspaces.value || workspaces.value.length < 1,
);
// An unlicensed instance can only ever have created one workspace, so more than
// one means the licence lapsed rather than never having existed.
const hasLapsedWorkspaces = computed(
    () => !hasUnlimitedWorkspaces.value && workspaces.value.length > 1,
);
// Workspaces hold database credentials, so provisioning is a granted capability
// rather than something every member has.
const canProvisionWorkspaces = computed(() =>
    usePermissionsStore().permissions.includes("create_collimato_workspace"),
);

const ownedCount = computed(
    () =>
        workspaces.value.filter((ws) => !ws.created_by || ws.created_by === userStore.user?.id)
            .length,
);
const sharedCount = computed(
    () =>
        workspaces.value.filter((ws) => ws.created_by && ws.created_by !== userStore.user?.id)
            .length,
);

const ownershipTabs = computed(() => [
    {
        value: "all",
        label: t.value("collimato.your_workspaces.filter.all"),
        count: workspaces.value.length,
    },
    {
        value: "mine",
        label: t.value("collimato.your_workspaces.filter.mine"),
        count: ownedCount.value,
    },
    {
        value: "shared",
        label: t.value("collimato.your_workspaces.filter.shared"),
        count: sharedCount.value,
    },
]);

const filteredWorkspaces = computed(() => {
    let result = [...workspaces.value];

    const q = searchQuery.value.trim().toLowerCase();

    if (q) {
        result = result.filter(
            (ws) => ws.name.toLowerCase().includes(q) || ws.description?.toLowerCase().includes(q),
        );
    }

    if (ownerFilter.value === "mine") {
        result = result.filter((ws) => !ws.created_by || ws.created_by === userStore.user?.id);
    } else if (ownerFilter.value === "shared") {
        result = result.filter((ws) => ws.created_by && ws.created_by !== userStore.user?.id);
    }

    switch (sortBy.value.value) {
        case "oldest":
            result.sort((a, b) => a.created_at - b.created_at);
            break;
        case "name_asc":
            result.sort((a, b) => a.name.localeCompare(b.name));
            break;
        case "name_desc":
            result.sort((a, b) => b.name.localeCompare(a.name));
            break;
        default:
            result.sort((a, b) => b.created_at - a.created_at);
    }

    return result;
});

function statusConfig(status) {
    const configs = {
        active: { badge: "bg-green-100 text-green-700" },
        draft: { badge: "bg-amber-100 text-amber-700" },
        inactive: { badge: "bg-gray-100 text-gray-600" },
        finished: { badge: "bg-emerald-100 text-emerald-700" },
    };

    return configs[status] ?? configs["active"];
}

function openWorkspace(workspace) {
    if (workspace.status === "draft") {
        router.push({ name: "workspace-setup", query: { workspaceId: workspace.id } });

        return;
    }

    router.push({ name: "collimato-workspace", params: { workspaceId: workspace.id } });
}

function openDeleteDialog(item) {
    itemToDelete.value = item;
    deleteDialog.value = true;
}

function deleteWorkspace() {
    collimatoService
        .deleteWorkspace(itemToDelete.value.id)
        .then(() => {
            workspaces.value = workspaces.value.filter((item) => item.id !== itemToDelete.value.id);
            itemToDelete.value = null;
        })
        .catch((error) => {
            alertStore.showError(extractErrorMessage(error));
        });
    deleteDialog.value = false;
}

onMounted(() => {
    collimatoService.getWorkspaces().then((response) => {
        workspaces.value = response.data;
        loaded.value = true;
    });
});
</script>
