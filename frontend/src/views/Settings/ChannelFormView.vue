<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div class="flex h-full flex-col">
        <div v-if="loaded" class="flex-1 overflow-y-auto">
            <div class="mx-auto max-w-3xl px-6 py-6">
                <div class="space-y-6">
                    <div>
                        <RouterLink
                            :to="{ name: 'channel-list' }"
                            class="inline-flex items-center gap-x-1 text-sm font-medium text-gray-500 hover:text-gray-700"
                        >
                            <ChevronLeftIcon class="size-5" aria-hidden="true" />
                            {{ t("common.button.back") }}
                        </RouterLink>

                        <div class="mt-3 flex items-center gap-x-3">
                            <h2 class="truncate text-xl font-semibold text-gray-900">
                                {{ oldChannel.displayname }}
                            </h2>
                            <span
                                v-if="archived"
                                class="inline-flex shrink-0 items-center rounded-md bg-gray-50 px-2 py-0.5 text-xs font-medium text-gray-600 ring-1 ring-inset ring-gray-500/10"
                            >
                                {{ t("settings.channels.archived") }}
                            </span>
                        </div>
                        <p class="mt-2 flex items-center gap-x-1.5 text-sm text-gray-500">
                            <UsersIcon class="size-4 text-gray-400" aria-hidden="true" />
                            {{ channel.channel_members.length }}
                        </p>
                    </div>

                    <div v-if="archived" class="rounded-md bg-yellow-50 p-4">
                        <div class="flex">
                            <ArchiveBoxIcon
                                class="size-5 shrink-0 text-yellow-400"
                                aria-hidden="true"
                            />
                            <p class="ml-3 text-sm text-yellow-700">
                                {{ t("settings.channel_settings.archived_notice") }}
                            </p>
                        </div>
                    </div>

                    <SectionCard
                        :title="t('settings.channel_settings.details.title')"
                        :description="t('settings.channel_settings.details.subtitle')"
                    >
                        <div class="space-y-4 p-4">
                            <div>
                                <label
                                    for="channel-name"
                                    class="block text-sm font-medium text-gray-900"
                                >
                                    {{ t("settings.channel_settings.details.name") }}
                                </label>
                                <input
                                    id="channel-name"
                                    v-model="channel.displayname"
                                    type="text"
                                    :disabled="archived"
                                    :placeholder="
                                        t('settings.channel_settings.details.name_placeholder')
                                    "
                                    class="mt-2 block w-full rounded-md border-0 px-3 py-1.5 text-sm text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 disabled:cursor-not-allowed disabled:bg-gray-50 disabled:text-gray-500"
                                />
                            </div>
                            <div>
                                <label
                                    for="channel-description"
                                    class="block text-sm font-medium text-gray-900"
                                >
                                    {{ t("settings.channel_settings.details.description") }}
                                </label>
                                <textarea
                                    id="channel-description"
                                    v-model="channel.description"
                                    rows="3"
                                    maxlength="255"
                                    :disabled="archived"
                                    :placeholder="
                                        t(
                                            'settings.channel_settings.details.description_placeholder',
                                        )
                                    "
                                    class="mt-2 block w-full rounded-md border-0 px-3 py-1.5 text-sm text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 disabled:cursor-not-allowed disabled:bg-gray-50 disabled:text-gray-500"
                                />
                                <p class="mt-1 text-right text-xs text-gray-500">
                                    {{ (channel.description ?? "").length }}/255
                                </p>
                            </div>
                        </div>
                    </SectionCard>

                    <section class="space-y-3">
                        <MembersTable
                            status
                            :members="members"
                            :columns="memberColumns"
                            :itemsPerPage="MEMBERS_PER_PAGE"
                        >
                            <template #toolbar>
                                <div class="flex flex-wrap items-center justify-between gap-3">
                                    <div>
                                        <h3 class="text-sm font-semibold text-gray-900">
                                            {{ t("settings.channel_settings.members.title") }}
                                        </h3>
                                        <p class="mt-0.5 text-xs text-gray-500">
                                            {{ t("settings.channel_settings.members.subtitle") }}
                                        </p>
                                    </div>
                                    <BaseButton
                                        size="small"
                                        :prepend-icon="PlusIcon"
                                        :isDisabled="archived"
                                        @click="addUserDialog = true"
                                    >
                                        {{ t("settings.channel_settings.add_user") }}
                                    </BaseButton>
                                </div>
                            </template>

                            <template #role="{ item }">
                                <DropdownMenu
                                    :disabled="archived"
                                    :items="[
                                        {
                                            label: t(
                                                'settings.channel_settings.make_channel_admin',
                                            ),
                                            value: 'admin',
                                        },
                                        {
                                            label: t(
                                                'settings.channel_settings.make_channel_member',
                                            ),
                                            value: 'member',
                                        },
                                    ]"
                                    @select="setRole(item, $event)"
                                >
                                    <template v-slot:button>
                                        <div
                                            class="inline-flex items-center gap-x-1 rounded-md px-2 py-1 text-sm text-gray-700 hover:bg-gray-50"
                                        >
                                            {{ item.role }}
                                            <ChevronDownIcon
                                                class="size-4 text-gray-400"
                                                aria-hidden="true"
                                            />
                                        </div>
                                    </template>
                                </DropdownMenu>
                            </template>

                            <template #actions="{ item }">
                                <button
                                    type="button"
                                    :disabled="archived"
                                    class="text-red-600 hover:text-red-500 disabled:cursor-not-allowed disabled:text-gray-400"
                                    @click="
                                        deleteDialog = true;
                                        itemToDelete = item;
                                    "
                                >
                                    {{ t("common.button.remove") }}
                                </button>
                            </template>
                        </MembersTable>

                        <AddUserDialog
                            v-model="addUserDialog"
                            :channel="channel"
                            @add="addUser"
                            @add-groups="addGroups"
                        />

                        <ConfirmDialog
                            :open="deleteDialog"
                            :title="t('settings.channel_settings.user_delete_dialog.title')"
                            :message="t('settings.channel_settings.user_delete_dialog.confirm')"
                            :confirm-label="t('common.button.remove')"
                            @confirm="deleteUser(itemToDelete)"
                            @close="
                                deleteDialog = false;
                                itemToDelete = null;
                            "
                        />
                    </section>

                    <SectionCard
                        v-if="can('archive_channels') || can('delete_channels')"
                        :title="t('settings.channel_settings.actions.title')"
                    >
                        <div class="divide-y divide-gray-100">
                            <div
                                v-if="can('archive_channels')"
                                class="flex items-center justify-between gap-x-6 px-4 py-3"
                            >
                                <div>
                                    <p class="text-sm font-medium text-gray-900">
                                        {{
                                            channel?.deleted
                                                ? t(
                                                      "settings.channel_settings.archive.unarchive_button",
                                                  )
                                                : t("settings.channel_settings.archive.button")
                                        }}
                                    </p>
                                    <p class="mt-0.5 text-xs text-gray-500">
                                        {{
                                            channel?.deleted
                                                ? t(
                                                      "settings.channel_settings.archive.unarchive_description",
                                                  )
                                                : t("settings.channel_settings.archive.description")
                                        }}
                                    </p>
                                </div>
                                <button
                                    type="button"
                                    class="shrink-0 rounded-md px-3 py-1.5 text-sm font-semibold text-gray-700 ring-1 ring-inset ring-gray-300 hover:bg-gray-50"
                                    @click="archiveDialog = true"
                                >
                                    {{
                                        channel?.deleted
                                            ? t(
                                                  "settings.channel_settings.archive.unarchive_button",
                                              )
                                            : t("settings.channel_settings.archive.button")
                                    }}
                                </button>
                            </div>

                            <div
                                v-if="can('delete_channels')"
                                class="flex items-center justify-between gap-x-6 px-4 py-3"
                            >
                                <div>
                                    <p class="text-sm font-medium text-red-700">
                                        {{ t("settings.channel_settings.delete.button") }}
                                    </p>
                                    <p class="mt-0.5 text-xs text-gray-500">
                                        {{ t("settings.channel_settings.delete.description") }}
                                    </p>
                                </div>
                                <button
                                    type="button"
                                    class="shrink-0 rounded-md bg-red-600 px-3 py-2 text-sm font-semibold text-white shadow-xs hover:bg-red-500"
                                    @click="deleteChannelDialog = true"
                                >
                                    {{ t("settings.channel_settings.delete.button") }}
                                </button>
                            </div>
                        </div>
                    </SectionCard>
                </div>
            </div>
        </div>
        <div v-else class="flex flex-1 items-center justify-center">
            <BaseSpinner />
        </div>

        <div
            v-if="loaded"
            class="flex shrink-0 items-center justify-end gap-x-3 border-t bg-gray-50 px-6 py-4"
        >
            <button
                type="button"
                class="rounded-md px-3 py-2 text-sm font-semibold text-gray-700 hover:text-gray-900"
                @click="router.push({ name: 'channel-list' })"
            >
                {{ t("common.button.cancel") }}
            </button>
            <BaseButton :is-disabled="disableButton() || archived" @click="updateChannel">
                {{ t("common.button.save") }}
            </BaseButton>
        </div>

        <ArchiveChannelDialog
            v-model="archiveDialog"
            :channelName="channel?.displayname"
            :unarchive="!!channel?.deleted"
            @confirm="confirmArchive"
        />
        <DeleteChannelDialog
            v-model="deleteChannelDialog"
            :channelName="channel?.displayname"
            @confirm="confirmDeleteChannel"
        />
    </div>
