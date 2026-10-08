// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { marked } from "marked";
import EmojiConvertor from "emoji-js";
import chatService from "@/services/chatService";
import { useUserStore } from "@/store/user";
import { useChannelsStore } from "@/store/channels";
import { channelRoles } from "@/constants/channels";
import { systemRoles } from "@/constants/system";
import DOMPurify from "dompurify";

const emojiConvertor = new EmojiConvertor();

emojiConvertor.init_env();
emojiConvertor.replace_mode = "css"; // Render emojis as images
emojiConvertor.img_set = "google";
emojiConvertor.img_sets.google.path = "/emoji-datasource/img/google/64/";
emojiConvertor.include_title = true;
emojiConvertor.include_text = true;

const convertEmoji = (text) => emojiConvertor.replace_colons(text);

const pendingLinkPreviewUpdates = new Map();

const CLEAN_LEADING_ZW = /^(?:\u200B|\u200C|\u200D|\u200E|\u200F|\u2060|\uFEFF)+/; // strip leading zero-widths
const SAFE_URL_PATTERN = /^(https?:\/\/)/i;

marked.setOptions({
    breaks: true,
    gfm: true,
});

function escapeHtml(text) {
    return text.replace(/&/g, "&amp;").replace(/</g, "&lt;").replace(/>/g, "&gt;");
}

