<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div
        class="flex shrink-0 flex-wrap items-center justify-between gap-3 border-b border-gray-200 py-4"
    >
        <div class="min-w-0">
            <nav class="flex min-w-0 flex-wrap items-center gap-x-1 text-sm text-gray-500">
                <button
                    type="button"
                    class="max-w-[12rem] truncate font-medium text-gray-900 hover:text-indigo-600"
                    @click="emit('navigate', -1)"
                >
                    {{ rootName }}
                </button>
                <template v-for="(crumb, i) in folderPath" :key="crumb.id">
                    <span aria-hidden="true">/</span>
                    <button
                        type="button"
                        class="max-w-[12rem] truncate hover:text-gray-700"
                        @click="emit('navigate', i)"
                    >
                        {{ crumb.name }}
                    </button>
                </template>
            </nav>
            <p class="mt-1 text-xs text-gray-500">
                {{ t("public_share.shared_by") }}
                {{ share.User.Name }} {{ share.User.LastName }} ·
                {{
                    t("public_share.items", {
                        count: itemCount,
                    })
                }}
            </p>
        </div>
        <div class="flex shrink-0 items-center gap-2">
            <Listbox
                :model-value="sort"
                as="div"
                class="relative"
                @update:model-value="emit('update:sort', $event)"
            >
                <ListboxButton
                    class="flex items-center gap-x-1 rounded-md bg-white py-2 pl-3 pr-2 text-left text-sm text-gray-700 shadow-sm ring-1 ring-inset ring-gray-300 focus:outline-none focus:ring-2 focus:ring-indigo-600"
                >
                    <span class="truncate">{{ currentSortLabel }}</span>
                    <ChevronUpDownIcon class="h-5 w-5 text-gray-400" aria-hidden="true" />
                </ListboxButton>
                <transition
                    leave-active-class="transition ease-in duration-100"
                    leave-from-class="opacity-100"
                    leave-to-class="opacity-0"
                >
                    <ListboxOptions
                        class="absolute right-0 z-20 mt-1 w-44 overflow-auto rounded-md bg-white py-1 text-sm shadow-lg ring-1 ring-black/5 focus:outline-none"
                    >
                        <ListboxOption
                            v-for="option in sortOptions"
                            :key="option.value"
                            :value="option.value"
                            as="template"
                            v-slot="{ active, selected }"
                        >
                            <li
                                :class="[
                                    active ? 'bg-indigo-600 text-white' : 'text-gray-900',
                                    'relative cursor-pointer select-none py-2 pl-3 pr-9',
                                ]"
                            >
                                <span
                                    :class="[
                                        selected ? 'font-semibold' : 'font-normal',
                                        'block truncate',
                                    ]"
                                    >{{ option.label }}</span
                                >
                                <span
                                    v-if="selected"
                                    :class="[
                                        active ? 'text-white' : 'text-indigo-600',
                                        'absolute inset-y-0 right-0 flex items-center pr-3',
                                    ]"
                                >
                                    <CheckIcon class="h-5 w-5" aria-hidden="true" />
                                </span>
                            </li>
                        </ListboxOption>
                    </ListboxOptions>
                </transition>
            </Listbox>
            <div class="flex rounded-md bg-gray-100 p-0.5">
                <button
                    type="button"
                    @click="emit('update:viewMode', 'grid')"
                    :class="[
                        'rounded p-1.5',
                        viewMode === 'grid'
                            ? 'bg-white text-gray-900 shadow-sm'
                            : 'text-gray-500 hover:text-gray-700',
                    ]"
                >
                    <Squares2X2Icon class="h-5 w-5" aria-hidden="true" />
                </button>
                <button
                    type="button"
                    @click="emit('update:viewMode', 'list')"
                    :class="[
                        'rounded p-1.5',
                        viewMode === 'list'
                            ? 'bg-white text-gray-900 shadow-sm'
                            : 'text-gray-500 hover:text-gray-700',
                    ]"
                >
                    <Bars3Icon class="h-5 w-5" aria-hidden="true" />
                </button>
            </div>
            <a
                v-if="share.AllowDownload"
                :href="publicShareService.downloadUrl(token, currentChildId)"
                class="rounded-md bg-white px-3 py-2 text-sm font-semibold text-indigo-600 shadow-sm ring-1 ring-inset ring-indigo-200 hover:bg-indigo-50"
            >
                {{ t("public_share.download_all") }}
            </a>
            <button
                v-if="share.AllowUpload"
                type="button"
                @click="emit('upload')"
                :disabled="uploading"
                class="rounded-md bg-indigo-600 px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-indigo-500 disabled:opacity-50"
            >
                {{ uploadLabel }}
            </button>
        </div>
    </div>
</template>

<script setup>
import { computed } from "vue";
import { Listbox, ListboxButton, ListboxOptions, ListboxOption } from "@headlessui/vue";
import { Squares2X2Icon, Bars3Icon } from "@heroicons/vue/24/outline";
import { ChevronUpDownIcon, CheckIcon } from "@heroicons/vue/20/solid";
import { t } from "@/i18n";
import publicShareService from "@/services/publicShareService";

const props = defineProps({
    share: {
        type: Object,
        required: true,
    },
    token: {
        type: String,
        required: true,
    },
    rootName: {
        type: String,
        default: "",
    },
    folderPath: {
        type: Array,
        default: () => [],
    },
    itemCount: {
        type: Number,
        default: 0,
    },
    currentChildId: {
        type: String,
        default: "",
    },
    sort: {
        type: String,
        default: "name-asc",
    },
    viewMode: {
        type: String,
        default: "grid",
    },
    uploading: {
        type: Boolean,
        default: false,
    },
    uploadLabel: {
        type: String,
        default: "",
    },
});

const emit = defineEmits(["navigate", "update:sort", "update:viewMode", "upload"]);

const sortOptions = computed(() => [
    { value: "name-asc", label: t.value("public_share.sort_name_asc") },
    { value: "name-desc", label: t.value("public_share.sort_name_desc") },
    { value: "size-asc", label: t.value("public_share.sort_size_asc") },
    { value: "size-desc", label: t.value("public_share.sort_size_desc") },
]);
const currentSortLabel = computed(
    () => sortOptions.value.find((o) => o.value === props.sort)?.label ?? "",
);
</script>
