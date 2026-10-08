<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div class="flex min-w-0 items-center">
        <div :class="avatarClass">
            <UserAvatar :user-id="userId" />
        </div>
        <div :class="textClass">
            <span
                v-if="loading"
                class="inline-block h-[0.8em] w-20 animate-pulse rounded bg-gray-200 align-middle"
            />
            <slot v-else-if="user" :user="user">{{ label }}</slot>
            <template v-else>{{ t("common.label.unknown_user") }}</template>
        </div>
    </div>
</template>

<script setup>
import { computed } from "vue";
import { t } from "@/i18n/index.js";
import { useUserState } from "@/composables/useUser";
import UserAvatar from "@/components/UserAvatar.vue";

const props = defineProps({
    userId: {
        type: String,
        default: "",
    },
    avatarClass: {
        type: String,
        default: "size-8 shrink-0",
    },
    textClass: {
        type: String,
        default: "truncate text-sm text-gray-900",
    },
    text: {
        type: String,
        default: "fullName",
        validator: (value) =>
            ["fullName", "email", "username", "fullNameWithEmail"].includes(value),
    },
});

const { user, loading } = useUserState(() => props.userId || null);

const label = computed(() => {
    const u = user.value;
    const fullName = `${u.name ?? ""} ${u.lastname ?? ""}`.trim();

    let value;

    switch (props.text) {
        case "email":
            value = u.email;
            break;
        case "username":
            value = u.username;
            break;
        case "fullNameWithEmail":
            value = u.email ? `${fullName} (${u.email})`.trim() : fullName;
            break;
        default:
            value = fullName;
    }

    return value || t.value("common.label.unknown_user");
});
</script>
