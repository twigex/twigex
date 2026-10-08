<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div>
        <div class="rounded-md border border-gray-200 bg-white px-3 py-1">
            <div class="flex flex-wrap items-center gap-2">
                <span
                    v-for="user in selectedList"
                    :key="user.id"
                    class="inline-flex items-center gap-1 rounded-md bg-gray-100 px-2 py-2 text-xs font-medium text-gray-700"
                >
                    <div class="h-4 w-4 shrink-0">
                        <UserAvatar :user="user" />
                    </div>
                    <span class="truncate max-w-[120px]">
                        {{ user.name }} {{ user.lastname }}
                    </span>
                    <button
                        class="ml-1 text-gray-500 hover:text-gray-700"
                        @click.stop="remove(user)"
                    >
                        ×
                    </button>
                </span>

                <input
                    v-if="multiple || !selectedList.length"
                    ref="input"
                    class="flex-1 min-w-[18ch] bg-transparent text-sm text-gray-900 placeholder-gray-400 border-0 outline-none focus:outline-none focus:ring-0"
                    spellcheck="false"
                    :placeholder="placeholder || t('user_picker.search_placeholder')"
                    :value="search.query.value"
                    @input="onInput"
                />
            </div>
        </div>

        <div class="mt-4 h-60 overflow-y-auto">
            <div v-if="search.loading.value" class="py-6 text-center text-sm text-gray-500">
                {{ t("user_picker.searching") }}
            </div>

            <div
                v-else-if="visibleResults.length === 0"
                class="py-6 text-center text-sm text-gray-500"
            >
                {{ t("common.label.no_users") }}
            </div>

            <div v-else class="space-y-1">
                <div
                    v-for="user in visibleResults"
                    :key="user.id"
                    class="flex cursor-pointer items-center gap-3 rounded-md px-3 py-2 hover:bg-gray-100"
                    @click="toggle(user)"
                >
                    <span
                        v-if="multiple"
                        class="flex h-[18px] w-[18px] shrink-0 items-center justify-center rounded border"
                        :class="
                            isSelected(user) ? 'bg-indigo-600 border-indigo-600' : 'border-gray-300'
                        "
                    >
                        <svg
                            v-if="isSelected(user)"
                            viewBox="0 0 20 20"
                            class="h-4 w-4 text-white"
                            fill="currentColor"
                        >
                            <path
                                fill-rule="evenodd"
                                d="M16.7 5.3a1 1 0 0 1 0 1.4l-7.2 7.2a1 1 0 0 1-1.4 0L3.3 9.2a1 1 0 1 1 1.4-1.4l3.1 3.1 6.5-6.5a1 1 0 0 1 1.4 0Z"
                                clip-rule="evenodd"
                            />
                        </svg>
                    </span>
                    <span
                        v-else
                        class="flex h-[18px] w-[18px] shrink-0 items-center justify-center rounded-full border-2"
                        :class="
                            isSelected(user) ? 'border-indigo-600 bg-indigo-600' : 'border-gray-300'
                        "
                    >
                        <span v-if="isSelected(user)" class="h-2 w-2 rounded-full bg-white" />
                    </span>

                    <div class="h-8 w-8 shrink-0">
                        <UserAvatar :user="user" />
                    </div>

                    <div class="min-w-0 flex-1">
                        <div class="truncate text-sm font-medium text-gray-900">
                            {{ user.name }} {{ user.lastname }}
                        </div>
                        <div class="truncate text-xs text-gray-500">
                            {{ user.email }}
                        </div>
                    </div>
                </div>
            </div>
        </div>
    </div>
</template>

<script setup>
import { t } from "@/i18n/index.js";
import { ref, computed, onMounted } from "vue";
import UserAvatar from "@/components/UserAvatar.vue";
import { useUserSearch } from "@/composables/useUserSearch";

const props = defineProps({
    modelValue: { type: [Array, Object], default: () => [] },
    multiple: { type: Boolean, default: true },
    excludeIds: { type: [Array, Set], default: () => [] },
    placeholder: { type: String, default: "" },
    limit: { type: Number, default: 20 },
});

const emit = defineEmits(["update:modelValue"]);

const search = useUserSearch(props.limit);
const input = ref(null);

onMounted(() => {
    search.search("");
    input.value?.focus();
});

const excludeSet = computed(() =>
    props.excludeIds instanceof Set ? props.excludeIds : new Set(props.excludeIds),
);

const selectedList = computed(() => {
    if (props.multiple) return props.modelValue ?? [];

    return props.modelValue ? [props.modelValue] : [];
});

const visibleResults = computed(() =>
    search.results.value.filter((u) => !excludeSet.value.has(u.id)),
);

function onInput(e) {
    search.search(e.target.value);
}

function isSelected(user) {
    return selectedList.value.some((u) => u.id === user.id);
}

function toggle(user) {
    if (props.multiple) {
        const next = isSelected(user)
            ? selectedList.value.filter((u) => u.id !== user.id)
            : [...selectedList.value, user];

        emit("update:modelValue", next);
    } else {
        emit("update:modelValue", isSelected(user) ? null : user);
    }
}

function remove(user) {
    if (props.multiple) {
        emit(
            "update:modelValue",
            selectedList.value.filter((u) => u.id !== user.id),
        );
    } else {
        emit("update:modelValue", null);
    }
}
</script>
