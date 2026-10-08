// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { ref } from "vue";

// Shared by every component on the page, so an item dragged from one view can
// be dropped on the sidebar or the breadcrumbs.
const dragged = ref(null);

// How far the pointer moves before a press becomes a drag, so a click still
// opens the item.
const DRAG_THRESHOLD = 5;

// Native drag and drop froze the page, as it did on the Kanban board, so a
// drag follows the pointer instead. A drop target is any element with
// data-move-target set to the target's id. Only the element under the pointer
// gets data-drop-target, because the same folder can show in several places.
let press = null;
let active = null;
let ghost = null;
let targetEl = null;

function showGhost(label, count, x, y) {
    ghost = document.createElement("div");
    ghost.textContent = label;
    Object.assign(ghost.style, {
        position: "fixed",
        left: "0",
        top: "0",
        zIndex: "70",
        pointerEvents: "none",
        maxWidth: "16rem",
        overflow: "hidden",
        textOverflow: "ellipsis",
        whiteSpace: "nowrap",
        padding: "0.375rem 0.75rem",
        borderRadius: "0.375rem",
        background: "#fff",
        boxShadow: "0 4px 12px rgb(0 0 0 / 0.15)",
        fontSize: "0.75rem",
        fontWeight: "500",
        lineHeight: "1.25rem",
        color: "#111827",
    });

    if (count > 1) {
        const badge = document.createElement("span");

        badge.textContent = count;
        Object.assign(badge.style, {
            marginLeft: "0.5rem",
            padding: "0 0.375rem",
            borderRadius: "9999px",
            background: "#4f46e5",
            color: "#fff",
            fontSize: "0.6875rem",
            fontWeight: "600",
        });
        ghost.appendChild(badge);
    }

    document.body.appendChild(ghost);
    moveGhost(x, y);
}

function moveGhost(x, y) {
    if (ghost) ghost.style.transform = `translate(${x + 12}px, ${y + 12}px)`;
}

// The click that ends a drag would open the item under the pointer.
function swallowNextClick() {
    const stop = (event) => {
        event.stopPropagation();
        event.preventDefault();
    };

    window.addEventListener("click", stop, { capture: true, once: true });
    setTimeout(() => window.removeEventListener("click", stop, { capture: true }), 0);
}

function targetAt(x, y) {
    const el = document.elementFromPoint(x, y)?.closest("[data-move-target]");
    const target = el ? active.resolveTarget(el.dataset.moveTarget, dragged.value) : null;

    return target ? { target, el } : null;
}

function markTarget(el) {
    if (el === targetEl) return;

    targetEl?.removeAttribute("data-drop-target");
    targetEl = el;
    targetEl?.setAttribute("data-drop-target", "");
}

function onPointerMove(event) {
    if (!press) return;

    if (!dragged.value) {
        if (Math.hypot(event.clientX - press.x, event.clientY - press.y) < DRAG_THRESHOLD) return;

        const { item, label, count } = press.options.start();

        active = press.options;
        dragged.value = item;
        document.body.style.userSelect = "none";
        showGhost(label, count, event.clientX, event.clientY);
    }

    moveGhost(event.clientX, event.clientY);
    markTarget(targetAt(event.clientX, event.clientY)?.el ?? null);
}

function finish() {
    window.removeEventListener("pointermove", onPointerMove);
    window.removeEventListener("pointerup", onPointerUp);
    window.removeEventListener("keydown", onKeyDown);
    ghost?.remove();
    ghost = null;
    document.body.style.userSelect = "";
    press = null;
    active = null;
    dragged.value = null;
    markTarget(null);
}

function onPointerUp(event) {
    const item = dragged.value;
    const options = active;
    const hit = item ? targetAt(event.clientX, event.clientY) : null;

    finish();
    if (!item) return;

    swallowNextClick();
    if (hit) options.drop(item, hit.target);
}

function onKeyDown(event) {
    if (event.key !== "Escape" || !dragged.value) return;

    swallowNextClick();
    finish();
}

export function usePointerDrag() {
    // Only a mouse drags: on a touch screen the same movement scrolls.
    // start() runs once the pointer passes the threshold and returns the
    // dragged item with its ghost label; resolveTarget(id, item) returns the
    // drop target for a data-move-target id, or null when it can't be used.
    function startPress(event, options) {
        if (event.pointerType !== "mouse" || event.button !== 0) return;

        finish();
        press = { x: event.clientX, y: event.clientY, options };
        window.addEventListener("pointermove", onPointerMove);
        window.addEventListener("pointerup", onPointerUp);
        window.addEventListener("keydown", onKeyDown);
    }

    return {
        dragged,
        startPress,
    };
}
