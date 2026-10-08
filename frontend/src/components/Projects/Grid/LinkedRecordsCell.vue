<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div class="flex min-w-0 flex-1 items-center gap-x-1">
        <button
            v-if="records.length"
            type="button"
            class="min-w-0 truncate rounded-full px-2 py-0.5 text-xs font-medium"
            :class="
                records[0].restricted
                    ? 'cursor-default bg-gray-50 italic text-gray-400'
                    : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
            "
            :title="titleOf(records[0])"
            @click.stop="open(records[0])"
        >
            {{ nameOf(records[0]) }}
        </button>

        <Menu v-if="records.length > 1" as="div" class="flex flex-none" @click.stop>
            <MenuButton
                ref="button"
                class="rounded-full bg-white px-1.5 py-0.5 text-xs font-medium text-gray-600 ring-1 ring-inset ring-gray-300 hover:bg-gray-50 focus:outline-none focus-visible:ring-2 focus-visible:ring-indigo-600"
                :title="t('projects.grid_view.show_linked_records')"
            >
                +{{ records.length - 1 }}
            </MenuButton>
            <Teleport to="body">
                <transition
                    enter-active-class="transition ease-out duration-100"
                    enter-from-class="transform opacity-0 scale-95"
                    enter-to-class="transform opacity-100 scale-100"
                    leave-active-class="transition ease-in duration-75"
                    leave-from-class="transform opacity-100 scale-100"
                    leave-to-class="transform opacity-0 scale-95"
                >
                    <MenuItems
                        ref="floating"
                        :style="floatingStyles"
                        class="z-[60] max-h-60 w-56 overflow-y-auto rounded-md bg-white py-1 shadow-lg ring-1 ring-black/5 focus:outline-none"
                    >
                        <p
                            class="px-4 pb-1 pt-1.5 text-xs font-semibold uppercase tracking-wide text-gray-500"
                        >
                            {{
                                t("projects.grid_view.linked_records_count", {
                                    count: records.length,
                                })
                            }}
                        </p>
                        <MenuItem
                            v-for="(record, index) in records"
                            :key="record.id ?? index"
                            v-slot="{ active }"
                        >
                            <button
                                type="button"
                                :class="[
                                    record.restricted
                                        ? 'cursor-default italic text-gray-400'
                                        : active
                                          ? 'bg-gray-100 text-gray-900'
                                          : 'text-gray-700',
                                    'block w-full truncate px-4 py-2 text-left text-sm',
                                ]"
                                :title="titleOf(record)"
                                @click="open(record)"
                            >
                                {{ nameOf(record) }}
                            </button>
                        </MenuItem>
                    </MenuItems>
                </transition>
            </Teleport>
        </Menu>
    </div>
</template>

<script setup>
import { ref } from "vue";
import { Menu, MenuButton, MenuItem, MenuItems } from "@headlessui/vue";
import { t } from "@/i18n/index.js";
import { useAnchoredPopup } from "@/composables/useAnchoredPopup";

defineProps({
    records: { type: Array, default: () => [] },
});

const emit = defineEmits(["open"]);

const button = ref(null);
const { floating, floatingStyles } = useAnchoredPopup({
    placement: "bottom-start",
    anchor: button,
});

// A linked task the user may not see comes without its name, and is not
// opened.
const nameOf = (record) =>
    record?.restricted
        ? t.value("projects.grid_view.restricted_record")
        : record?.name || t.value("projects.grid_view.unnamed_record");

const titleOf = (record) =>
    record?.restricted ? t.value("projects.grid_view.restricted_record_hint") : nameOf(record);

function open(record) {
    if (!record?.restricted) emit("open", record);
}
</script>
