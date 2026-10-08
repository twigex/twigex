<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div class="flex h-screen flex-col bg-gray-50">
        <header
            class="flex shrink-0 items-center gap-2 border-b border-gray-200 bg-white px-4 py-3 sm:px-6"
        >
            <CloudArrowDownIcon
                class="h-6 w-6 text-indigo-600"
                aria-hidden="true"
            />
            <span class="truncate text-sm font-semibold text-gray-900">
                {{ host }}
            </span>
        </header>

        <div class="min-h-0 flex-1">
            <div
                v-if="!showBrowser"
                class="flex h-full items-center justify-center overflow-y-auto p-4"
            >
                <div
                    v-if="state === 'loading'"
                    class="flex items-center gap-3 text-gray-500"
                >
                    <ArrowPathIcon
                        class="h-5 w-5 animate-spin"
                        aria-hidden="true"
                    />
                    <span class="text-sm">{{ t("public_share.loading") }}</span>
                </div>

                <PublicShareErrorCard
                    v-else-if="state === 'error'"
                    :message="errorMessage"
                />

                <PublicSharePasswordForm
                    v-else-if="state === 'password'"
                    v-model:password="password"
                    :error="passwordError"
                    :unlocking="unlocking"
                    @submit="unlock"
                />

                <PublicShareFileCard
                    v-else-if="state === 'ready' && !isFolder"
                    :share="share"
                    :token="token"
                    @preview="previewRoot"
                    @open-office="openInOffice()"
                />

                <PublicShareUploadDropzone
                    v-else-if="state === 'ready' && isFolder && share.AllowUpload"
                    v-model:dragOver="dragOver"
                    :uploading="uploading"
                    :upload-label="uploadLabel"
                    @upload="triggerUpload"
                    @drop="handleDrop"
                />
            </div>

            <div v-else class="flex h-full flex-col px-4 sm:px-6">
                <PublicShareToolbar
                    v-model:sort="sort"
                    v-model:viewMode="viewMode"
                    :share="share"
                    :token="token"
                    :root-name="rootName"
                    :folder-path="folderPath"
                    :item-count="children.length"
                    :current-child-id="currentChildId"
                    :uploading="uploading"
                    :upload-label="uploadLabel"
                    @navigate="navigateTo"
                    @upload="triggerUpload"
                />

                <div
                    class="relative min-h-0 flex-1 overflow-y-auto py-4"
                    @dragover.prevent
                    @dragenter.prevent="dragOver = true"
                    @dragleave.self="dragOver = false"
                    @drop.prevent="handleDrop"
                >
                    <div
                        v-if="dragOver && share.AllowUpload"
                        class="pointer-events-none absolute inset-2 z-10 flex items-center justify-center rounded-lg border-2 border-dashed border-indigo-400 bg-indigo-50/60 text-sm font-medium text-indigo-600"
                    >
                        {{ t("public_share.drop_to_upload") }}
                    </div>

                    <div
                        v-if="navigating"
                        class="absolute inset-0 z-10 flex items-center justify-center bg-white/60"
                    >
                        <ArrowPathIcon
                            class="h-6 w-6 animate-spin text-gray-400"
                            aria-hidden="true"
                        />
                    </div>

                    <p
                        v-if="share.Message"
                        class="mb-4 whitespace-pre-line rounded-lg bg-white p-3 text-sm text-gray-600 ring-1 ring-gray-200"
                    >
                        {{ share.Message }}
                    </p>

                    <div
                        v-if="!children.length"
                        class="flex flex-col items-center justify-center py-16 text-center"
                    >
                        <FolderOpenIcon
                            class="h-10 w-10 text-gray-300"
                            aria-hidden="true"
                        />
                        <p class="mt-2 text-sm text-gray-500">
                            {{ t("public_share.empty") }}
                        </p>
                    </div>

                    <PublicShareItemGrid
                        v-else-if="viewMode === 'grid'"
                        :items="pagedChildren"
                        :token="token"
                        :failed-thumbs="failedThumbs"
                        @child-click="onChildClick"
                        @open="openChild"
                        @thumb-failed="thumbFailed"
                    />

                    <PublicShareItemList
                        v-else
                        :items="pagedChildren"
                        :token="token"
                        :failed-thumbs="failedThumbs"
                        @child-click="onChildClick"
                        @open="openChild"
                        @thumb-failed="thumbFailed"
                    />
                </div>

                <PublicSharePager
                    v-if="pageCount > 1"
                    v-model:page="page"
                    :page-count="pageCount"
                />
            </div>
        </div>

        <input
            ref="fileInput"
            type="file"
            class="hidden"
            @change="handleUpload"
        />

        <MediaViewer />
        <CollaboraOffice />
        <EuroOfficeEditor />
    </div>
</template>

<script setup>
import { onMounted } from "vue";
import { useRoute } from "vue-router";
import { ArrowPathIcon, FolderOpenIcon, CloudArrowDownIcon } from "@heroicons/vue/24/outline";
import { t } from "@/i18n";
import { usePublicShareBrowser } from "@/composables/files/usePublicShareBrowser";
import { usePublicShareUnlock } from "@/composables/files/usePublicShareUnlock";
import { usePublicShareOpen } from "@/composables/files/usePublicShareOpen";
import { usePublicShareUpload } from "@/composables/files/usePublicShareUpload";
import CollaboraOffice from "@/components/Office/CollaboraOffice.vue";
import EuroOfficeEditor from "@/components/Office/EuroOfficeEditor.vue";
import MediaViewer from "@/components/MediaViewer.vue";
import PublicShareErrorCard from "@/components/PublicShare/PublicShareErrorCard.vue";
import PublicSharePasswordForm from "@/components/PublicShare/PublicSharePasswordForm.vue";
import PublicShareFileCard from "@/components/PublicShare/PublicShareFileCard.vue";
import PublicShareUploadDropzone from "@/components/PublicShare/PublicShareUploadDropzone.vue";
import PublicShareToolbar from "@/components/PublicShare/PublicShareToolbar.vue";
import PublicShareItemGrid from "@/components/PublicShare/PublicShareItemGrid.vue";
import PublicShareItemList from "@/components/PublicShare/PublicShareItemList.vue";
import PublicSharePager from "@/components/PublicShare/PublicSharePager.vue";

const route = useRoute();
const token = route.params.token;
const host = window.location.host;

const {
    state,
    errorMessage,
    share,
    rootName,
    folderPath,
    failedThumbs,
    viewMode,
    sort,
    page,
    navigating,
    isFolder,
    children,
    showBrowser,
    currentChildId,
    sortedChildren,
    pageCount,
    pagedChildren,
    thumbFailed,
    load,
    openFolder,
    navigateTo,
} = usePublicShareBrowser(token);

const { password, passwordError, unlocking, unlock } = usePublicShareUnlock(token, load);

const { onChildClick, openChild, previewRoot, openInOffice } = usePublicShareOpen({
    token,
    share,
    sortedChildren,
    openFolder,
});

const { dragOver, uploading, fileInput, uploadLabel, triggerUpload, handleUpload, handleDrop } =
    usePublicShareUpload({
        token,
        share,
        currentChildId,
        load,
    });

onMounted(load);
</script>
