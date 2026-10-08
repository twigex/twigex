<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div class="flex h-full flex-col">
        <div v-if="!details" class="flex flex-1 items-center justify-center">
            <BaseSpinner />
        </div>

        <div v-else class="flex-1 overflow-y-auto">
            <div class="mx-auto max-w-3xl space-y-6 px-6 py-6">
                <div>
                    <RouterLink
                        :to="{ name: 'projects-settings-list' }"
                        class="inline-flex items-center gap-x-1 text-sm font-medium text-gray-500 hover:text-gray-700"
                    >
                        <ChevronLeftIcon class="size-5" aria-hidden="true" />
                        {{ t("common.button.back") }}
                    </RouterLink>

                    <div class="mt-3 flex flex-wrap items-start justify-between gap-4">
                        <div class="flex min-w-0 items-center gap-x-4">
                            <LetterAvatar
                                :id="details.workspace.id"
                                :name="details.workspace.title"
                                class="size-12 shrink-0 rounded-lg text-base"
                            />
                            <div class="min-w-0">
                                <h2 class="truncate text-xl font-semibold text-gray-900">
                                    {{ details.workspace.title }}
                                </h2>
                                <div
                                    class="mt-1 flex flex-wrap items-center gap-x-6 gap-y-1 text-sm text-gray-500"
                                >
                                    <span class="flex items-center gap-x-1.5">
                                        <UsersIcon
                                            class="size-4 text-gray-400"
                                            aria-hidden="true"
                                        />
                                        {{ details.members.length }}
                                    </span>
                                    <span class="flex items-center gap-x-1.5">
                                        <CalendarIcon
                                            class="size-4 text-gray-400"
                                            aria-hidden="true"
                                        />
                                        {{ getDate(details.workspace.created_at) }}
                                    </span>
                                </div>
                            </div>
                        </div>
                        <BaseButton
                            v-if="!isWorkspaceAdmin"
                            variant="secondary"
                            :prepend-icon="ShieldCheckIcon"
                            @click="joinWorkspace"
                        >
                            {{ t("settings.projects.join") }}
                        </BaseButton>
                    </div>
                </div>

                <SectionCard :title="t('settings.projects.general')">
                    <div class="space-y-4 p-4">
                        <div>
                            <label
                                for="project-name"
                                class="block text-sm font-medium text-gray-900"
                            >
                                {{ t("settings.projects.name") }}
                            </label>
                            <input
                                id="project-name"
                                v-model="name"
                                type="text"
                                class="mt-2 block w-full rounded-md border-0 px-3 py-1.5 text-sm text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600"
                            />
                        </div>
                        <div>
                            <label
                                for="project-description"
                                class="block text-sm font-medium text-gray-900"
                            >
                                {{ t("settings.projects.description") }}
                            </label>
                            <textarea
                                id="project-description"
                                v-model="description"
                                rows="3"
                                class="mt-2 block w-full rounded-md border-0 px-3 py-1.5 text-sm text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600"
                            />
                        </div>
                    </div>
                </SectionCard>

                <section class="space-y-3">
                    <MembersTable
                        :members="details.members"
                        :columns="memberColumns"
                        :itemsPerPage="MEMBERS_PER_PAGE"
                    >
                        <template #toolbar>
                            <div class="flex items-center justify-between gap-3">
                                <h3 class="text-sm font-semibold text-gray-900">
                                    {{ t("settings.projects.members") }}
                                </h3>
                                <BaseButton
                                    size="small"
                                    :prepend-icon="PlusIcon"
                                    @click="memberDialog = true"
                                >
                                    {{ t("settings.projects.add_members") }}
                                </BaseButton>
                            </div>
                        </template>

                        <template #role="{ item }">
                            <div class="flex flex-wrap gap-1">
                                <span
                                    v-for="role in splitRoles(item.role)"
                                    :key="role"
                                    class="inline-flex items-center rounded-md bg-gray-50 px-1.5 py-0.5 text-xs font-medium text-gray-600 ring-1 ring-inset ring-gray-500/10"
                                >
                                    {{ roleLabel(role) }}
                                </span>
                            </div>
                        </template>

                        <template #actions="{ item }">
                            <button
                                v-if="rolesLicensed"
                                type="button"
                                class="text-indigo-600 hover:text-indigo-500"
                                @click="editMemberRoles(item)"
                            >
                                {{ t("common.button.edit") }}
                            </button>
                            <button
                                type="button"
                                class="text-red-600 hover:text-red-500"
                                @click="askRemoveMember(item)"
                            >
                                {{ t("common.button.remove") }}
                            </button>
                        </template>
                    </MembersTable>
                </section>

                <SectionCard
                    v-if="hasGroups || details.groups.length > 0"
                    :title="t('settings.projects.groups')"
                >
                    <template #action>
                        <BaseButton
                            v-if="hasGroups"
                            size="small"
                            variant="secondary"
                            :prepend-icon="PlusIcon"
                            @click="groupDialog = true"
                        >
                            {{ t("settings.projects.add_groups") }}
                        </BaseButton>
                    </template>
                    <p
                        v-if="details.groups.length === 0"
                        class="px-4 py-6 text-center text-sm text-gray-500"
                    >
                        {{ t("settings.projects.no_groups") }}
                    </p>
                    <ul v-else role="list" class="divide-y divide-gray-100">
                        <li
                            v-for="group in details.groups"
                            :key="group.group_id"
                            class="flex items-center justify-between gap-x-6 px-4 py-3"
                        >
                            <button
                                type="button"
                                class="flex min-w-0 items-center gap-x-3 text-left"
                                @click="openGroupMembers(group)"
                            >
                                <div
                                    class="flex size-9 shrink-0 items-center justify-center rounded-full bg-gray-100"
                                >
                                    <UserGroupIcon
                                        class="size-5 text-gray-500"
                                        aria-hidden="true"
                                    />
                                </div>
                                <div class="min-w-0">
                                    <p class="truncate text-sm font-medium text-gray-900">
                                        {{ group.name }}
                                    </p>
                                    <p class="truncate text-xs text-gray-500">
                                        {{
                                            t("collimato.users.groups_table.member_count", {
                                                count: group.member_count ?? 0,
                                            })
                                        }}
                                    </p>
                                </div>
                            </button>
                            <div class="flex shrink-0 items-center gap-x-4">
                                <div class="hidden flex-wrap justify-end gap-1 sm:flex">
                                    <span
                                        v-for="role in group.roles || []"
                                        :key="role"
                                        class="inline-flex items-center rounded-md bg-gray-50 px-1.5 py-0.5 text-xs font-medium text-gray-600 ring-1 ring-inset ring-gray-500/10"
                                    >
                                        {{ roleLabel(role) }}
                                    </span>
                                </div>
                                <button
                                    v-if="rolesLicensed"
                                    type="button"
                                    class="text-sm font-medium text-indigo-600 hover:text-indigo-500"
                                    @click="editGroupRoles(group)"
                                >
                                    {{ t("common.button.edit") }}
                                </button>
                                <button
                                    type="button"
                                    class="text-sm font-medium text-red-600 hover:text-red-500"
                                    @click="askRemoveGroup(group)"
                                >
                                    {{ t("common.button.remove") }}
                                </button>
                            </div>
                        </li>
                    </ul>
                </SectionCard>

                <div
                    class="flex flex-wrap items-center justify-between gap-4 rounded-md border border-red-200 bg-red-50/50 px-4 py-4"
                >
                    <div>
                        <h3 class="text-sm font-semibold text-gray-900">
                            {{ t("settings.projects.delete.title") }}
                        </h3>
                        <p class="mt-0.5 text-xs text-gray-500">
                            {{ t("settings.projects.delete.description") }}
                        </p>
                    </div>
                    <BaseButton
                        color="bg-red-600 text-white hover:bg-red-500"
                        @click="deleteDialog = true"
                    >
                        {{ t("common.button.delete") }}
                    </BaseButton>
                </div>
            </div>
        </div>

        <div
            v-if="details"
            class="flex shrink-0 items-center justify-end gap-x-3 border-t bg-gray-50 px-6 py-4"
        >
            <button
                type="button"
                class="rounded-md px-3 py-2 text-sm font-semibold text-gray-700 hover:text-gray-900"
                @click="router.push({ name: 'projects-settings-list' })"
            >
                {{ t("common.button.cancel") }}
            </button>
            <BaseButton
                :is-disabled="!detailsChanged || !name.trim()"
                :is-loading="saving"
                @click="saveDetails"
            >
                {{ t("common.button.save") }}
            </BaseButton>
        </div>

        <UserPickerDialog
            v-model="memberDialog"
            :title="t('settings.projects.add_members')"
            :exclude-ids="details ? details.members.map((m) => m.user_id) : []"
            @confirm="addMembers"
        />
        <RoleDialog
            v-model="memberRoleDialog"
            :user="{ role: splitRoles(memberToEdit?.role).join(',') }"
            :roles="details?.roles ?? []"
            scope="projects"
            @add="updateMemberRoles"
        />
        <RoleDialog
            v-model="groupRoleDialog"
            :user="{ role: (groupToEdit?.roles || []).join(',') }"
            :roles="details?.roles ?? []"
            scope="projects"
            @add="updateGroupRoles"
        />
        <AddGroupDialog
            v-model="groupDialog"
            :roles="details?.roles ?? []"
            scope="projects"
            @add="addGroups"
        />
        <GroupMembersDialog
            v-model="groupMembersOpen"
            :group-id="groupMembersGroup?.group_id ?? ''"
            :group-name="groupMembersGroup?.name ?? ''"
        />
        <ConfirmDialog
            :open="removeMemberDialog"
            :title="t('collimato.users.remove_user_dialog.title')"
            :message="t('collimato.users.remove_user_dialog.confirm')"
            :confirm-label="t('common.button.remove')"
            @confirm="removeMember"
            @close="removeMemberDialog = false"
        />
        <ConfirmDialog
            :open="removeGroupDialog"
            :title="t('collimato.users.remove_group_dialog.title')"
            :message="t('collimato.users.remove_group_dialog.confirm')"
            :confirm-label="t('common.button.remove')"
            @confirm="removeGroup"
            @close="removeGroupDialog = false"
        />
        <ConfirmDialog
            :open="deleteDialog"
            :title="
                t('settings.projects.delete_dialog.title', { name: details?.workspace.title ?? '' })
            "
            :message="t('settings.projects.delete_dialog.message')"
            :confirm-label="t('common.button.delete')"
            @confirm="deleteWorkspace"
            @close="deleteDialog = false"
        />
    </div>
