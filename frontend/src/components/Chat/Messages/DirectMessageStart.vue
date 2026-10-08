<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div class="flex flex-col items-center justify-center py-12 px-6 text-center">
        <div class="mb-4 size-16">
            <UserAvatar :user-id="partnerId" />
        </div>
        <h2 class="text-lg font-semibold text-gray-900 mb-1">
            {{ t("channels.direct_start.title", { name: partnerName }) }}
        </h2>
        <p class="text-sm text-gray-500 max-w-sm">
            {{ t("channels.direct_start.description", { name: partnerName }) }}
        </p>
    </div>
</template>

<script setup>
import { computed } from "vue";
import { t } from "@/i18n/index.js";
import { useUserStore } from "@/store/user";
import { useUserState } from "@/composables/useUser";
import UserAvatar from "@/components/UserAvatar.vue";

const props = defineProps({
    channel: {
        type: Object,
        required: true,
    },
});

const userStore = useUserStore();

const partnerId = computed(() => {
    const members = props.channel.channel_members || [];
    const other = members.find((member) => member.user_id !== userStore.user?.id);

    return other?.user_id ?? userStore.user?.id ?? "";
});

const { user: partner, loading } = useUserState(() => partnerId.value || null);

const partnerName = computed(() => {
    if (loading.value) return "";

    const name = `${partner.value?.name ?? ""} ${partner.value?.lastname ?? ""}`.trim();

    return name || t.value("common.label.unknown_user");
});
</script>
