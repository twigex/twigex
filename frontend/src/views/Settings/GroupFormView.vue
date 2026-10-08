<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div class="flex h-full flex-col">
        <div class="flex-1 overflow-y-auto">
            <div class="mx-auto max-w-3xl space-y-6 px-6 py-6">
                <div>
                    <RouterLink
                        :to="{ name: 'groups' }"
                        class="inline-flex items-center gap-x-1 text-sm font-medium text-gray-500 hover:text-gray-700"
                    >
                        <ChevronLeftIcon class="size-5" aria-hidden="true" />
                        {{ t("common.button.back") }}
                    </RouterLink>

                    <h2 class="mt-3 truncate text-xl font-semibold text-gray-900">
                        {{ isEditing ? savedName : t("settings.headers.new_group.title") }}
                    </h2>
                    <p
                        v-if="isEditing"
                        class="mt-2 flex items-center gap-x-1.5 text-sm text-gray-500"
                    >
                        <UsersIcon class="size-4 text-gray-400" aria-hidden="true" />
                        {{ memberCount }}
                    </p>
                    <p v-else class="mt-2 text-sm text-gray-500">
                        {{ t("settings.headers.new_group.description") }}
                    </p>
                </div>

                <SectionCard :title="t('settings.groups.form.general')">
                    <div class="space-y-4 p-4">
                        <div>
                            <label for="group-name" class="block text-sm font-medium text-gray-900">
                                {{ t("settings.groups.form.name") }}
                            </label>
                            <input
                                id="group-name"
                                v-model="name"
                                type="text"
                                required
                                class="mt-2 block w-full rounded-md border-0 px-3 py-1.5 text-sm text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600"
                            />
                        </div>
                        <div>
                            <label
                                for="group-description"
                                class="block text-sm font-medium text-gray-900"
                            >
                                {{ t("settings.groups.form.description") }}
                            </label>
                            <textarea
                                id="group-description"
                                v-model="description"
                                rows="3"
                                class="mt-2 block w-full rounded-md border-0 px-3 py-1.5 text-sm text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600"
                            />
                        </div>
                    </div>
                </SectionCard>

                <SectionCard
                    :title="t('settings.groups.form.roles')"
                    :description="t('settings.groups.form.roles_hint')"
                >
                    <p
                        v-if="availableRoles.length === 0"
                        class="px-4 py-6 text-center text-sm text-gray-500"
                    >
                        {{ t("settings.groups.form.no_roles") }}
                    </p>
                    <ul v-else role="list" class="divide-y divide-gray-100">
                        <li
                            v-for="role in availableRoles"
                            :key="role.name"
                            class="relative flex items-start gap-x-3 px-4 py-3 hover:bg-gray-50"
                        >
                            <div class="flex h-6 items-center">
                                <input
                                    :id="`role-${role.name}`"
                                    v-model="selectedRoles"
                                    type="checkbox"
                                    :value="role.name"
                                    class="size-4 rounded border-gray-300 text-indigo-600 focus:ring-indigo-600"
                                />
                            </div>
                            <label :for="`role-${role.name}`" class="min-w-0 flex-1 cursor-pointer">
                                <span class="block text-sm font-medium text-gray-900">
                                    {{ roleLabel(role, "system") }}
                                </span>
                                <span
                                    v-if="roleDescription(role, 'system')"
                                    class="block truncate text-xs text-gray-500"
                                >
                                    {{ roleDescription(role, "system") }}
                                </span>
                            </label>
                        </li>
                    </ul>
                </SectionCard>

                <section v-if="isEditing" class="space-y-3">
                    <MembersTable
                        v-model:sort="memberSort"
                        :members="members"
                        :columns="memberColumns"
                        :loading="membersLoading"
                        :itemsPerPage="MEMBERS_PER_PAGE"
                        :totalItems="memberTotal"
                        :currentPage="memberPage"
                        @update:currentPage="changeMemberPage"
                    >
                        <template #toolbar>
                            <div class="flex flex-wrap items-center justify-between gap-3">
                                <h3 class="text-sm font-semibold text-gray-900">
                                    {{ t("settings.groups.form.members") }}
                                </h3>
                                <div class="flex items-center gap-x-3">
                                    <div v-if="memberCount > 0" class="relative w-56">
                                        <MagnifyingGlassIcon
                                            class="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-gray-400"
                                            aria-hidden="true"
                                        />
                                        <input
                                            v-model="memberSearch"
                                            type="search"
                                            :placeholder="t('settings.groups.form.search_members')"
                                            class="block w-full rounded-md border-0 py-1.5 pl-9 pr-3 text-sm text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600"
                                        />
                                    </div>
                                    <BaseButton
                                        size="small"
                                        :prepend-icon="PlusIcon"
                                        @click="addDialogOpen = true"
                                    >
                                        {{ t("settings.groups.form.add_members") }}
                                    </BaseButton>
                                </div>
                            </div>
                        </template>

                        <template #joined_at="{ item }">
                            {{ getDate(item.joined_at) }}
                        </template>

                        <template #actions="{ item }">
                            <button
                                type="button"
                                class="text-red-600 hover:text-red-500"
                                @click="askRemove(item)"
                            >
                                {{ t("common.button.remove") }}
                            </button>
                        </template>

                        <template #empty>
                            {{
                                memberCount === 0
                                    ? t("settings.groups.form.no_members")
                                    : t("settings.groups.form.no_match")
                            }}
                        </template>
                    </MembersTable>
                </section>
            </div>
        </div>

        <div class="flex shrink-0 items-center justify-end gap-x-3 border-t bg-gray-50 px-6 py-4">
            <button
                type="button"
                class="rounded-md px-3 py-2 text-sm font-semibold text-gray-700 hover:text-gray-900"
                @click="router.push({ name: 'groups' })"
            >
                {{ t("common.button.cancel") }}
            </button>
            <BaseButton :is-disabled="!name.trim()" :is-loading="saving" @click="save">
                {{ isEditing ? t("common.button.save") : t("common.button.create") }}
            </BaseButton>
        </div>

        <ConfirmDialog
            :open="removeDialogOpen"
            :title="t('settings.groups.form.confirm_remove_title')"
            :message="
                t('settings.groups.form.confirm_remove_description', { name: memberToRemoveName })
            "
            :confirm-label="t('common.button.remove')"
            @confirm="confirmRemove"
            @close="removeDialogOpen = false"
        />

        <UserPickerDialog
            v-model="addDialogOpen"
            :title="t('settings.groups.add_dialog.title')"
            :subtitle="name"
            @confirm="addPicked"
        />
    </div>
