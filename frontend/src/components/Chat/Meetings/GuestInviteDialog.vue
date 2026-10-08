<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <TransitionRoot as="template" :show="true">
        <Dialog class="relative z-[60]" @close="emit('close')">
            <TransitionChild
                as="template"
                enter="ease-out duration-300"
                enter-from="opacity-0"
                enter-to="opacity-100"
                leave="ease-in duration-200"
                leave-from="opacity-100"
                leave-to="opacity-0"
            >
                <div class="fixed inset-0 bg-gray-500/75 transition-opacity" />
            </TransitionChild>

            <div class="fixed inset-0 z-10 w-screen overflow-y-auto">
                <div class="flex min-h-full items-center justify-center p-4">
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
                            class="relative w-full max-w-lg transform overflow-hidden rounded-xl bg-white shadow-xl transition-all"
                        >
                            <div class="border-b border-gray-200 px-6 py-5">
                                <div class="flex items-center gap-3">
                                    <div
                                        class="flex h-9 w-9 shrink-0 items-center justify-center rounded-lg bg-indigo-100"
                                    >
                                        <UserPlusIcon class="h-5 w-5 text-indigo-600" />
                                    </div>
                                    <div>
                                        <DialogTitle
                                            as="h3"
                                            class="text-base font-semibold leading-6 text-gray-900"
                                        >
                                            {{ t("guest_invite.title") }}
                                        </DialogTitle>
                                        <p class="text-sm text-gray-500">
                                            {{ t("guest_invite.subtitle") }}
                                        </p>
                                    </div>
                                </div>
                            </div>

                            <div class="px-6 py-5">
                                <div
                                    class="flex flex-wrap gap-2 rounded-md p-2 ring-1 ring-inset focus-within:ring-2"
                                    :class="
                                        error
                                            ? 'ring-red-400 focus-within:ring-red-500'
                                            : 'ring-gray-300 focus-within:ring-indigo-600'
                                    "
                                >
                                    <span
                                        v-for="email in emails"
                                        :key="email"
                                        class="inline-flex items-center gap-1 rounded-md bg-indigo-50 px-2 py-1 text-sm text-indigo-700"
                                    >
                                        {{ email }}
                                        <button
                                            type="button"
                                            @click="removeEmail(email)"
                                            class="text-indigo-400 hover:text-indigo-600"
                                        >
                                            <XMarkIcon class="h-3.5 w-3.5" />
                                        </button>
                                    </span>
                                    <input
                                        v-model="draft"
                                        type="email"
                                        :placeholder="t('guest_invite.placeholder')"
                                        class="min-w-[8rem] flex-1 border-0 p-1 text-sm focus:outline-none focus:ring-0"
                                        @keydown.enter.prevent="commitDraft"
                                        @keydown="
                                            (e) =>
                                                e.key === ',' && (e.preventDefault(), commitDraft())
                                        "
                                        @keydown.space.prevent="commitDraft"
                                        @input="
                                            error = '';
                                            success = '';
                                        "
                                        @blur="commitDraft"
                                    />
                                </div>
                                <p
                                    class="mt-2 text-xs"
                                    :class="
                                        error
                                            ? 'text-red-600'
                                            : success
                                              ? 'text-green-600'
                                              : 'text-gray-400'
                                    "
                                >
                                    {{ error || success || t("guest_invite.hint") }}
                                </p>

                                <div class="mt-5">
                                    <label class="block text-sm font-medium text-gray-700">
                                        {{ t("guest_invite.people_label") }}
                                    </label>
                                    <div class="relative mt-1">
                                        <input
                                            v-model="userQuery"
                                            type="text"
                                            :placeholder="t('meetings.schedule.people_placeholder')"
                                            @input="onUserQuery"
                                            class="block w-full rounded-md border-0 py-2 px-3 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm"
                                        />
                                        <ul
                                            v-if="userResults.length"
                                            class="absolute z-10 mt-1 max-h-40 w-full overflow-auto rounded-md bg-white py-1 text-sm shadow-lg ring-1 ring-black/5"
                                        >
                                            <li
                                                v-for="u in userResults"
                                                :key="u.id"
                                                @click="addUser(u)"
                                                class="cursor-pointer px-3 py-2 text-gray-700 hover:bg-indigo-600 hover:text-white"
                                            >
                                                {{ u.name }}
                                            </li>
                                        </ul>
                                    </div>
                                    <div
                                        v-if="pickedUsers.length"
                                        class="mt-2 flex flex-wrap gap-2"
                                    >
                                        <span
                                            v-for="u in pickedUsers"
                                            :key="u.id"
                                            class="inline-flex items-center gap-1 rounded-md bg-indigo-50 px-2 py-1 text-sm text-indigo-700"
                                        >
                                            {{ u.name }}
                                            <button
                                                type="button"
                                                @click="removeUser(u.id)"
                                                class="text-indigo-400 hover:text-indigo-600"
                                            >
                                                <XMarkIcon class="h-3.5 w-3.5" />
                                            </button>
                                        </span>
                                    </div>
                                </div>

                                <div class="mt-5">
                                    <label class="block text-sm font-medium text-gray-700">
                                        {{ t("guest_invite.groups_label") }}
                                    </label>
                                    <div class="relative mt-1">
                                        <input
                                            v-model="groupQuery"
                                            type="text"
                                            :placeholder="t('guest_invite.groups_placeholder')"
                                            @input="onGroupQuery"
                                            class="block w-full rounded-md border-0 py-2 px-3 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm"
                                        />
                                        <ul
                                            v-if="groupResults.length"
                                            class="absolute z-10 mt-1 max-h-40 w-full overflow-auto rounded-md bg-white py-1 text-sm shadow-lg ring-1 ring-black/5"
                                        >
                                            <li
                                                v-for="g in groupResults"
                                                :key="g.id"
                                                @click="addGroup(g)"
                                                class="cursor-pointer px-3 py-2 text-gray-700 hover:bg-indigo-600 hover:text-white"
                                            >
                                                {{ g.name }}
                                            </li>
                                        </ul>
                                    </div>
                                    <div
                                        v-if="pickedGroups.length"
                                        class="mt-2 flex flex-wrap gap-2"
                                    >
                                        <span
                                            v-for="g in pickedGroups"
                                            :key="g.id"
                                            class="inline-flex items-center gap-1 rounded-md bg-indigo-50 px-2 py-1 text-sm text-indigo-700"
                                        >
                                            {{ g.name }}
                                            <button
                                                type="button"
                                                @click="removeGroup(g.id)"
                                                class="text-indigo-400 hover:text-indigo-600"
                                            >
                                                <XMarkIcon class="h-3.5 w-3.5" />
                                            </button>
                                        </span>
                                    </div>
                                </div>

                                <div class="mt-5">
                                    <label class="block text-sm font-medium text-gray-700">
                                        {{ t("guest_invite.expiry_label") }}
                                    </label>
                                    <Listbox v-model="selectedExpiry">
                                        <div class="relative mt-1">
                                            <ListboxButton
                                                class="relative w-full cursor-default rounded-md bg-white py-2 pl-3 pr-10 text-left text-sm shadow-sm ring-1 ring-inset ring-gray-300 focus:outline-none focus:ring-2 focus:ring-indigo-600"
                                            >
                                                <span class="block truncate text-gray-900">
                                                    {{ t(selectedExpiry.labelKey) }}
                                                </span>
                                                <span
                                                    class="pointer-events-none absolute inset-y-0 right-0 flex items-center pr-2"
                                                >
                                                    <ChevronUpDownIcon
                                                        class="h-5 w-5 text-gray-400"
                                                        aria-hidden="true"
                                                    />
                                                </span>
                                            </ListboxButton>

                                            <ListboxOptions
                                                class="absolute z-10 mt-1 max-h-60 w-full overflow-auto rounded-md bg-white py-1 text-base shadow-lg ring-1 ring-black/5 focus:outline-none sm:text-sm"
                                            >
                                                <ListboxOption
                                                    v-for="option in EXPIRY_OPTIONS"
                                                    :key="option.value"
                                                    v-slot="{ active, selected }"
                                                    as="template"
                                                    :value="option"
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
                                                                selected
                                                                    ? 'font-semibold'
                                                                    : 'font-normal',
                                                                'block truncate',
                                                            ]"
                                                        >
                                                            {{ t(option.labelKey) }}
                                                        </span>
                                                        <span
                                                            v-if="selected"
                                                            :class="[
                                                                active
                                                                    ? 'text-white'
                                                                    : 'text-indigo-600',
                                                                'absolute inset-y-0 right-0 flex items-center pr-4',
                                                            ]"
                                                        >
                                                            <CheckIcon
                                                                class="h-5 w-5"
                                                                aria-hidden="true"
                                                            />
                                                        </span>
                                                    </li>
                                                </ListboxOption>
                                            </ListboxOptions>
                                        </div>
                                    </Listbox>
                                </div>
                            </div>

                            <div class="border-t border-gray-200 px-6 py-5">
                                <h4 class="text-sm font-medium text-gray-700">
                                    {{ t("guest_invite.active_title") }}
                                </h4>

                                <p v-if="!invites.length" class="mt-2 text-sm text-gray-400">
                                    {{ t("guest_invite.none") }}
                                </p>

                                <ul v-else class="mt-2 divide-y divide-gray-100">
                                    <li
                                        v-for="link in invites"
                                        :key="link.id"
                                        class="flex items-center justify-between gap-3 py-2"
                                    >
                                        <div class="min-w-0">
                                            <p class="truncate text-sm text-gray-900">
                                                {{ link.invited_email }}
                                            </p>
                                            <p class="text-xs text-gray-400">
                                                {{ expiresInLabel(link.expires_at) }}
                                                ·
                                                {{
                                                    link.used_at
                                                        ? t("guest_invite.used")
                                                        : t("guest_invite.unused")
                                                }}
                                            </p>
                                        </div>
                                        <div class="flex shrink-0 items-center">
                                            <button
                                                type="button"
                                                :title="
                                                    copiedId === link.id
                                                        ? t('guest_invite.copied')
                                                        : t('guest_invite.copy')
                                                "
                                                @click="copyLink(link)"
                                                class="rounded-md p-1.5 text-gray-400 transition-colors hover:bg-gray-100 hover:text-gray-600"
                                                :class="
                                                    copiedId === link.id ? 'text-green-600' : ''
                                                "
                                            >
                                                <CheckIcon
                                                    v-if="copiedId === link.id"
                                                    class="h-4 w-4"
                                                />
                                                <ClipboardDocumentIcon v-else class="h-4 w-4" />
                                            </button>
                                            <button
                                                type="button"
                                                :title="t('guest_invite.revoke')"
                                                :disabled="revoking === link.id"
                                                @click="revoke(link)"
                                                class="rounded-md p-1.5 text-gray-400 transition-colors hover:bg-red-50 hover:text-red-600 disabled:opacity-50"
                                            >
                                                <TrashIcon class="h-4 w-4" />
                                            </button>
                                        </div>
                                    </li>
                                </ul>
                            </div>

                            <div
                                class="flex flex-row-reverse gap-3 border-t border-gray-200 px-6 py-4"
                            >
                                <button
                                    type="button"
                                    @click="sendInvites"
                                    :disabled="(!emails.length && !pickedUsers.length) || sending"
                                    class="inline-flex items-center justify-center gap-2 rounded-md bg-indigo-600 px-4 py-2 text-sm font-semibold text-white shadow-sm hover:bg-indigo-500 disabled:cursor-not-allowed disabled:opacity-50"
                                >
                                    {{ t("guest_invite.send") }}
                                </button>
                                <button
                                    type="button"
                                    @click="emit('close')"
                                    class="inline-flex items-center justify-center rounded-md bg-white px-4 py-2 text-sm font-semibold text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 hover:bg-gray-50"
                                >
                                    {{ t("guest_invite.cancel") }}
                                </button>
                            </div>
                        </DialogPanel>
                    </TransitionChild>
                </div>
            </div>
        </Dialog>
    </TransitionRoot>
