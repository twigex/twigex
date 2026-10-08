<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <span class="relative inline-block h-full w-full">
        <img
            v-if="photo && !photoFailed"
            class="h-full w-full flex-none rounded-full bg-gray-50"
            :src="photoSrc"
            alt=""
            @error="photoFailed = true"
        />
        <span
            v-else-if="loading"
            class="block h-full w-full animate-pulse rounded-full bg-gray-200"
        />
        <span
            v-else-if="!displayName"
            class="block h-full w-full overflow-hidden rounded-full bg-gray-100"
        >
            <svg
                class="h-full w-full text-gray-300"
                fill="currentColor"
                viewBox="0 0 24 24"
                aria-hidden="true"
            >
                <path
                    d="M24 20.993V24H0v-2.996A14.977 14.977 0 0112.004 15c4.904 0 9.26 2.354 11.996 5.993zM16.002 8.999a4 4 0 11-8 0 4 4 0 018 0z"
                />
            </svg>
        </span>
        <span
            v-else
            class="flex h-full w-full items-center justify-center rounded-full text-white"
            :class="colorClass"
        >
            <svg class="h-full w-full" viewBox="0 0 10 10" aria-hidden="true">
                <text
                    x="5"
                    y="5"
                    text-anchor="middle"
                    dominant-baseline="central"
                    font-size="6"
                    font-weight="400"
                    fill="currentColor"
                >
                    {{ initial }}
                </text>
            </svg>
        </span>
        <span
            v-if="status"
            class="absolute right-0 top-0 block size-[20%] translate-x-[25%] -translate-y-[25%] transform rounded-full ring-2 ring-white"
            :class="statusColors[getUserStatus()]"
        />
    </span>
</template>

<script setup>
import { computed, ref, watch } from "vue";
import { useUserStore } from "@/store/user";
import { useUserState } from "@/composables/useUser";
import { colorClassFor, initialsFor } from "@/utils/avatar";

const props = defineProps({
    user: {
        type: Object,
        default: null,
    },
    userId: {
        type: String,
        default: "",
    },
    name: {
        type: String,
        default: "",
    },
    status: {
        type: Boolean,
        default: false,
    },
});

const statusColors = {
    online: "bg-green-500",
    away: "bg-yellow-500",
    offline: "bg-gray-300",
    dnd: "bg-red-500",
};

const userStore = useUserStore();

const id = computed(() => props.user?.id || props.userId);

const { user: lookedUpUser, loading } = useUserState(() =>
    props.user ? null : props.userId || null,
);

const resolved = computed(() => {
    if (id.value && userStore.user?.id === id.value) {
        return userStore.user;
    }

    if (props.user) {
        return Object.hasOwn(props.user, "photo")
            ? props.user
            : userStore.getUserById(id.value) || props.user;
    }

    return lookedUpUser.value || {};
});

const photo = computed(() => Boolean(resolved.value.photo));
const photoSrc = computed(() => userStore.getPhotoSrc(id.value, resolved.value.photo));
const photoFailed = ref(false);

watch([() => resolved.value.photo, photoSrc], () => {
    photoFailed.value = false;
});

const displayName = computed(() =>
    (props.name || props.user?.name || resolved.value.name || "").trim(),
);

const initial = computed(() => initialsFor(displayName.value));

const colorClass = computed(() => colorClassFor(id.value));

function getUserStatus() {
    return userStore.getUserStatus(id.value).status;
}
</script>
