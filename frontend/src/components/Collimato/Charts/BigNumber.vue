<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <dl class="w-full flex flex-wrap items-end justify-center gap-x-10 gap-y-4">
        <div v-for="entry in values" :key="entry.label" class="flex flex-col items-center">
            <dd class="text-3xl font-medium leading-10 tracking-tight text-gray-900">
                {{ entry.formatted }}
            </dd>
            <dt v-if="values.length > 1" class="text-sm text-gray-500 text-center">
                {{ entry.label }}
            </dt>
        </div>
    </dl>
</template>

<script setup>
import { computed } from "vue";

const props = defineProps({
    options: {
        type: Object,
        required: true,
    },
});

const values = computed(() => props.options?.values ?? []);
</script>
