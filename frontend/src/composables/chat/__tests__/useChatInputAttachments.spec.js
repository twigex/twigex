// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { describe, it, expect, vi, beforeEach } from "vitest";
import { useChatInputAttachments } from "@/composables/chat/useChatInputAttachments";
import chatService from "@/services/chatService";

vi.mock("vue-router", () => ({
    useRoute: () => ({ params: { chatId: "c1" } }),
}));

vi.mock("@/services/chatService", () => ({
    default: {
        uploadFile: vi.fn(),
    },
}));

function pendingUpload() {
    let resolve;
    let reject;

    chatService.uploadFile.mockReturnValueOnce(
        new Promise((res, rej) => {
            resolve = res;
            reject = rej;
        }),
    );

    return { resolve, reject };
}

const file = (name) => ({
    name,
    type: "application/pdf",
    size: 10,
});

describe("useChatInputAttachments", () => {
    beforeEach(() => {
        chatService.uploadFile.mockReset();
    });

    it("fills in the preview when the upload finishes", async () => {
        const upload = pendingUpload();
        const { filePreviews, processFiles } = useChatInputAttachments({ message: () => null });

        processFiles([file("a.pdf")]);
        upload.resolve({ data: { id: "f1", storage_id: "s1" } });
        await Promise.resolve();
        await Promise.resolve();

        expect(filePreviews.value[0].id).toBe("f1");
        expect(filePreviews.value[0].progress).toBe(100);
    });

    it("ignores an upload that finishes after its preview was removed", async () => {
        const upload = pendingUpload();
        const { filePreviews, processFiles, removeFilePreview } = useChatInputAttachments({
            message: () => null,
        });

        processFiles([file("a.pdf")]);
        const preview = filePreviews.value[0];

        removeFilePreview(0);

        upload.resolve({ data: { id: "f1", storage_id: "s1" } });
        await Promise.resolve();
        await Promise.resolve();

        expect(preview.id).toBeNull();
        expect(filePreviews.value).toEqual([]);
    });

    it("reports no error for a failed upload whose preview was cleared", async () => {
        const upload = pendingUpload();
        const { filePreviews, uploadErrors, processFiles } = useChatInputAttachments({
            message: () => null,
        });

        processFiles([file("a.pdf")]);
        filePreviews.value = [];

        upload.reject(new Error("gone"));
        await Promise.resolve();
        await Promise.resolve();

        expect(uploadErrors.value).toEqual([]);
    });
});
