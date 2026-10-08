<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <TransitionRoot as="template" :show="true">
        <Dialog class="relative z-50" @close="emit('close')">
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
                            class="relative w-full max-w-md transform overflow-hidden rounded-xl bg-white shadow-xl transition-all"
                        >
                            <div class="border-b border-gray-200 px-6 py-5">
                                <DialogTitle
                                    as="h3"
                                    class="text-base font-semibold leading-6 text-gray-900"
                                >
                                    {{ dialogTitle }}
                                </DialogTitle>
                            </div>

                            <form class="space-y-4 px-6 py-5" @submit.prevent="submit">
                                <div>
                                    <label class="block text-sm font-medium text-gray-700">
                                        {{ t("meetings.schedule.name_label") }}
                                    </label>
                                    <input
                                        v-model="title"
                                        type="text"
                                        maxlength="120"
                                        :placeholder="t('meetings.schedule.name_placeholder')"
                                        class="mt-1 block w-full rounded-md border-0 py-2 px-3 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm"
                                    />
                                </div>

                                <div v-if="!isInstant">
                                    <label class="block text-sm font-medium text-gray-700">
                                        {{ t("meetings.schedule.time_label") }}
                                    </label>
                                    <MeetingDateTimeField v-model="when" class="mt-1" />
                                </div>

                                <div v-if="!isInstant">
                                    <label class="block text-sm font-medium text-gray-700">
                                        {{ t("meetings.schedule.duration_label") }}
                                    </label>
                                    <Listbox v-model="selectedDuration">
                                        <div class="relative mt-1">
                                            <ListboxButton
                                                class="relative w-full cursor-default rounded-md bg-white py-2 pl-3 pr-10 text-left text-sm shadow-sm ring-1 ring-inset ring-gray-300 focus:outline-none focus:ring-2 focus:ring-indigo-600"
                                            >
                                                <span class="block truncate">
                                                    {{ t(selectedDuration.labelKey) }}
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
                                                class="absolute z-10 mt-1 max-h-60 w-full overflow-auto rounded-md bg-white py-1 text-sm shadow-lg ring-1 ring-black/5 focus:outline-none"
                                            >
                                                <ListboxOption
                                                    v-for="option in DURATION_OPTIONS"
                                                    :key="option.value"
                                                    :value="option"
                                                    v-slot="{ active, selected }"
                                                    as="template"
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
                                                                'absolute inset-y-0 right-0 flex items-center pr-3',
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

                                    <div
                                        v-if="selectedDuration.value === 'custom'"
                                        class="mt-2 flex items-center gap-2"
                                    >
                                        <input
                                            v-model="customMinutes"
                                            type="number"
                                            min="1"
                                            class="w-24 rounded-md border-0 py-2 px-3 text-sm text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 focus:ring-2 focus:ring-inset focus:ring-indigo-600"
                                        />
                                        <span class="text-sm text-gray-500">
                                            {{ t("meetings.schedule.duration_minutes") }}
                                        </span>
                                    </div>
                                </div>

                                <div>
                                    <label class="block text-sm font-medium text-gray-700">
                                        {{ t("meetings.schedule.members_label") }}
                                    </label>
                                    <p class="text-xs text-gray-400">
                                        {{ t("meetings.schedule.members_hint") }}
                                    </p>

                                    <p v-if="membersLoading" class="mt-1 text-sm text-gray-400">
                                        {{ t("meetings.schedule.members_loading") }}
                                    </p>
                                    <p v-else-if="membersError" class="mt-1 text-sm text-red-600">
                                        {{ t("meetings.schedule.members_error") }}
                                    </p>
                                    <p
                                        v-else-if="!members.length"
                                        class="mt-1 text-sm text-gray-400"
                                    >
                                        {{ t("meetings.schedule.members_empty") }}
                                    </p>
                                    <div
                                        v-else
                                        class="mt-1 max-h-40 overflow-y-auto rounded-md border border-gray-300 divide-y divide-gray-200"
                                    >
                                        <label
                                            v-for="member in members"
                                            :key="member.id"
                                            class="flex cursor-pointer items-center gap-2 px-3 py-2 text-sm text-gray-700 hover:bg-gray-50"
                                        >
                                            <input
                                                type="checkbox"
                                                :value="member.id"
                                                v-model="selectedInvitees"
                                                class="h-4 w-4 rounded border-gray-300 text-indigo-600 focus:ring-indigo-600"
                                            />
                                            {{ member.name }}
                                        </label>
                                    </div>
                                </div>

                                <div>
                                    <label class="block text-sm font-medium text-gray-700">
                                        {{ t("meetings.schedule.people_label") }}
                                    </label>
                                    <p class="text-xs text-gray-400">
                                        {{ t("meetings.schedule.people_hint") }}
                                    </p>
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

                                <div>
                                    <label class="block text-sm font-medium text-gray-700">
                                        {{ t("meetings.schedule.groups_label") }}
                                    </label>
                                    <p class="text-xs text-gray-400">
                                        {{ t("meetings.schedule.groups_hint") }}
                                    </p>
                                    <div class="relative mt-1">
                                        <input
                                            v-model="groupQuery"
                                            type="text"
                                            :placeholder="t('meetings.schedule.groups_placeholder')"
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

                                <div>
                                    <label class="block text-sm font-medium text-gray-700">
                                        {{ t("meetings.schedule.invite_label") }}
                                    </label>
                                    <div
                                        class="mt-1 flex flex-wrap gap-2 rounded-md p-2 ring-1 ring-inset focus-within:ring-2"
                                        :class="
                                            emailError
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
                                            v-model="emailDraft"
                                            type="email"
                                            :placeholder="t('meetings.schedule.invite_placeholder')"
                                            class="min-w-[8rem] flex-1 border-0 p-1 text-sm focus:outline-none focus:ring-0"
                                            @keydown.enter.prevent="commitEmail"
                                            @keydown="
                                                (e) =>
                                                    e.key === ',' &&
                                                    (e.preventDefault(), commitEmail())
                                            "
                                            @keydown.space.prevent="commitEmail"
                                            @input="emailError = ''"
                                            @blur="commitEmail"
                                        />
                                    </div>
                                    <p v-if="emailError" class="mt-1 text-xs text-red-600">
                                        {{ emailError }}
                                    </p>
                                </div>

                                <div>
                                    <label class="block text-sm font-medium text-gray-700">
                                        {{ t("meetings.schedule.password_label") }}
                                    </label>
                                    <input
                                        v-model="guestPassword"
                                        type="text"
                                        autocomplete="off"
                                        :placeholder="t('meetings.schedule.password_placeholder')"
                                        class="mt-1 block w-full rounded-md border-0 py-2 px-3 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm"
                                    />
                                    <p class="text-xs text-gray-400">
                                        {{ t("meetings.schedule.password_hint") }}
                                    </p>
                                </div>

                                <p v-if="error" class="text-xs text-red-600">
                                    {{ error }}
                                </p>

                                <div class="flex flex-row-reverse gap-3 pt-2">
                                    <button
                                        type="submit"
                                        :disabled="saving"
                                        class="inline-flex items-center justify-center rounded-md bg-indigo-600 px-4 py-2 text-sm font-semibold text-white hover:bg-indigo-500 disabled:opacity-50"
                                    >
                                        {{ submitLabel }}
                                    </button>
                                    <button
                                        type="button"
                                        @click="emit('close')"
                                        class="inline-flex items-center justify-center rounded-md bg-white px-4 py-2 text-sm font-semibold text-gray-900 ring-1 ring-inset ring-gray-300 hover:bg-gray-50"
                                    >
                                        {{ t("common.button.cancel") }}
                                    </button>
                                </div>
                            </form>
                        </DialogPanel>
                    </TransitionChild>
                </div>
            </div>
        </Dialog>
    </TransitionRoot>
