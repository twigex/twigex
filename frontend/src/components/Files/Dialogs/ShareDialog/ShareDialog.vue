<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <TransitionRoot as="template" :show="modelValue">
        <Dialog as="div" class="relative z-50" @close="closeDialog()">
            <TransitionChild
                as="template"
                enter="ease-out duration-300"
                enter-from="opacity-0"
                enter-to="opacity-100"
                leave="ease-in duration-200"
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
                            class="relative flex max-h-[85vh] transform flex-col rounded-lg bg-white px-4 pb-4 pt-5 text-left shadow-xl transition-all w-full sm:my-8 sm:w-full sm:max-w-lg sm:p-6"
                        >
                            <div class="flex min-h-0 flex-1 flex-col">
                                <div
                                    class="mt-3 flex min-h-0 w-full flex-1 flex-col text-center sm:mt-0 sm:text-left"
                                >
                                    <div class="shrink-0">
                                        <div class="justify-between flex flex-row">
                                            <DialogTitle
                                                as="h3"
                                                class="text-base truncate font-semibold leading-6 text-gray-900 mr-5"
                                                >{{ t("files.share_dialog.title") }}
                                                {{
                                                    dialogStore.shareDialog.file
                                                        ? dialogStore.shareDialog.file.name
                                                        : ""
                                                }}</DialogTitle
                                            >
                                        </div>
                                        <ShareTabSwitcher
                                            v-if="
                                                dialogStore.shareDialog.editPermission !==
                                                    'share' && canShare
                                            "
                                            v-model:active-tab="activeTab"
                                        />
                                    </div>

                                    <div class="min-h-0 flex-1 overflow-y-auto px-1 pb-1">
                                        <div v-show="activeTab === 'people'" class="mt-2">
                                            <SharePeoplePicker
                                                v-if="
                                                    dialogStore.shareDialog.editPermission !==
                                                        'share' && canShare
                                                "
                                                v-model:selected-users="selectedUsers"
                                                v-model:selected-groups="selectedGroups"
                                                :exclude-user-ids="excludeUserIds"
                                                :exclude-group-ids="excludeGroupIds"
                                                :placeholder="
                                                    t('common.label.search_people_groups')
                                                "
                                            />

                                            <Transition
                                                enter-active-class="transition-opacity duration-500"
                                                enter-from-class="opacity-0"
                                                enter-to-class="opacity-100"
                                                leave-active-class="transition-opacity duration-0"
                                                leave-from-class="opacity-0"
                                                leave-to-class="opacity-0"
                                            >
                                                <div
                                                    v-if="
                                                        selectionCount <= 0 &&
                                                        dialogStore.shareDialog.editPermission !=
                                                            'share'
                                                    "
                                                    class="mt-5"
                                                >
                                                    <ShareAccessList
                                                        :owner-id="
                                                            dialogStore.shareDialog.file?.owner
                                                        "
                                                        :file-details="fileDetails"
                                                        :show-user-menu="
                                                            (user) =>
                                                                canShare ||
                                                                user.id === userStore.user.id
                                                        "
                                                        :show-group-menu="canShare"
                                                        :user-menu-items="userMenuItems"
                                                        :group-menu-items="groupMenuItems"
                                                        @menu-select="(item) => item.action()"
                                                        @open-group="openGroupMembers"
                                                    />
                                                </div>
                                            </Transition>
                                            <TransitionRoot
                                                :show="
                                                    selectionCount > 0 ||
                                                    dialogStore.shareDialog.editPermission ==
                                                        'share'
                                                "
                                                enter="transition-opacity duration-500"
                                                enter-from="opacity-0"
                                                enter-to="opacity-100"
                                                leave="transition-opacity duration-0"
                                                leave-from="opacity-100"
                                                leave-to="opacity-0"
                                            >
                                                <UserView
                                                    :expiry="dialogStore.shareDialog.expiry"
                                                    :access-level="
                                                        dialogStore.shareDialog.accessLevel
                                                    "
                                                    @update:expiry="
                                                        dialogStore.shareDialog.expiry = $event
                                                    "
                                                    @update:access-level="
                                                        dialogStore.shareDialog.accessLevel = $event
                                                    "
                                                />
                                            </TransitionRoot>
                                        </div>

                                        <div v-show="activeTab === 'link'" class="mt-2">
                                            <template v-if="!showLinkForm">
                                                <ShareLinkList
                                                    :links="fileDetails?.sharedLinks ?? []"
                                                    :access-label="accessLabel"
                                                    @edit="startEditLink"
                                                    @copy="copyLink"
                                                    @remove="removeLink"
                                                    @create="openLinkForm"
                                                />
                                            </template>

                                            <template v-else>
                                                <h3
                                                    class="mb-2 text-sm font-semibold text-gray-900"
                                                >
                                                    {{
                                                        editingLink
                                                            ? t(
                                                                  "files.share_dialog.link.edit_title",
                                                              )
                                                            : t("files.share_dialog.link.new_title")
                                                    }}
                                                </h3>
                                                <LinkView
                                                    :key="editingLink || 'new'"
                                                    v-model="linkConfig"
                                                    :is-folder="isFolder"
                                                />
                                            </template>
                                        </div>
                                    </div>
                                </div>
                            </div>
                            <div
                                class="mt-5 grid shrink-0 gap-3 sm:mt-4"
                                :class="showPrimaryAction ? 'grid-cols-2' : 'grid-cols-1'"
                            >
                                <button
                                    @click="
                                        activeTab === 'link' && showLinkForm
                                            ? backToLinkList()
                                            : closeDialog()
                                    "
                                    type="button"
                                    class="rounded-md bg-gray-100 px-3 py-2 text-sm font-semibold hover:bg-gray-200"
                                >
                                    {{
                                        activeTab !== "link"
                                            ? t("common.button.cancel")
                                            : showLinkForm
                                              ? t("common.button.back")
                                              : t("common.button.close")
                                    }}
                                </button>

                                <button
                                    v-if="showPrimaryAction"
                                    @click="
                                        activeTab === 'people'
                                            ? shareFile()
                                            : editingLink
                                              ? updateLink()
                                              : createLink()
                                    "
                                    :disabled="
                                        activeTab === 'people'
                                            ? selectionCount <= 0 &&
                                              dialogStore.shareDialog.editPermission != 'share'
                                            : creatingLink
                                    "
                                    class="rounded-md bg-indigo-600 px-3 py-2 text-sm font-semibold text-white hover:bg-indigo-500 disabled:opacity-50"
                                >
                                    {{
                                        activeTab === "people"
                                            ? dialogStore.shareDialog.editPermission != ""
                                                ? t("common.button.save")
                                                : t("common.button.share")
                                            : editingLink
                                              ? t("common.button.save")
                                              : t("files.share_dialog.link.create")
                                    }}
                                </button>
                            </div>
                        </DialogPanel>
                    </TransitionChild>
                </div>
            </div>

            <!-- Nested inside the share Dialog so HeadlessUI's dialog stack
                     handles it: closing this viewer only closes
                     the viewer, not the share dialog underneath it. -->
            <GroupMembersDialog
                v-model="groupMembersOpen"
                :group-id="groupMembersGroup?.id ?? ''"
                :group-name="groupMembersGroup?.name ?? ''"
            />
            <ConfirmDialog
                :open="confirm.open"
                :title="t('files.share_dialog.inherited_title')"
                :message="confirm.message"
                @confirm="runConfirm"
                @close="confirm.open = false"
            />
        </Dialog>
    </TransitionRoot>