</template>

<script setup>
import { t } from "@/i18n/index";

import { computed, onMounted, ref } from "vue";
import { RouterLink, useRoute, useRouter } from "vue-router";
import workspaceService from "@/services/workspaceService";
import useDateOperations from "@/composables/useDateOperations.js";
import { useAlertStore } from "@/store/alerts";
import { useSettingsStore } from "@/store/settings";
import { useUserStore } from "@/store/user";
import { extractErrorMessage } from "@/utils/errors";
import BaseButton from "@/components/BaseButton.vue";
import BaseSpinner from "@/components/BaseSpinner.vue";
import ConfirmDialog from "@/components/ConfirmDialog.vue";
import LetterAvatar from "@/components/LetterAvatar.vue";
import MembersTable from "@/components/MembersTable.vue";
import SectionCard from "@/components/SectionCard.vue";
import UserPickerDialog from "@/components/UserPickerDialog.vue";
import GroupMembersDialog from "@/components/GroupMembersDialog.vue";
import AddGroupDialog from "@/components/Collimato/Dialogs/AddGroupDialog.vue";
import RoleDialog from "@/components/Collimato/Dialogs/RoleDialog.vue";
import { roleLabel as labelForRole, roleNameLabel } from "@/utils/roleLabels";
import { UserGroupIcon } from "@heroicons/vue/24/outline";
import {
    CalendarIcon,
    ChevronLeftIcon,
    PlusIcon,
    ShieldCheckIcon,
    UsersIcon,
} from "@heroicons/vue/20/solid";