</template>

<script setup>
import { t } from "@/i18n/index";

import { computed, onMounted, ref, watch } from "vue";
import { RouterLink, useRoute, useRouter } from "vue-router";
import { useSettingsStore } from "@/store/settings";
import { useAlertStore } from "@/store/alerts";
import roleService from "@/services/roleService";
import groupService from "@/services/groupService";
import MembersTable from "@/components/MembersTable.vue";
import useDateOperations from "@/composables/useDateOperations.js";
import UserPickerDialog from "@/components/UserPickerDialog.vue";
import BaseButton from "@/components/BaseButton.vue";
import ConfirmDialog from "@/components/ConfirmDialog.vue";
import SectionCard from "@/components/SectionCard.vue";
import { ChevronLeftIcon, MagnifyingGlassIcon, PlusIcon, UsersIcon } from "@heroicons/vue/20/solid";
import { roleDescription, roleLabel } from "@/utils/roleLabels";

const route = useRoute();
const router = useRouter();
const alertStore = useAlertStore();

const groupId = computed(() => route.params.id ?? "");
const isEditing = computed(() => Boolean(groupId.value));

const name = ref("");
const savedName = ref("");
const description = ref("");

// System roles granted to every group member. availableRoles excludes
// system_admin: a group must never be able to grant full bypass (the backend
// rejects it too).
const availableRoles = ref([]);
const selectedRoles = ref([]);

const MEMBERS_PER_PAGE = 15;
const SEARCH_DEBOUNCE_MS = 250;

const members = ref([]);
const memberCount = ref(0);
const memberTotal = ref(0);
const memberPage = ref(1);
const memberSearch = ref("");
const memberSort = ref({ key: "name", desc: false });
const membersLoading = ref(isEditing.value);
const addDialogOpen = ref(false);

const { getDate } = useDateOperations();

const memberColumns = computed(() => [
    {
        key: "joined_at",
        label: t.value("members_table.joined"),
        sortable: true,
        hiddenBelow: "sm",
    },
]);

let memberRequest = 0;
let searchTimer = null;

const saving = ref(false);

