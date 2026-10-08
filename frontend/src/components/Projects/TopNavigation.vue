<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div class="flex flex-col w-full" ref="listboxWrapper">
        <div class="tn-row1">
            <template v-if="route.params.tid != undefined">
                <span class="tn-avatar">{{
                    tableName ? tableName.charAt(0).toUpperCase() : "?"
                }}</span>
                <span class="tn-table-name" :title="tableName">{{ tableName }}</span>
            </template>
            <FolderNavigation v-else />

            <div v-if="route.params.tid != undefined" class="tn-actions">
                <FilterMenu v-if="showFilters" @apply="handleApplyFilters" />

                <Popover
                    v-if="route.path.includes('/grid/')"
                    style="display: inline-flex; align-items: center"
                    v-slot="{ close }"
                >
                    <PopoverButton
                        :as="BaseButton"
                        size="small"
                        variant="secondary"
                        :prepend-icon="AdjustmentsHorizontalIcon"
                        class="relative focus:outline-none focus-visible:outline-none"
                    >
                        <span
                            v-if="sortActive"
                            class="absolute -right-0.5 -top-0.5 h-2 w-2 rounded-full bg-green-500 ring-2 ring-white"
                        />
                        <span class="tn-btn-label">{{ t("projects.task_sort.sort") }}</span>
                    </PopoverButton>
                    <transition
                        enter-active-class="transition ease-out duration-100"
                        enter-from-class="transform opacity-0 scale-95"
                        enter-to-class="transform opacity-100 scale-100"
                        leave-active-class="transition ease-in duration-75"
                        leave-from-class="transform opacity-100 scale-100"
                        leave-to-class="transform opacity-0 scale-95"
                    >
                        <teleport to="body">
                            <PopoverPanel
                                class="fixed top-[100px] left-[calc(50%-250px)] z-40 w-[500px] rounded-md bg-white shadow-lg ring-1 ring-black/5"
                            >
                                <TaskSort @close="close" />
                            </PopoverPanel>
                        </teleport>
                    </transition>
                </Popover>

                <BaseButton
                    v-if="currentView && route.name !== 'gantt-view'"
                    type="button"
                    size="small"
                    variant="secondary"
                    :prepend-icon="ShareIcon"
                    :is-disabled="!canManageViews"
                    @click="openShareDialog()"
                >
                    <span class="tn-btn-label">{{ t("projects.view.share") }}</span>
                </BaseButton>

                <BaseButton
                    v-if="route.name !== 'gantt-view' && route.name !== 'calendar-view'"
                    type="button"
                    size="small"
                    variant="secondary"
                    :prepend-icon="EyeSlashIcon"
                    :is-disabled="!rolesStore.hasPermissionToUpdateWorkspaceView"
                    @click="rightSideNavigation = !rightSideNavigation"
                >
                    <span class="tn-btn-label">{{
                        t("projects.navigation.custom_cards_navigation.title")
                    }}</span>
                </BaseButton>

                <BaseButton
                    v-if="canCreateTask"
                    type="button"
                    size="small"
                    class="ml-1"
                    :prepend-icon="PlusIcon"
                    @click="workspaceStore.openNewTaskDialog()"
                >
                    <span class="tn-btn-label">{{ t("projects.new_task.button") }}</span>
                </BaseButton>
            </div>
        </div>

        <div v-if="route.params.tid != undefined" class="tn-row2">
            <!-- View tabs, overflow hidden, tabs that don't fit are clipped -->
            <div class="tn-tabs" ref="tabsEl">
                <div
                    v-for="(item, idx) in flatViewTabs"
                    :key="item.kind === 'create' ? 'create-' + item.group.type : item.view.id"
                    class="tn-tab-wrap"
                    :data-tab-idx="idx"
                >
                    <button
                        v-if="item.kind === 'create'"
                        class="tn-tab"
                        :class="{ 'tn-tab--active': isTypeActive(item.group.type) }"
                        :style="{ '--tab-color': item.group.color, '--tab-bg': item.group.bg }"
                        :disabled="
                            !['gantt', 'calendar'].includes(item.group.type) && !canCreateViews
                        "
                        @click="
                            (['gantt', 'calendar'].includes(item.group.type) || canCreateViews) &&
                            createView(item.group.type)
                        "
                    >
                        <component :is="item.group.icon" class="h-3.5 w-3.5 shrink-0" />
                        <span>{{ item.group.label }}</span>
                    </button>
                    <div
                        v-else
                        class="tn-tab"
                        :class="{ 'tn-tab--active': isViewActive(item.view) }"
                        :style="{ '--tab-color': item.group.color, '--tab-bg': item.group.bg }"
                        @click="changeSelected(item.view, false)"
                    >
                        <component :is="item.group.icon" class="h-3.5 w-3.5 shrink-0" />
                        <span>{{ item.view.name }}</span>
                        <ViewActionsMenu
                            v-if="
                                !item.view.is_default &&
                                canManageViews &&
                                item.group.type !== 'gantt'
                            "
                            :view="item.view"
                            :can-edit="rolesStore.hasPermissionToUpdateWorkspaceView"
                            :can-delete="rolesStore.hasPermissionToDeleteTableView"
                            @toggle-visibility="viewActions?.toggleVisibility($event)"
                            @edit="viewActions?.rename($event)"
                            @delete="viewActions?.remove($event)"
                        />
                    </div>
                </div>
            </div>

            <!-- ... button, outside tn-tabs, only when views overflow -->
            <div v-if="overflowItems.length > 0" class="tn-ov-wrap">
                <button class="tn-ov-btn" @click.stop="overflowOpen = !overflowOpen">
                    <EllipsisHorizontalIcon class="h-3.5 w-3.5" />
                </button>
                <div
                    v-if="overflowOpen"
                    class="tn-dropdown tn-ov-menu rounded-md bg-white shadow-lg ring-1 ring-black/5"
                >
                    <button
                        v-for="item in overflowItems"
                        :key="
                            item.kind === 'create'
                                ? 'ov-c-' + item.group.type
                                : 'ov-v-' + item.view.id
                        "
                        class="tn-tab tn-ov-item"
                        :class="{
                            'tn-tab--active':
                                item.kind === 'view'
                                    ? isViewActive(item.view)
                                    : isTypeActive(item.group.type),
                        }"
                        :style="{ '--tab-color': item.group.color, '--tab-bg': item.group.bg }"
                        @click="pickOverflow(item)"
                    >
                        <component :is="item.group.icon" class="h-3.5 w-3.5 shrink-0" />
                        <span>{{ item.kind === "view" ? item.view.name : item.group.label }}</span>
                    </button>
                </div>
            </div>

            <!-- New view, always visible, outside tn-tabs -->
            <Menu v-if="canCreateViews" as="div" class="relative flex-shrink-0">
                <MenuButton
                    :as="BaseButton"
                    size="small"
                    variant="secondary"
                    :prepend-icon="PlusIcon"
                >
                    {{ t("projects.top_navigation.new_view") }}
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
                        class="absolute left-0 z-20 mt-2 w-48 origin-top-left rounded-md bg-white py-1 shadow-lg ring-1 ring-black/5 focus:outline-none"
                    >
                        <MenuItem
                            v-for="vt in CREATABLE_VIEW_TYPES"
                            :key="vt.type"
                            v-slot="{ active }"
                        >
                            <button
                                type="button"
                                :class="[
                                    active ? 'bg-gray-100 text-gray-900' : 'text-gray-700',
                                    'group flex w-full items-center px-4 py-2 text-sm',
                                ]"
                                @click="createView(vt.type)"
                            >
                                <component
                                    :is="vt.icon"
                                    class="mr-3 h-5 w-5 text-gray-400 group-hover:text-gray-500"
                                    aria-hidden="true"
                                />
                                {{ vt.label }}
                            </button>
                        </MenuItem>
                    </MenuItems>
                </transition>
            </Menu>
        </div>

        <ViewActions ref="viewActions" />

        <FormDialog
            :open="kanbanViewDialog"
            :title="t('projects.top_navigation.create_new_kanban_view')"
            :confirm-label="t('projects.create_kanban_view.create_view')"
            :loading="creatingView"
            @confirm="handleCreateKanbanView"
            @close="closeKanbanDialog"
        >
            <input
                v-model="newKanbanViewName"
                type="text"
                class="block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6"
                :placeholder="t('projects.create_kanban_view.view_name')"
            />
            <div class="mt-4">
                <label class="block text-sm font-medium leading-6 text-gray-900">{{
                    t("projects.create_kanban_view.title")
                }}</label>
                <p class="text-xs text-gray-500 mb-2">
                    {{ t("projects.create_kanban_view.description") }}
                </p>
                <BaseSelect
                    :model-value="newKanbanGroupField?.name ?? ''"
                    :options="
                        kanbanHeaders.map((h) => ({
                            value: h.name,
                            label: h.display_name || h.name,
                        }))
                    "
                    @update:model-value="
                        (name) =>
                            (newKanbanGroupField =
                                kanbanHeaders.find((h) => h.name === name) ?? null)
                    "
                />
            </div>
            <div class="mt-4 flex items-center gap-3">
                <button
                    type="button"
                    @click="newKanbanViewIsPublic = !newKanbanViewIsPublic"
                    :class="[
                        newKanbanViewIsPublic ? 'bg-indigo-600' : 'bg-gray-200',
                        'relative inline-flex h-5 w-9 flex-shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors duration-200 ease-in-out focus:outline-none focus-visible:ring-2 focus-visible:ring-indigo-600 focus-visible:ring-offset-2',
                    ]"
                >
                    <span
                        :class="[
                            newKanbanViewIsPublic ? 'translate-x-4' : 'translate-x-0',
                            'pointer-events-none inline-block h-4 w-4 transform rounded-full bg-white shadow ring-0 transition duration-200 ease-in-out',
                        ]"
                    />
                </button>
                <span class="text-sm text-gray-700">
                    {{
                        newKanbanViewIsPublic
                            ? t("projects.view.visibility_public")
                            : t("projects.view.visibility_private")
                    }}
                </span>
            </div>
            <div v-if="creatingView" class="mt-4 flex items-center gap-2 text-sm text-gray-500">
                <div
                    class="animate-spin rounded-full h-4 w-4 border-2 border-indigo-600 border-t-transparent flex-shrink-0"
                ></div>
                <span>{{ t("projects.create_kanban_view.building") }}</span>
            </div>
        </FormDialog>

        <FormDialog
            :open="gridViewDialog"
            :title="t('projects.top_navigation.create_new_grid_view')"
            :confirm-label="t('projects.create_kanban_view.create_view')"
            :loading="creatingView"
            @confirm="handleCreateGridView"
            @close="closeCreatViewDialog"
        >
            <input
                v-model="newGridViewName"
                type="text"
                name="gridViewName"
                id="gridViewName"
                class="block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6"
                :placeholder="t('projects.top_navigation.enter_grid_view_name')"
            />
            <div class="mt-4 flex items-center gap-3">
                <button
                    type="button"
                    @click="newGridViewIsPublic = !newGridViewIsPublic"
                    :class="[
                        newGridViewIsPublic ? 'bg-indigo-600' : 'bg-gray-200',
                        'relative inline-flex h-5 w-9 flex-shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors duration-200 ease-in-out focus:outline-none focus-visible:ring-2 focus-visible:ring-indigo-600 focus-visible:ring-offset-2',
                    ]"
                >
                    <span
                        :class="[
                            newGridViewIsPublic ? 'translate-x-4' : 'translate-x-0',
                            'pointer-events-none inline-block h-4 w-4 transform rounded-full bg-white shadow ring-0 transition duration-200 ease-in-out',
                        ]"
                    />
                </button>
                <span class="text-sm text-gray-700">
                    {{
                        newGridViewIsPublic
                            ? t("projects.view.visibility_public")
                            : t("projects.view.visibility_private")
                    }}
                </span>
            </div>
        </FormDialog>

        <TransitionRoot appear :show="shareDialogOpen" as="template">
            <Dialog as="div" @close="closeShareDialog" class="relative z-50">
                <TransitionChild
                    as="template"
                    enter="duration-300 ease-out"
                    enter-from="opacity-0"
                    enter-to="opacity-100"
                    leave="duration-200 ease-in"
                    leave-from="opacity-100"
                    leave-to="opacity-0"
                >
                    <div class="fixed inset-0 bg-gray-500 bg-opacity-75 transition-opacity" />
                </TransitionChild>
                <div class="fixed inset-0 z-10 w-screen overflow-y-auto">
                    <div
                        class="flex min-h-full items-end justify-center p-4 text-center sm:items-center sm:p-0"
                    >
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
                                class="relative w-full transform overflow-visible rounded-lg bg-white px-4 pb-4 pt-5 text-left shadow-xl transition-all sm:my-8 sm:max-w-lg sm:p-6"
                            >
                                <DialogTitle
                                    as="h3"
                                    class="text-base font-semibold leading-6 text-gray-900"
                                >
                                    {{ t("projects.view.share_dialog_title") }}
                                </DialogTitle>
                                <div
                                    v-if="selectedShareUsersModel.length"
                                    class="mt-3 flex flex-wrap gap-1.5"
                                >
                                    <span
                                        v-for="user in selectedShareUsersModel"
                                        :key="user.id"
                                        class="inline-flex items-center gap-1.5 rounded-full bg-indigo-50 pl-1 pr-2 py-0.5 text-xs font-medium text-indigo-700 ring-1 ring-inset ring-indigo-200"
                                    >
                                        <div class="h-4 w-4 shrink-0">
                                            <UserAvatar :user="user" />
                                        </div>
                                        {{ user.name }} {{ user.lastname }}
                                        <button
                                            @click="toggleShareUser(user)"
                                            class="ml-0.5 hover:text-indigo-900"
                                        >
                                            ×
                                        </button>
                                    </span>
                                </div>
                                <div class="mt-2">
                                    <div
                                        class="w-full flex items-center rounded-md bg-white text-gray-900 ring-1 ring-inset ring-gray-300 focus-within:ring-2 focus-within:ring-indigo-600 shadow-sm sm:text-sm sm:leading-6"
                                    >
                                        <input
                                            ref="shareSearchInput"
                                            v-model="shareQuery"
                                            type="text"
                                            class="flex-1 bg-transparent border-none outline-none focus:outline-none focus:ring-0 py-1.5 px-3 shadow-none text-sm"
                                            :placeholder="
                                                t('projects.view.share_search_placeholder')
                                            "
                                            @click="openShareList"
                                            @input="onShareQueryInput"
                                        />
                                        <button
                                            class="pr-2 pl-1 text-gray-400 hover:text-gray-600"
                                            @click.stop="toggleShareList"
                                        >
                                            <ChevronUpDownIcon class="h-5 w-5" aria-hidden="true" />
                                        </button>
                                    </div>
                                    <!-- Results list, normal flow for working IntersectionObserver -->
                                    <div
                                        v-if="shareIsOpen"
                                        ref="shareListEl"
                                        class="mt-1 max-h-56 w-full overflow-auto rounded-md bg-white py-1 text-base shadow-lg ring-1 ring-black ring-opacity-5 sm:text-sm"
                                    >
                                        <ul>
                                            <li
                                                v-for="person in shareResults"
                                                :key="person.id"
                                                class="relative cursor-default select-none py-2 pl-3 pr-9 hover:bg-indigo-600 hover:text-white text-gray-900"
                                                :class="
                                                    isShareSelected(person)
                                                        ? 'bg-indigo-50 font-semibold'
                                                        : ''
                                                "
                                                @click="toggleShareUser(person)"
                                            >
                                                <div class="flex items-center">
                                                    <div class="h-6 w-6 shrink-0">
                                                        <UserAvatar :user="person" />
                                                    </div>
                                                    <span class="ml-3 truncate"
                                                        >{{ person.name }}
                                                        {{ person.lastname }}</span
                                                    >
                                                </div>
                                                <span
                                                    v-if="isShareSelected(person)"
                                                    class="absolute inset-y-0 right-0 flex items-center pr-4 text-indigo-600"
                                                >
                                                    <CheckIcon class="h-5 w-5" />
                                                </span>
                                            </li>
                                            <li ref="shareSentinel" class="h-1" />
                                            <li
                                                v-if="shareLoading"
                                                class="py-2 text-center text-xs text-gray-400"
                                            >
                                                {{ t("common.label.loading") }}
                                            </li>
                                            <li
                                                v-else-if="!shareResults.length"
                                                class="py-2 text-center text-xs text-gray-400"
                                            >
                                                {{ t("common.label.no_users") }}
                                            </li>
                                        </ul>
                                    </div>
                                </div>
                                <div class="mt-5 sm:mt-6 sm:flex sm:flex-row-reverse">
                                    <button
                                        type="button"
                                        class="inline-flex w-full justify-center rounded-md bg-indigo-600 px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-indigo-500 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600 sm:ml-3 sm:w-auto"
                                        @click="saveShareDialog"
                                    >
                                        {{ t("common.button.save") }}
                                    </button>
                                    <button
                                        type="button"
                                        class="mt-3 inline-flex w-full justify-center rounded-md bg-white px-3 py-2 text-sm font-semibold text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 hover:bg-gray-50 sm:mt-0 sm:w-auto"
                                        @click="closeShareDialog"
                                    >
                                        {{ t("common.button.cancel") }}
                                    </button>
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
import { t } from "@/i18n/index.js";
import { computed, watch, ref, onMounted, onBeforeUnmount, nextTick } from "vue";
import { useRouter, useRoute } from "vue-router";
import { useWorkspaceStore } from "@/store/workspaces";
import UserAvatar from "@/components/UserAvatar.vue";
import { useUserStore } from "@/store/user";
import { useWorkspaceRolesStore } from "@/store/workspaceRoles";
import ViewActions from "@/components/Projects/Menu/ViewActions.vue";
import ViewActionsMenu from "./Menu/ViewActionsMenu.vue";
import workspaceService from "@/services/workspaceService";
import FolderNavigation from "@/components/Projects/Navigation/FolderNavigation.vue";
import TaskSort from "@/components/Projects/Filters/TaskSort.vue";
import { ChevronUpDownIcon, CheckIcon } from "@heroicons/vue/20/solid";
import { useAlertStore } from "@/store/alerts";
import { extractErrorMessage } from "@/utils/errors";
import { workspaceTables } from "@/utils/projects/tree";
import FilterMenu from "./Menu/FilterMenu.vue";
import FormDialog from "@/components/FormDialog.vue";
import BaseButton from "@/components/BaseButton.vue";
import BaseSelect from "@/components/BaseSelect.vue";
import { Dialog, DialogPanel, DialogTitle, TransitionChild, TransitionRoot } from "@headlessui/vue";
import {
    Menu,
    MenuButton,
    MenuItem,
    MenuItems,
    Popover,
    PopoverButton,
    PopoverPanel,
} from "@headlessui/vue";
import {
    PlusIcon,
    AdjustmentsHorizontalIcon,
    EyeSlashIcon,
    ChartBarIcon,
    ShareIcon,
    EllipsisHorizontalIcon,
    TableCellsIcon,
    ViewColumnsIcon,
    CalendarDaysIcon,
} from "@heroicons/vue/24/outline";