const ADMIN_ROLE = "admin";
const MEMBERS_PER_PAGE = 20;

const route = useRoute();
const router = useRouter();
const alertStore = useAlertStore();
const settingsStore = useSettingsStore();
const userStore = useUserStore();
const { getDate } = useDateOperations();

const rolesLicensed = computed(() => settingsStore.getLicenseFeature("workspace_roles"));
const hasGroups = computed(() => settingsStore.getLicenseFeature("groups"));

const workspaceId = route.params.id;

const details = ref(null);
const name = ref("");
const description = ref("");
const saving = ref(false);

const memberDialog = ref(false);
const memberRoleDialog = ref(false);
const groupRoleDialog = ref(false);
const groupDialog = ref(false);
const groupMembersOpen = ref(false);
const removeMemberDialog = ref(false);
const removeGroupDialog = ref(false);
const deleteDialog = ref(false);

const memberToRemove = ref(null);
const memberToEdit = ref(null);
const groupToEdit = ref(null);
const groupMembersGroup = ref(null);

const memberColumns = computed(() => [
    {
        key: "role",
        label: t.value("members_table.role"),
        sortable: true,
        hiddenBelow: "sm",
    },
]);

const detailsChanged = computed(() => {
    if (!details.value) return false;

    return (
        name.value !== details.value.workspace.title ||
        description.value !== details.value.workspace.description
    );
});

