<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div
        v-if="!can('send_message')"
        class="flex items-center justify-center rounded-xl border border-gray-200 bg-gray-50 px-4 py-3"
    >
        <span class="text-sm text-gray-400">{{ t("channels.input.no_permission") }}</span>
    </div>

    <div v-else class="relative w-full">
        <!-- Reply banner -->
        <ChatInputReplyBar v-if="reply" :reply="reply" @cancel="emits('cancel-reply')" />

        <!-- Main input card -->
        <div
            class="rounded-xl border border-gray-200 bg-white transition-colors focus-within:border-indigo-400 focus-within:ring-1 focus-within:ring-indigo-400"
            :class="reply ? 'rounded-t-none' : ''"
        >
            <!-- Formatting toolbar -->
            <ChatInputFormatToolbar @format="applyFormat" />

            <!-- Textarea -->
            <textarea
                v-model="newMessage"
                @keydown="handleKeydown"
                @input="handleInput"
                @paste="handlePaste"
                ref="messageInput"
                :rows="calculatedRows"
                class="block w-full resize-none border-0 bg-transparent px-3 pt-2.5 pb-1 text-sm text-gray-900 placeholder:text-gray-400 focus:outline-none focus:ring-0"
                :placeholder="t('channels.input.message_placeholder')"
            />

            <div
                v-if="showMessageCounter"
                class="px-3 pb-1 text-right text-xs"
                :class="isOverMessageLimit ? 'text-red-600' : 'text-gray-400'"
            >
                {{ messageRuneCount }} / {{ MAX_MESSAGE_RUNES }}
            </div>

            <!-- File attachment previews -->
            <ChatInputAttachments
                v-if="filePreviews.length"
                :previews="filePreviews"
                @remove="removeFilePreview"
            />

            <!-- Toolbar -->
            <div class="flex items-center justify-between px-2 pb-2">
                <!-- Left: attach + emoji -->
                <div class="flex items-center gap-x-0.5">
                    <button
                        @click="triggerFileUpload"
                        type="button"
                        class="flex h-8 w-8 items-center justify-center rounded-md text-gray-400 hover:bg-gray-100 hover:text-gray-600 transition-colors"
                    >
                        <PaperClipIcon class="size-4" aria-hidden="true" />
                    </button>
                    <ExpressionPicker
                        v-model="showExpressionPicker"
                        class="bottom-full"
                        :disable-gifs="props.message != null"
                        @select-emoji="addEmoji"
                        @select-gif="addGif"
                        @close="showExpressionPicker = false"
                    />
                </div>

                <!-- Right: send / edit actions -->
                <div class="flex items-center gap-x-1.5">
                    <!-- Edit mode -->
                    <template v-if="props.message != null">
                        <button
                            @click="((removedFiles = []), emits('cancel'))"
                            class="rounded-md px-3 py-1.5 text-sm font-medium text-gray-600 hover:bg-gray-100 transition-colors"
                        >
                            {{ t("common.button.cancel") }}
                        </button>
                        <button
                            :disabled="!canPost()"
                            @click="sendMessage"
                            class="rounded-md px-3 py-1.5 text-sm font-semibold text-white transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
                            :class="canPost() ? 'bg-indigo-600 hover:bg-indigo-500' : 'bg-gray-300'"
                        >
                            {{ t("common.button.update") }}
                        </button>
                    </template>

                    <!-- Compose mode: icon send button -->
                    <button
                        v-else
                        :disabled="!canPost()"
                        @click="sendMessage"
                        class="flex h-8 w-8 items-center justify-center rounded-md transition-colors"
                        :class="
                            canPost()
                                ? 'bg-indigo-600 hover:bg-indigo-500 text-white'
                                : 'text-gray-300 cursor-not-allowed'
                        "
                    >
                        <PaperAirplaneIcon class="size-4" aria-hidden="true" />
                    </button>
                </div>
            </div>
        </div>

        <!-- Mention dropdown -->
        <ChatInputMentionList
            v-if="showMentionDropdown && (usersInChannel.length > 0 || matchingSpecials.length > 0)"
            :users="usersInChannel"
            :specials="matchingSpecials"
            :selected="mentionSelected"
            @select="mentions.insert"
        />

        <input
            type="file"
            @change="onFileChange"
            name="files"
            ref="fileInput"
            class="hidden"
            multiple
        />
    </div>

    <ChatInputFileErrors
        :size-errors="error"
        :upload-errors="uploadErrors"
        :size-limit="fileSize"
        @dismiss-size="error = []"
        @dismiss-upload="uploadErrors = []"
    />
</template>

<script setup>
import { t } from "@/i18n/index.js";

