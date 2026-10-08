<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <TransitionRoot as="template" :show="modelValue">
        <Dialog class="relative z-50" @close="cancel">
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
                            class="relative transform overflow-hidden rounded-lg bg-white p-6 text-left shadow-xl transition-all w-full max-w-sm"
                        >
                            <div class="flex items-center justify-between mb-4">
                                <DialogTitle as="h3" class="text-base font-semibold text-gray-900">
                                    {{ t("settings.profile_info.adjust_photo") }}
                                </DialogTitle>
                                <button
                                    type="button"
                                    class="rounded-md bg-white text-gray-400 hover:text-gray-500 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2"
                                    @click="cancel"
                                >
                                    <XMarkIcon class="size-5" />
                                </button>
                            </div>

                            <!-- Cropper area -->
                            <div
                                ref="containerRef"
                                class="relative mx-auto bg-gray-900 rounded-lg"
                                style="width: 300px; height: 300px; overflow: hidden"
                            >
                                <img ref="imgRef" class="hidden" :src="objectUrl" alt="" />
                            </div>

                            <p class="mt-3 text-center text-xs text-gray-400">
                                {{ t("settings.profile_info.photo_crop_hint") }}
                            </p>

                            <!-- Actions -->
                            <div class="mt-5 flex gap-3 justify-end">
                                <button
                                    type="button"
                                    class="rounded-md bg-white px-3 py-2 text-sm font-semibold text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 hover:bg-gray-50"
                                    :disabled="uploading"
                                    @click="cancel"
                                >
                                    {{ t("common.button.cancel") }}
                                </button>
                                <button
                                    type="button"
                                    class="rounded-md px-3 py-2 text-sm font-semibold text-white shadow-sm focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600"
                                    :class="
                                        uploading
                                            ? 'bg-gray-300'
                                            : 'bg-indigo-600 hover:bg-indigo-500'
                                    "
                                    :disabled="uploading"
                                    @click="confirm"
                                >
                                    {{
                                        uploading
                                            ? t("settings.profile_info.photo_uploading")
                                            : t("settings.profile_info.apply_photo")
                                    }}
                                </button>
                            </div>
                        </DialogPanel>
                    </TransitionChild>
                </div>
            </div>
        </Dialog>
    </TransitionRoot>
</template>

<script setup>
import { ref, watch, nextTick } from "vue";
import Cropper from "cropperjs";
import { Dialog, DialogPanel, DialogTitle, TransitionChild, TransitionRoot } from "@headlessui/vue";
import { XMarkIcon } from "@heroicons/vue/24/outline";
import { t } from "@/i18n/index.js";
import { useAlertStore } from "@/store/alerts";
import userService from "@/services/userService";

const props = defineProps({
    modelValue: { type: Boolean, required: true },
    file: { type: File, default: null },
});

const emit = defineEmits(["update:modelValue", "uploaded"]);

const imgRef = ref(null);
const containerRef = ref(null);
const objectUrl = ref("");
const uploading = ref(false);
let cropper = null;

watch(
    () => props.modelValue,
    async (shown) => {
        if (shown && props.file) {
            objectUrl.value = URL.createObjectURL(props.file);
            await nextTick();
            // Delay past the 300ms dialog enter transition so getBoundingClientRect()
            // returns correct values (canvas reports ~285px during scale-95→scale-100).
            setTimeout(initCropper, 310);
        } else {
            destroyCropper();
        }
    },
);

async function initCropper() {
    if (!imgRef.value) return;
    cropper = new Cropper(imgRef.value, {
        container: containerRef.value,
        template: `
            <cropper-canvas background style="width:300px;height:300px;display:block;">
                <cropper-image scalable translatable initial-center-size="cover"></cropper-image>
                <cropper-shade hidden></cropper-shade>
                <cropper-handle action="select" plain></cropper-handle>
                <cropper-selection
                    initial-coverage="1"
                    aspect-ratio="1"
                    movable="false"
                    resizable="false"
                    zoomable="false"
                >
                    <cropper-handle action="move" theme-color="rgba(255,255,255,0.35)"></cropper-handle>
                </cropper-selection>
            </cropper-canvas>
        `,
    });

    await nextTick();
    applyCircleStyle();

    const img = cropper.getCropperImage();

    if (img) {
        await img.$ready();
        img.$center("cover");
    }
}

function applyCircleStyle() {
    if (!containerRef.value) return;
    const selection = containerRef.value.querySelector("cropper-selection");

    if (selection) {
        selection.style.cssText += `
            border-radius: 50%;
            box-shadow: 0 0 0 9999px rgba(0, 0, 0, 0.55);
            overflow: hidden;
        `;
    }
}

function destroyCropper() {
    if (cropper) {
        cropper.destroy();
        cropper = null;
    }

    if (objectUrl.value) {
        URL.revokeObjectURL(objectUrl.value);
        objectUrl.value = "";
    }
}

async function confirm() {
    const selection = cropper?.getCropperSelection();

    if (!selection) return;

    uploading.value = true;
    try {
        const canvas = await selection.$toCanvas({
            width: 256,
            height: 256,
        });

        const blob = await new Promise((resolve) => canvas.toBlob(resolve, "image/jpeg", 0.92));
        const formData = new FormData();

        formData.append("photo", blob, "photo.jpg");

        const res = await userService.uploadPhoto(formData);

        emit("uploaded", res.data);
        emit("update:modelValue", false);
    } catch {
        useAlertStore().showError(t.value("settings.profile_info.photo_upload_failed"));
    } finally {
        uploading.value = false;
    }
}

function cancel() {
    if (!uploading.value) {
        destroyCropper();
        emit("update:modelValue", false);
    }
}
</script>

<style scoped>
:deep(cropper-canvas) {
    display: block;
    width: 300px !important;
    height: 300px !important;
}
</style>