</template>

<script setup>
import { ref, computed, onMounted } from "vue";
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
import { XMarkIcon } from "@heroicons/vue/24/outline";
import { ChevronUpDownIcon, CheckIcon } from "@heroicons/vue/20/solid";
import MeetingDateTimeField from "@/components/Chat/Meetings/MeetingDateTimeField.vue";
import { t } from "@/i18n";
import chatService from "@/services/chatService";
import userService from "@/services/userService";
import { useAlertStore } from "@/store/alerts";
import { useUserStore } from "@/store/user";
import { useMeetingsStore } from "@/store/meetings";
import groupService from "@/services/groupService";

const props = defineProps({
    channelId: {
        type: String,
        required: true,
    },
    mode: {
        type: String,
        default: "schedule",
        validator: (v) => ["schedule", "instant", "edit"].includes(v),
    },
    // The meeting being edited; required when mode is "edit".
    meeting: {
        type: Object,
        default: null,
    },
});

// Only "started" (instant meeting) is joined automatically by the caller.
const emit = defineEmits(["started", "created", "updated", "close"]);

const alertStore = useAlertStore();
const userStore = useUserStore();
const meetingsStore = useMeetingsStore();

const SCHEDULED_INVITE_EXPIRY_HOURS = 24;
const EMAIL_REGEX = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;

const isInstant = computed(() => props.mode === "instant");
const isEdit = computed(() => props.mode === "edit");

