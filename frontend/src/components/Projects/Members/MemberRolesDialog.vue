<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <TransitionRoot as="template" :show="open">
        <div>
            <Dialog class="relative z-50" @close="open = false">
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
                                class="relative transform overflow-visible rounded-lg bg-white px-4 pb-4 pt-5 text-left shadow-xl transition-all sm:my-8 sm:w-full sm:max-w-lg sm:p-6"
                            >
                                <div class="sm:flex sm:items-start">
                                    <div
                                        class="mx-auto flex h-12 w-12 shrink-0 items-center justify-center rounded-full bg-indigo-100 sm:mx-0 sm:h-10 sm:w-10"
                                    >
                                        <ChevronUpDownIcon
                                            class="h-6 w-6 text-indigo-600"
                                            aria-hidden="true"
                                        />
                                    </div>
                                    <div
                                        class="mt-3 text-center sm:ml-4 sm:mt-0 sm:text-left w-full"
                                    >
                                        <DialogTitle
                                            as="h3"
                                            class="text-base font-semibold leading-6 text-gray-900"
                                        >
                                            {{ title }}
                                        </DialogTitle>
                                        <div class="mt-2">
                                            <p class="text-sm text-gray-500">
                                                {{ message }}
                                            </p>

                                            <Listbox v-model="roles" multiple>
                                                <div class="mt-4">
                                                    <ListboxLabel
                                                        class="block text-sm font-medium text-gray-700"
                                                    >
                                                        {{
                                                            t(
                                                                "projects.workspace_members.select_roles",
                                                            )
                                                        }}
                                                    </ListboxLabel>
                                                    <div class="relative mt-1">
                                                        <ListboxButton
                                                            class="relative w-full cursor-default rounded-md bg-white py-2 pl-3 pr-10 text-left shadow-sm ring-1 ring-inset ring-gray-300 focus:outline-none focus:ring-2 focus:ring-indigo-500 sm:text-sm"
                                                        >
                                                            <span class="block truncate">
                                                                {{
                                                                    roles.length > 0
                                                                        ? roles
                                                                              .map(roleName)
                                                                              .join(", ")
                                                                        : t(
                                                                              "projects.workspace_members.select_roles",
                                                                          )
                                                                }}
                                                            </span>
                                                            <span
                                                                class="pointer-events-none absolute inset-y-0 right-0 flex items-center pr-2"
                                                            >
                                                                <ChevronUpDownIcon
                                                                    class="h-5 w-5 text-gray-400"
                                                                    aria-hidden="true"
                                                                />
                                                            </span>
                                                        </ListboxButton>

                                                        <ListboxOptions
                                                            class="absolute z-50 mt-1 max-h-60 w-full overflow-auto rounded-md bg-white py-1 text-base shadow-lg ring-1 ring-black ring-opacity-5 focus:outline-none sm:text-sm"
                                                        >
                                                            <ListboxOption
                                                                v-for="role in options"
                                                                :key="role"
                                                                :value="role"
                                                                as="template"
                                                                v-slot="{ selected, active }"
                                                            >
                                                                <li
                                                                    :class="[
                                                                        'relative cursor-default select-none py-2 pl-10 pr-4',
                                                                        active
                                                                            ? 'bg-indigo-600 text-white'
                                                                            : 'text-gray-900',
                                                                    ]"
                                                                >
                                                                    <span
                                                                        :class="[
                                                                            'block truncate',
                                                                            selected
                                                                                ? 'font-medium'
                                                                                : 'font-normal',
                                                                        ]"
                                                                    >
                                                                        {{ roleName(role) }}
                                                                    </span>
                                                                    <span
                                                                        v-if="selected"
                                                                        class="absolute inset-y-0 left-0 flex items-center pl-3 text-indigo-600"
                                                                    >
                                                                        <CheckIcon
                                                                            class="h-5 w-5"
                                                                            aria-hidden="true"
                                                                        />
                                                                    </span>
                                                                </li>
                                                            </ListboxOption>
                                                        </ListboxOptions>
                                                    </div>

                                                    <div v-if="roles.length > 0" class="mt-3">
                                                        <p class="text-sm text-gray-600 mb-2">
                                                            {{
                                                                t(
                                                                    "projects.workspace_members.selected_roles",
                                                                )
                                                            }}
                                                        </p>
                                                        <div class="flex flex-wrap gap-2">
                                                            <span
                                                                v-for="role in roles"
                                                                :key="role"
                                                                class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-indigo-100 text-indigo-800"
                                                            >
                                                                {{ roleName(role) }}
                                                                <button
                                                                    type="button"
                                                                    @click.stop="removeRole(role)"
                                                                    class="ml-1.5 text-indigo-600 hover:text-indigo-900"
                                                                >
                                                                    ×
                                                                </button>
                                                            </span>
                                                        </div>
                                                    </div>
                                                </div>
                                            </Listbox>
                                        </div>
                                    </div>
                                </div>

                                <div class="mt-5 sm:mt-4 sm:flex sm:flex-row-reverse">
                                    <button
                                        type="button"
                                        class="inline-flex w-full justify-center rounded-md bg-indigo-600 px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-indigo-500 sm:ml-3 sm:w-auto"
                                        @click="$emit('confirm')"
                                    >
                                        {{ t("projects.workspace_members.update_role") }}
                                    </button>
                                    <button
                                        type="button"
                                        class="mt-3 inline-flex w-full justify-center rounded-md bg-white px-3 py-2 text-sm font-semibold text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 hover:bg-gray-50 sm:mt-0 sm:w-auto"
                                        @click="open = false"
                                    >
                                        {{ t("common.button.cancel") }}
                                    </button>
                                </div>
                            </DialogPanel>
                        </TransitionChild>
                    </div>
                </div>
            </Dialog>
        </div>
    </TransitionRoot>
</template>

<script setup>
import { t } from "@/i18n/index.js";
import { CheckIcon, ChevronUpDownIcon } from "@heroicons/vue/20/solid";
import {
    Dialog,
    DialogPanel,
    DialogTitle,
    Listbox,
    ListboxButton,
    ListboxLabel,
    ListboxOption,
    ListboxOptions,
    TransitionChild,
    TransitionRoot,
} from "@headlessui/vue";

// The roles a member or a group holds in a workspace, picked from the
// workspace's roles. Confirming is left to the members page, which knows
// whose roles they are.
defineProps({
    title: { type: String, default: "" },
    message: { type: String, default: "" },
    options: { type: Array, default: () => [] },
    roleName: { type: Function, required: true },
});
defineEmits(["confirm"]);

const open = defineModel("open", { type: Boolean, default: false });
const roles = defineModel("roles", { type: Array, default: () => [] });

const removeRole = (role) => {
    roles.value = roles.value.filter((r) => r !== role);
};
</script>
