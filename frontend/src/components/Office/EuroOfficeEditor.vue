<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <TransitionRoot as="template" :show="officeStore.open && officeStore.type === 'eurooffice'">
        <div class="fixed inset-0 z-50 flex items-center justify-center">
            <TransitionChild
                as="template"
                enter="ease-out duration-300"
                enter-from="opacity-0"
                enter-to="opacity-100"
                leave="ease-in duration-200"
                leave-from="opacity-100"
                leave-to="opacity-0"
            >
                <div class="fixed inset-0 bg-gray-500 bg-opacity-75 transition-opacity"></div>
            </TransitionChild>

            <TransitionChild
                as="template"
                enter="ease-out duration-300"
                enter-from="opacity-0 translate-y-4 sm:translate-y-0 sm:scale-95"
                enter-to="opacity-100 translate-y-0 sm:scale-100"
                leave="ease-in duration-200"
                leave-from="opacity-100 translate-y-0 sm:scale-100"
                leave-to="opacity-0 translate-y-4 sm:translate-y-0 sm:scale-95"
            >
                <div class="fixed inset-0 flex items-center justify-center">
                    <div class="relative w-full h-full bg-white shadow-xl">
                        <div id="eurooffice-editor" class="w-full h-full"></div>
                    </div>
                </div>
            </TransitionChild>
        </div>
    </TransitionRoot>
</template>

<script setup>
import { TransitionChild, TransitionRoot } from "@headlessui/vue";
import { useOfficeStore } from "@/store/office";
import { watch, onUnmounted } from "vue";

const officeStore = useOfficeStore();

let editorInstance = null;
let scriptEl = null;

function destroyEditor() {
    if (editorInstance) {
        try {
            editorInstance.destroyEditor();
        } catch (_) {}

        editorInstance = null;
    }

    if (scriptEl) {
        scriptEl.remove();
        scriptEl = null;
    }
}

function loadEditor(host, config) {
    destroyEditor();

    scriptEl = document.createElement("script");
    scriptEl.src = host + "/web-apps/apps/api/documents/api.js";
    scriptEl.onload = () => {
        if (!window.DocsAPI) return;
        editorInstance = new window.DocsAPI.DocEditor("eurooffice-editor", {
            ...config,
            editorConfig: {
                ...config.editorConfig,
                customization: {
                    ...config.editorConfig?.customization,
                    close: { visible: true },
                },
            },
            events: {
                onRequestClose: () => {
                    officeStore.closeOffice();
                },
            },
        });
    };

    document.head.appendChild(scriptEl);
}

watch(
    () => officeStore.open,
    (isOpen) => {
        if (isOpen && officeStore.type === "eurooffice" && officeStore.config) {
            loadEditor(officeStore.host, officeStore.config);
        } else if (!isOpen) {
            destroyEditor();
        }
    },
);

onUnmounted(() => {
    destroyEditor();
});
</script>