</template>

<script setup>
import { t } from "@/i18n/index.js";

import { computed, ref, watch } from "vue";
import { useDialogStore } from "@/store/dialogs";
import { useUserStore } from "@/store/user";
import fileService from "@/services/fileService";
import shareService from "@/services/shareService";
import { useAlertStore } from "@/store/alerts";
import { useDetailsStore } from "@/store/details";
import { extractErrorMessage } from "@/utils/errors";
import UserView from "@/components/Files/Dialogs/ShareDialog/UserView.vue";
import LinkView from "@/components/Files/Dialogs/ShareDialog/LinkView.vue";
import SharePeoplePicker from "@/components/Files/Dialogs/ShareDialog/SharePeoplePicker.vue";
import ShareTabSwitcher from "@/components/Files/Dialogs/ShareDialog/ShareTabSwitcher.vue";
import ShareAccessList from "@/components/Files/Dialogs/ShareDialog/ShareAccessList.vue";
import ShareLinkList from "@/components/Files/Dialogs/ShareDialog/ShareLinkList.vue";
import GroupMembersDialog from "@/components/GroupMembersDialog.vue";
import ConfirmDialog from "@/components/ConfirmDialog.vue";
import { Dialog, DialogPanel, DialogTitle, TransitionChild, TransitionRoot } from "@headlessui/vue";

