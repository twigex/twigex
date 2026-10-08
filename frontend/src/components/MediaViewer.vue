<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <TransitionRoot as="template" :show="mediaStore.show">
        <Dialog as="div" class="relative z-50" @close="closeViewer">
            <TransitionChild
                as="template"
                enter="ease-out duration-200"
                enter-from="opacity-0"
                enter-to="opacity-100"
                leave="ease-in duration-150"
                leave-from="opacity-100"
                leave-to="opacity-0"
            >
                <div class="fixed inset-0 bg-black/90" />
            </TransitionChild>

            <div class="fixed inset-0 flex flex-col">
                <TransitionChild
                    as="template"
                    enter="ease-out duration-200"
                    enter-from="opacity-0 scale-95"
                    enter-to="opacity-100 scale-100"
                    leave="ease-in duration-150"
                    leave-from="opacity-100 scale-100"
                    leave-to="opacity-0 scale-95"
                >
                    <DialogPanel class="flex flex-col h-full w-full outline-none">
                        <!-- Top bar -->
                        <div class="flex items-center justify-between px-4 py-3 shrink-0">
                            <div class="flex items-center gap-2 min-w-0">
                                <span class="text-sm text-white/70 truncate">{{
                                    currentMedia?.name ?? currentMedia?.type ?? ""
                                }}</span>
                                <span
                                    v-if="mediaStore.files.length > 1"
                                    class="text-xs text-white/40 shrink-0 tabular-nums"
                                >
                                    {{ mediaStore.index + 1 }}&thinsp;/&thinsp;{{
                                        mediaStore.files.length
                                    }}
                                </span>
                                <span v-if="pdfNumPages > 0" class="text-xs text-white/40 shrink-0">
                                    &mdash;&nbsp;{{
                                        pdfNumPages === 1
                                            ? t("media_viewer.pdf.one_page")
                                            : t("media_viewer.pdf.n_pages", {
                                                  n: pdfNumPages,
                                              })
                                    }}
                                </span>
                            </div>
                            <div class="flex items-center gap-1 shrink-0 ml-4">
                                <template v-if="currentMediaType === 'pdf'">
                                    <button
                                        @click="zoomOut"
                                        type="button"
                                        :disabled="pdfLoading"
                                        class="rounded-md p-1.5 text-white/60 hover:text-white hover:bg-white/10 transition-colors disabled:opacity-30 disabled:pointer-events-none"
                                    >
                                        <MagnifyingGlassMinusIcon
                                            class="size-5"
                                            aria-hidden="true"
                                        />
                                    </button>
                                    <span
                                        class="text-xs text-white/40 w-10 text-center tabular-nums"
                                        >{{ Math.round(zoomScale * 100) }}%</span
                                    >
                                    <button
                                        @click="zoomIn"
                                        type="button"
                                        :disabled="pdfLoading"
                                        class="rounded-md p-1.5 text-white/60 hover:text-white hover:bg-white/10 transition-colors disabled:opacity-30 disabled:pointer-events-none"
                                    >
                                        <MagnifyingGlassPlusIcon
                                            class="size-5"
                                            aria-hidden="true"
                                        />
                                    </button>
                                    <div class="w-px h-5 bg-white/20 mx-1" />
                                </template>
                                <button
                                    @click="closeViewer"
                                    type="button"
                                    class="rounded-md p-1.5 text-white/60 hover:text-white hover:bg-white/10 transition-colors"
                                >
                                    <XMarkIcon class="size-5" aria-hidden="true" />
                                </button>
                            </div>
                        </div>

                        <!-- Media area -->
                        <div class="flex-1 flex items-center justify-center min-h-0 relative">
                            <button
                                v-if="mediaStore.files.length > 1"
                                @click="prevMedia"
                                type="button"
                                class="absolute left-3 z-10 rounded-full p-2 bg-black/50 text-white hover:bg-black/70 transition-colors"
                            >
                                <ArrowLeftIcon class="size-5" aria-hidden="true" />
                            </button>

                            <img
                                v-if="currentMediaType === 'image'"
                                :src="currentMedia.url"
                                alt=""
                                class="max-h-full max-w-full object-contain"
                            />

                            <video
                                v-else-if="currentMedia?.kind === 'gifv'"
                                :src="currentMedia.url"
                                :poster="currentMedia.poster || ''"
                                autoplay
                                loop
                                muted
                                playsinline
                                class="max-h-full max-w-full object-contain rounded"
                            />

                            <video
                                v-else-if="currentMediaType === 'video'"
                                ref="videoPlayer"
                                class="max-h-full max-w-full"
                                controls
                                autoplay
                            >
                                <source
                                    :src="currentMedia.url"
                                    :type="mediaType(currentMedia.type)"
                                />
                            </video>

                            <div
                                v-if="currentMediaType === 'pdf'"
                                ref="pdfWrapper"
                                class="h-full w-full relative overflow-y-auto"
                            >
                                <div
                                    v-if="pdfLoading"
                                    class="absolute inset-0 flex items-center justify-center"
                                >
                                    <ArrowPathIcon
                                        class="size-8 text-white/40 animate-spin"
                                        aria-hidden="true"
                                    />
                                </div>
                                <div
                                    v-else-if="pdfError"
                                    class="absolute inset-0 flex items-center justify-center"
                                >
                                    <p class="text-sm text-white/60">
                                        {{ t("media_viewer.pdf.error_loading") }}
                                    </p>
                                </div>
                                <div
                                    ref="pdfContainer"
                                    class="flex flex-col items-center px-4 py-4"
                                />
                            </div>

                            <button
                                v-if="mediaStore.files.length > 1"
                                @click="nextMedia"
                                type="button"
                                class="absolute right-3 z-10 rounded-full p-2 bg-black/50 text-white hover:bg-black/70 transition-colors"
                            >
                                <ArrowRightIcon class="size-5" aria-hidden="true" />
                            </button>
                        </div>

                        <!-- Dot indicators -->
                        <div
                            v-if="mediaStore.files.length > 1"
                            class="flex items-center justify-center gap-1.5 py-3 shrink-0"
                        >
                            <button
                                v-for="(_, i) in mediaStore.files"
                                :key="i"
                                @click="goToIndex(i)"
                                class="w-1.5 h-1.5 rounded-full transition-all"
                                :class="
                                    i === mediaStore.index
                                        ? 'bg-white scale-125'
                                        : 'bg-white/30 hover:bg-white/60'
                                "
                            />
                        </div>
                    </DialogPanel>
                </TransitionChild>
            </div>
        </Dialog>
    </TransitionRoot>