</template>

<script setup>
import { ref, onMounted } from "vue";
import {
    Dialog,
    DialogPanel,
    DialogTitle,
    TransitionChild,
    TransitionRoot,
    Listbox,
    ListboxButton,
    ListboxOptions,
    ListboxOption,
} from "@headlessui/vue";
import {
    UserPlusIcon,
    XMarkIcon,
    TrashIcon,
    ClipboardDocumentIcon,
} from "@heroicons/vue/24/outline";
import { ChevronUpDownIcon, CheckIcon } from "@heroicons/vue/20/solid";
import { t } from "@/i18n";
import chatService from "@/services/chatService";
import userService from "@/services/userService";
import groupService from "@/services/groupService";

const props = defineProps({
    channelId: {
        type: String,
        required: true,
    },
    meetingId: {
        type: String,
        required: true,
    },
});

const emit = defineEmits(["close"]);

const EXPIRY_OPTIONS = [
    { value: 1, labelKey: "guest_invite.expiry_1h" },
    { value: 4, labelKey: "guest_invite.expiry_4h" },
    { value: 24, labelKey: "guest_invite.expiry_24h" },
    { value: 168, labelKey: "guest_invite.expiry_7d" },
];

const emails = ref([]);
const draft = ref("");
const sending = ref(false);
const error = ref("");
const success = ref("");
const selectedExpiry = ref(EXPIRY_OPTIONS[1]);
const invites = ref([]);
const revoking = ref("");
const copiedId = ref("");

