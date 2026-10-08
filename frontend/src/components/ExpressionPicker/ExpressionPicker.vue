<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div v-if="isMobile">
        <button
            @click="((open = true), $emit('click'))"
            type="button"
            class="rounded-full p-1 text-gray-400 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600"
        >
            <FaceSmileIcon class="h-6 w-6" aria-hidden="true" />
        </button>
        <TransitionRoot as="template" :show="open">
            <Dialog class="relative z-10" @close="open = false">
                <TransitionChild
                    as="template"
                    enter="ease-out duration-300"
                    enter-from="opacity-0"
                    enter-to=""
                    leave="ease-in duration-200"
                    leave-from=""
                    leave-to="opacity-0"
                >
                    <div class="fixed inset-0 bg-gray-500/75 transition-opacity"></div>
                </TransitionChild>

                <div class="fixed inset-0 z-10 w-screen overflow-y-auto">
                    <div
                        class="flex min-h-full items-end justify-center p-4 text-center sm:items-center sm:p-0"
                    >
                        <TransitionChild
                            as="template"
                            enter="ease-out duration-300"
                            enter-from="opacity-0 translate-y-4 sm:translate-y-0 sm:scale-95"
                            enter-to=" translate-y-0 sm:scale-100"
                            leave="ease-in duration-200"
                            leave-from=" translate-y-0 sm:scale-100"
                            leave-to="opacity-0 translate-y-4 sm:translate-y-0 sm:scale-95"
                        >
                            <DialogPanel
                                class="relative transform overflow-hidden rounded-lg bg-white px-2 pt-1 pb-4 text-left shadow-xl transition-all sm:my-8 sm:w-full sm:max-w-lg sm:p-6"
                            >
                                <div>
                                    <div class="mt-3 text-center md:mt-0">
                                        <div class="flex flex-row items-center justify-between">
                                            <DialogTitle
                                                as="h3"
                                                class="text-base font-semibold text-gray-900"
                                                >Select emoji</DialogTitle
                                            >
                                            <button
                                                type="button"
                                                class="rounded-md bg-white text-gray-400 hover:text-gray-500 focus:outline-2 focus:outline-offset-2 focus:outline-indigo-600"
                                                @click="open = false"
                                            >
                                                <span class="sr-only">Close</span>
                                                <XMarkIcon class="size-6" aria-hidden="true" />
                                            </button>
                                        </div>

                                        <div class="mt-2">
                                            <ExpressionPickerTabs
                                                @select-emoji="
                                                    ($emit('select-emoji', $event), (open = false))
                                                "
                                                @select-gif="
                                                    ($emit('select-gif', $event), (open = false))
                                                "
                                            />
                                        </div>
                                    </div>
                                </div>
                            </DialogPanel>
                        </TransitionChild>
                    </div>
                </div>
            </Dialog>
        </TransitionRoot>
    </div>
    <Popover v-else v-slot="{ close }" class="relative">
        <PopoverButton
            @click="$emit('click')"
            ref="reference"
            class="hover:bg-gray-100 text-gray-400 size-8 rounded-md outline-none items-center justify-center flex"
        >
            <FaceSmileIcon class="h-6 w-6" aria-hidden="true" />
        </PopoverButton>

        <transition
            enter-active-class="transition duration-200 ease-out"
            enter-from-class="translate-y-1 opacity-0"
            enter-to-class="translate-y-0 opacity-100"
            leave-active-class="transition duration-150 ease-in"
            leave-from-class="translate-y-0 opacity-100"
            leave-to-class="translate-y-1 opacity-0"
        >
            <PopoverPanel
                ref="floating"
                :style="floatingStyles"
                class="z-50 w-96 max-w-[calc(100vw-1rem)]"
            >
                <div class="overflow-hidden rounded-lg shadow-lg bg-white p-4">
                    <ExpressionPickerTabs
                        :disable-gifs="props.disableGifs"
                        @select-emoji="($emit('select-emoji', $event), close())"
                        @select-gif="($emit('select-gif', $event), close())"
                    />
                </div>
            </PopoverPanel>
        </transition>
    </Popover>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from "vue";
import ExpressionPickerTabs from "@/components/ExpressionPicker/ExpressionPickerTabs.vue";
import {
    Dialog,
    DialogPanel,
    DialogTitle,
    TransitionChild,
    TransitionRoot,
    Popover,
    PopoverButton,
    PopoverPanel,
} from "@headlessui/vue";
import { FaceSmileIcon } from "@heroicons/vue/24/outline";
import { XMarkIcon } from "@heroicons/vue/24/outline";
import { useFloating, offset, flip, shift } from "@floating-ui/vue";
import { autoUpdate } from "@floating-ui/dom";
import { warmEmojiSheet } from "@/utils/emoji";

const props = defineProps({
    disableGifs: {
        type: Boolean,
        default: false,
    },
});

defineEmits(["update:modelValue", "select-emoji", "select-gif", "click"]);

const reference = ref(null);
const floating = ref(null);
const { floatingStyles } = useFloating(reference, floating, {
    // Fixed, so a scrolling or overflow-hidden panel cannot clip the picker;
    // autoUpdate still keeps it on its button while that panel scrolls.
    strategy: "fixed",
    placement: "top-start",
    transform: false,
    middleware: [
        offset(8),
        flip({ fallbackPlacements: ["bottom-start", "top-end", "bottom-end", "left"] }),
        shift({ padding: 8 }),
    ],
    whileElementsMounted: autoUpdate,
});
const isMobile = ref(window.innerWidth < 640);
const open = ref(false);

function handleResize() {
    isMobile.value = window.innerWidth < 640;
}

onMounted(() => {
    window.addEventListener("resize", handleResize);
    warmEmojiSheet();
});

onUnmounted(() => {
    window.removeEventListener("resize", handleResize);
});
</script>