const gridViewDialog = ref(false);
const newGridViewName = ref("");
const newGridViewIsPublic = ref(false);
const prevSelected = ref("");

const kanbanViewDialog = ref(false);
const newKanbanViewName = ref("");
const newKanbanGroupField = ref(null);
const newKanbanViewIsPublic = ref(false);
const kanbanHeaders = ref([]);
const creatingView = ref(false);

const openDropdownType = ref(null);

const ALL_VIEW_TYPES = [
    { type: "grid", label: "Grid", icon: TableCellsIcon, color: "#1d4ed8", bg: "#dbeafe" },
    { type: "kanban", label: "Kanban", icon: ViewColumnsIcon, color: "#7c3aed", bg: "#ede9fe" },
    { type: "gantt", label: "Gantt", icon: ChartBarIcon, color: "#d97706", bg: "#fef3c7" },
    {
        type: "calendar",
        label: "Calendar",
        icon: CalendarDaysIcon,
        color: "#be185d",
        bg: "#fce7f3",
    },
];

const CREATABLE_VIEW_TYPES = ALL_VIEW_TYPES.filter((vt) => ["grid", "kanban"].includes(vt.type));

const router = useRouter();
const route = useRoute();
const workspaceStore = useWorkspaceStore();
const userStore = useUserStore();
const rolesStore = useWorkspaceRolesStore();

