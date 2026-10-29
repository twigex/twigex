<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div class="mt-5 space-y-4">
        <h3 class="text-base font-semibold leading-6 text-gray-600">
            {{ t("files.share_dialog.permissions") }}
        </h3>

        <div>
            <label class="mb-1 block text-sm font-medium text-gray-900">
                {{ t("files.share_dialog.expires") }}
            </label>
            <div class="flex items-center gap-2">
                <Popover as="div" class="relative flex-1">
                    <PopoverButton
                        ref="reference"
                        class="flex w-full items-center rounded-md bg-white py-2 pl-3 pr-10 text-left text-sm text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 focus:outline-none focus:ring-2 focus:ring-indigo-600"
                    >
                        <span
                            class="block flex-1 truncate"
                            :class="localExpiry ? 'text-gray-900' : 'text-gray-400'"
                        >
                            {{ localExpiry || t("files.share_dialog.no_expiration") }}
                        </span>
                        <span
                            class="pointer-events-none absolute inset-y-0 right-0 flex items-center pr-3"
                        >
                            <CalendarDaysIcon class="h-5 w-5 text-gray-400" aria-hidden="true" />
                        </span>
                    </PopoverButton>
                    <Teleport to="body">
                        <PopoverPanel
                            ref="floating"
                            :style="floatingStyles"
                            class="z-[100] rounded-xl bg-white p-3 shadow-lg ring-1 ring-black/5"
                            v-slot="{ close }"
                        >
                            <DatePicker
                                :modelValue="localExpiry"
                                @update:modelValue="
                                    (val) => {
                                        localExpiry = val;
                                        close();
                                    }
                                "
                            />
                        </PopoverPanel>
                    </Teleport>
                </Popover>
                <button
                    v-if="localExpiry"
                    type="button"
                    class="flex h-9 w-9 shrink-0 items-center justify-center rounded-md text-gray-400 ring-1 ring-inset ring-gray-300 hover:bg-gray-50 hover:text-gray-600"
                    :title="t('common.button.clear')"
                    @click="localExpiry = ''"
                >
                    <XMarkIcon class="h-5 w-5" aria-hidden="true" />
                </button>
            </div>
        </div>

        <div>
            <label class="mb-1 block text-sm font-medium text-gray-900">
                {{ t("files.share_dialog.permission_level") }}
            </label>
            <div class="divide-y divide-gray-200 rounded-lg border border-gray-200">
                <label
                    v-for="opt in levelOptions"
                    :key="opt.value"
                    class="flex cursor-pointer items-center gap-3 px-4 py-3"
                >
                    <input
                        type="radio"
                        class="h-4 w-4 text-indigo-600 focus:ring-indigo-600"
                        :value="opt.value"
                        v-model="localAccessLevel"
                    />
                    <span class="flex min-w-0 flex-1 flex-col">
                        <span class="text-sm font-medium text-gray-900">
                            {{ opt.label }}
                        </span>
                        <span class="text-sm text-gray-500">
                            {{ opt.description }}
                        </span>
                    </span>
                </label>
            </div>
        </div>
    </div>
</template>

<script setup>
import { t } from "@/i18n/index.js";

import { ref, computed } from "vue";
import { Popover, PopoverButton, PopoverPanel } from "@headlessui/vue";
import { useFloating, flip, shift, offset, autoUpdate } from "@floating-ui/vue";
import { CalendarDaysIcon, XMarkIcon } from "@heroicons/vue/24/outline";
import DatePicker from "@/components/DatePicker/DatePicker.vue";
import useDateOperations from "@/composables/useDateOperations.js";

const props = defineProps({
    expiry: {
        type: Number, // unix seconds (UTC), 0 = no expiration
        default: 0,
    },
    accessLevel: {
        type: String,
        default: "viewer",
    },
});

const emit = defineEmits(["update:expiry", "update:accessLevel"]);

const levelOptions = computed(() => [
    {
        value: "viewer",
        label: t.value("files.share_dialog.role_viewer"),
        description: t.value("files.share_dialog.role_viewer_description"),
    },
    {
        value: "editor",
        label: t.value("files.share_dialog.role_editor"),
        description: t.value("files.share_dialog.role_editor_description"),
    },
    {
        value: "manager",
        label: t.value("files.share_dialog.role_manager"),
        description: t.value("files.share_dialog.role_manager_description"),
    },
]);

const reference = ref(null);
const floating = ref(null);
const { floatingStyles } = useFloating(reference, floating, {
    strategy: "fixed",
    middleware: [offset(4), flip(), shift({ padding: 8 })],
    whileElementsMounted: autoUpdate,
});

const { getDateInputValue, getEndOfDayTimestamp } = useDateOperations();

const localExpiry = computed({
    get() {
        return getDateInputValue(props.expiry, false);
    },
    set(ymd) {
        emit("update:expiry", getEndOfDayTimestamp(ymd));
    },
});

const localAccessLevel = computed({
    get: () => props.accessLevel,
    set: (v) => emit("update:accessLevel", v),
});
</script>
