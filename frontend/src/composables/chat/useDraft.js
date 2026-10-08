// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

const DRAFT_KEY_PREFIX = "chat_draft_";

export function useDraft() {
    function saveDraft(channelId, text) {
        if (!channelId) return;
        if (text) {
            localStorage.setItem(`${DRAFT_KEY_PREFIX}${channelId}`, text);
        } else {
            localStorage.removeItem(`${DRAFT_KEY_PREFIX}${channelId}`);
        }
    }

    function loadDraft(channelId) {
        if (!channelId) return "";

        return localStorage.getItem(`${DRAFT_KEY_PREFIX}${channelId}`) ?? "";
    }

    function clearDraft(channelId) {
        if (!channelId) return;
        localStorage.removeItem(`${DRAFT_KEY_PREFIX}${channelId}`);
    }

    return { saveDraft, loadDraft, clearDraft };
}