const canCreateTask = computed(
    () =>
        ["grid-view", "kanban-view", "calendar-view", "gantt-view"].includes(route.name) &&
        rolesStore.hasPermissionToCreateTask,
);
const canManageViews = computed(
    () =>
        rolesStore.hasPermissionToUpdateWorkspaceView || rolesStore.hasPermissionToDeleteTableView,
);
const canCreateViews = computed(() => rolesStore.hasPermissionToCreateWorkspaceView);
const isListboxOpen = ref(false);
const listboxWrapper = ref(null);
const showFilterPopover = ref(false);
const selectedFilters = ref([]);

const sortActive = computed({
    get() {
        return workspaceStore.getSortActive;
    },
    set(value) {
        workspaceStore.setSortActive(value);
    },
});

// All view types always visible; clicking a tab with 0 views triggers creation
const ALWAYS_VISIBLE_TYPES = ["grid"];

const viewTypeGroups = computed(() => {
    const result = [];

    for (const vt of ALL_VIEW_TYPES) {
        const hasPermission = rolesStore.hasViewPermission(vt.type);
        let views = typesCreated.value.filter((v) => v.view_type === vt.type);

        if (!hasPermission) {
            if (vt.type === "grid") {
                // Always show the default grid view even without explicit grid permission
                views = views.filter((v) => v.is_default);
                if (views.length === 0) continue;
            } else {
                continue;
            }
        }

        const explicitlyPermitted = rolesStore.userPermissions.includes("views_" + vt.type);
        const showAsEmpty =
            ALWAYS_VISIBLE_TYPES.includes(vt.type) ||
            (["gantt", "calendar"].includes(vt.type) && explicitlyPermitted);

        if (showAsEmpty || views.length > 0) {
            result.push({ ...vt, views });
        }
    }

    return result;
});

