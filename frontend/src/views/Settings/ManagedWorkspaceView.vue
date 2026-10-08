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
                        :to="{ name: 'collimato-settings-list' }"
                        class="inline-flex items-center gap-x-1 text-sm font-medium text-gray-500 hover:text-gray-700"
                    >
                        <ChevronLeftIcon class="size-5" aria-hidden="true" />
                        {{ t("common.button.back") }}
                    </RouterLink>

                    <div class="mt-3 flex flex-wrap items-start justify-between gap-4">
                        <div class="min-w-0">
                            <div class="flex items-center gap-x-3">
                                <h2 class="truncate text-xl font-semibold text-gray-900">
                                    {{ details.workspace.name }}
                                </h2>
                                <span
                                    v-if="details.workspace.status === 'draft'"
                                    class="inline-flex shrink-0 items-center rounded-md bg-yellow-50 px-2 py-0.5 text-xs font-medium text-yellow-800 ring-1 ring-inset ring-yellow-600/20"
                                >
                                    {{ t("settings.collimato_workspaces.draft") }}
                                </span>
                            </div>
                            <div
                                class="mt-2 flex flex-wrap items-center gap-x-6 gap-y-1 text-sm text-gray-500"
                            >
                                <span class="flex items-center gap-x-1.5">
                                    <UsersIcon class="size-4 text-gray-400" aria-hidden="true" />
                                    {{
                                        t("collimato.users.groups_table.member_count", {
                                            count: details.users.length,
                                        })
                                    }}
                                </span>
                                <span class="flex items-center gap-x-1.5">
                                    <CalendarIcon class="size-4 text-gray-400" aria-hidden="true" />
                                    {{ getDate(details.workspace.created_at) }}
                                </span>
                            </div>
                        </div>
                        <BaseButton
                            v-if="!isWorkspaceAdmin"
                            variant="secondary"
                            :prepend-icon="ShieldCheckIcon"
                            @click="joinWorkspace"
                        >
                            {{ t("settings.collimato_workspace.join") }}
                        </BaseButton>
                    </div>
                </div>

                <SectionCard :title="t('settings.collimato_workspace.general')">
                    <div class="space-y-4 p-4">
                        <div>
                            <label
                                for="workspace-name"
                                class="block text-sm font-medium text-gray-900"
                            >
                                {{ t("settings.collimato_workspace.name") }}
                            </label>
                            <input
                                id="workspace-name"
                                v-model="name"
                                type="text"
                                class="mt-2 block w-full rounded-md border-0 px-3 py-1.5 text-sm text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600"
                            />
                        </div>
                        <div>
                            <label
                                for="workspace-description"
                                class="block text-sm font-medium text-gray-900"
                            >
                                {{ t("settings.collimato_workspace.description") }}
                            </label>
                            <textarea
                                id="workspace-description"
                                v-model="description"
                                rows="3"
                                class="mt-2 block w-full rounded-md border-0 px-3 py-1.5 text-sm text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600"
                            />
                        </div>
                    </div>
                </SectionCard>

                <SectionCard :title="t('settings.collimato_workspace.created_by')">
                    <div class="px-4 py-4">
                        <UserAvatarWithText
                            :user-id="details.workspace.created_by"
                            text="fullNameWithEmail"
                            avatar-class="size-9 shrink-0"
                            text-class="ml-3 truncate text-sm font-medium text-gray-900"
                        />
                    </div>
                </SectionCard>

                <section class="space-y-3">
                    <MembersTable
                        :members="directMembers"
                        :columns="memberColumns"
                        :itemsPerPage="MEMBERS_PER_PAGE"
                    >
                        <template #toolbar>
                            <div class="flex items-center justify-between gap-3">
                                <h3 class="text-sm font-semibold text-gray-900">
                                    {{ t("settings.collimato_workspace.members") }}
                                </h3>
                                <BaseButton
                                    size="small"
                                    :prepend-icon="PlusIcon"
                                    @click="memberDialog = true"
                                >
                                    {{ t("settings.collimato_workspace.add_members") }}
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
                                    {{ roleNameLabel(role, "collimato") }}
                                </span>
                            </div>
                        </template>

                        <template #actions="{ item }">
                            <button
                                v-if="hasCollimatoRoles"
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
                    :title="t('settings.collimato_workspace.groups')"
                >
                    <template #action>
                        <BaseButton
                            v-if="hasGroups"
                            size="small"
                            variant="secondary"
                            :prepend-icon="PlusIcon"
                            @click="groupDialog = true"
                        >
                            {{ t("settings.collimato_workspace.add_groups") }}
                        </BaseButton>
                    </template>
                    <p
                        v-if="details.groups.length === 0"
                        class="px-4 py-6 text-center text-sm text-gray-500"
                    >
                        {{ t("settings.collimato_workspace.no_groups") }}
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
                                        {{ roleNameLabel(role, "collimato") }}
                                    </span>
                                </div>
                                <button
                                    v-if="hasGroups"
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
                            {{ t("settings.collimato_workspace.delete.title") }}
                        </h3>
                        <p class="mt-0.5 text-xs text-gray-500">
                            {{ t("settings.collimato_workspace.delete.description") }}
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
                @click="router.push({ name: 'collimato-settings-list' })"
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
            :title="t('settings.collimato_workspace.add_members')"
            :exclude-ids="directMembers.map((m) => m.user_id)"
            @confirm="addMembers"
        />
        <RoleDialog
            v-model="memberRoleDialog"
            :user="memberToEdit ?? {}"
            :roles="details?.roles ?? []"
            @add="updateMemberRoles"
        />
        <RoleDialog
            v-model="groupRoleDialog"
            :user="{ role: (groupToEdit?.roles || []).join(',') }"
            :roles="details?.roles ?? []"
            @add="updateGroupRoles"
        />
        <AddGroupDialog v-model="groupDialog" :roles="details?.roles ?? []" @add="addGroups" />
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
                t('settings.collimato_workspaces.delete_dialog.title', {
                    name: details?.workspace.name ?? '',
                })
            "
            :message="t('settings.collimato_workspaces.delete_dialog.message')"
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
import collimatoService from "@/services/collimatoService";
import useDateOperations from "@/composables/useDateOperations.js";
import { useAlertStore } from "@/store/alerts";
import { useSettingsStore } from "@/store/settings";
import { useUserStore } from "@/store/user";
import { extractErrorMessage } from "@/utils/errors";
import BaseButton from "@/components/BaseButton.vue";
import BaseSpinner from "@/components/BaseSpinner.vue";
import ConfirmDialog from "@/components/ConfirmDialog.vue";
import SectionCard from "@/components/SectionCard.vue";
import MembersTable from "@/components/MembersTable.vue";
import { roleNameLabel } from "@/utils/roleLabels";
import UserAvatarWithText from "@/components/UserAvatarWithText.vue";
import UserPickerDialog from "@/components/UserPickerDialog.vue";
import GroupMembersDialog from "@/components/GroupMembersDialog.vue";
import AddGroupDialog from "@/components/Collimato/Dialogs/AddGroupDialog.vue";
import RoleDialog from "@/components/Collimato/Dialogs/RoleDialog.vue";
import { UserGroupIcon } from "@heroicons/vue/24/outline";
import {
    CalendarIcon,
    ChevronLeftIcon,
    PlusIcon,
    ShieldCheckIcon,
    UsersIcon,
} from "@heroicons/vue/20/solid";