const dialogTitle = computed(() => {
    if (isInstant.value) return t.value("meetings.start.title");
    if (isEdit.value) return t.value("meetings.edit.title");

    return t.value("meetings.schedule.title");
});

const submitLabel = computed(() => {
    if (isInstant.value) return t.value("meetings.start.submit");
    if (isEdit.value) return t.value("meetings.edit.submit");

    return t.value("meetings.schedule.submit");
});

const DURATION_OPTIONS = [
    { value: 30, labelKey: "meetings.schedule.duration_30m" },
    { value: 60, labelKey: "meetings.schedule.duration_1h" },
    { value: 90, labelKey: "meetings.schedule.duration_90m" },
    { value: 120, labelKey: "meetings.schedule.duration_2h" },
    { value: "custom", labelKey: "meetings.schedule.duration_custom" },
];

const title = ref("");
const when = ref("");
const selectedDuration = ref(DURATION_OPTIONS[1]);
const customMinutes = ref(60);

const resolvedDuration = computed(() =>
    selectedDuration.value.value === "custom"
        ? Math.max(1, Math.round(Number(customMinutes.value) || 0))
        : selectedDuration.value.value,
);
const emails = ref([]);
const emailDraft = ref("");
const emailError = ref("");
// Guest links the meeting already had (edit mode), diffed against `emails` on save.
const originalGuestLinks = ref([]);
const selectedInvitees = ref([]);

// Org-wide invitees from beyond this channel.
const userQuery = ref("");
const userResults = ref([]);
const pickedUsers = ref([]);
let userSearchTimer = null;

const invitees = computed(() => [
    ...new Set([...selectedInvitees.value, ...pickedUsers.value.map((u) => u.id)]),
]);

