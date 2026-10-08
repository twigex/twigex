<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div ref="emojiPickerRef" class="w-full h-96 flex flex-col">
        <div
            v-if="settingsStore.chatSettings.gif && !props.disableGifs"
            class="flex flex-row gap-x-2 mb-2"
        >
            <button
                @click="selected = 'emoji'"
                type="button"
                class="inline-flex items-center gap-x-1.5 rounded-md px-2.5 py-1.5 text-sm text-gray-600 font-semibold shadow-xs focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600"
                :class="selected === 'emoji' ? 'bg-indigo-100 ' : '  hover:bg-gray-100'"
            >
                {{ t("channels.expression_picker.emojis") }}
            </button>

            <button
                @click="selected = 'gif'"
                type="button"
                class="inline-flex items-center gap-x-1.5 rounded-md px-2.5 py-1.5 text-sm text-gray-600 font-semibold shadow-xs focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600"
                :class="selected === 'gif' ? 'bg-indigo-100 ' : 'hover:bg-gray-100'"
            >
                {{ t("channels.expression_picker.gifs") }}
            </button>
        </div>

        <div class="flex flex-col h-full max-h-96 min-h-0">
            <EmojiGrid v-if="selected === 'emoji'" @select-emoji="$emit('select-emoji', $event)" />
            <GifGrid
                v-else-if="selected == 'gif' && !props.disableGifs"
                @select-gif="$emit('select-gif', $event)"
            />
        </div>
    </div>
</template>

<script setup>
import { t } from "@/i18n/index.js";
import { ref } from "vue";
import EmojiGrid from "@/components/ExpressionPicker/EmojiGrid.vue";
import GifGrid from "@/components/ExpressionPicker/GifGrid.vue";
import { useSettingsStore } from "@/store/settings";

const props = defineProps({
    disableGifs: {
        type: Boolean,
        default: false,
    },
});

defineEmits(["select-emoji", "select-gif"]);

const settingsStore = useSettingsStore();
const selected = ref("emoji");
</script>