import { ref, onMounted, onBeforeUnmount, computed, nextTick, watch } from "vue";
import { useRoute } from "vue-router";
import ExpressionPicker from "@/components/ExpressionPicker/ExpressionPicker.vue";
import ChatInputReplyBar from "@/components/Chat/Messages/ChatInputReplyBar.vue";
import ChatInputFormatToolbar from "@/components/Chat/Messages/ChatInputFormatToolbar.vue";
import ChatInputAttachments from "@/components/Chat/Messages/ChatInputAttachments.vue";
import ChatInputMentionList from "@/components/Chat/Messages/ChatInputMentionList.vue";
import ChatInputFileErrors from "@/components/Chat/Messages/ChatInputFileErrors.vue";
import { PaperClipIcon, PaperAirplaneIcon } from "@heroicons/vue/24/solid";
import chatService from "@/services/chatService";
import { toComposerText, fromComposerText } from "@/utils/mentions";
import { useMentionAutocomplete } from "@/composables/useMentionAutocomplete";
import { useChatInputAttachments } from "@/composables/chat/useChatInputAttachments";
import { useChatInputFormat } from "@/composables/chat/useChatInputFormat";
import { useUserStore } from "@/store/user";
import { usePermissions } from "@/composables/usePermissions";
import { useDraft } from "@/composables/chat/useDraft";
import { connection } from "@/js/websocket";
import TurndownService from "turndown";
import { gfm } from "turndown-plugin-gfm";

const props = defineProps({
    reply: {
        type: Object,
        default: null,
    },
    message: {
        type: Object,
        default: null,
    },
    busy: {
        type: Boolean,
        default: false,
    },
});

const emits = defineEmits(["send-message", "cancel", "update", "cancel-reply"]);

const calculatedRows = computed(() => {
    const lines = newMessage.value.split("\n").length;

    return Math.min(lines, 6);
});

// gfm plugin keeps tables, strikethrough, and task lists intact when pasting rich text.
// Fenced code blocks preserve the language label; the default indented style drops it.
var turndownService = new TurndownService({ codeBlockStyle: "fenced" });

turndownService.use(gfm);

// Round-trips highlights so pasted <mark> matches how we render ==highlight==.
turndownService.addRule("highlight", {
    filter: ["mark"],
    replacement: (content) => `==${content}==`,
});

// Copying a rendered code block would otherwise paste the Copy button's label into the code.
turndownService.addRule("stripCopyBtn", {
    filter: (node) => node.nodeName === "BUTTON" && node.classList.contains("copy-btn"),
    replacement: () => "",
});

const { saveDraft, loadDraft, clearDraft } = useDraft();

const userStore = useUserStore();
const { can } = usePermissions();
const route = useRoute();

// Maps each mention shown in the textarea back to its stored id, so an edit
// cannot drop a mention of someone who no longer has a name.
const mentionRestore = ref([]);

// Left as names: one written before ids existed may already name the wrong
// person.
const mentionExisting = ref([]);

function mentionUsername(id) {
    return props.message?.mentions?.[id]?.username || userStore.usersMap[id]?.username || null;
}

// An unresolved handle stays a name and still notifies by name.
function mentionID(handle) {
    if (handle === "all" || handle === "here") return null;

    return userStore.getUserIdByUsername(handle);
}

const newMessage = ref("");
const messageInput = ref(null);
const fileInput = ref(null);

// The counter and the limit measure this, not the textarea: an id is far
// longer than the name shown in its place.
const storedMessage = computed(() =>
    fromComposerText(newMessage.value, mentionRestore.value, mentionID, mentionExisting.value),
);

// Must match the backend cap (posts.message TEXT column, utf8mb4 worst case).
const MAX_MESSAGE_RUNES = 16383;
// Spread counts code points so an emoji costs one, matching Go's RuneCountInString.
const messageRuneCount = computed(() => [...storedMessage.value].length);
const isOverMessageLimit = computed(() => messageRuneCount.value > MAX_MESSAGE_RUNES);
const showMessageCounter = computed(() => messageRuneCount.value > MAX_MESSAGE_RUNES - 500);
// Channel members, including those reachable through a group, are found by
// the server, so the composer scales the way rendering messages does. The
// search returns one page.
const mentions = useMentionAutocomplete({
    text: newMessage,
    input: () => messageInput.value,
    pageSize: 25,
    specials: ["all", "here"],
    search: async (query, offset) => {
        if (offset > 0) return [];
        const res = await chatService.searchChannelMembers(route.params.chatId, query);
        const members = res.data || [];

        userStore.addUsers(members);

        return members;
    },
});
const {
    visible: showMentionDropdown,
    results: usersInChannel,
    selected: mentionSelected,
    matchingSpecials,
} = mentions;

const { applyFormat } = useChatInputFormat({
    text: newMessage,
    input: () => messageInput.value,
});

const {
    filePreviews,
    error,
    uploadErrors,
    removedFiles,
    fileSize,
    clearFileErrors,
    processFiles,
    removeFilePreview,
    loadMessageFiles,
} = useChatInputAttachments({ message: () => props.message });

const showExpressionPicker = ref(false);
const sent = ref(true);
const selectedGif = ref(null);