// Flat list: one entry per individual view, or one creation-entry per type with no views.
const flatViewTabs = computed(() => {
    const result = [];

    for (const group of viewTypeGroups.value) {
        if (group.views.length === 0) {
            result.push({ kind: "create", group });
        } else {
            for (const view of group.views) {
                result.push({ kind: "view", group, view });
            }
        }
    }

    return result;
});

// Overflow, after flatViewTabs, no v-for refs
const overflowOpen = ref(false);
const ovNewViewOpen = ref(false);
const tabsEl = ref(null);
const hiddenIdx = ref(new Set());
const overflowItems = computed(() => flatViewTabs.value.filter((_, i) => hiddenIdx.value.has(i)));
let tabRO = null;

function checkHidden() {
    const el = tabsEl.value;

    if (!el) return;
    const cRight = el.getBoundingClientRect().right;
    const next = new Set();

    el.querySelectorAll("[data-tab-idx]").forEach((n) => {
        if (n.getBoundingClientRect().right > cRight + 2) next.add(Number(n.dataset.tabIdx));
    });
    hiddenIdx.value = next;
}

watch(tabsEl, (el) => {
    if (tabRO) tabRO.disconnect();
    if (!el) return;
    nextTick(() => {
        tabRO = new ResizeObserver(checkHidden);
        tabRO.observe(el);
        checkHidden();
    });
});

