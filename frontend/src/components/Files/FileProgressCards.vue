<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div
        aria-live="assertive"
        class="pointer-events-none fixed inset-0 z-50 flex items-end justify-end px-4 py-6 sm:p-6"
    >
        <div class="flex w-full max-w-sm flex-col items-end gap-3">
            <transition
                enter-active-class="transform ease-out duration-300 transition"
                enter-from-class="translate-y-2 opacity-0 sm:translate-x-2"
                enter-to-class="translate-y-0 opacity-100 sm:translate-x-0"
                leave-active-class="transition ease-in duration-100"
                leave-from-class="opacity-100"
                leave-to-class="opacity-0"
            >
                <FileProgressCard
                    v-if="moves.length > 0"
                    class="pointer-events-auto"
                    :title="moveTitle"
                    :items="moves"
                    @close="closeMoves"
                >
                    <template #icon>
                        <FolderArrowDownIcon
                            class="h-6 w-6 flex-shrink-0 text-indigo-500"
                            aria-hidden="true"
                        />
                    </template>
                </FileProgressCard>
            </transition>

            <transition
                enter-active-class="transform ease-out duration-300 transition"
                enter-from-class="translate-y-2 opacity-0 sm:translate-x-2"
                enter-to-class="translate-y-0 opacity-100 sm:translate-x-0"
                leave-active-class="transition ease-in duration-100"
                leave-from-class="opacity-100"
                leave-to-class="opacity-0"
            >
                <FileProgressCard
                    v-if="uploads.length > 0"
                    class="pointer-events-auto"
                    :title="uploadTitle"
                    :items="uploads"
                    @close="closeUploads"
                >
                    <template #icon="{ item }">
                        <component
                            :is="fileTypes[item.type]?.icon ?? fileTypes['not-found'].icon"
                            :class="[
                                'h-6 w-6 flex-shrink-0',
                                fileTypes[item.type]?.color ?? fileTypes['not-found'].color,
                            ]"
                            aria-hidden="true"
                        />
                    </template>
                </FileProgressCard>
            </transition>
        </div>
    </div>
</template>

<script setup>
import { onMounted } from "vue";
import FileProgressCard from "@/components/Files/FileProgressCard.vue";
import fileTypes from "@/constants/fileTypes";
import { useFileProgress } from "@/composables/files/useFileProgress";
import { useJobsStore } from "@/store/jobs";
import { FolderArrowDownIcon } from "@heroicons/vue/24/outline";

const { uploads, moves, uploadTitle, moveTitle, closeUploads, closeMoves } = useFileProgress();

// A move started before a reload is still running on the server.
onMounted(() => useJobsStore().startPolling());
</script>