</template>

<script setup>
import { t } from "@/i18n/index.js";

import { ref, computed, onMounted } from "vue";
import { RouterLink, useRoute, useRouter } from "vue-router";
import { ArchiveBoxIcon } from "@heroicons/vue/24/outline";
import { ChevronDownIcon, ChevronLeftIcon, PlusIcon, UsersIcon } from "@heroicons/vue/20/solid";
import BaseButton from "@/components/BaseButton.vue";
import BaseSpinner from "@/components/BaseSpinner.vue";
import ConfirmDialog from "@/components/ConfirmDialog.vue";
import MembersTable from "@/components/MembersTable.vue";
import AddUserDialog from "@/components/Chat/Dialogs/AddUserDialog.vue";
import DropdownMenu from "@/components/DropdownMenu.vue";
import chatService from "@/services/chatService";
import { usePermissions } from "@/composables/usePermissions";
import { useAlertStore } from "@/store/alerts";
import DeleteChannelDialog from "@/components/Chat/Dialogs/DeleteChannelDialog.vue";
import SectionCard from "@/components/SectionCard.vue";
import ArchiveChannelDialog from "@/components/Chat/Dialogs/ArchiveChannelDialog.vue";
import { useUserStore } from "@/store/user";
import { useUsers } from "@/composables/useUser";