// watch(flatViewTabs) moved to onMounted to avoid TDZ with typesCreated

function pickOverflow(item) {
    if (item.kind === "view") changeSelected(item.view, false);
    else createView(item.group.type);
    overflowOpen.value = false;
    ovNewViewOpen.value = false;
}

function isTypeActive(type) {
    switch (type) {
        case "grid":
            return route.path.includes("/grid/");
        case "kanban":
            return route.path.includes("/kanban/");
        case "gantt":
            return route.name === "gantt-view";
        case "calendar":
            return route.name === "calendar-view";
        default:
            return false;
    }
}

function isViewActive(view) {
    return view.id != null && String(view.id) === String(route.params.fid);
}

async function createView(type) {
    const { id: routeId, tid: routeTid } = route.params;

    openDropdownType.value = null;
    switch (type) {
        case "grid":
            gridViewDialog.value = true;
            break;
        case "kanban": {
            try {
                const res = await workspaceStore.fetchWorkspaceNav(routeId);
                const tbl = workspaceTables(res?.data).find((t) => t.id === routeTid);

                kanbanHeaders.value = tbl?.headers?.filter((h) => h.single_select === true) ?? [];
            } catch {
                kanbanHeaders.value = [];
            }

            newKanbanViewName.value = "";
            newKanbanGroupField.value = kanbanHeaders.value[0] ?? null;
            kanbanViewDialog.value = true;
            break;
        }

        case "gantt":
            // GanttView only reads id+tid. fid is not used by the component
            await router.push({
                name: "gantt-view",
                params: { id: routeId, tid: routeTid, fid: "gantt" },
            });
            break;
        case "calendar":
            await router.push({
                name: "calendar-view",
                params: { id: routeId, tid: routeTid, fid: "calendar" },
            });
            break;
    }
}

const showFilters = computed(
    () =>
        route.path.includes("/grid/") ||
        route.name === "gantt-view" ||
        route.name === "calendar-view",
);

const closeCreatViewDialog = () => {
    gridViewDialog.value = false;
    newGridViewName.value = "";
    newGridViewIsPublic.value = false;
};

const closeKanbanDialog = () => {
    kanbanViewDialog.value = false;
    newKanbanViewName.value = "";
    newKanbanGroupField.value = null;
    newKanbanViewIsPublic.value = false;
};

function saveNewView(view, closeDialog, openView) {
    if (!view.name) {
        useAlertStore().showError(
            t.value("projects.top_navigation.please_enter_a_valid_view_name"),
        );

        return;
    }

    creatingView.value = true;

    workspaceService
        .createNewView({
            workspace_id: route.params.id,
            table_id: route.params.tid,
            ...view,
        })
        .then((response) => {
            const newView = response.data;

            typesCreated.value.push(newView);

            closeDialog();
            openView(newView);
        })
        .catch((error) => {
            useAlertStore().showError(
                extractErrorMessage(
                    error,
                    t.value("projects.top_navigation.failed_to_create_view"),
                ),
            );
        })
        .finally(() => {
            creatingView.value = false;
        });
}