// Org-user invite (search seam shared with the schedule dialog).
const userQuery = ref("");
const userResults = ref([]);
const pickedUsers = ref([]);
let userSearchTimer = null;

function onUserQuery() {
    clearTimeout(userSearchTimer);
    const q = userQuery.value.trim();

    if (!q) {
        userResults.value = [];

        return;
    }

    userSearchTimer = setTimeout(async () => {
        const picked = new Set(pickedUsers.value.map((u) => u.id));
        const { data: results } = await userService.search(q);

        userResults.value = results
            .filter((u) => !picked.has(u.id))
            .map((u) => ({
                id: u.id,
                name: `${u.name} ${u.lastname}`.trim() || u.username,
            }));
    }, 250);
}

function addUser(user) {
    if (!pickedUsers.value.some((u) => u.id === user.id)) {
        pickedUsers.value.push(user);
    }

    userQuery.value = "";
    userResults.value = [];
}

function removeUser(id) {
    pickedUsers.value = pickedUsers.value.filter((u) => u.id !== id);
}

const groupQuery = ref("");
const groupResults = ref([]);
const pickedGroups = ref([]);
let groupSearchTimer = null;

function onGroupQuery() {
    clearTimeout(groupSearchTimer);
    const q = groupQuery.value.trim().toLowerCase();

    groupSearchTimer = setTimeout(async () => {
        try {
            const { data } = await groupService.search(q, 20);

            if (groupQuery.value.trim().toLowerCase() !== q) return;

            const pickedIds = new Set(pickedGroups.value.map((g) => g.id));

            groupResults.value = (data ?? [])
                .filter((g) => !pickedIds.has(g.id))
                .map((g) => ({ id: g.id, name: g.name }));
        } catch {
            groupResults.value = [];
        }
    }, 250);
}

