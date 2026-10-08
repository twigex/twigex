<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <li class="flex items-center justify-between gap-x-6 py-1">
        <div class="flex min-w-0 gap-x-4">
            <div class="h-12 w-12 flex-none">
                <UserAvatar :user="user" />
            </div>
            <div class="min-w-0 flex-auto">
                <p class="text-sm font-semibold leading-6 text-gray-900 text-left">
                    {{ user.name + " " + user.lastname }}
                </p>
                <p class="mt-1 truncate text-xs leading-5 text-gray-500 text-left">
                    {{ user.email }}
                </p>
            </div>
        </div>
    </li>
</template>

<script setup>
import UserAvatar from "@/components/UserAvatar.vue";

defineProps({
    user: {
        type: Object,
        required: true,
    },
});
</script>