function handleCreateKanbanView() {
    saveNewView(
        {
            name: newKanbanViewName.value.trim(),
            view_type: "kanban",
            is_public: newKanbanViewIsPublic.value,
            parent_table_id: newKanbanGroupField.value?.linked_id ?? null,
            group_field: newKanbanGroupField.value?.name ?? "",
        },
        closeKanbanDialog,
        (newView) => {
            if (rolesStore.userPermissions.includes("views_kanban")) {
                router.push({
                    name: "kanban-view",
                    params: {
                        id: newView.workspace_id,
                        tid: newView.table_id,
                        fid: newView.id,
                    },
                });

                return;
            }

            let defaultGridId = typesCreated.value.find((v) => v.is_default)?.id;

            if (!defaultGridId && workspaceDetails.value) {
                defaultGridId = workspaceTables(workspaceDetails.value).find(
                    (t) => t.id === route.params.tid,
                )?.options?.[0]?.id;
            }

            if (defaultGridId) {
                router.push({
                    name: "grid-view",
                    params: { id: route.params.id, tid: route.params.tid, fid: defaultGridId },
                });
            }
        },
    );
}

function handleCreateGridView() {
    saveNewView(
        {
            name: newGridViewName.value.trim(),
            view_type: "grid",
            is_public: newGridViewIsPublic.value,
        },
        closeCreatViewDialog,
        (newView) => {
            if (rolesStore.hasViewPermission("grid")) {
                router.push({
                    name: "grid-view",
                    params: { id: route.params.id, tid: route.params.tid, fid: newView.id },
                });
            }
        },
    );
}

const shareDialogOpen = ref(false);
const selectedShareUsersModel = ref([]);
const shareQuery = ref("");
const shareIsOpen = ref(false);
const shareResults = ref([]);
const shareLoading = ref(false);
const shareOffset = ref(0);
const shareHasMore = ref(true);
const shareListEl = ref(null);
const shareSentinel = ref(null);
const shareSearchInput = ref(null);
const SHARE_PAGE = 20;
let shareDebounce = null;
let shareObserver = null;

function openShareList() {
    if (shareIsOpen.value) return;
    shareIsOpen.value = true;
    resetShareList();
    nextTick(() => shareSearchInput.value?.focus());
}

function closeShareList() {
    shareIsOpen.value = false;
    shareQuery.value = "";
    shareResults.value = [];
}

function toggleShareList() {
    shareIsOpen.value ? closeShareList() : openShareList();
}

function onShareQueryInput() {
    clearTimeout(shareDebounce);
    shareDebounce = setTimeout(resetShareList, 250);
}

function resetShareList() {
    shareResults.value = [];
    shareOffset.value = 0;
    shareHasMore.value = true;
    fetchSharePage();
}

async function fetchSharePage() {
    if (shareLoading.value || !shareHasMore.value || !route.params.id) return;
    shareLoading.value = true;
    try {
        const res = await workspaceService.searchWorkspaceMembers(
            route.params.id,
            shareQuery.value,
            SHARE_PAGE,
            shareOffset.value,
        );
        const batch = res.data ?? [];

        userStore.addUsers(batch);
        shareResults.value.push(...batch);
        shareOffset.value += batch.length;
        shareHasMore.value = batch.length === SHARE_PAGE;
    } catch {
        shareHasMore.value = false;
    } finally {
        shareLoading.value = false;
    }
}

watch(shareResults, async () => {
    await nextTick();
    if (!shareSentinel.value || !shareListEl.value) return;
    if (shareObserver) shareObserver.disconnect();
    shareObserver = new IntersectionObserver(
        (e) => {
            if (e[0].isIntersecting) fetchSharePage();
        },
        { root: shareListEl.value, threshold: 0.1 },
    );
    shareObserver.observe(shareSentinel.value);
});
function toggleShareUser(user) {
    const idx = selectedShareUsersModel.value.findIndex((u) => u.id === user.id);

    if (idx >= 0) selectedShareUsersModel.value.splice(idx, 1);
    else selectedShareUsersModel.value.push(user);
}

function isShareSelected(user) {
    return selectedShareUsersModel.value.some((u) => u.id === user.id);
}

function closeShareDialog() {
    shareDialogOpen.value = false;
    selectedShareUsersModel.value = [];
    closeShareList();
    if (shareObserver) shareObserver.disconnect();
}

async function openShareDialog() {
    shareDialogOpen.value = true;
    shareIsOpen.value = false;
    shareQuery.value = "";
    shareResults.value = [];
    try {
        const sharesRes = await workspaceService.getViewShares({
            workspace_id: route.params.id,
            table_id: route.params.tid,
            view_id: route.params.fid,
        });
        const sharedIds = sharesRes.data?.user_ids || [];

        if (sharedIds.length) {
            await userStore.ensureUsers(sharedIds);
            selectedShareUsersModel.value = sharedIds
                .map((id) => userStore.usersMap[id])
                .filter(Boolean);
        }
    } catch (error) {
        useAlertStore().showError(extractErrorMessage(error));
    }
}

