<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div
        class="absolute bottom-full left-0 mb-2 z-20 w-80 rounded-xl border border-gray-200 bg-white shadow-lg overflow-hidden"
    >
        <div class="max-h-64 overflow-y-auto divide-y divide-gray-100">
            <!-- Channel members -->
            <div v-if="props.users.length > 0">
                <div
                    class="px-3 pt-3 pb-1 text-xs font-semibold uppercase tracking-wider text-gray-400"
                >
                    {{ t("channels.input.channel_members") }}
                </div>
                <div
                    v-for="(user, index) in props.users"
                    :key="user.username"
                    @click="emits('select', user.username)"
                    class="flex items-center gap-x-3 px-3 py-2 cursor-pointer hover:bg-gray-50 transition-colors"
                    :class="{ 'bg-gray-50': props.selected === index }"
                >
                    <div class="h-7 w-7 shrink-0">
                        <UserAvatar :user="user" status />
                    </div>
                    <div class="min-w-0">
                        <span class="text-sm font-medium text-gray-800">@{{ user.username }}</span>
                        <span class="ml-1.5 text-xs text-gray-500 truncate"
                            >{{ user.name }} {{ user.lastname }}</span
                        >
                    </div>
                </div>
            </div>

            <!-- Special mentions -->
            <div v-if="props.specials.length > 0">
                <div
                    class="px-3 pt-3 pb-1 text-xs font-semibold uppercase tracking-wider text-gray-400"
                >
                    {{ t("channels.input.special_mentions") }}
                </div>
                <div
                    v-for="(name, index) in props.specials"
                    :key="name"
                    @click="emits('select', name)"
                    class="flex items-center gap-x-3 px-3 py-2 cursor-pointer hover:bg-gray-50 transition-colors"
                    :class="{ 'bg-gray-50': props.selected === props.users.length + index }"
                >
                    <div
                        class="flex h-7 w-7 shrink-0 items-center justify-center rounded-full bg-indigo-100"
                    >
                        <span class="text-xs font-bold text-indigo-600">@</span>
                    </div>
                    <div>
                        <span class="text-sm font-medium text-gray-800">@{{ name }}</span>
                    </div>
                </div>
            </div>
        </div>
    </div>
</template>

<script setup>
import { t } from "@/i18n/index.js";
import UserAvatar from "@/components/UserAvatar.vue";

const props = defineProps({
    users: {
        type: Array,
        required: true,
    },
    specials: {
        type: Array,
        required: true,
    },
    selected: {
        type: Number,
        default: 0,
    },
});

const emits = defineEmits(["select"]);
</script>
