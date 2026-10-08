<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div class="sticky bottom-0 bg-white px-4 py-4">
        <div class="flex items-start gap-x-3">
            <div class="mt-0.5 size-6 shrink-0">
                <UserAvatar :user="currentUser" />
            </div>
            <form @submit.prevent="addComment" class="relative min-w-0 flex-1">
                <div
                    class="border-b border-gray-200 pb-px focus-within:border-b-2 focus-within:border-indigo-600 focus-within:pb-0"
                >
                    <label for="comment" class="sr-only">{{
                        t("projects.task_comments.add_your_comment")
                    }}</label>
                    <textarea
                        rows="3"
                        name="comment"
                        id="comment"
                        ref="commentInput"
                        v-model="newComment"
                        class="block w-full resize-none border-0 bg-transparent p-0 text-base text-gray-900 placeholder:text-gray-400 focus:ring-0 sm:text-sm/6"
                        :placeholder="t('projects.task_comments.added_a_comment')"
                        required
                        @input="detectMention"
                        @keydown.down="moveInMentions($event, 1)"
                        @keydown.up="moveInMentions($event, -1)"
                        @keydown.enter.exact.prevent="handleEnterKey"
                        @keydown.shift.enter.prevent="insertNewline"
                    />
                </div>

                <ul
                    v-if="mentionDropdownVisible"
                    ref="mentionListEl"
                    role="listbox"
                    class="absolute left-0 z-50 max-h-72 w-96 max-w-full overflow-auto rounded-md bg-white py-1 text-sm shadow-lg ring-1 ring-black/5 focus:outline-none"
                    :class="mentionDropdownPosition.top ? 'mt-2' : 'mb-2'"
                    :style="mentionDropdownPosition"
                >
                    <template v-if="matchingSpecials.length">
                        <li
                            role="presentation"
                            class="bg-gray-100 px-3 py-2 text-xs font-semibold text-gray-900"
                        >
                            {{ t("projects.task_comments.special_mentions") }}
                        </li>
                        <li
                            v-for="(name, index) in matchingSpecials"
                            :key="name"
                            role="option"
                            :aria-selected="mentionSelectedIndex === index"
                            class="flex cursor-pointer select-none items-center gap-x-3 px-3 py-2"
                            :class="
                                mentionSelectedIndex === index
                                    ? 'bg-indigo-600 text-white'
                                    : 'text-gray-900'
                            "
                            @mouseenter="mentions.selected.value = index"
                            @click="mentions.insert(name)"
                        >
                            <span
                                class="flex size-6 flex-none items-center justify-center rounded-full text-xs font-semibold"
                                :class="
                                    mentionSelectedIndex === index
                                        ? 'bg-indigo-500 text-white'
                                        : 'bg-indigo-50 text-indigo-600'
                                "
                                >@</span
                            >
                            <span class="truncate font-medium">@{{ name }}</span>
                            <span
                                class="ml-auto truncate text-xs"
                                :class="
                                    mentionSelectedIndex === index
                                        ? 'text-indigo-200'
                                        : 'text-gray-500'
                                "
                            >
                                {{ t("projects.task_comments.mention_all_members") }}
                            </span>
                        </li>
                    </template>

                    <li
                        role="presentation"
                        class="bg-gray-100 px-3 py-2 text-xs font-semibold text-gray-900"
                    >
                        {{ t("projects.task_comments.project_members") }}
                    </li>
                    <li
                        v-for="(user, index) in mentionSearchResults"
                        :key="user.id"
                        role="option"
                        :aria-selected="mentionSelectedIndex === matchingSpecials.length + index"
                        class="flex cursor-pointer select-none items-center gap-x-3 px-3 py-2"
                        :class="
                            mentionSelectedIndex === matchingSpecials.length + index
                                ? 'bg-indigo-600 text-white'
                                : 'text-gray-900'
                        "
                        @mouseenter="mentions.selected.value = matchingSpecials.length + index"
                        @click="mentions.insert(user.username)"
                    >
                        <div class="size-6 flex-none">
                            <UserAvatar :user="user" />
                        </div>
                        <span class="truncate font-medium"
                            >{{ user.name }} {{ user.lastname }}</span
                        >
                        <span
                            class="truncate"
                            :class="
                                mentionSelectedIndex === matchingSpecials.length + index
                                    ? 'text-indigo-200'
                                    : 'text-gray-500'
                            "
                            >@{{ user.username }}</span
                        >
                    </li>
                    <li
                        v-if="
                            mentionSearchResults.length === 0 && mentionQuery && !mentionPageLoading
                        "
                        role="presentation"
                        class="px-3 py-2 text-gray-500"
                    >
                        {{ t("projects.task_comments.no_matching_members") }}
                    </li>
                    <li ref="mentionSentinel" role="presentation" class="h-px" />
                    <li
                        v-if="mentionPageLoading"
                        role="presentation"
                        class="px-3 py-1.5 text-center text-xs text-gray-500"
                    >
                        {{ t("projects.task_comments.loading_members") }}
                    </li>
                </ul>

                <div class="flex items-center justify-between pt-2">
                    <div class="flex items-center gap-x-1">
                        <button
                            type="button"
                            class="flex size-8 items-center justify-center rounded-md text-gray-400 hover:bg-gray-100 hover:text-gray-600"
                            :title="t('projects.task_comments.mention_someone')"
                            @click="startMention"
                        >
                            <AtSymbolIcon class="size-5" aria-hidden="true" />
                            <span class="sr-only">{{
                                t("projects.task_comments.mention_someone")
                            }}</span>
                        </button>
                        <ExpressionPicker disable-gifs @select-emoji="insertAtCaret" />
                    </div>
                    <div class="flex items-center gap-x-3">
                        <span
                            v-if="showCommentCounter"
                            class="text-xs tabular-nums"
                            :class="
                                isOverCommentLimit ? 'font-medium text-red-600' : 'text-gray-500'
                            "
                        >
                            {{ commentRuneCount }} / {{ MAX_COMMENT_RUNES }}
                        </span>
                        <button
                            type="submit"
                            :disabled="isOverCommentLimit"
                            class="inline-flex items-center rounded-md bg-indigo-600 px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-indigo-500 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600 disabled:cursor-not-allowed disabled:opacity-50"
                        >
                            {{ t("projects.task_comments.comment") }}
                        </button>
                    </div>
                </div>
            </form>
        </div>
    </div>
