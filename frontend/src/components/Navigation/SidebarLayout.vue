<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div class="flex h-full w-full">
        <aside
            class="hidden lg:block flex-shrink-0 overflow-hidden border-r bg-white transition-[width] duration-300 ease-in-out"
            :style="{ width: collapsed ? '3.5rem' : '18rem' }"
        >
            <div
                class="sticky top-14 flex h-[calc(100vh-3.5rem)] flex-col overflow-y-auto overflow-x-hidden"
            >
                <div
                    :class="[
                        collapsed ? 'justify-center' : 'justify-between pl-5 pr-3',
                        'flex h-11 shrink-0 items-center gap-x-2',
                    ]"
                >
                    <span v-if="!collapsed" class="truncate text-sm font-semibold text-gray-900">
                        {{ title }}
                    </span>
                    <button
                        type="button"
                        class="flex size-7 shrink-0 items-center justify-center rounded-md bg-indigo-50 text-indigo-600 hover:bg-indigo-100 hover:text-indigo-700"
                        :title="toggleLabel"
                        :aria-label="toggleLabel"
                        :aria-expanded="!collapsed"
                        @click="navigationStore.open = !navigationStore.open"
                    >
                        <ChevronDoubleRightIcon
                            v-if="collapsed"
                            class="size-5"
                            aria-hidden="true"
                        />
                        <ChevronDoubleLeftIcon v-else class="size-5" aria-hidden="true" />
                    </button>
                </div>

                <!-- One element in both states, so the navigation is not
                     mounted again, and its data not loaded again, on toggle. -->
                <div class="relative flex min-h-0 flex-1 flex-col">
                    <button
                        v-if="collapsed && canScrollUp"
                        type="button"
                        class="absolute inset-x-0 top-0 z-20 flex h-8 items-center justify-center bg-gradient-to-b from-white via-white to-transparent text-gray-400 hover:text-indigo-600"
                        :title="t('main.navigation.scroll_up')"
                        :aria-label="t('main.navigation.scroll_up')"
                        @click="scrollStrip(-1)"
                    >
                        <ChevronUpIcon class="size-5" aria-hidden="true" />
                    </button>

                    <div
                        ref="scroller"
                        :class="[
                            collapsed
                                ? 'w-14 px-2 [scrollbar-width:none] [&::-webkit-scrollbar]:hidden'
                                : 'w-72 px-3',
                            'isolate flex min-h-0 flex-1 flex-col overflow-y-auto bg-white pb-4',
                        ]"
                        @scroll="updateScrollButtons"
                    >
                        <slot name="sidebar" :mini="collapsed" />
                    </div>

                    <button
                        v-if="collapsed && canScrollDown"
                        type="button"
                        class="absolute inset-x-0 bottom-0 z-20 flex h-8 items-center justify-center bg-gradient-to-t from-white via-white to-transparent text-gray-400 hover:text-indigo-600"
                        :title="t('main.navigation.scroll_down')"
                        :aria-label="t('main.navigation.scroll_down')"
                        @click="scrollStrip(1)"
                    >
                        <ChevronDownIcon class="size-5" aria-hidden="true" />
                    </button>
                </div>
            </div>
        </aside>

        <div class="flex-1 min-w-0 h-full">
            <slot />
        </div>

        <MobileDrawer>
            <div class="w-72 h-full flex">
                <div class="flex grow flex-col overflow-y-auto bg-white px-3 pb-4">
                    <slot name="sidebar" :mini="false" />
                </div>
            </div>
        </MobileDrawer>
    </div>
</template>

<script setup>
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from "vue";
import {
    ChevronDoubleLeftIcon,
    ChevronDoubleRightIcon,
    ChevronDownIcon,
    ChevronUpIcon,
} from "@heroicons/vue/20/solid";
import { t } from "@/i18n/index.js";
import { useNavigationStore } from "@/store/navigation";
import MobileDrawer from "@/components/Navigation/MobileDrawer.vue";

// The sidebar slot gets mini, to draw its icon-only form.
defineProps({
    title: {
        type: String,
        default: "",
    },
});

const navigationStore = useNavigationStore();

const collapsed = computed(() => !navigationStore.open);

const toggleLabel = computed(() =>
    collapsed.value
        ? t.value("main.navigation.expand_sidebar")
        : t.value("main.navigation.collapse_sidebar"),
);

const scroller = ref(null);
const canScrollUp = ref(false);
const canScrollDown = ref(false);

function updateScrollButtons() {
    const el = scroller.value;

    if (!el) return;

    canScrollUp.value = el.scrollTop > 0;
    canScrollDown.value = el.scrollTop + el.clientHeight < el.scrollHeight - 1;
}

function scrollStrip(direction) {
    const el = scroller.value;

    if (!el) return;

    el.scrollBy({ top: direction * (el.clientHeight - 64), behavior: "smooth" });
}

let resizeObserver = null;
let mutationObserver = null;

watch(collapsed, () => nextTick(updateScrollButtons));

// A count, not a flag: on a route change the next section's layout mounts
// before the previous one unmounts. Logout resets the store while a layout is
// still mounted, so the count must not go below zero.
onMounted(() => {
    navigationStore.sidebars++;

    resizeObserver = new ResizeObserver(updateScrollButtons);
    resizeObserver.observe(scroller.value);

    mutationObserver = new MutationObserver(updateScrollButtons);
    mutationObserver.observe(scroller.value, { childList: true, subtree: true });

    updateScrollButtons();
});

onUnmounted(() => {
    navigationStore.sidebars = Math.max(0, navigationStore.sidebars - 1);

    resizeObserver?.disconnect();
    mutationObserver?.disconnect();
});
</script>
