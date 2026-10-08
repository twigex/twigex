<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <TransitionRoot as="template" :show="modelValue">
        <Dialog class="relative z-50" @close="close">
            <TransitionChild
                as="template"
                enter="ease-out duration-200"
                enter-from="opacity-0"
                enter-to="opacity-100"
                leave="ease-in duration-150"
                leave-from="opacity-100"
                leave-to="opacity-0"
            >
                <div class="fixed inset-0 bg-gray-900/40 backdrop-blur-sm" />
            </TransitionChild>

            <div class="fixed inset-0 z-10 flex items-center justify-center p-4">
                <TransitionChild
                    as="template"
                    enter="ease-out duration-200"
                    enter-from="opacity-0 scale-95"
                    enter-to="opacity-100 scale-100"
                    leave="ease-in duration-150"
                    leave-from="opacity-100 scale-100"
                    leave-to="opacity-0 scale-95"
                >
                    <DialogPanel
                        class="w-full max-w-xl overflow-hidden rounded-xl bg-white text-gray-900 shadow-2xl ring-1 ring-black/5"
                    >
                        <header class="flex items-center justify-between px-5 pt-5">
                            <DialogTitle
                                class="flex items-center gap-2 text-lg font-semibold text-gray-900"
                            >
                                <UserGroupIcon class="h-5 w-5 text-indigo-600" aria-hidden="true" />
                                {{ groupName }}
                            </DialogTitle>
                            <button
                                type="button"
                                class="rounded-md p-2 text-gray-400 hover:bg-gray-100 hover:text-gray-700"
                                @click="close"
                            >
                                ×
                            </button>
                        </header>

                        <div class="px-5 pt-4">
                            <div class="rounded-md border border-gray-200 bg-white px-3 py-2">
                                <input
                                    class="w-full bg-transparent text-sm text-gray-900 placeholder-gray-400 border-0 outline-none focus:outline-none focus:ring-0"
                                    spellcheck="false"
                                    :placeholder="t('group_members.search_placeholder')"
                                    :value="query"
                                    @input="onInput"
                                />
                            </div>
                        </div>

                        <div class="mt-3 px-5 pb-5">
                            <div class="relative h-72">
                                <div
                                    v-if="loading && members.length === 0"
                                    class="absolute inset-0 flex items-center justify-center text-indigo-600"
                                >
                                    <BaseSpinner size="lg" />
                                </div>

                                <div
                                    v-else-if="members.length === 0"
                                    class="absolute inset-0 flex items-center justify-center text-sm text-gray-500"
                                >
                                    {{ t("group_members.empty") }}
                                </div>

                                <RecycleScroller
                                    v-else
                                    class="h-full overflow-y-auto"
                                    :items="members"
                                    :item-size="56"
                                    :buffer="200"
                                    key-field="id"
                                    v-slot="{ item }"
                                    @scroll-end="loadMore"
                                >
                                    <div
                                        class="flex items-center gap-3 px-1 py-2"
                                        style="height: 56px"
                                    >
                                        <div class="h-8 w-8 shrink-0">
                                            <UserAvatar
                                                :user="item.user_info"
                                                :user-id="item.user_id"
                                            />
                                        </div>
                                        <div class="min-w-0 flex-1">
                                            <div class="truncate text-sm font-medium text-gray-900">
                                                {{ item.user_info?.name }}
                                                {{ item.user_info?.lastname }}
                                            </div>
                                            <div class="truncate text-xs text-gray-500">
                                                {{ item.user_info?.email }}
                                            </div>
                                        </div>
                                    </div>
                                </RecycleScroller>

                                <!-- Load-more overlay: absolute, so the list
                                     never reflows when paging in more rows. -->
                                <transition
                                    enter-active-class="transition ease-out duration-150"
                                    enter-from-class="opacity-0 translate-y-1"
                                    enter-to-class="opacity-100 translate-y-0"
                                    leave-active-class="transition ease-in duration-100"
                                    leave-from-class="opacity-100"
                                    leave-to-class="opacity-0"
                                >
                                    <div
                                        v-if="loading && members.length > 0"
                                        class="pointer-events-none absolute inset-x-0 bottom-0 flex justify-center pb-2"
                                    >
                                        <span
                                            class="flex items-center justify-center rounded-full bg-white/90 p-1.5 text-indigo-600 shadow-sm ring-1 ring-black/5 backdrop-blur-sm"
                                        >
                                            <BaseSpinner size="sm" />
                                        </span>
                                    </div>
                                </transition>
                            </div>
                        </div>
                    </DialogPanel>
                </TransitionChild>
            </div>
        </Dialog>
    </TransitionRoot>
</template>

<script setup>
import { t } from "@/i18n/index.js";
import { watch } from "vue";
import { Dialog, DialogPanel, DialogTitle, TransitionChild, TransitionRoot } from "@headlessui/vue";
import { UserGroupIcon } from "@heroicons/vue/24/outline";
import { RecycleScroller } from "vue-virtual-scroller";
import "vue-virtual-scroller/dist/vue-virtual-scroller.css";
import BaseSpinner from "@/components/BaseSpinner.vue";
import UserAvatar from "@/components/UserAvatar.vue";
import { useGroupMembers } from "@/composables/useGroupMembers";

const props = defineProps({
    modelValue: { type: Boolean, required: true },
    groupId: { type: String, default: "" },
    groupName: { type: String, default: "" },
});

const emit = defineEmits(["update:modelValue"]);

const { members, query, loading, start, search, loadMore, reset } = useGroupMembers(
    () => props.groupId,
);

watch(
    () => props.modelValue,
    (open) => {
        if (open) start();
        else reset();
    },
);

function onInput(e) {
    search(e.target.value);
}

function close() {
    emit("update:modelValue", false);
}
</script>