</template>

<script setup>
import { ref, computed, watch, onMounted, onBeforeUnmount, onUnmounted, nextTick } from "vue";
import { t } from "@/i18n/index.js";
import * as pdfjsLib from "pdfjs-dist";
import pdfWorker from "pdfjs-dist/build/pdf.worker.mjs?url";
import { useMediaStore } from "@/store/media";
import { Dialog, DialogPanel, TransitionChild, TransitionRoot } from "@headlessui/vue";
import {
    XMarkIcon,
    ArrowLeftIcon,
    ArrowRightIcon,
    ArrowPathIcon,
    MagnifyingGlassPlusIcon,
    MagnifyingGlassMinusIcon,
} from "@heroicons/vue/24/outline";

pdfjsLib.GlobalWorkerOptions.workerSrc = pdfWorker;

const mediaStore = useMediaStore();
const videoPlayer = ref(null);
const pdfWrapper = ref(null);
const pdfContainer = ref(null);
const zoomScale = ref(1);
const pdfDocGetter = ref(() => null);
const pdfNumPages = ref(0);
const pdfLoading = ref(false);
const pdfError = ref(false);
const currentMediaType = ref("unknown");

let zoomRenderTimer = null;

const currentMedia = computed(() => {
    if (!mediaStore.files.length) return null;

    return mediaStore.files[mediaStore.index];
});

watch(
    () => currentMedia.value,
    async () => {
        if (!currentMedia.value) {
            currentMediaType.value = "unknown";
        } else if (currentMedia.value.type.includes("image")) {
            currentMediaType.value = "image";
        } else if (currentMedia.value.type.includes("video")) {
            currentMediaType.value = "video";
        } else if (currentMedia.value.type.includes("pdf")) {
            currentMediaType.value = "pdf";
        } else {
            currentMediaType.value = "unknown";
        }

        if (currentMediaType.value === "pdf") {
            zoomScale.value = 1;
            pdfNumPages.value = 0;
            await loadPDF(currentMedia.value.url);
        }
    },
);

watch(zoomScale, () => {
    if (currentMediaType.value !== "pdf" || pdfDocGetter.value() === null) return;
    if (zoomRenderTimer) clearTimeout(zoomRenderTimer);
    zoomRenderTimer = setTimeout(async () => {
        const wrapper = pdfWrapper.value;
        const scrollRatio = wrapper
            ? wrapper.scrollTop / Math.max(wrapper.scrollHeight - wrapper.clientHeight, 1)
            : 0;

        await renderAllPages(currentMedia.value.id);
        await nextTick();
        if (wrapper) {
            wrapper.scrollTop =
                scrollRatio * Math.max(wrapper.scrollHeight - wrapper.clientHeight, 1);
        }
    }, 150);
});