async function saveShareDialog() {
    try {
        await workspaceService.setViewShares({
            workspace_id: route.params.id,
            table_id: route.params.tid,
            view_id: route.params.fid,
            user_ids: selectedShareUsersModel.value.map((u) => u.id),
        });
        useAlertStore().showSuccess(t.value("projects.view.shared_successfully"));
        closeShareDialog();
    } catch (error) {
        useAlertStore().showError(extractErrorMessage(error));
    }
}

function handleApplyFilters(filters) {
    selectedFilters.value = filters;
    showFilterPopover.value = false;

    const groups = filters.groups ?? [];
    const flatFilters = filters.flatFilters ?? [];

    // Update store so GridView's fetchPage picks up new filters
    workspaceStore.setSavedGroups(groups);
    workspaceStore.setSavedFlatFilters(flatFilters);
    workspaceStore.setDefaultGroupFilters(groups);
    workspaceStore.setDefaultFlatFilters(flatFilters);
    workspaceStore.setFilterActive(groups.length > 0 || flatFilters.length > 0);

    // Bump token. GridView watcher calls fetchPage(1) which handles totalRows + currentPage
    workspaceStore.bumpGridReloadToken();
}

const selected = computed({
    get() {
        return typeof workspaceStore.getSelectedView === "string"
            ? workspaceStore.getSelectedView
            : (workspaceStore.getSelectedView?.name ?? "Unknown");
    },
    set(value) {
        workspaceStore.setSelectedView(value);
    },
});

const typesCreated = computed({
    get() {
        return workspaceStore.getViewTypes;
    },
    set(value) {
        workspaceStore.setViewTypes(value);
    },
});

function changeSelected(view, newView) {
    prevSelected.value = selected.value;
    selected.value = view.name;

    setTimeout(() => {
        navigateToView(view, newView);
    }, 50);

    isListboxOpen.value = false;
}

const viewActions = ref(null);

function handleClickOutside(event) {
    if (listboxWrapper.value && !listboxWrapper.value.contains(event.target)) {
        isListboxOpen.value = false;
        openDropdownType.value = null;
        overflowOpen.value = false;
        ovNewViewOpen.value = false;
    }
}

onMounted(() => {
    document.addEventListener("click", handleClickOutside);
    watch(flatViewTabs, () => nextTick(checkHidden));
});

onBeforeUnmount(() => {
    document.removeEventListener("click", handleClickOutside);
    if (tabRO) {
        tabRO.disconnect();
        tabRO = null;
    }
});

const VIEW_ROUTES = {
    grid: "grid-view",
    kanban: "kanban-view",
    gantt: "gantt-view",
    calendar: "calendar-view",
};

async function navigateToView(view, newView) {
    const { id, tid } = route.params;

    if (view.id === undefined) {
        if (view.name === "Grid View" && newView) {
            gridViewDialog.value = true;
        } else if (view.name === "Grid View") {
            await router.push({ name: "grid-view", params: { id, tid } });
        }

        return;
    }

    const name = VIEW_ROUTES[view.view_type];

    if (name) {
        if (view.view_type === "grid") {
            selectedView.value = view;
            selected.value = view.name;
        }

        await router.push({ name, params: { id, tid, fid: view.id } });
    }

    sortActive.value = false;
}

const selectedView = computed({
    get() {
        return workspaceStore.getSelectedViewData;
    },
    set(value) {
        workspaceStore.setSelectedViewData(value);
    },
});

const rightSideNavigation = computed({
    get() {
        return workspaceStore.getRightNavigation;
    },
    set(value) {
        workspaceStore.setRightNavigation(value);
    },
});

const workspaceName = ref("");
const tableName = ref("");

async function fetchWorkspaceData() {
    const { id: workspaceId, tid: tableId } = route.params;

    if (!workspaceId || !tableId) {
        return { workspace: null, workspaceName: "", tableName: "" };
    }

    try {
        const workspace = (await workspaceStore.fetchWorkspaceNav(workspaceId)).data;

        if (!workspace || typeof workspace !== "object") {
            throw new TypeError("Invalid workspace details received.");
        }

        workspaceDetails.value = workspace;

        const table = workspaceTables(workspace).find((tbl) => tbl.id === tableId);

        return {
            workspace,
            workspaceName: workspace.title,
            tableName: table
                ? table.display_name?.String?.trim() || table.name
                : t.value("projects.full_task_navigation.unknown_table"),
        };
    } catch (error) {
        useAlertStore().showError(
            error?.response?.data?.message ||
                error?.message ||
                t.value("projects.top_navigation.unexpected_error_while_loading_workspace"),
        );

        return { workspace: null, workspaceName: "Error", tableName: "Error" };
    }
}

const workspaceDetails = computed({
    get() {
        return workspaceStore.getWorkspaceDetails;
    },
    set(value) {
        workspaceStore.setWorkspaceDetails(value);
    },
});

const currentView = computed(() => {
    if (!route.params.fid || !workspaceDetails.value?.tables) return null;
    for (const table of workspaceDetails.value.tables) {
        const found = table.views?.find((v) => v.id === route.params.fid);

        if (found) return found;
    }

    return null;
});