const isWorkspaceAdmin = computed(() => {
    const me = details.value?.members.find((m) => m.user_id === userStore.user?.id);

    return splitRoles(me?.role).includes(ADMIN_ROLE);
});

function splitRoles(role) {
    return (role ?? "").split(" ").filter(Boolean);
}

function roleLabel(roleName) {
    const role = (details.value?.roles ?? []).find((r) => r.name === roleName);

    return role ? labelForRole(role, "projects") : roleNameLabel(roleName, "projects");
}

async function load() {
    try {
        const response = await workspaceService.getManaged(workspaceId);

        details.value = response.data;
        name.value = response.data.workspace.title;
        description.value = response.data.workspace.description;
    } catch (error) {
        alertStore.showError(extractErrorMessage(error));
        router.push({ name: "projects-settings-list" });
    }
}

async function saveDetails() {
    saving.value = true;

    try {
        await workspaceService.updateManaged(workspaceId, {
            name: name.value,
            description: description.value,
        });
        await load();

        alertStore.showSuccess(t.value("settings.projects.success.saved"));
    } catch (error) {
        alertStore.showError(extractErrorMessage(error));
    } finally {
        saving.value = false;
    }
}

async function joinWorkspace() {
    try {
        await workspaceService.joinManaged(workspaceId);
        await load();

        alertStore.showSuccess(t.value("settings.projects.success.joined"));
    } catch (error) {
        alertStore.showError(extractErrorMessage(error));
    }
}

async function addMembers(users) {
    try {
        await workspaceService.addManagedMembers(
            workspaceId,
            users.map((u) => u.id),
        );
        await load();

        alertStore.showSuccess(t.value("collimato.users.success.user_added"));
    } catch (error) {
        alertStore.showError(extractErrorMessage(error));
    }
}

function editMemberRoles(member) {
    memberToEdit.value = member;
    memberRoleDialog.value = true;
}

async function updateMemberRoles(roles) {
    try {
        await workspaceService.updateManagedMemberRole(
            workspaceId,
            memberToEdit.value.user_id,
            roles.join(" "),
        );
        await load();

        alertStore.showSuccess(t.value("collimato.users.success.user_roles_updated"));
    } catch (error) {
        alertStore.showError(extractErrorMessage(error));
    }
}

function askRemoveMember(member) {
    memberToRemove.value = member;
    removeMemberDialog.value = true;
}

async function removeMember() {
    removeMemberDialog.value = false;

    try {
        await workspaceService.removeManagedMember(workspaceId, memberToRemove.value.user_id);
        await load();

        alertStore.showSuccess(t.value("collimato.users.success.user_removed"));
    } catch (error) {
        alertStore.showError(extractErrorMessage(error));
    }
}

async function addGroups({ groupIds, roles }) {
    groupDialog.value = false;

    try {
        await workspaceService.addManagedGroups(workspaceId, {
            group_ids: groupIds,
            roles,
        });
        await load();

        alertStore.showSuccess(t.value("collimato.users.success.group_added"));
    } catch (error) {
        alertStore.showError(extractErrorMessage(error));
    }
}

function editGroupRoles(group) {
    groupToEdit.value = group;
    groupRoleDialog.value = true;
}

async function updateGroupRoles(roles) {
    try {
        await workspaceService.updateManagedGroupRoles(
            workspaceId,
            groupToEdit.value.group_id,
            roles,
        );
        await load();

        alertStore.showSuccess(t.value("collimato.users.success.group_roles_updated"));
    } catch (error) {
        alertStore.showError(extractErrorMessage(error));
    }
}

function askRemoveGroup(group) {
    groupToEdit.value = group;
    removeGroupDialog.value = true;
}

async function removeGroup() {
    removeGroupDialog.value = false;

    try {
        await workspaceService.removeManagedGroup(workspaceId, groupToEdit.value.group_id);
        await load();

        alertStore.showSuccess(t.value("collimato.users.success.group_removed"));
    } catch (error) {
        alertStore.showError(extractErrorMessage(error));
    }
}

function openGroupMembers(group) {
    groupMembersGroup.value = group;
    groupMembersOpen.value = true;
}

async function deleteWorkspace() {
    deleteDialog.value = false;

    try {
        await workspaceService.deleteManaged(workspaceId);

        alertStore.showSuccess(t.value("settings.projects.deleted"));
        router.push({ name: "projects-settings-list" });
    } catch (error) {
        alertStore.showError(extractErrorMessage(error));
    }
}

onMounted(load);
</script>