const route = useRoute();
const router = useRouter();
const alertStore = useAlertStore();
const settingsStore = useSettingsStore();
const userStore = useUserStore();
const { getDate } = useDateOperations();

const hasCollimatoRoles = computed(() => settingsStore.getLicenseFeature("collimato_roles"));
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

const memberToEdit = ref(null);
const groupToEdit = ref(null);
const groupMembersGroup = ref(null);

const MEMBERS_PER_PAGE = 20;

const memberColumns = computed(() => [
    {
        key: "role",
        label: t.value("members_table.roles"),
        sortable: true,
        hiddenBelow: "sm",
    },
]);

const directMembers = computed(() => (details.value?.users ?? []).filter((u) => !u.via_group));

const detailsChanged = computed(() => {
    if (!details.value) return false;

    return (
        name.value !== details.value.workspace.name ||
        description.value !== details.value.workspace.description
    );
});

const isWorkspaceAdmin = computed(() => {
    const me = directMembers.value.find((m) => m.user_id === userStore.user?.id);

    return splitRoles(me?.role).includes("workspace_admin");
});

function splitRoles(role) {
    return (role ?? "")
        .split(",")
        .map((r) => r.trim())
        .filter(Boolean);
}

async function load() {
    try {
        const response = await collimatoService.getWorkspaceDetails(workspaceId);

        details.value = response.data;
        name.value = response.data.workspace.name;
        description.value = response.data.workspace.description;
    } catch (error) {
        alertStore.showError(extractErrorMessage(error));
        router.push({ name: "collimato-settings-list" });
    }
}

async function saveDetails() {
    saving.value = true;

    try {
        await collimatoService.updateManagedWorkspace(workspaceId, {
            name: name.value,
            description: description.value,
        });
        await load();

        alertStore.showSuccess(t.value("settings.collimato_workspace.success.saved"));
    } catch (error) {
        alertStore.showError(extractErrorMessage(error));
    } finally {
        saving.value = false;
    }
}

async function joinWorkspace() {
    try {
        await collimatoService.joinManagedWorkspace(workspaceId);
        await load();

        alertStore.showSuccess(t.value("settings.collimato_workspace.success.joined"));
    } catch (error) {
        alertStore.showError(extractErrorMessage(error));
    }
}

async function addMembers(users) {
    try {
        await collimatoService.addManagedWorkspaceUsers(
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
        await collimatoService.updateManagedWorkspaceUserRoles(
            workspaceId,
            memberToEdit.value.user_id,
            roles,
        );
        await load();

        alertStore.showSuccess(t.value("collimato.users.success.user_roles_updated"));
    } catch (error) {
        alertStore.showError(extractErrorMessage(error));
    }
}

function askRemoveMember(member) {
    memberToEdit.value = member;
    removeMemberDialog.value = true;
}

async function removeMember() {
    removeMemberDialog.value = false;

    try {
        await collimatoService.removeManagedWorkspaceUser(workspaceId, memberToEdit.value.user_id);
        await load();

        alertStore.showSuccess(t.value("collimato.users.success.user_removed"));
    } catch (error) {
        alertStore.showError(extractErrorMessage(error));
    }
}

async function addGroups({ groupIds, roles }) {
    groupDialog.value = false;

    try {
        await collimatoService.addManagedWorkspaceGroups(workspaceId, {
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
        await collimatoService.updateManagedWorkspaceGroupRoles(
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
        await collimatoService.removeManagedWorkspaceGroup(workspaceId, groupToEdit.value.group_id);
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
        await collimatoService.deleteManagedWorkspace(workspaceId);

        alertStore.showSuccess(t.value("settings.collimato_workspaces.deleted"));
        router.push({ name: "collimato-settings-list" });
    } catch (error) {
        alertStore.showError(extractErrorMessage(error));
    }
}

onMounted(load);
</script>
