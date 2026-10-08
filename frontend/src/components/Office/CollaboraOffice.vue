<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <TransitionRoot as="template" :show="officeStore.open && officeStore.type === 'collabora'">
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
                        <iframe
                            :src="officeStore.url"
                            frameborder="0"
                            class="w-full h-full"
                        ></iframe>
                    </div>
                </div>
            </TransitionChild>
        </div>
    </TransitionRoot>
</template>

<script setup>
import { TransitionChild, TransitionRoot } from "@headlessui/vue";
import { useOfficeStore } from "@/store/office";
import { onMounted, onUnmounted } from "vue";

const officeStore = useOfficeStore();

const handlePostMessage = (event) => {
    let data;

    try {
        data = JSON.parse(event.data);
    } catch {
        return;
    }

    if (data?.MessageId === "UI_Close") {
        officeStore.closeOffice();
    }
};

onMounted(() => {
    window.addEventListener("message", handlePostMessage);
});

onUnmounted(() => {
    window.removeEventListener("message", handlePostMessage);
});
</script>