const loaded = ref(false);
const deleteDialog = ref(false);
const itemToDelete = ref(null);
const addUserDialog = ref(false);
const userStore = useUserStore();
const route = useRoute();
const router = useRouter();
const channel = ref(null);
const deleteChannelDialog = ref(false);
const archiveDialog = ref(false);
const archived = computed(() => !!channel.value?.deleted);

function setArchivedAt(at) {
    channel.value.deleted = at;
    oldChannel.value.deleted = at;
}

function confirmArchive() {
    if (channel.value.deleted) {
        chatService.unarchiveChannel(channel.value.id).then(() => {
            setArchivedAt(0);
        });

        return;
    }

    chatService.archiveChannel(channel.value.id).then(() => {
        setArchivedAt(Math.floor(Date.now() / 1000));
    });
}

const { can } = usePermissions();

// Deleting archives the channel and hands the rest to a job, so the row is
// still listed, as archived, until that finishes.
function confirmDeleteChannel() {
    chatService.deleteChannel(channel.value.id).then(() => {
        useAlertStore().showSuccess(t.value("settings.channel_settings.delete.started"));
        router.push({ name: "channels" });
    });
}

// Trigger a lazy load so member names resolve in the list.
useUsers(() => channel.value?.channel_members?.map((m) => m.user_id) ?? []);
const oldChannel = ref(null);
const usersToRemove = ref([]);
const usersToAdd = ref([]);
const groupsToAdd = ref([]);