const props = defineProps({
    modelValue: {
        type: Boolean,
        required: true,
    },
});

const emit = defineEmits([
    "close",
    "share",
    "unshare",
    "unshare-group",
    "update-share",
    "link-share",
    "update-link",
]);

const dialogStore = useDialogStore();
const detailsStore = useDetailsStore();
const userStore = useUserStore();
const alertStore = useAlertStore();
const selectedUsers = ref([]);
const selectedGroups = ref([]);
const sharedUserIds = ref([]);
const sharedGroupIds = ref([]);
const groupMembersOpen = ref(false);
const groupMembersGroup = ref(null);
const activeTab = ref("people");
const linkConfig = ref({});
const creatingLink = ref(false);
const editingLink = ref(null);
const showLinkForm = ref(false);

const isFolder = computed(() => dialogStore.shareDialog.file?.isFolder ?? false);

const isOwner = computed(() => dialogStore.shareDialog.file?.owner === userStore.user.id);

// Only an owner or manager may share/manage, so only they see the picker and menus.
const canShare = computed(() => isOwner.value || fileDetails.value?.accessLevel === "manager");

// Non-sharers get a read-only dialog, so there is no primary action for them.
const showPrimaryAction = computed(
    () => (activeTab.value === "people" && canShare.value) || showLinkForm.value,
);

const selectionCount = computed(() => selectedUsers.value.length + selectedGroups.value.length);

// Hide already-shared people/groups and the current user from the picker.
const excludeUserIds = computed(() => [userStore.user.id, ...sharedUserIds.value]);
const excludeGroupIds = computed(() => sharedGroupIds.value);

function openGroupMembers(group) {
    groupMembersGroup.value = group;
    groupMembersOpen.value = true;
}

const fileDetails = ref(null);
const confirm = ref({ open: false, message: "", action: null });

watch(
    () => props.modelValue,
    async () => {
        if (!props.modelValue) return;

        const { data: details } = await fileService.getDetails(dialogStore.shareDialog.file.id);

        fileDetails.value = details;

        // Exclude only people granted directly on this item; inherited users
        // stay selectable so they can be granted more access on this subtree.
        sharedUserIds.value = (details.sharedUsers ?? [])
            .filter((u) => !u.inherited)
            .map((u) => u.id);
        sharedGroupIds.value = (details.sharedGroups ?? [])
            .filter((g) => !g.inherited)
            .map((g) => g.id);

        // Opened straight into editing a link (from the Details sidebar).
        if (dialogStore.shareDialog.editLink) {
            activeTab.value = "link";
            startEditLink(dialogStore.shareDialog.editLink);
            dialogStore.shareDialog.editLink = null;
        }
    },
);

const SHARE_TYPE_USER = 1;
const SHARE_TYPE_GROUP = 2;

function userMenuItems(user) {
    const items = [];

    if (canShare.value) {
        items.push({
            label: t.value("files.share_dialog.edit_permissions"),
            action: () => editShare(user, SHARE_TYPE_USER),
        });
    }

    items.push({
        label: t.value("files.share_dialog.unshare"),
        action: () => unshareFile(user.id),
    });

    return items;
}

function groupMenuItems(group) {
    return [
        {
            label: t.value("files.share_dialog.edit_permissions"),
            action: () => editShare(group, SHARE_TYPE_GROUP),
        },
        {
            label: t.value("files.share_dialog.unshare_group"),
            action: () => unshareGroupFromFile(group.id),
        },
    ];
}

