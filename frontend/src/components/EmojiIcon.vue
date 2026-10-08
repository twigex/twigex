<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <span
        v-if="emoji"
        role="img"
        :aria-label="shortName"
        :title="`:${shortName}:`"
        class="inline-block size-[1.375em] select-text overflow-hidden bg-contain bg-no-repeat align-[-0.3em] text-transparent"
        :style="{ backgroundImage: `url(${emojiImageUrl(emoji)})` }"
        >{{ emojiCharacter(emoji) }}</span
    >
    <template v-else>:{{ shortName }}:</template>
</template>

<script setup>
import { computed } from "vue";
import { emojiByShortName, emojiCharacter, emojiImageUrl } from "@/utils/emoji";

const props = defineProps({
    shortName: { type: String, required: true },
});

const emoji = computed(() => emojiByShortName(props.shortName));
</script>