</template>

<script setup>
import { t } from "@/i18n/index.js";
import { ref, onMounted, onBeforeUnmount, computed, watch, nextTick } from "vue";
import UserAvatar from "@/components/UserAvatar.vue";
import ExpressionPicker from "@/components/ExpressionPicker/ExpressionPicker.vue";
import { AtSymbolIcon } from "@heroicons/vue/24/outline";
import { useUserStore } from "@/store/user";
import { useWorkspaceStore } from "@/store/workspaces";
import workspaceService from "@/services/workspaceService";
import { useRoute } from "vue-router";
import { useAlertStore } from "@/store/alerts";
import { extractErrorMessage } from "@/utils/errors";
import { useMentionAutocomplete } from "@/composables/useMentionAutocomplete";
import { fromComposerText } from "@/utils/mentions";
import { toCommentRow } from "@/utils/projects/taskComment";

const emit = defineEmits(["added"]);

const route = useRoute();
const workspaceStore = useWorkspaceStore();
const userStore = useUserStore();
const currentUser = userStore.user;
const newComment = ref("");
const commentInput = ref(null);

const mentions = useMentionAutocomplete({
    text: newComment,
    input: () => commentInput.value,
    specials: ["all"],
    specialsFirst: true,
    search: async (query, offset) => {
        if (!route.params.id) return [];
        const res = await workspaceService.searchWorkspaceMembers(
            route.params.id,
            query,
            20,
            offset,
        );
        const batch = res.data ?? [];

        userStore.addUsers(batch);

        return batch;
    },
});
const {
    visible: mentionDropdownVisible,
    query: mentionQuery,
    results: mentionSearchResults,
    selected: mentionSelectedIndex,
    loading: mentionPageLoading,
    matchingSpecials,
} = mentions;

