<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div
        class="flex items-center justify-between rounded-t-xl border border-b-0 border-gray-200 bg-indigo-50 px-3 py-2"
    >
        <div class="flex items-center gap-x-2 min-w-0">
            <div class="w-0.5 h-4 shrink-0 rounded-full bg-indigo-400" />
            <span class="text-xs text-gray-500 shrink-0">
                {{ t("channels.input.replying_to") }}
            </span>
            <span class="text-xs font-semibold text-gray-800 truncate">
                {{ userStore.getUserFullName(props.reply.user_id) }}
            </span>
        </div>
        <button
            @click="emits('cancel')"
            type="button"
            class="ml-2 shrink-0 rounded-md p-1 text-gray-400 hover:bg-indigo-100 hover:text-gray-600 transition-colors"
        >
            <XMarkIcon class="size-4" aria-hidden="true" />
        </button>
    </div>
</template>

<script setup>
import { t } from "@/i18n/index.js";
import { XMarkIcon } from "@heroicons/vue/24/solid";
import { useUserStore } from "@/store/user";
import { useUser } from "@/composables/useUser";

const props = defineProps({
    reply: {
        type: Object,
        required: true,
    },
});

const emits = defineEmits(["cancel"]);

const userStore = useUserStore();

// Trigger a lazy load so the reply author's name resolves in the template.
useUser(() => props.reply?.user_id);
</script>
