<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <span
        class="inline-block animate-spin rounded-full border-current border-t-transparent"
        :class="[sizeClass, borderClass]"
        role="status"
        :aria-label="label"
    />
</template>

<script setup>
import { computed } from "vue";

const props = defineProps({
    // sm | md | lg | xl
    size: { type: String, default: "md" },
    label: { type: String, default: "Loading" },
});

const sizeClass = computed(
    () =>
        ({
            sm: "h-4 w-4",
            md: "h-6 w-6",
            lg: "h-8 w-8",
            xl: "h-10 w-10",
        })[props.size] ?? "h-6 w-6",
);

const borderClass = computed(() => (props.size === "sm" ? "border-2" : "border-[3px]"));
</script>
