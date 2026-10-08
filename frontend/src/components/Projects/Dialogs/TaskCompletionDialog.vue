<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <TransitionRoot as="template" :show="!!question">
        <Dialog as="div" class="relative z-[60]" @close="answer(dismissal)">
            <TransitionChild
                as="template"
                enter="ease-out duration-300"
                enter-from="opacity-0"
                enter-to="opacity-100"
                leave="ease-in duration-200"
                leave-from="opacity-100"
                leave-to="opacity-0"
            >
                <div class="fixed inset-0 bg-gray-500/75 transition-opacity" />
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
                            class="relative transform overflow-hidden rounded-lg bg-white px-4 pb-4 pt-5 text-left shadow-xl transition-all sm:my-8 sm:w-full sm:max-w-lg sm:p-6"
                        >
                            <div class="sm:flex sm:items-start">
                                <div
                                    class="mx-auto flex size-12 shrink-0 items-center justify-center rounded-full bg-green-100 sm:mx-0 sm:size-10"
                                >
                                    <CheckCircleIcon
                                        class="size-6 text-green-600"
                                        aria-hidden="true"
                                    />
                                </div>
                                <div class="mt-3 w-full text-center sm:ml-4 sm:mt-0 sm:text-left">
                                    <DialogTitle
                                        as="h3"
                                        class="text-base font-semibold text-gray-900"
                                    >
                                        {{ title }}
                                    </DialogTitle>
                                    <p class="mt-2 text-sm text-gray-500">{{ message }}</p>
                                </div>
                            </div>

                            <div class="mt-5 gap-3 sm:mt-4 sm:flex sm:flex-row-reverse">
                                <button
                                    v-for="(button, index) in buttons"
                                    :key="button.choice"
                                    type="button"
                                    class="mt-3 inline-flex w-full justify-center rounded-md px-3 py-2 text-sm font-semibold shadow-sm sm:mt-0 sm:w-auto"
                                    :class="
                                        index === 0
                                            ? 'bg-indigo-600 text-white hover:bg-indigo-500'
                                            : 'bg-white text-gray-900 ring-1 ring-inset ring-gray-300 hover:bg-gray-50'
                                    "
                                    @click="answer(button.choice)"
                                >
                                    {{ button.label }}
                                </button>
                            </div>
                        </DialogPanel>
                    </TransitionChild>
                </div>
            </div>
        </Dialog>
    </TransitionRoot>
</template>

<script setup>
import { computed } from "vue";
import { Dialog, DialogPanel, DialogTitle, TransitionChild, TransitionRoot } from "@headlessui/vue";
import { CheckCircleIcon } from "@heroicons/vue/24/outline";
import { t } from "@/i18n/index.js";
import { useTaskCompletion } from "@/composables/projects/useTaskCompletion";

const { question, answer } = useTaskCompletion();

const isParent = computed(() => question.value?.kind === "parent");

const dismissal = computed(() => (isParent.value ? "not_now" : "cancel"));

const title = computed(() =>
    isParent.value
        ? t.value("projects.task_completion.parent_title")
        : t.value("projects.task_completion.subtasks_title"),
);

const message = computed(() =>
    isParent.value
        ? t.value("projects.task_completion.parent_message", { name: question.value?.name || "" })
        : t.value("projects.task_completion.subtasks_message", {
              count: question.value?.count || 0,
          }),
);

const buttons = computed(() =>
    isParent.value
        ? [
              { choice: "complete", label: t.value("projects.task_completion.complete_parent") },
              { choice: "not_now", label: t.value("projects.task_completion.not_now") },
          ]
        : [
              { choice: "all", label: t.value("projects.task_completion.complete_all") },
              { choice: "only", label: t.value("projects.task_completion.only_this_task") },
              { choice: "cancel", label: t.value("common.button.cancel") },
          ],
);
</script>