function addGroup(group) {
    if (!pickedGroups.value.some((g) => g.id === group.id)) {
        pickedGroups.value.push(group);
    }

    groupQuery.value = "";
    groupResults.value = [];
}

function removeGroup(id) {
    pickedGroups.value = pickedGroups.value.filter((g) => g.id !== id);
}

const EMAIL_REGEX = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;

function expiresInLabel(expiresAt) {
    const seconds = expiresAt - Math.floor(Date.now() / 1000);

    if (seconds <= 0) return t.value("guest_invite.expired");

    const hours = Math.round(seconds / 3600);

    if (hours < 1) {
        const minutes = Math.max(1, Math.round(seconds / 60));

        return t.value("guest_invite.expires_in_minutes", { count: minutes });
    }

    if (hours < 48) {
        return t.value("guest_invite.expires_in_hours", { count: hours });
    }

    return t.value("guest_invite.expires_in_days", {
        count: Math.round(hours / 24),
    });
}

async function loadInvites() {
    try {
        const { data } = await chatService.getGuestInvites(props.channelId, props.meetingId);

        invites.value = data || [];
    } catch {
        invites.value = [];
    }
}

async function revoke(link) {
    revoking.value = link.id;
    try {
        await chatService.revokeGuestInvite(props.channelId, props.meetingId, link.id);
        invites.value = invites.value.filter((l) => l.id !== link.id);
    } catch {
        error.value = t.value("guest_invite.revoke_error");
    } finally {
        revoking.value = "";
    }
}

function commitDraft() {
    const value = draft.value.trim().replace(/,$/, "").trim();

    if (!value) {
        draft.value = "";

        return;
    }

    if (!EMAIL_REGEX.test(value)) {
        error.value = t.value("guest_invite.invalid_email");

        return;
    }

    if (!emails.value.includes(value)) {
        emails.value.push(value);
    }

    draft.value = "";
    error.value = "";
}

function removeEmail(email) {
    emails.value = emails.value.filter((e) => e !== email);
}

async function sendInvites() {
    commitDraft();
    if (error.value) return;
    if (!emails.value.length && !pickedUsers.value.length && !pickedGroups.value.length) return;

    sending.value = true;
    try {
        if (pickedUsers.value.length) {
            await chatService.inviteMembers(
                props.channelId,
                props.meetingId,
                pickedUsers.value.map((u) => u.id),
            );
            pickedUsers.value = [];
        }

        if (pickedGroups.value.length) {
            await chatService.addMeetingGroups(
                props.channelId,
                props.meetingId,
                pickedGroups.value.map((g) => g.id),
            );
            pickedGroups.value = [];
        }

        if (emails.value.length) {
            await chatService.inviteGuests(
                props.channelId,
                props.meetingId,
                emails.value,
                selectedExpiry.value.value,
            );
            emails.value = [];
            draft.value = "";
            await loadInvites();
        }

        success.value = t.value("guest_invite.success");
    } catch {
        error.value = t.value("guest_invite.error");
    } finally {
        sending.value = false;
    }
}

async function copyLink(link) {
    const url = `${window.location.origin}/guest/video/${link.id}`;

    try {
        await navigator.clipboard.writeText(url);
        copiedId.value = link.id;
        setTimeout(() => {
            if (copiedId.value === link.id) copiedId.value = "";
        }, 2000);
    } catch {
        error.value = t.value("guest_invite.copy_error");
    }
}

onMounted(loadInvites);
</script>