// marked v13 renderers receive a single token object, not positional args.
const renderer = {
    html({ text }) {
        return text.replace(/</g, "&lt;").replace(/>/g, "&gt;");
    },

    paragraph({ tokens }) {
        return `<p class="m-0 whitespace-pre-wrap">${this.parser.parseInline(tokens)}</p>`;
    },

    link({ href, title, tokens }) {
        const text = this.parser.parseInline(tokens);

        if (!href) {
            return text;
        }

        const safeHref = href.trim();
        const isMailto = safeHref.startsWith("mailto:");

        if (!isMailto && !SAFE_URL_PATTERN.test(safeHref)) {
            return `[${text}](${href})`;
        }

        const titleAttr = title ? ` title="${title}"` : "";
        const targetAttr = isMailto ? "" : ` target="_blank" rel="noopener noreferrer"`;

        return `<a href="${safeHref}"${targetAttr}${titleAttr} class="text-blue-600 hover:underline">${text}</a>`;
    },

    code({ text, lang }) {
        // Emit a neutral placeholder; <CodeBlock> owns highlighting, the language
        // label, and its own copy button. The raw code rides through as escaped
        // text so the html->component walker can read it back verbatim.
        const safeLang = (lang || "").replace(/[^\w#+.-]/g, "");

        return `<pre class="md-code-block" data-lang="${safeLang}"><code>${escapeHtml(text)}</code></pre>`;
    },

    codespan({ text }) {
        return `<code class="font-mono text-sm bg-gray-100 border border-gray-200 text-gray-800 rounded px-1 py-0.5">${text}</code>`;
    },

    image({ href, title, text }) {
        if (!href || !SAFE_URL_PATTERN.test(href.trim())) {
            return `[${text}](${href})`;
        }

        const titleAttr = title ? ` title="${title}"` : "";

        return `<a href="${href}" target="_blank" rel="noopener noreferrer"${titleAttr} class="text-blue-600 hover:underline">${text}</a>`;
    },
};

// Obsidian-style ==highlight==; the \S lookarounds keep "a == b" comparisons from matching.
const highlightExtension = {
    name: "highlight",
    level: "inline",
    start(src) {
        return src.match(/==/)?.index;
    },
    tokenizer(src) {
        const match = /^==(?=\S)([\s\S]*?\S)==/.exec(src);

        if (match) {
            return {
                type: "highlight",
                raw: match[0],
                text: match[1],
                tokens: this.lexer.inlineTokens(match[1]),
            };
        }
    },
    renderer(token) {
        return `<mark class="bg-yellow-200 rounded px-0.5">${this.parser.parseInline(token.tokens)}</mark>`;
    },
};

marked.use({
    useNewRenderer: true,
    renderer,
    extensions: [highlightExtension],
});

function applyLinkPreview(post, linkMetadata) {
    post.metadata = post.metadata || {};
    post.metadata.links = post.metadata.links || [];

    post.metadata.links.push(linkMetadata);
}

function flushPendingLinkPreviews(post) {
    const queued = pendingLinkPreviewUpdates.get(post.id);

    if (!queued) return;

    queued.forEach((meta) => applyLinkPreview(post, meta));
    pendingLinkPreviewUpdates.delete(post.id);
}

export default function useChatOperations() {
    const channelsStore = useChannelsStore();

    const getDirectChannelName = (userStore, channel) => {
        let me = userStore.user;

        for (let i = 0; i < channel.channel_members.length; i++) {
            if (me.id != channel.channel_members[i].user_id) {
                let user = userStore.getUserById(channel.channel_members[i].user_id);

                return user ? user.name + `(${user.email})` : "";
            }
        }
    };

    // Runs inside render and as a watcher source, so it must not write:
    // assigning channel.posts here re-triggers the effect that called it.
    const getCurrentChannel = (id) => channelsStore.channels.find((c) => c.id === id);

    function preserveExtraBlankLines(input) {
        return input.replace(/\r\n/g, "\n").replace(/\n{3,}/g, (m) => {
            const extra = m.length - 2;

            return "\n\n" + "\u00A0\n\n".repeat(extra);
        });
    }

    const renderMarkdown = (text) => {
        const cleaned = text.replace(CLEAN_LEADING_ZW, "");

        const prepared = preserveExtraBlankLines(cleaned);

        const html = marked(prepared);

        const sanitized = DOMPurify.sanitize(html, {
            ADD_ATTR: ["target"],
        });

        return convertEmoji(sanitized);
    };

    // replace_colons passes anything it does not recognise through untouched,
    // so a reaction stored before it was validated still reaches the DOM.
    const renderEmoji = (text) =>
        DOMPurify.sanitize(convertEmoji(text), {
            ALLOWED_TAGS: ["span"],
            ALLOWED_ATTR: ["class", "style", "title", "data-codepoints"],
        });

    const isChannelAdmin = (channel) => {
        const userStore = useUserStore();

        if (userStore.user.role === systemRoles.Admin) {
            return true;
        }

        return channel.channel_members.some(
            (member) => member.user_id === userStore.user.id && member.role === channelRoles.Admin,
        );
    };

    const addUser = (users, channel) => {
        chatService
            .addUsersToChannel({
                id: channel.id,
                users: users.map((user) => user.id),
            })
            .then((response) => {
                const members = response.data;

                members.forEach((m) => {
                    const idx = channel.channel_members.findIndex(
                        (existing) => existing.user_id === m.user_id,
                    );

                    if (idx === -1) {
                        channel.channel_members.push(m);
                    } else {
                        channel.channel_members.splice(idx, 1, m);
                    }
                });
            });
    };

    const addGroups = (groups, channel) => {
        const ids = groups.map((g) => g.id);

        return chatService.addGroupsToChannel(channel.id, ids).then(async () => {
            try {
                const cgRes = await chatService.getChannelGroups(channel.id);

                channel.channel_groups = cgRes.data ?? [];
            } catch {}

            try {
                const res = await chatService.getChannels();
                const updated = (res.data ?? []).find((c) => c.id === channel.id);

                if (updated) {
                    channel.channel_members = updated.channel_members;
                }
            } catch {
                // Best-effort refresh; backend already accepted the attach.
            }
        });
    };

    const addPostsToChannel = (channelId, posts) => {
        const channel = channelsStore.getChannelById(channelId);

        if (!channel) return;

        if (!channel.posts) channel.posts = [];

        const list = Array.isArray(posts) ? posts : [posts];

        list.forEach((post) => {
            if (channel.posts.some((p) => p.id === post.id)) return;

            const pendingIndex = post.pending_post_id
                ? channel.posts.findIndex((p) => p.id === post.pending_post_id)
                : -1;

            if (pendingIndex !== -1) {
                channel.posts.splice(pendingIndex, 1, post);
            } else {
                channel.posts.push(post);
            }

            flushPendingLinkPreviews(post);
        });
    };

    const handleLinkPreviewUpdate = (msg) => {
        const { channel_id, post_id, link_metadata } = msg.data;
        const channel = channelsStore.getChannelById(channel_id);

        if (!channel) return;

        const post = channel.posts?.find((p) => p.id === post_id);

        if (!post) {
            if (!pendingLinkPreviewUpdates.has(post_id)) {
                pendingLinkPreviewUpdates.set(post_id, []);
            }

            pendingLinkPreviewUpdates.get(post_id).push(link_metadata);

            return;
        }

        applyLinkPreview(post, link_metadata);
    };

    return {
        getDirectChannelName,
        getCurrentChannel,
        renderMarkdown,
        renderEmoji,
        isChannelAdmin,
        addUser,
        addGroups,
        addPostsToChannel,
        handleLinkPreviewUpdate,
        applyLinkPreview,
    };
}