onMounted(() => {
    // The first page of members is read up front, so the mentions in what is
    // typed can be resolved before the list is ever opened.
    mentions.reset();
});

const activityDeps = {
    getUser: userStore.getUserById,
};

const mentionDropdownPosition = ref({ bottom: "100%", left: "0px" });
const mentionListEl = ref(null);
const mentionSentinel = ref(null);
let mentionObserver = null;

// The list reads the next page of members when its end scrolls into view.
async function setupMentionObserver() {
    await nextTick();
    await nextTick(); // double tick ensures v-if DOM is ready
    if (!mentionSentinel.value || !mentionListEl.value) return;
    if (mentionObserver) mentionObserver.disconnect();
    mentionObserver = new IntersectionObserver(
        (e) => {
            if (e[0].isIntersecting) mentions.loadMore();
        },
        { root: mentionListEl.value, threshold: 0.1 },
    );
    mentionObserver.observe(mentionSentinel.value);
}

watch(mentionSearchResults, setupMentionObserver);
watch(mentionDropdownVisible, (v) => {
    if (v) setupMentionObserver();
});
onBeforeUnmount(() => mentionObserver?.disconnect());

const adjustMentionDropdownPosition = (event) => {
    const textarea = event.target;
    const rect = textarea.getBoundingClientRect();
    const dropdownHeight = 150;

    if (window.innerHeight - rect.bottom < dropdownHeight) {
        mentionDropdownPosition.value = { bottom: "100%", left: "0px" };
    } else {
        mentionDropdownPosition.value = { top: "100%", left: "0px" };
    }
};

const mentionedUserIds = ref([]);

const detectMention = async (event) => {
    mentions.onInput();
    if (mentionDropdownVisible.value) {
        await nextTick();
        adjustMentionDropdownPosition(event);
    }
};

const extractMentionedUserIds = () => {
    mentionedUserIds.value = [];

    const mentionMatches = newComment.value.match(/@([\p{L}\p{M}0-9_-]+)/gu);

    if (mentionMatches) {
        mentionMatches.forEach((mention) => {
            const id = resolveUsername(mention.slice(1));

            if (id && !mentionedUserIds.value.includes(id)) {
                mentionedUserIds.value.push(id);
            }
        });
    }
};

// The mention dropdown searches users the store does not hold, so the store's
// own getter would miss exactly the handles most likely to be mentioned.
const idsByUsername = computed(() => {
    const ids = new Map();

    for (const u of Object.values(userStore.usersMap)) {
        if (u?.username) ids.set(u.username, u.id);
    }

    return ids;
});

const resolveUsername = (handle) => {
    if (!handle || handle === "all" || handle === "here") return null;

    const known = idsByUsername.value.get(handle);

    if (known) return known;
    const match = mentionSearchResults.value.find((u) => u?.username === handle);

    return match ? match.id : null;
};

// The counter measures this, not the textarea: an id is far longer than the
// handle shown in its place.
const storedComment = computed(() => fromComposerText(newComment.value, [], resolveUsername));

// Must match the backend cap (workspace_comments.content TEXT, utf8mb4 worst case).
const MAX_COMMENT_RUNES = 16383;
const commentRuneCount = computed(() => [...storedComment.value].length);
const isOverCommentLimit = computed(() => commentRuneCount.value > MAX_COMMENT_RUNES);
const showCommentCounter = computed(() => commentRuneCount.value > MAX_COMMENT_RUNES - 500);