async function addToView(workspace, tableID, viewID) {
    try {
        if (!workspace || typeof workspace !== "object") return;

        let table = workspace.tables?.find((tbl) => tbl.id === tableID);

        if (!table && Array.isArray(workspace.folders)) {
            for (const folder of workspace.folders) {
                if (Array.isArray(folder.tables)) {
                    const found = folder.tables.find((tbl) => tbl.id === tableID);

                    if (found) {
                        table = found;
                        break;
                    }
                }
            }
        }

        if (!table) {
            typesCreated.value = [];

            return;
        }

        const gridView =
            Array.isArray(table.options) && table.options.length > 0
                ? {
                      ...table.options[0],
                      name: "Grid View",
                      view_type: table.options[0].view_type || "grid",
                      is_default: true,
                  }
                : null;

        typesCreated.value = gridView ? [gridView] : [];

        if (Array.isArray(table.views) && table.views.length > 0) {
            const uniqueViews = table.views.filter(
                (view) => !typesCreated.value.some((existingView) => existingView.id === view.id),
            );

            if (uniqueViews.length > 0) {
                typesCreated.value = [...typesCreated.value, ...uniqueViews];
            }

            if (viewID) {
                const selectedView = typesCreated.value.find((view) => view.id === viewID);

                if (selectedView) {
                    selected.value = selectedView.name;
                }
            }
        }
    } catch (error) {
        useAlertStore().showError(extractErrorMessage(error));
    }
}

watch(
    () => [route.params.id, route.params.tid, route.params.fid, route.name],
    () => {
        (async () => {
            const path = route.path;
            const { tid: tableID, fid: viewID } = route.params;

            const data = await fetchWorkspaceData();

            workspaceName.value = data.workspaceName;
            tableName.value = data.tableName;

            await addToView(data.workspace, tableID, viewID);

            if (path.includes("/grid/")) {
                const gridView = typesCreated.value.find(
                    (view) => view.view_type === "grid" && view.id === viewID,
                );

                selected.value = gridView ? gridView.name : "Grid View";
            } else if (path.includes("/kanban/")) {
                const kanbanView = typesCreated.value.find(
                    (view) => view.view_type === "kanban" && view.id === viewID,
                );

                selected.value = kanbanView ? kanbanView.name : "Kanban View";
            }
        })();
    },
    { immediate: true },
);
</script>

<style scoped>
.divider {
    height: 1px;
    background: #e5e7eb;
    margin: 8px 0;
}
.truncate {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    max-width: 350px;
}
.tn-row1 {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 4px 8px 3px;
    flex-shrink: 0;
    width: 100%;
    container-type: inline-size;
}
@container (max-width: 680px) {
    .tn-btn-label {
        display: none;
    }
}
.tn-avatar {
    display: inline-flex;
    height: 32px;
    width: 32px;
    align-items: center;
    justify-content: center;
    border-radius: 8px;
    background-color: #3b82f6;
    color: #fff;
    font-size: 14px;
    font-weight: 700;
    flex-shrink: 0;
}
.tn-table-name {
    font-size: 18px;
    font-weight: 600;
    color: #111827;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    min-width: 0;
    flex: 1;
    line-height: 1.2;
}
.tn-row2 {
    display: flex;
    align-items: center;
    padding: 4px 6px;
    height: 44px;
    flex-shrink: 0;
    border-top: 1px solid #f0f0f0;
    background: #f9fafb;
    gap: 6px;
}
.tn-tabs {
    display: flex;
    align-items: center;
    gap: 6px;
    flex: 0 1 auto;
    min-width: 0;
    overflow: hidden;
}
.tn-ov-wrap {
    position: relative;
    flex-shrink: 0;
}
.tn-ov-btn {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    padding: 8px;
    background: #fff;
    border: none;
    border-radius: 4px;
    box-shadow:
        inset 0 0 0 1px #d1d5db,
        0 1px 2px 0 rgb(0 0 0 / 0.05);
    cursor: pointer;
    color: #6b7280;
}
.tn-ov-btn:hover {
    background: #f9fafb;
    color: #111827;
}
.tn-ov-menu {
    right: 0;
    left: auto;
    min-width: 160px;
}
.tn-ov-item {
    width: 100%;
    max-width: unset;
    justify-content: flex-start;
    box-shadow: none;
}
.tn-ov-item.tn-tab--active {
    box-shadow: none;
}
.tn-tab-wrap {
    position: relative;
    flex-shrink: 0;
}
/* Sized and coloured like a small secondary Button; the view type's colour
   is on its icon, and on the ring of the tab that is open. */
.tn-tab {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    padding: 6px 10px;
    font-size: 14px;
    line-height: 20px;
    font-weight: 500;
    color: #374151;
    background: #fff;
    border: none;
    border-radius: 4px;
    box-shadow:
        inset 0 0 0 1px #d1d5db,
        0 1px 2px 0 rgb(0 0 0 / 0.05);
    cursor: pointer;
    white-space: nowrap;
}
.tn-tab > svg:first-child {
    width: 16px;
    height: 16px;
    color: var(--tab-color, #6b7280);
}
.tn-tab:hover {
    background: #f9fafb;
    color: #111827;
}
.tn-tab--active {
    color: #111827;
    background: var(--tab-bg, #eef2ff);
    box-shadow:
        inset 0 0 0 1px var(--tab-color, #4f46e5),
        0 1px 2px 0 rgb(0 0 0 / 0.05);
}
.tn-tab--active:hover {
    background: var(--tab-bg, #eef2ff);
}
.tn-dropdown {
    position: absolute;
    top: calc(100% + 4px);
    left: 0;
    z-index: 200;
    min-width: 180px;
    padding: 4px;
}
.tn-actions {
    display: flex;
    align-items: center;
    gap: 4px;
    flex-shrink: 0;
    margin-left: auto;
}
</style>