const MEMBERS_PER_PAGE = 20;

const memberColumns = computed(() => [
    {
        key: "role",
        label: t.value("members_table.role"),
        sortable: true,
    },
]);

const members = computed(() =>
    channel.value.channel_members
        .map((member) => ({ ...member, user_info: userStore.getUserById(member.user_id) }))
        .filter((member) => member.user_info),
);

async function updateChannel() {
    const id = channel.value.id;
    const oldMembers = oldChannel.value.channel_members;
    const wasMember = (userId) => oldMembers.some((member) => member.user_id === userId);

    try {
        await chatService.updateChannel(id, {
            displayname: channel.value.displayname,
            description: channel.value.description,
        });

        const added = usersToAdd.value.filter((userId) => !wasMember(userId));

        if (added.length > 0) {
            await chatService.addUsersToChannel({ id, users: added });
        }

        if (groupsToAdd.value.length > 0) {
            await chatService.addGroupsToChannel(
                id,
                groupsToAdd.value.map((group) => group.id),
            );
        }

        const removals = usersToRemove.value
            .filter(wasMember)
            .map((userId) => chatService.removeUserFromChannel({ id, user: userId }));

        const roleChanges = channel.value.channel_members
            .filter((member) => {
                const old = oldMembers.find((m) => m.user_id === member.user_id);

                return old ? old.role !== member.role : member.role !== "member";
            })
            .map((member) =>
                chatService.changeUserRole({ id, user: member.user_id, role: member.role }),
            );

        await Promise.all([...removals, ...roleChanges]);

        const { data } = await chatService.getChannelById(id);

        channel.value = data;
        oldChannel.value = JSON.parse(JSON.stringify(data));
        usersToAdd.value = [];
        usersToRemove.value = [];
        groupsToAdd.value = [];

        useAlertStore().showSuccess(t.value("settings.channel_settings.saved"));
    } catch {
        useAlertStore().showError(t.value("settings.channel_settings.save_failed"));
    }
}

function addUser(users) {
    addUserDialog.value = false;

    userStore.addUsers(users);

    for (const user of users) {
        if (channel.value.channel_members.some((member) => member.user_id === user.id)) {
            continue;
        }

        channel.value.channel_members.push({ user_id: user.id, role: "member" });

        if (usersToRemove.value.includes(user.id)) {
            usersToRemove.value = usersToRemove.value.filter((id) => id !== user.id);
        } else {
            usersToAdd.value.push(user.id);
        }
    }
}

function addGroups(groups) {
    addUserDialog.value = false;

    for (const group of groups) {
        if (!groupsToAdd.value.some((g) => g.id === group.id)) groupsToAdd.value.push(group);
    }
}

function deleteUser(item) {
    channel.value.channel_members = channel.value.channel_members.filter(
        (member) => member.user_id !== item.user_id,
    );

    if (!usersToAdd.value.includes(item.user_id)) {
        usersToRemove.value.push(item.user_id);
    } else {
        usersToAdd.value = usersToAdd.value.filter((id) => id !== item.user_id);
    }

    deleteDialog.value = false;
    itemToDelete.value = null;
}

function setRole(item, event) {
    channel.value.channel_members.forEach((member) => {
        if (member.user_id === item.user_id) {
            member.role = event.value;
        }
    });
}

function disableButton() {
    return (
        JSON.stringify(channel.value) === JSON.stringify(oldChannel.value) &&
        usersToRemove.value.length === 0 &&
        usersToAdd.value.length === 0 &&
        groupsToAdd.value.length === 0
    );
}

onMounted(async () => {
    chatService.getChannelById(route.params.id).then((response) => {
        channel.value = response.data;
        oldChannel.value = JSON.parse(JSON.stringify(response.data));

        loaded.value = true;
    });
});
</script>
