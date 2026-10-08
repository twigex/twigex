// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { describe, it, expect, vi } from "vitest";
import { handleEvent, onProjectChange } from "@/js/websocket";

vi.mock("@/services/userService", () => ({ default: {} }));
vi.mock("@/services/chatService", () => ({ default: {} }));
vi.mock("@/services/notificationService", () => ({ default: {} }));

describe("project_change", () => {
    it("reaches every listener until it stops listening", () => {
        const first = vi.fn();
        const second = vi.fn();
        const stopFirst = onProjectChange(first);
        const stopSecond = onProjectChange(second);
        const change = { type: "task_updated", workspace_id: "w1", table_id: "t1" };

        handleEvent({ event: "project_change", app: "projects", data: change });
        stopFirst();
        handleEvent({ event: "project_change", app: "projects", data: change });
        stopSecond();

        expect(first).toHaveBeenCalledTimes(1);
        expect(first).toHaveBeenCalledWith(change);
        expect(second).toHaveBeenCalledTimes(2);
    });
});