function onUserQuery() {
    clearTimeout(userSearchTimer);
    const q = userQuery.value.trim();

    if (!q) {
        userResults.value = [];

        return;
    }

    userSearchTimer = setTimeout(async () => {
        const memberIds = new Set(members.value.map((m) => m.id));
        const pickedIds = new Set(pickedUsers.value.map((u) => u.id));
        const { data: results } = await userService.search(q);

        userResults.value = results
            .filter(
                (u) => u.id !== userStore.user?.id && !memberIds.has(u.id) && !pickedIds.has(u.id),
            )
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

const guestPassword = ref("");
const members = ref([]);
const membersLoading = ref(true);
const membersError = ref(false);
const error = ref("");
const saving = ref(false);

async function loadMembers() {
    membersLoading.value = true;
    membersError.value = false;
    try {
        const { data } = await chatService.getChannelMembers(props.channelId);

        members.value = (data || [])
            .filter((m) => m.id !== userStore.user?.id)
            .map((m) => ({
                id: m.id,
                name: `${m.name} ${m.lastname}`.trim() || m.username,
            }));
    } catch {
        membersError.value = true;
        members.value = [];
    } finally {
        membersLoading.value = false;
    }
}

// Local "YYYY-MM-DDTHH:MM" value (the format DatePicker emits and `new Date(...)` parses as local time).
function toLocalInput(unixSeconds) {
    const d = new Date(unixSeconds * 1000);
    const pad = (n) => String(n).padStart(2, "0");

    return (
        `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}` +
        `T${pad(d.getHours())}:${pad(d.getMinutes())}`
    );
}

async function prefillFromMeeting() {
    const m = props.meeting;

    if (!m) return;

    title.value = m.title || "";
    if (m.scheduled_at) {
        when.value = toLocalInput(m.scheduled_at);
    }

    const match = DURATION_OPTIONS.find((o) => o.value === m.duration_minutes);

    if (match) {
        selectedDuration.value = match;
    } else {
        selectedDuration.value = DURATION_OPTIONS.find((o) => o.value === "custom");
        customMinutes.value = m.duration_minutes;
    }

    // Non-member invitees become removable chips so they're visible and not silently carried along.
    const memberIds = new Set(members.value.map((u) => u.id));
    const inviteeIds = m.invitees || [];

    selectedInvitees.value = inviteeIds.filter((id) => memberIds.has(id));
    const nonMembers = inviteeIds.filter((id) => !memberIds.has(id));

    if (nonMembers.length) {
        let byId = new Map();

        try {
            const { data: found } = await userService.byIds(nonMembers);

            byId = new Map(found.map((u) => [u.id, u]));
        } catch {
            byId = new Map();
        }

        pickedUsers.value = nonMembers.map((id) => {
            const u = byId.get(id);

            return {
                id,
                name: u ? `${u.name} ${u.lastname}`.trim() || u.username || id : id,
            };
        });
    }

    try {
        const { data } = await chatService.getGuestInvites(props.channelId, m.id);
        const links = (data || []).map((l) => ({
            id: l.id,
            email: l.invited_email,
        }));

        originalGuestLinks.value = links;
        emails.value = [...new Set(links.map((l) => l.email).filter(Boolean))];
    } catch {
        originalGuestLinks.value = [];
    }

    try {
        const { data } = await chatService.getMeetingGroups(props.channelId, m.id);

        pickedGroups.value = (data || []).map((g) => ({
            id: g.group_id,
            name: g.name,
        }));
    } catch {
        pickedGroups.value = [];
    }
}

onMounted(async () => {
    await loadMembers();
    if (isEdit.value) {
        await prefillFromMeeting();
    }
});

function commitEmail() {
    const value = emailDraft.value.trim().replace(/,$/, "").trim();

    if (!value) {
        emailDraft.value = "";

        return;
    }

    if (!EMAIL_REGEX.test(value)) {
        emailError.value = t.value("meetings.schedule.invalid_email");

        return;
    }

    if (!emails.value.includes(value)) {
        emails.value.push(value);
    }

    emailDraft.value = "";
    emailError.value = "";
}

function removeEmail(email) {
    emails.value = emails.value.filter((e) => e !== email);
}

function resolveScheduledAt() {
    const scheduledAt = Math.floor(new Date(when.value).getTime() / 1000);

    if (!when.value || Number.isNaN(scheduledAt)) {
        error.value = t.value("meetings.schedule.invalid_time");

        return null;
    }

    if (scheduledAt <= Math.floor(Date.now() / 1000)) {
        error.value = t.value("meetings.schedule.past_time");

        return null;
    }

    return scheduledAt;
}

async function submit() {
    error.value = "";
    commitEmail();
    if (emailError.value) {
        return;
    }

    if (isEdit.value) {
        await submitEdit();

        return;
    }

    let scheduledAt = 0;

    if (!isInstant.value) {
        scheduledAt = resolveScheduledAt();
        if (scheduledAt === null) {
            return;
        }
    }

    saving.value = true;
    try {
        const { data: meeting } = await chatService.createMeeting(props.channelId, {
            scheduledAt,
            title: title.value.trim(),
            durationMinutes: isInstant.value ? 0 : resolvedDuration.value,
            invitees: invitees.value,
            groups: pickedGroups.value.map((g) => g.id),
            guestPassword: guestPassword.value.trim(),
        });

        if (emails.value.length) {
            await chatService.inviteGuests(
                props.channelId,
                meeting.id,
                emails.value,
                SCHEDULED_INVITE_EXPIRY_HOURS,
            );
        }

        // The host isn't sent the notification that refreshes other clients, so refresh here.
        meetingsStore.loadMyMeetings();

        if (isInstant.value) {
            emit("started", meeting);
        } else {
            alertStore.showSuccess(t.value("meetings.schedule.success"));
            emit("created", meeting);
        }
    } catch {
        error.value = isInstant.value
            ? t.value("meetings.error.create")
            : t.value("meetings.schedule.error");
    } finally {
        saving.value = false;
    }
}

async function submitEdit() {
    const scheduledAt = resolveScheduledAt();

    if (scheduledAt === null) {
        return;
    }

    saving.value = true;
    try {
        const { data: meeting } = await chatService.updateMeeting(
            props.channelId,
            props.meeting.id,
            {
                scheduledAt,
                title: title.value.trim(),
                durationMinutes: resolvedDuration.value,
                invitees: invitees.value,
                groups: pickedGroups.value.map((g) => g.id),
                guestPassword: guestPassword.value.trim(),
            },
        );

        await syncGuestInvites(props.meeting.id);

        meetingsStore.loadMyMeetings();
        alertStore.showSuccess(t.value("meetings.edit.success"));
        emit("updated", meeting);
    } catch {
        error.value = t.value("meetings.error.edit");
    } finally {
        saving.value = false;
    }
}

async function syncGuestInvites(meetingId) {
    const originalEmails = new Set(originalGuestLinks.value.map((l) => l.email));
    const current = new Set(emails.value);

    const added = emails.value.filter((e) => !originalEmails.has(e));

    if (added.length) {
        await chatService.inviteGuests(
            props.channelId,
            meetingId,
            added,
            SCHEDULED_INVITE_EXPIRY_HOURS,
        );
    }

    const removed = originalGuestLinks.value.filter((l) => !current.has(l.email));

    await Promise.all(
        removed.map((l) => chatService.revokeGuestInvite(props.channelId, meetingId, l.id)),
    );
}
</script>
