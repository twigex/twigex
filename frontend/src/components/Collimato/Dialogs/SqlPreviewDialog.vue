<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <TransitionRoot as="template" :show="modelValue">
        <Dialog class="relative z-50" @close="$emit('update:modelValue', false)">
            <TransitionChild
                as="template"
                enter="ease-out duration-300"
                enter-from="opacity-0"
                enter-to="opacity-100"
                leave="ease-in duration-200"
                leave-from="opacity-100"
                leave-to="opacity-0"
            >
                <div class="fixed inset-0 bg-gray-500 bg-opacity-75 transition-opacity" />
            </TransitionChild>

            <div class="fixed inset-0 z-10 w-screen overflow-y-auto">
                <div
                    class="flex min-h-full items-end justify-center p-4 text-center sm:items-center sm:p-0"
                >
                    <TransitionChild
                        as="template"
                        enter="ease-out duration-300"
                        enter-from="opacity-0 translate-y-4 sm:translate-y-0 sm:scale-95"
                        enter-to="opacity-100 translate-y-0 sm:scale-100"
                        leave="ease-in duration-200"
                        leave-from="opacity-100 translate-y-0 sm:scale-100"
                        leave-to="opacity-0 translate-y-4 sm:translate-y-0 sm:scale-95"
                    >
                        <DialogPanel
                            class="relative transform overflow-hidden rounded-lg bg-white px-2 pb-1 pt-5 text-left shadow-xl transition-all w-full sm:my-8 sm:w-full sm:max-w-3xl sm:p-6 w-full"
                        >
                            <div class="absolute right-0 top-0 pr-4 pt-2 sm:block">
                                <button
                                    type="button"
                                    class="rounded-md bg-white text-gray-400 hover:text-gray-500 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2"
                                    @click="$emit('update:modelValue', false)"
                                >
                                    <span class="sr-only">Close</span>
                                    <XMarkIcon class="h-6 w-6" aria-hidden="true" />
                                </button>
                            </div>
                            <div
                                class="flex flex-row items-center justify-center sm:items-center mt-3"
                            >
                                <div class="mt-3 text-center sm:ml-1 sm:mt-0 sm:text-left w-full">
                                    <div class="h-[60vh] w-full">
                                        <div
                                            ref="aceEditor"
                                            :style="{
                                                height: '100%',
                                                width: '100%',
                                            }"
                                        ></div>
                                    </div>
                                </div>
                            </div>
                        </DialogPanel>
                    </TransitionChild>
                </div>
            </div>
        </Dialog>
    </TransitionRoot>
</template>

<script setup>
import { watch, ref, nextTick } from "vue";
import { Dialog, DialogPanel, TransitionChild, TransitionRoot } from "@headlessui/vue";
import { XMarkIcon } from "@heroicons/vue/24/outline";
import ace from "ace-builds";
import "ace-builds/src-noconflict/mode-sql";
import "ace-builds/src-noconflict/theme-dracula";

const props = defineProps({
    modelValue: {
        type: Boolean,
        required: true,
    },
    message: {
        type: String,
        required: true,
    },
});

defineEmits(["update:modelValue"]);

const aceEditor = ref(null);
let editor = null;

watch(
    () => props.modelValue,
    (newValue) => {
        if (newValue) {
            nextTick(() => {
                initEditor();
            });
        }
    },
);

watch(
    () => props.message,
    (newValue) => {
        if (editor) {
            editor.setValue(formatSQL(newValue), 1);
        }
    },
);

function initEditor() {
    editor = ace.edit(aceEditor.value);
    editor.setTheme("ace/theme/dracula");
    editor.getSession().setMode("ace/mode/sql");
    editor.setValue(formatSQL(props.message), 1);
    editor.setFontSize(14);
    editor.clearSelection();
    editor.setOptions({
        showLineNumbers: false, // Remove line numbers
        showGutter: false, // Remove the gutter (where line numbers and breakpoints appear)
        readOnly: true, // Make it read-only
        highlightActiveLine: false, // Disable the highlight on the active line
        highlightGutterLine: false, // Disable gutter line highlight
        useWorker: false, // Disable web worker for syntax checking
        wrap: true, // Enable word wrapping
    });
}

function formatSQL(sql) {
    return sql;
}
</script>
