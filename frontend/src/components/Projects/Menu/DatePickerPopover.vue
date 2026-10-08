<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <Teleport to="body">
        <div v-if="props.visible" ref="floating" :style="floatingStyles" class="z-[60]">
            <div>
                <ul
                    class="w-full rounded-md bg-white py-1 text-base shadow-lg ring-1 ring-black/5 focus:outline-none sm:text-sm"
                    style="overflow: visible"
                >
                    <div
                        class="mt-2 relative overflow-visible rounded-md mx-auto flex justify-center"
                        :style="{ width: targetW + 'px', height: targetH + 'px' }"
                    >
                        <div
                            ref="contentEl"
                            :style="{
                                transform: `scale(${scale})`,
                                transformOrigin: 'top center',
                                width: 'max-content',
                            }"
                        >
                            <DatePicker v-model="selectedDate" />
                        </div>
                    </div>

                    <div class="mt-4 px-4 pb-2 sm:flex sm:items-center sm:justify-between">
                        <div class="mb-3 sm:mb-0">
                            <button
                                type="button"
                                class="inline-flex items-center justify-center rounded-md bg-white px-3 py-1.5 text-sm font-medium text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
                                @click="clearDate"
                                :title="t('projects.menu.date_picker_dialog.date_clear_title')"
                            >
                                {{ t("common.button.clear") }}
                            </button>
                        </div>

                        <div class="sm:flex sm:flex-row-reverse sm:space-x-reverse sm:space-x-3">
                            <button
                                type="button"
                                class="inline-flex w-full sm:w-auto items-center justify-center rounded-md bg-indigo-600 px-3 py-1.5 text-sm font-medium text-white shadow-sm hover:bg-indigo-500"
                                @click="saveDate"
                            >
                                {{ t("common.button.save") }}
                            </button>

                            <button
                                type="button"
                                class="mt-3 sm:mt-0 inline-flex w-full sm:w-auto items-center justify-center rounded-md bg-white px-3 py-1.5 text-sm font-medium text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 hover:bg-gray-50"
                                @click="closePopover"
                            >
                                {{ t("common.button.cancel") }}
                            </button>
                        </div>
                    </div>
                </ul>
            </div>
        </div>
    </Teleport>
</template>

<script setup>
import { t } from "@/i18n/index.js";
import { ref, onMounted, onBeforeUnmount, nextTick, watch, toRef } from "vue";
import DatePicker from "@/components/DatePicker/DatePicker.vue";
import { useAnchoredPopup } from "@/composables/useAnchoredPopup";

const props = defineProps({
    visible: { type: Boolean, default: false },
    anchor: { type: Object, default: null },
    activePopoverItem: {
        type: Object,
        default: () => ({ label: "", id: null }),
    },
    // The day the field holds, YYYY-MM-DD, which the picker opens on.
    value: { type: String, default: "" },
});

const emit = defineEmits(["update:visible", "save"]);

const { floating, floatingStyles } = useAnchoredPopup({ anchor: toRef(props, "anchor") });

// Make the frame bigger
const targetW = 300;
const targetH = 300;

const contentEl = ref(null);
const scale = ref(1);
const selectedDate = ref(null);
let ro;

function closePopover() {
    emit("update:visible", false);
}

// A click elsewhere cancels. One on the cell that opened the picker is left
// to the cell, which closes it, or it would close and open again.
function onPointerDown(event) {
    if (!props.visible) return;

    const panel = floating.value?.$el ?? floating.value;

    if (panel?.contains?.(event.target) || props.anchor?.contains?.(event.target)) return;

    closePopover();
}

function onKeyDown(event) {
    if (props.visible && event.key === "Escape") closePopover();
}

function clearDate() {
    emit("save", "");
    emit("update:visible", false);
    selectedDate.value = null;
}

function todayYMD() {
    const d = new Date();
    const y = d.getFullYear();
    const m = String(d.getMonth() + 1).padStart(2, "0");
    const day = String(d.getDate()).padStart(2, "0");

    return `${y}-${m}-${day}`;
}

function saveDate() {
    emit("save", selectedDate.value || todayYMD());
    emit("update:visible", false);
}

function measureAndScale() {
    const el = contentEl.value;

    if (!el) return;

    const prev = el.style.transform;

    el.style.transform = "scale(1)";

    const naturalW = el.scrollWidth || el.getBoundingClientRect().width;
    const naturalH = el.scrollHeight || el.getBoundingClientRect().height;

    scale.value = Math.min(targetW / naturalW, targetH / naturalH, 1);

    el.style.transform = prev;
}

onMounted(async () => {
    await nextTick();
    measureAndScale();

    ro = new ResizeObserver(() => measureAndScale());
    if (contentEl.value) ro.observe(contentEl.value);

    window.addEventListener("resize", measureAndScale);
    document.addEventListener("pointerdown", onPointerDown, true);
    document.addEventListener("keydown", onKeyDown);
});

watch(
    () => props.visible,
    async (v) => {
        if (v) {
            selectedDate.value = props.value || null;
            await nextTick();
            measureAndScale();
        } else {
            selectedDate.value = null;
        }
    },
);

// Moving to another date field while open starts the picker on that one.
watch(
    () => props.value,
    (value) => {
        if (props.visible) selectedDate.value = value || null;
    },
);

onBeforeUnmount(() => {
    if (ro && contentEl.value) ro.unobserve(contentEl.value);
    window.removeEventListener("resize", measureAndScale);
    document.removeEventListener("pointerdown", onPointerDown, true);
    document.removeEventListener("keydown", onKeyDown);
});
</script>
