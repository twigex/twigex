<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <li
        class="group relative flex cursor-pointer flex-col rounded-lg bg-white shadow-sm ring-1 ring-black/5 transition-shadow hover:shadow-md"
    >
        <div class="flex items-start gap-x-3 px-4 pt-4">
            <slot name="icon" />
            <div class="min-w-0 flex-auto">
                <p class="truncate text-sm font-semibold text-gray-900" :title="title">
                    {{ title }}
                </p>
                <div v-if="$slots.subtitle" class="mt-0.5">
                    <slot name="subtitle" />
                </div>
            </div>
            <div
                v-if="$slots.actions"
                class="flex flex-none items-center gap-x-1 opacity-0 focus-within:opacity-100 group-hover:opacity-100 [@media(hover:none)]:opacity-100"
                @click.stop
            >
                <slot name="actions" />
            </div>
        </div>

        <p
            v-if="descriptionFallback"
            :class="[
                description ? 'text-gray-600' : 'text-gray-400',
                'mt-3 line-clamp-2 min-h-[2.5rem] px-4 text-sm',
            ]"
        >
            {{ description || descriptionFallback }}
        </p>

        <slot />

        <div
            v-if="$slots.footer"
            class="mt-4 flex items-center justify-between gap-x-2 border-t border-gray-100 px-4 py-2.5"
        >
            <slot name="footer" />
        </div>
    </li>
</template>

<script setup>
// descriptionFallback turns the description line on, for cards whose items
// have descriptions; it is shown while the item has none.
defineProps({
    title: {
        type: String,
        default: "",
    },
    description: {
        type: String,
        default: "",
    },
    descriptionFallback: {
        type: String,
        default: "",
    },
});
</script>
