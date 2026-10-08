<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div class="w-full max-w-lg rounded-xl bg-white p-8 shadow-sm ring-1 ring-gray-200">
        <div class="flex items-start gap-4">
            <span class="flex h-12 w-12 shrink-0 items-center justify-center rounded-lg bg-gray-50">
                <component
                    :is="rootMeta.icon"
                    :class="['h-6 w-6', rootMeta.color]"
                    aria-hidden="true"
                />
            </span>
            <div class="min-w-0 flex-1">
                <h1 class="truncate text-lg font-semibold text-gray-900">
                    {{ share.Name }}
                </h1>
                <p class="mt-0.5 text-sm text-gray-500">
                    {{ t("public_share.shared_by") }}
                    {{ share.User.Name }} {{ share.User.LastName }}
                </p>
            </div>
            <div class="flex shrink-0 items-center gap-2">
                <button
                    v-if="share.AllowView && previewable(share.Type)"
                    type="button"
                    @click="emit('preview')"
                    class="rounded-md bg-white px-3 py-2 text-sm font-semibold text-indigo-600 shadow-sm ring-1 ring-inset ring-indigo-200 hover:bg-indigo-50"
                >
                    {{ t("public_share.preview") }}
                </button>
                <button
                    v-else-if="share.AllowView && isOfficeType(share.Type)"
                    type="button"
                    @click="emit('openOffice')"
                    class="rounded-md bg-white px-3 py-2 text-sm font-semibold text-indigo-600 shadow-sm ring-1 ring-inset ring-indigo-200 hover:bg-indigo-50"
                >
                    {{ t("public_share.open") }}
                </button>
                <a
                    v-if="share.AllowDownload"
                    :href="publicShareService.downloadUrl(token)"
                    class="rounded-md bg-indigo-600 px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-indigo-500"
                >
                    {{ t("public_share.download") }}
                </a>
            </div>
        </div>
        <p
            v-if="share.Message"
            class="mt-4 whitespace-pre-line rounded-lg bg-gray-50 p-3 text-sm text-gray-600"
        >
            {{ share.Message }}
        </p>
    </div>
</template>

<script setup>
import { computed } from "vue";
import { t } from "@/i18n";
import fileTypes from "@/constants/fileTypes";
import publicShareService from "@/services/publicShareService";
import { isOfficeType, previewable, typeMeta } from "@/utils/files/publicShareTypes";

const props = defineProps({
    share: {
        type: Object,
        required: true,
    },
    token: {
        type: String,
        required: true,
    },
});

const emit = defineEmits(["preview", "openOffice"]);

const rootMeta = computed(() =>
    props.share
        ? typeMeta({ IsFolder: props.share.IsFolder ?? false, Type: props.share.Type })
        : fileTypes["not-found"],
);
</script>