function editShare(item, shareType) {
    withInheritedConfirm(item, () => {
        dialogStore.shareDialog.editItem = item;
        dialogStore.shareDialog.editShareType = shareType;
        // Groups serialize access_level (snake), users accessLevel (camel).
        dialogStore.shareDialog.accessLevel = item.accessLevel ?? item.access_level ?? "viewer";
        dialogStore.shareDialog.expiry = item.expiration;

        dialogStore.shareDialog.editPermission = "share";
    });
}

function shareFile() {
    if (dialogStore.shareDialog.editPermission == "share") {
        emit("update-share", {
            id: dialogStore.shareDialog.editItem.id,
            fileID: dialogStore.shareDialog.file.id,
            shareType: dialogStore.shareDialog.editShareType,
            accessLevel: dialogStore.shareDialog.accessLevel,
            expiry: dialogStore.shareDialog.expiry,
        });
        closeDialog();

        return;
    }

    // Most-permissive resolution means a subfolder grant can only raise access,
    // never lower it below what a user already inherits, so block that no-op.
    const rank = { viewer: 1, editor: 2, manager: 3 };
    const chosenRank = rank[dialogStore.shareDialog.accessLevel] ?? 0;
    const downgraded = selectedUsers.value
        .map((u) =>
            (fileDetails.value?.sharedUsers ?? []).find((x) => x.id === u.id && x.inherited),
        )
        .find((row) => row && (rank[row.accessLevel] ?? 0) > chosenRank);

    if (downgraded) {
        alertStore.showError(
            t.value("files.share_dialog.no_downgrade", {
                folder: downgraded.grantedBy,
            }),
        );

        return;
    }

    emit("share", {
        file: dialogStore.shareDialog.file.id,
        users: selectedUsers.value.map((u) => u.id),
        groups: selectedGroups.value.map((g) => g.id),
        accessLevel: dialogStore.shareDialog.accessLevel,
        expiry: dialogStore.shareDialog.expiry,
    });

    if (fileDetails.value) {
        fileDetails.value.sharedUsers = [
            ...(fileDetails.value.sharedUsers ?? []),
            ...selectedUsers.value,
        ];
        fileDetails.value.sharedGroups = [
            ...(fileDetails.value.sharedGroups ?? []),
            ...selectedGroups.value,
        ];
    }

    closeDialog();
}

// Keep the Details sidebar's link list in sync with the dialog so the user
// doesn't have to refresh after creating/editing/deleting a link.
function syncLinksToDetails() {
    if (detailsStore.details && detailsStore.file?.id === dialogStore.shareDialog.file?.id) {
        detailsStore.details.sharedLinks = [...(fileDetails.value.sharedLinks ?? [])];
    }
}

async function createLink() {
    if (creatingLink.value) return;
    creatingLink.value = true;
    try {
        const { data } = await shareService.shareLink({
            item: dialogStore.shareDialog.file.id,
            ...linkConfig.value,
        });

        if (!fileDetails.value.sharedLinks) {
            fileDetails.value.sharedLinks = [];
        }

        fileDetails.value.sharedLinks.push(data);
        syncLinksToDetails();
        copyLink(data.id);
        backToLinkList();
    } catch (error) {
        alertStore.showError(extractErrorMessage(error));
    } finally {
        creatingLink.value = false;
    }
}

function openLinkForm() {
    editingLink.value = null;
    linkConfig.value = {};
    showLinkForm.value = true;
}

function backToLinkList() {
    showLinkForm.value = false;
    editingLink.value = null;
    linkConfig.value = {};
}

function startEditLink(link) {
    editingLink.value = link.id;
    linkConfig.value = {
        allowView: link.allowView,
        allowDownload: link.allowDownload,
        allowUpload: link.allowUpload,
        allowEdit: link.allowEdit,
        passwordProtected: link.passwordProtected,
        password: "",
        expiration: link.expiration,
        maxDownloads: link.maxDownloads,
        message: link.message,
    };
    showLinkForm.value = true;
}

function cancelEdit() {
    editingLink.value = null;
    linkConfig.value = {};
}

