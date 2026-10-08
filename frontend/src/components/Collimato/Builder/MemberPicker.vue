<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div>
        <p v-if="members === null" class="rounded-md bg-gray-50 px-3 py-2 text-xs text-gray-500">
            {{ t("collimato.view_builder.all_members") }}
        </p>
        <template v-else-if="members.length">
            <div class="mb-1 flex items-center gap-3 text-xs">
                <button
                    type="button"
                    class="font-medium text-indigo-600 hover:text-indigo-500"
                    @click="$emit('all')"
                >
                    {{ t("collimato.view_builder.select_all") }}
                </button>
                <button
                    type="button"
                    class="font-medium text-gray-500 hover:text-gray-700"
                    @click="$emit('clear')"
                >
                    {{ t("collimato.view_builder.clear") }}
                </button>
                <span class="ml-auto text-gray-400">
                    {{
                        t("collimato.view_builder.selected_count", {
                            n: selected.length,
                            total: members.length,
                        })
                    }}
                </span>
            </div>
            <div class="flex flex-wrap gap-2">
                <label
                    v-for="m in members"
                    :key="m"
                    class="flex cursor-pointer items-center gap-1.5 rounded-md px-2 py-1 text-sm ring-1 ring-inset"
                    :class="
                        selected.includes(m)
                            ? 'bg-indigo-50 text-indigo-700 ring-indigo-200'
                            : 'text-gray-600 ring-gray-200 hover:bg-gray-50'
                    "
                >
                    <input
                        type="checkbox"
                        class="h-3.5 w-3.5 rounded border-gray-300 text-indigo-600 focus:ring-indigo-600"
                        :checked="selected.includes(m)"
                        @change="$emit('toggle', m)"
                    />
                    <span class="font-mono">{{ m }}</span>
                </label>
            </div>
        </template>
        <p v-else class="text-sm text-gray-400">
            {{ t("collimato.view_builder.no_members") }}
        </p>
    </div>
</template>

<script setup>
import { t } from "@/i18n/index.js";

defineProps({
    // Selectable member names, or null when the source cube's members are unknown.
    members: { type: Array, default: null },
    selected: { type: Array, default: () => [] },
});
defineEmits(["toggle", "all", "clear"]);
</script>