const removeDialogOpen = ref(false);
const memberToRemove = ref(null);

const memberToRemoveName = computed(() => {
    const m = memberToRemove.value;

    if (!m) return "";

    return `${m.user_info?.name ?? ""} ${m.user_info?.lastname ?? ""}`.trim();
});

watch(memberSort, () => {
    memberPage.value = 1;
    loadMembers();
});

watch(memberSearch, () => {
    clearTimeout(searchTimer);

    searchTimer = setTimeout(() => {
        memberPage.value = 1;
        loadMembers();
    }, SEARCH_DEBOUNCE_MS);
});

async function loadRoles() {
    try {
        const { data } = await roleService.getRoles();

        availableRoles.value = (data ?? []).filter((r) => r.name !== "system_admin");
    } catch {
        availableRoles.value = [];
    }
}

async function loadGroup() {
    try {
        const { data: group } = await groupService.get(groupId.value);

        name.value = group.name;
        savedName.value = group.name;
        description.value = group.description ?? "";
        memberCount.value = group.member_count ?? 0;
        selectedRoles.value = group.roles ?? [];

        return true;
    } catch {
        alertStore.showError(t.value("settings.groups.retrieval_failed"));
        router.push({ name: "groups" });

        return false;
    }
}

async function loadMembers() {
    const request = ++memberRequest;

    membersLoading.value = true;

    try {
        const { data } = await groupService.members(
            groupId.value,
            memberSearch.value.trim(),
            MEMBERS_PER_PAGE,
            (memberPage.value - 1) * MEMBERS_PER_PAGE,
            memberSort.value,
        );

        if (request !== memberRequest) return;

        members.value = data.items ?? [];
        memberTotal.value = data.total ?? 0;
    } catch {
        if (request !== memberRequest) return;

        alertStore.showError(t.value("settings.groups.retrieval_failed"));
    } finally {
        if (request === memberRequest) membersLoading.value = false;
    }

    if (members.value.length === 0 && memberPage.value > 1) {
        memberPage.value -= 1;
        await loadMembers();
    }
}

function changeMemberPage(page) {
    memberPage.value = page;
    loadMembers();
}

async function refreshMembers() {
    try {
        const { data: group } = await groupService.get(groupId.value);

        memberCount.value = group.member_count ?? 0;
    } catch {
        alertStore.showError(t.value("settings.groups.retrieval_failed"));
    }

    await loadMembers();
}

async function save() {
    if (!name.value.trim()) return;
    saving.value = true;
    try {
        if (isEditing.value) {
            await groupService.update(groupId.value, {
                name: name.value.trim(),
                description: description.value.trim(),
                roles: selectedRoles.value,
            });
            alertStore.showSuccess(t.value("settings.groups.updated"));
        } else {
            const { data: created } = await groupService.create({
                name: name.value.trim(),
                description: description.value.trim(),
                roles: selectedRoles.value,
            });

            savedName.value = created.name;

            alertStore.showSuccess(t.value("settings.groups.created"));
            router.replace({
                name: "edit-group",
                params: { id: created.id },
            });

            return;
        }

        router.push({ name: "groups" });
    } catch {
        alertStore.showError(
            isEditing.value
                ? t.value("settings.groups.update_failed")
                : t.value("settings.groups.create_failed"),
        );
    } finally {
        saving.value = false;
    }
}

async function addPicked(users) {
    if (!users || users.length === 0) return;

    try {
        await groupService.addMembers(
            groupId.value,
            users.map((u) => u.id),
        );
    } catch {
        alertStore.showError(t.value("settings.groups.add_member_failed"));

        return;
    }

    await refreshMembers();
}

function askRemove(member) {
    memberToRemove.value = member;
    removeDialogOpen.value = true;
}

async function confirmRemove() {
    removeDialogOpen.value = false;

    try {
        await groupService.removeMember(groupId.value, memberToRemove.value.user_id);
    } catch {
        alertStore.showError(t.value("settings.groups.remove_member_failed"));

        return;
    }

    await refreshMembers();
}

onMounted(async () => {
    if (!useSettingsStore().getLicenseFeature("groups")) {
        router.push({ name: "groups" });

        return;
    }

    if (!isEditing.value) {
        await loadRoles();

        return;
    }

    const [, loaded] = await Promise.all([loadRoles(), loadGroup()]);

    if (loaded) await loadMembers();
});
</script>