const handleKeydown = (event) => {
    // While the mention list is open, up and down move through it and Enter
    // takes the highlighted person instead of sending a half-typed mention.
    if (
        mentions.entries.value.length > 0 &&
        showMentionDropdown.value &&
        (event.key === "ArrowDown" || event.key === "ArrowUp")
    ) {
        event.preventDefault();
        mentions.move(event.key === "ArrowDown" ? 1 : -1);

        return;
    }

    if (event.key === "Enter") {
        if (event.shiftKey) {
            return;
        }

        event.preventDefault();

        if (mentions.chooseSelected()) return;
        sendMessage();

        return;
    }
};

const handleInput = () => {
    if (sent.value == true) {
        sent.value = false;

        if (connection.readyState == WebSocket.OPEN) {
            connection.send(
                JSON.stringify({
                    event: "typing",
                    data: {
                        channel_id: route.params.chatId,
                    },
                }),
            );

            setTimeout(() => {
                sent.value = true;
            }, 1000);
        }
    }

    if (!props.message) {
        saveDraft(route.params.chatId, newMessage.value);
    }

    mentions.onInput();
};

const handlePaste = (event) => {
    const clipboard = event.clipboardData;
    const html = clipboard.getData("text/html");

    const inputText = turndownService.turndown(html);

    // Handle pasted files (images, pdfs, etc.)
    if (clipboard.files && clipboard.files.length > 0) {
        event.preventDefault();
        handlePastedFiles(clipboard.files);

        return;
    }

    if (inputText) {
        event.preventDefault();
        insertAtCursor(messageInput.value, inputText);
    }
};

function insertAtCursor(el, text) {
    const start = el.selectionStart;
    const end = el.selectionEnd;
    const value = el.value;

    const newValue = value.substring(0, start) + text + value.substring(end);

    el.value = newValue;
    el.selectionStart = el.selectionEnd = start + text.length;
    newMessage.value = newValue; // update v-model
}

const addEmoji = (emoji) => {
    const inputEl = messageInput.value;
    const cursorPos = inputEl.selectionStart;

    newMessage.value =
        newMessage.value.slice(0, cursorPos) + emoji + newMessage.value.slice(cursorPos);

    showExpressionPicker.value = false;

    nextTick(() => {
        const newCursorPos = cursorPos + emoji.length;

        inputEl.focus();
        inputEl.setSelectionRange(newCursorPos, newCursorPos);
    });
};

const addGif = (gif) => {
    selectedGif.value = gif;
    sendMessage();
};

const triggerFileUpload = () => {
    fileInput.value?.click();
};

function onFileChange(event) {
    processFiles(event.target.files);
    event.target.value = "";
}

function handlePastedFiles(fileList) {
    processFiles(fileList);
}

function canPost() {
    if (props.busy || isOverMessageLimit.value) {
        return false;
    }

    for (let i = 0; i < filePreviews.value.length; i++) {
        if (!filePreviews.value[i].id) {
            return false;
        }
    }

    if (newMessage.value.length < 1 && filePreviews.value.length < 1 && !selectedGif.value) {
        return false;
    }

    return true;
}

const sendMessage = () => {
    if (!canPost()) {
        return;
    }

    mentions.close();

    if (props.message) {
        let attachments = filePreviews.value;

        if (props.message.metadata && props.message.metadata.files) {
            attachments = filePreviews.value.filter((file) => {
                return !props.message.metadata.files.some((f) => f.id === file.id);
            });
        }

        emits("update", {
            id: props.message.id,
            message: storedMessage.value,
            attachments: attachments,
            removedFiles: removedFiles.value,
            gif: selectedGif.value,
        });

        selectedGif.value = null;
        clearFileErrors();

        return;
    }

    emits("send-message", {
        message: storedMessage.value,
        reply: props.reply ? props.reply.id : "",
        attachments: filePreviews.value,
        gif: selectedGif.value,
    });

    newMessage.value = "";
    filePreviews.value = [];
    removedFiles.value = [];
    selectedGif.value = null;
    clearFileErrors();
    clearDraft(route.params.chatId);
};

watch(
    () => props.reply,
    (newReply) => {
        if (newReply) {
            nextTick(() => messageInput.value?.focus());
        }
    },
);

watch(
    () => route.params.chatId,
    (newId, oldId) => {
        if (props.message) return;
        saveDraft(oldId, newMessage.value);
        newMessage.value = loadDraft(newId);
        filePreviews.value = [];
        clearFileErrors();
    },
);

onBeforeUnmount(() => {
    if (!props.message) {
        saveDraft(route.params.chatId, newMessage.value);
    }
});

onMounted(() => {
    if (props.message) {
        const composed = toComposerText(props.message.message, mentionUsername);

        newMessage.value = composed.text;
        mentionRestore.value = composed.restore;
        mentionExisting.value = composed.existing;

        loadMessageFiles(props.message);
    } else {
        newMessage.value = loadDraft(route.params.chatId);
    }
});
</script>