const addComment = async () => {
    if (!newComment.value.trim() || isOverCommentLimit.value) return;

    mentionedUserIds.value = [];

    const workspaces = Array.isArray(workspaceStore.workspaces) ? workspaceStore.workspaces : [];

    const currentWorkspace =
        workspaces.find((ws) => String(ws.id) === String(route.params.id)) || null;

    const members = Array.isArray(currentWorkspace?.members) ? currentWorkspace.members : [];

    const memberUserIds = new Set(members.map((m) => m.user_id));

    if (newComment.value.includes("@all")) {
        mentionedUserIds.value = Array.from(memberUserIds);
    } else {
        extractMentionedUserIds();
    }

    // VALIDATION: Check mentioned users are workspace members.
    // Use usersMap (includes all search-loaded users) + mentionSearchResults as sources.
    const mentionMatches = newComment.value.match(/@([\p{L}\p{M}0-9_-]+)/gu) || [];
    const invalidMentions = [];
    const usernameToUser = new Map();

    for (const u of mentionSearchResults.value) usernameToUser.set(u.username, u);
    for (const u of Object.values(userStore.usersMap)) {
        if (u?.username) usernameToUser.set(u.username, u);
    }

    mentionMatches.forEach((mention) => {
        const username = mention.slice(1);

        if (username === "all") return;
        // usernameToUser is built from searchWorkspaceMembers results, if found, they ARE a member
        if (!usernameToUser.has(username)) {
            invalidMentions.push(`@${username}`);
        }
    });

    if (invalidMentions.length > 0) {
        useAlertStore().showError(
            `Cannot mention users who are not workspace members: ${invalidMentions.join(", ")}`,
        );

        return;
    }

    const selectedItem = workspaceStore.getSelectedItem;
    const data = {
        id: selectedItem.id,
        comment: storedComment.value,
        table_id: route.params.tid,
        workspace_id: route.params.id,
        task_id: selectedItem.id,
        mentioned_users: mentionedUserIds.value,
    };

    try {
        const response = await workspaceService.addTaskComment(data);

        if (response.status !== 200) {
            throw new Error("Network response was not ok");
        }

        workspaceStore.appendActivity({
            ...toCommentRow(response.data, activityDeps),
            person: {
                name: t.value("projects.task_comments.you"),
            },
            date: t.value("projects.task_comments.just_now"),
        });

        newComment.value = "";
        mentionedUserIds.value = [];

        emit("added");
    } catch (error) {
        useAlertStore().showError(extractErrorMessage(error, "Failed to add comment"));
    }
};

const handleEnterKey = () => {
    if (mentionDropdownVisible.value) {
        mentions.chooseSelected();
    } else {
        addComment();
    }
};

// Up and down move through the mention list while it is open, and through
// the comment's lines otherwise.
const moveInMentions = (event, direction) => {
    if (!mentionDropdownVisible.value) return;
    event.preventDefault();
    mentions.move(direction);
};

const insertAtCaret = async (text) => {
    const el = commentInput.value;
    const val = newComment.value ?? "";
    const start = el?.selectionStart ?? val.length;
    const end = el?.selectionEnd ?? val.length;

    newComment.value = val.slice(0, start) + text + val.slice(end);

    await nextTick();
    el?.focus();
    el?.setSelectionRange(start + text.length, start + text.length);
};

// The mention list opens only for an @ that starts a word.
const startMention = async () => {
    const el = commentInput.value;
    const before = (newComment.value ?? "").slice(0, el?.selectionStart ?? newComment.value.length);

    await insertAtCaret(before === "" || /\s$/.test(before) ? "@" : " @");
    mentions.onInput();
};

const insertNewline = (e) => {
    const el = e.target;
    const start = el.selectionStart ?? 0;
    const end = el.selectionEnd ?? 0;
    const val = newComment.value ?? "";

    newComment.value = val.slice(0, start) + "\n" + val.slice(end);

    nextTick(() => {
        el.selectionStart = el.selectionEnd = start + 1;
    });
};
</script>