onMounted(() => window.addEventListener("keydown", handleKeydown));
onBeforeUnmount(stopVideo);
onUnmounted(() => window.removeEventListener("keydown", handleKeydown));

function handleKeydown(e) {
    if (!mediaStore.show) return;
    if (e.key === "ArrowRight") nextMedia();
    else if (e.key === "ArrowLeft") prevMedia();
}

async function loadPDF(pdfUrl) {
    pdfLoading.value = true;
    pdfError.value = false;
    try {
        await nextTick(); // ensure pdfWrapper is in the DOM
        const doc = await pdfjsLib.getDocument({
            url: pdfUrl,
            wasmUrl: `${import.meta.env.BASE_URL}pdfjs/wasm/`,
        }).promise;

        const naturalViewport = (await doc.getPage(1)).getViewport({
            scale: 1,
        });
        const containerWidth = pdfWrapper.value?.clientWidth ?? 0;

        if (containerWidth > 0) {
            zoomScale.value = parseFloat(
                Math.min((containerWidth - 32) / naturalViewport.width, 2).toFixed(2),
            );
        }

        pdfNumPages.value = doc.numPages;

        // Pass doc directly so pdfDocGetter stays null during the initial
        // render. The zoom watcher checks pdfDocGetter and exits early while
        // it is null, preventing a spurious re-render 150ms after load.
        await renderAllPages(currentMedia.value.id, doc);
        pdfDocGetter.value = () => doc;
    } catch (err) {
        pdfError.value = true;
        console.error("Failed to load PDF:", err);
    } finally {
        pdfLoading.value = false;
    }
}

async function renderAllPages(id, docArg) {
    const doc = docArg ?? pdfDocGetter.value();

    if (!doc) return;
    pdfContainer.value.innerHTML = "";
    for (let pageNum = 1; pageNum <= doc.numPages; pageNum++) {
        if (id !== currentMedia.value.id) {
            pdfContainer.value.innerHTML = "";
            break;
        }

        const canvas = document.createElement("canvas");

        canvas.style.marginBottom = "1rem";
        await renderPDFPage(doc, pageNum, canvas);
        pdfContainer.value.appendChild(canvas);
    }
}

async function renderPDFPage(doc, pageNum, canvas) {
    try {
        const page = await doc.getPage(pageNum);
        const dpr = window.devicePixelRatio || 1;
        const viewport = page.getViewport({ scale: zoomScale.value * dpr });

        canvas.width = viewport.width;
        canvas.height = viewport.height;
        canvas.style.width = `${viewport.width / dpr}px`;
        canvas.style.height = `${viewport.height / dpr}px`;

        await page.render({ canvasContext: canvas.getContext("2d"), viewport }).promise;
    } catch (err) {
        console.error("Error rendering PDF page:", err);
    }
}

function cleanupPDF() {
    if (zoomRenderTimer) {
        clearTimeout(zoomRenderTimer);
        zoomRenderTimer = null;
    }

    try {
        const doc = pdfDocGetter.value();

        if (doc?.destroy) doc.destroy();
    } catch (e) {
        console.error("Error cleaning up PDF:", e);
    }

    pdfDocGetter.value = () => null;
    pdfNumPages.value = 0;
    pdfLoading.value = false;
    pdfError.value = false;
    if (pdfContainer.value) pdfContainer.value.innerHTML = "";
}

function stopVideo() {
    const player = videoPlayer.value;

    if (!player) return;

    player.pause();

    const source = player.querySelector("source");

    if (source) source.removeAttribute("src");

    player.load();
}

function closeViewer() {
    stopVideo();
    cleanupPDF();
    mediaStore.close();
}

async function navigate(newIndex) {
    stopVideo();
    if (currentMediaType.value === "pdf") cleanupPDF();
    currentMediaType.value = "unknown";
    await nextTick();
    mediaStore.index = newIndex;
    zoomScale.value = 1;
}

async function nextMedia() {
    await navigate((mediaStore.index + 1) % mediaStore.files.length);
}

async function prevMedia() {
    await navigate((mediaStore.index - 1 + mediaStore.files.length) % mediaStore.files.length);
}

async function goToIndex(i) {
    if (i !== mediaStore.index) await navigate(i);
}

function zoomIn() {
    zoomScale.value = parseFloat((zoomScale.value + 0.1).toFixed(1));
}

function zoomOut() {
    if (zoomScale.value > 0.2) zoomScale.value = parseFloat((zoomScale.value - 0.1).toFixed(1));
}

function mediaType(mimeType) {
    if (mimeType === "video/quicktime") return "video/mp4";
    if (mimeType === "video/x-matroska") return "video/webm";

    return mimeType;
}
</script>
