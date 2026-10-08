// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { describe, it, expect, beforeEach, afterEach, vi } from "vitest";
import { usePointerDrag } from "@/composables/usePointerDrag";

function pointer(type, x, y) {
    return new MouseEvent(type, { clientX: x, clientY: y });
}

function press(startPress, options, extra = {}) {
    startPress({ pointerType: "mouse", button: 0, clientX: 0, clientY: 0, ...extra }, options);
}

let target;

beforeEach(() => {
    target = document.createElement("div");
    target.dataset.moveTarget = "folder-1";
    document.body.appendChild(target);
    document.elementFromPoint = vi.fn(() => target);
});

afterEach(() => {
    window.dispatchEvent(new KeyboardEvent("keydown", { key: "Escape" }));
    window.dispatchEvent(pointer("pointerup", 0, 0));
    target.remove();
});

function dragOptions(overrides = {}) {
    return {
        start: vi.fn(() => ({ item: { id: "file-1" }, label: "Report.pdf" })),
        resolveTarget: vi.fn((id) => ({ id })),
        drop: vi.fn(),
        ...overrides,
    };
}

describe("usePointerDrag", () => {
    it("keeps a short press a click", () => {
        const { startPress, dragged } = usePointerDrag();
        const options = dragOptions();

        press(startPress, options);
        window.dispatchEvent(pointer("pointermove", 2, 2));
        window.dispatchEvent(pointer("pointerup", 2, 2));

        expect(options.start).not.toHaveBeenCalled();
        expect(options.drop).not.toHaveBeenCalled();
        expect(dragged.value).toBeNull();
    });

    it("drops the dragged item on the target under the pointer", () => {
        const { startPress, dragged } = usePointerDrag();
        const options = dragOptions();

        press(startPress, options);
        window.dispatchEvent(pointer("pointermove", 20, 20));

        expect(dragged.value).toEqual({ id: "file-1" });
        expect(target.hasAttribute("data-drop-target")).toBe(true);

        window.dispatchEvent(pointer("pointerup", 20, 20));

        expect(options.drop).toHaveBeenCalledWith({ id: "file-1" }, { id: "folder-1" });
        expect(dragged.value).toBeNull();
        expect(target.hasAttribute("data-drop-target")).toBe(false);
    });

    it("marks only the element under the pointer when a folder shows twice", () => {
        const { startPress } = usePointerDrag();
        const twin = document.createElement("div");

        twin.dataset.moveTarget = "folder-1";
        document.body.appendChild(twin);

        press(startPress, dragOptions());
        window.dispatchEvent(pointer("pointermove", 20, 20));

        expect(target.hasAttribute("data-drop-target")).toBe(true);
        expect(twin.hasAttribute("data-drop-target")).toBe(false);

        document.elementFromPoint = vi.fn(() => twin);
        window.dispatchEvent(pointer("pointermove", 30, 30));

        expect(target.hasAttribute("data-drop-target")).toBe(false);
        expect(twin.hasAttribute("data-drop-target")).toBe(true);

        twin.remove();
    });

    it("does not drop on a target that resolveTarget refuses", () => {
        const { startPress } = usePointerDrag();
        const options = dragOptions({ resolveTarget: vi.fn(() => null) });

        press(startPress, options);
        window.dispatchEvent(pointer("pointermove", 20, 20));

        expect(target.hasAttribute("data-drop-target")).toBe(false);

        window.dispatchEvent(pointer("pointerup", 20, 20));

        expect(options.drop).not.toHaveBeenCalled();
    });

    it("cancels the drag on Escape", () => {
        const { startPress, dragged } = usePointerDrag();
        const options = dragOptions();

        press(startPress, options);
        window.dispatchEvent(pointer("pointermove", 20, 20));
        window.dispatchEvent(new KeyboardEvent("keydown", { key: "Escape" }));
        window.dispatchEvent(pointer("pointerup", 20, 20));

        expect(dragged.value).toBeNull();
        expect(options.drop).not.toHaveBeenCalled();
    });

    it("ignores touch so the page can scroll", () => {
        const { startPress, dragged } = usePointerDrag();
        const options = dragOptions();

        press(startPress, options, { pointerType: "touch" });
        window.dispatchEvent(pointer("pointermove", 20, 20));

        expect(options.start).not.toHaveBeenCalled();
        expect(dragged.value).toBeNull();
    });
});