async function updateLink() {
    if (creatingLink.value) return;
    creatingLink.value = true;
    try {
        await shareService.updateLinkShare({
            token: editingLink.value,
            ...linkConfig.value,
        });

        const idx = (fileDetails.value.sharedLinks ?? []).findIndex(
            (l) => l.id === editingLink.value,
        );

        if (idx !== -1) {
            fileDetails.value.sharedLinks[idx] = {
                ...fileDetails.value.sharedLinks[idx],
                allowView: linkConfig.value.allowView,
                allowDownload: linkConfig.value.allowDownload,
                allowUpload: linkConfig.value.allowUpload,
                allowEdit: linkConfig.value.allowEdit,
                passwordProtected: linkConfig.value.passwordProtected,
                maxDownloads: linkConfig.value.maxDownloads,
                message: linkConfig.value.message,
                expiration: linkConfig.value.expiration,
            };
        }

        syncLinksToDetails();
        backToLinkList();
        alertStore.showSuccess(t.value("files.share_dialog.link.updated"));
    } catch (error) {
        alertStore.showError(extractErrorMessage(error));
    } finally {
        creatingLink.value = false;
    }
}

async function removeLink(token) {
    try {
        await shareService.deleteLink(token);
        fileDetails.value.sharedLinks = (fileDetails.value.sharedLinks ?? []).filter(
            (link) => link.id !== token,
        );
        syncLinksToDetails();
        if (editingLink.value === token) {
            cancelEdit();
        }
    } catch (error) {
        alertStore.showError(extractErrorMessage(error));
    }
}

async function copyLink(token) {
    const url = `${window.location.origin}/share/${token}`;

    try {
        await navigator.clipboard.writeText(url);
        alertStore.showSuccess(t.value("files.share_dialog.link.copied"));
    } catch {
        alertStore.showError(t.value("files.share_dialog.link.copy_failed"));
    }
}

function accessLabel(link) {
    if (link.allowUpload && !link.allowView && !link.allowDownload) {
        return t.value("files.share_dialog.link.level_drop");
    }

    if (link.allowUpload) {
        return t.value("files.share_dialog.link.level_collaborate");
    }

    if (link.allowDownload) {
        return t.value("files.share_dialog.link.level_download");
    }

    return t.value("files.share_dialog.link.level_view");
}

// A grant on a parent folder is escalated to that folder, so confirm first;
// direct grants run immediately.
function withInheritedConfirm(item, action) {
    if (!item?.inherited) {
        action();

        return;
    }

    confirm.value = {
        open: true,
        message: t.value("files.share_dialog.inherited_confirm", {
            folder: item.grantedBy ?? item.granted_by,
        }),
        action,
    };
}

function runConfirm() {
    const action = confirm.value.action;

    confirm.value.open = false;
    if (action) {
        action();
    }
}

function unshareFile(userId) {
    const shared = (fileDetails.value?.sharedUsers ?? []).find((u) => u.id === userId);

    withInheritedConfirm(shared, () => {
        emit("unshare", {
            userId: userId,
            fileId: dialogStore.shareDialog.file.id,
        });

        sharedUserIds.value = sharedUserIds.value.filter((id) => id !== userId);
        fileDetails.value.sharedUsers = fileDetails.value.sharedUsers.filter(
            (u) => u.id !== userId,
        );
    });
}

function unshareGroupFromFile(groupId) {
    const shared = (fileDetails.value?.sharedGroups ?? []).find((g) => g.id === groupId);

    withInheritedConfirm(shared, () => {
        emit("unshare-group", {
            groupId: groupId,
            fileId: dialogStore.shareDialog.file.id,
        });

        sharedGroupIds.value = sharedGroupIds.value.filter((id) => id !== groupId);
        fileDetails.value.sharedGroups = (fileDetails.value.sharedGroups ?? []).filter(
            (g) => g.id !== groupId,
        );
    });
}

function closeDialog() {
    emit("close");

    setTimeout(() => {
        resetValues();
    }, 300);
}

function resetValues() {
    dialogStore.shareDialog.accessLevel = "viewer";
    dialogStore.shareDialog.expiry = null;
    sharedUserIds.value = [];
    sharedGroupIds.value = [];
    selectedUsers.value = [];
    selectedGroups.value = [];
    dialogStore.shareDialog.editPermission = "";
    activeTab.value = "people";
    linkConfig.value = {};
    editingLink.value = null;
    showLinkForm.value = false;
}
</script>
