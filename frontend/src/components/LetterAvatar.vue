<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <span
        class="flex shrink-0 items-center justify-center font-semibold text-white"
        :class="colorClassFor(id)"
        aria-hidden="true"
    >
        {{ initialsFor(name) }}
    </span>
</template>

<script setup>
import { colorClassFor, initialsFor } from "@/utils/avatar";

// Size, text size and rounding come from the class the caller passes.
defineProps({
    id: {
        type: String,
        default: "",
    },
    name: {
        type: String,
        default: "",
    },
});
</script>
