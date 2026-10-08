<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <TransitionRoot as="template" :show="modelValue">
        <Dialog class="relative z-50" @close="close">
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
                            class="relative transform overflow-hidden rounded-xl bg-white text-left shadow-xl transition-all w-full sm:my-8 sm:w-full sm:max-w-lg"
                        >
                            <div class="px-6 pt-6 pb-5 border-b border-gray-200">
                                <div class="flex items-center gap-x-3">
                                    <div
                                        class="flex size-10 shrink-0 items-center justify-center rounded-full bg-indigo-100"
                                    >
                                        <HashtagIcon
                                            class="size-5 text-indigo-600"
                                            aria-hidden="true"
                                        />
                                    </div>
                                    <DialogTitle
                                        as="h3"
                                        class="text-base font-semibold text-gray-900"
                                        >{{ t("channels.new_channel_dialog.title") }}</DialogTitle
                                    >
                                </div>
                            </div>

                            <div class="px-6 py-5 space-y-5">
                                <div>
                                    <label
                                        for="channel-name"
                                        class="block text-sm font-medium text-gray-900"
                                        >{{ t("channels.new_channel_dialog.channel_name") }}</label
                                    >
                                    <div class="mt-2">
                                        <input
                                            v-model="channelName"
                                            type="text"
                                            id="channel-name"
                                            name="channel-name"
                                            class="block w-full rounded-lg border-0 py-2 text-gray-900 shadow-sm ring-1 ring-inset placeholder:text-gray-400 focus:ring-2 focus:ring-inset sm:text-sm sm:leading-6"
                                            :class="
                                                v$.channelName.$error
                                                    ? 'ring-red-500 focus:ring-red-500'
                                                    : 'ring-gray-300 focus:ring-indigo-600'
                                            "
                                            :placeholder="
                                                t('channels.new_channel_dialog.channel_name')
                                            "
                                        />
                                        <p
                                            v-if="v$.channelName.$error"
                                            class="mt-1.5 text-sm text-red-600"
                                        >
                                            {{ t("common.error.required_field") }}
                                        </p>
                                    </div>
                                </div>

                                <div>
                                    <label class="block text-sm font-medium text-gray-900">{{
                                        t("channels.new_channel_dialog.channel_type")
                                    }}</label>
                                    <RadioGroup
                                        v-model="channelType"
                                        class="mt-2 grid grid-cols-2 gap-3"
                                    >
                                        <RadioGroupOption
                                            v-for="option in typeOptions"
                                            :key="option.value"
                                            :value="option.value"
                                            v-slot="{ checked }"
                                        >
                                            <div
                                                class="relative flex cursor-pointer rounded-lg border bg-white p-4 focus:outline-none transition-all"
                                                :class="
                                                    checked
                                                        ? 'border-indigo-600 ring-2 ring-indigo-600'
                                                        : 'border-gray-300 hover:border-gray-400'
                                                "
                                            >
                                                <div class="flex items-start gap-x-3">
                                                    <component
                                                        :is="option.icon"
                                                        class="size-5 mt-0.5 shrink-0"
                                                        :class="
                                                            checked
                                                                ? 'text-indigo-600'
                                                                : 'text-gray-400'
                                                        "
                                                        aria-hidden="true"
                                                    />
                                                    <div>
                                                        <p
                                                            class="text-sm font-semibold"
                                                            :class="
                                                                checked
                                                                    ? 'text-indigo-900'
                                                                    : 'text-gray-900'
                                                            "
                                                        >
                                                            {{ option.label }}
                                                        </p>
                                                        <p class="mt-0.5 text-xs text-gray-500">
                                                            {{ option.description }}
                                                        </p>
                                                    </div>
                                                </div>
                                                <CheckCircleIcon
                                                    v-if="checked"
                                                    class="absolute top-3 right-3 size-4 text-indigo-600"
                                                    aria-hidden="true"
                                                />
                                            </div>
                                        </RadioGroupOption>
                                    </RadioGroup>
                                </div>

                                <div>
                                    <label
                                        for="channel-description"
                                        class="block text-sm font-medium text-gray-900"
                                        >{{
                                            t("channels.new_channel_dialog.channel_description")
                                        }}</label
                                    >
                                    <div class="mt-2">
                                        <textarea
                                            v-model="channelDescription"
                                            id="channel-description"
                                            name="channel-description"
                                            rows="3"
                                            maxlength="255"
                                            class="block w-full rounded-lg border-0 py-2 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6 resize-none"
                                            :placeholder="
                                                t(
                                                    'channels.new_channel_dialog.channel_description_placeholder',
                                                )
                                            "
                                        />
                                        <p class="mt-1 text-xs text-gray-400 text-right">
                                            {{ channelDescription.length }}/255
                                        </p>
                                    </div>
                                </div>
                            </div>

                            <div class="px-6 py-4 bg-gray-50 flex flex-row-reverse gap-x-3">
                                <button
                                    type="button"
                                    class="inline-flex justify-center rounded-lg bg-indigo-600 px-4 py-2 text-sm font-semibold text-white shadow-xs hover:bg-indigo-500 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600"
                                    @click="createChannel"
                                >
                                    {{ t("channels.new_channel_dialog.create_button") }}
                                </button>
                                <button
                                    type="button"
                                    class="inline-flex justify-center rounded-lg bg-white px-4 py-2 text-sm font-semibold text-gray-900 shadow-xs ring-1 ring-inset ring-gray-300 hover:bg-gray-50"
                                    @click="close"
                                >
                                    {{ t("common.button.cancel") }}
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
import { t } from "@/i18n/index.js";

import { ref, computed } from "vue";
import {
    Dialog,
    DialogPanel,
    DialogTitle,
    TransitionChild,
    TransitionRoot,
    RadioGroup,
    RadioGroupOption,
} from "@headlessui/vue";
import {
    LockClosedIcon,
    GlobeAltIcon,
    HashtagIcon,
    CheckCircleIcon,
} from "@heroicons/vue/24/outline";
import { useVuelidate } from "@vuelidate/core";
import { required } from "@vuelidate/validators";

defineProps({
    modelValue: {
        type: Boolean,
        required: true,
    },
});

const emit = defineEmits(["create", "update:modelValue"]);

const channelName = ref("");
const channelType = ref("public");
const channelDescription = ref("");

const typeOptions = computed(() => [
    {
        value: "public",
        label: t.value("channels.new_channel_dialog.public_channel"),
        description: t.value("channels.new_channel_dialog.public_channel_description"),
        icon: GlobeAltIcon,
    },
    {
        value: "private",
        label: t.value("channels.new_channel_dialog.private_channel"),
        description: t.value("channels.new_channel_dialog.private_channel_description"),
        icon: LockClosedIcon,
    },
]);

const rules = {
    channelName: { required },
};

const v$ = useVuelidate(rules, { channelName });

async function createChannel() {
    const isValid = await v$.value.$validate();

    if (!isValid) return;

    emit("create", {
        name: channelName.value,
        description: channelDescription.value,
        type: channelType.value,
    });

    close();
}

function close() {
    v$.value.$reset();
    emit("update:modelValue", false);
    channelName.value = "";
    channelType.value = "public";
    channelDescription.value = "";
}
</script>
