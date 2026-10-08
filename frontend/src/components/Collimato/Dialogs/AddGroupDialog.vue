<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <TransitionRoot as="template" :show="modelValue">
        <Dialog as="div" class="relative z-50" @close="close">
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
                            class="relative transform overflow-hidden rounded-lg bg-white text-left shadow-xl transition-all sm:my-8 sm:w-full sm:max-w-md"
                        >
                            <div class="px-6 pt-6 pb-4">
                                <DialogTitle
                                    as="h3"
                                    class="text-base font-semibold leading-6 text-gray-900"
                                >
                                    {{ t("collimato.users.add_group_dialog.title") }}
                                </DialogTitle>
                                <p class="mt-1 text-sm text-gray-500">
                                    {{ t("collimato.users.add_group_dialog.description") }}
                                </p>

                                <div class="mt-4">
                                    <p
                                        class="text-xs font-medium uppercase tracking-wide text-gray-500"
                                    >
                                        {{ t("collimato.users.add_group_dialog.roles_label") }}
                                    </p>
                                    <div class="mt-2 flex flex-wrap gap-1.5">
                                        <button
                                            v-for="role in roles"
                                            :key="role.id"
                                            type="button"
                                            @click="toggleRole(role)"
                                            class="inline-flex items-center rounded-md px-2 py-1 text-xs font-medium ring-1 ring-inset"
                                            :class="
                                                isRoleSelected(role)
                                                    ? 'bg-indigo-600 text-white ring-indigo-600'
                                                    : 'bg-white text-gray-700 ring-gray-300 hover:bg-gray-50'
                                            "
                                        >
                                            {{ roleLabel(role, scope) }}
                                        </button>
                                        <span
                                            v-if="roles.length === 0"
                                            class="text-xs text-gray-400"
                                        >
                                            {{ t("collimato.users.add_group_dialog.no_roles") }}
                                        </span>
                                    </div>
                                    <p
                                        v-if="v$.selectedRoles.$error"
                                        class="mt-2 text-sm text-red-600"
                                    >
                                        {{
                                            t("collimato.users.add_group_dialog.error.select_role")
                                        }}
                                    </p>
                                </div>

                                <div class="relative mt-4">
                                    <MagnifyingGlassIcon
                                        class="pointer-events-none absolute left-2.5 top-1/2 -translate-y-1/2 h-4 w-4 text-gray-400"
                                        aria-hidden="true"
                                    />
                                    <input
                                        v-model="searchQuery"
                                        type="search"
                                        :placeholder="
                                            t('collimato.users.add_group_dialog.search_placeholder')
                                        "
                                        class="block w-full rounded-md border-0 py-1.5 pl-8 pr-3 text-sm text-gray-900 ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600"
                                    />
                                </div>

                                <div
                                    v-if="selectedGroups.length > 0"
                                    class="mt-3 flex flex-wrap gap-1.5"
                                >
                                    <span
                                        v-for="group in selectedGroups"
                                        :key="group.id"
                                        class="inline-flex items-center gap-x-1 rounded-full bg-indigo-100 px-2.5 py-1 text-xs font-medium text-indigo-700"
                                    >
                                        {{ group.name }}
                                        <button
                                            type="button"
                                            @click="deselect(group)"
                                            class="rounded-full p-0.5 hover:bg-indigo-200"
                                        >
                                            <XMarkIcon class="h-3 w-3" aria-hidden="true" />
                                        </button>
                                    </span>
                                </div>

                                <p
                                    v-if="v$.selectedGroups.$error"
                                    class="mt-2 text-sm text-red-600"
                                >
                                    {{ t("collimato.users.add_group_dialog.error.select_group") }}
                                </p>
                            </div>

                            <div class="border-t max-h-64 overflow-y-auto">
                                <div v-if="loading" class="flex items-center justify-center py-8">
                                    <div
                                        class="h-6 w-6 rounded-full border-2 border-indigo-500 border-t-transparent animate-spin"
                                    ></div>
                                </div>
                                <div
                                    v-else-if="filteredGroups.length === 0"
                                    class="flex flex-col items-center justify-center py-8 text-center px-4"
                                >
                                    <MagnifyingGlassIcon
                                        class="h-8 w-8 text-gray-300"
                                        aria-hidden="true"
                                    />
                                    <p class="mt-2 text-sm text-gray-500">
                                        {{ t("collimato.users.add_group_dialog.no_results") }}
                                    </p>
                                </div>
                                <ul v-else role="list" class="divide-y divide-gray-100">
                                    <li
                                        v-for="group in filteredGroups"
                                        :key="group.id"
                                        class="flex items-center gap-x-3 px-4 py-3 hover:bg-gray-50 cursor-pointer"
                                        @click="toggleGroup(group)"
                                    >
                                        <div
                                            class="flex h-8 w-8 shrink-0 items-center justify-center rounded-full"
                                            :class="
                                                isSelected(group) ? 'bg-indigo-500' : 'bg-gray-100'
                                            "
                                        >
                                            <UserGroupIcon
                                                class="h-4 w-4"
                                                :class="
                                                    isSelected(group)
                                                        ? 'text-white'
                                                        : 'text-gray-500'
                                                "
                                                aria-hidden="true"
                                            />
                                        </div>
                                        <div class="flex-1 min-w-0">
                                            <p class="truncate text-sm font-medium text-gray-900">
                                                {{ group.name }}
                                            </p>
                                            <p class="truncate text-xs text-gray-500">
                                                {{
                                                    t(
                                                        "collimato.users.add_group_dialog.member_count",
                                                        {
                                                            count: group.member_count ?? 0,
                                                        },
                                                    )
                                                }}
                                            </p>
                                        </div>
                                        <CheckIcon
                                            v-if="isSelected(group)"
                                            class="h-4 w-4 shrink-0 text-indigo-600"
                                            aria-hidden="true"
                                        />
                                    </li>
                                </ul>
                            </div>

                            <div
                                class="flex flex-row-reverse gap-x-3 border-t bg-gray-50 px-6 py-4"
                            >
                                <button
                                    type="button"
                                    @click="add"
                                    class="inline-flex items-center rounded-md bg-indigo-600 px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-indigo-500 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600"
                                >
                                    {{ t("common.button.add") }}
                                    <span
                                        v-if="selectedGroups.length > 0"
                                        class="ml-1.5 inline-flex items-center rounded-full bg-indigo-500 px-1.5 py-0.5 text-xs font-semibold"
                                    >
                                        {{ selectedGroups.length }}
                                    </span>
                                </button>
                                <button
                                    type="button"
                                    @click="close"
                                    class="inline-flex justify-center rounded-md bg-white px-3 py-2 text-sm font-semibold text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 hover:bg-gray-50"
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
import { ref, onMounted, watch } from "vue";
import { t } from "@/i18n/index.js";
import { useVuelidate } from "@vuelidate/core";
import { required } from "@vuelidate/validators";
import { useGroupSearch } from "@/composables/useGroupSearch";
import { Dialog, DialogPanel, DialogTitle, TransitionChild, TransitionRoot } from "@headlessui/vue";
import { CheckIcon, MagnifyingGlassIcon, XMarkIcon } from "@heroicons/vue/20/solid";
import { UserGroupIcon } from "@heroicons/vue/24/outline";
import { roleLabel } from "@/utils/roleLabels";

defineProps({
    modelValue: {
        type: Boolean,
        required: true,
    },
    roles: {
        type: Array,
        default: () => [],
    },
    scope: {
        type: String,
        default: "collimato",
    },
});

const emit = defineEmits(["add", "update:modelValue"]);

const groupSearch = useGroupSearch();
const filteredGroups = groupSearch.results;
const loading = groupSearch.loading;

const selectedGroups = ref([]);
const selectedRoles = ref([]);
const searchQuery = ref("");

const rules = {
    selectedGroups: { required },
    selectedRoles: { required },
};
const v$ = useVuelidate(rules, { selectedGroups, selectedRoles });

watch(searchQuery, (q) => groupSearch.search(q));

function isSelected(group) {
    return selectedGroups.value.some((g) => g.id === group.id);
}

function toggleGroup(group) {
    const idx = selectedGroups.value.findIndex((g) => g.id === group.id);

    if (idx >= 0) {
        selectedGroups.value.splice(idx, 1);
    } else {
        selectedGroups.value.push(group);
    }

    v$.value.selectedGroups.$reset();
}

function deselect(group) {
    selectedGroups.value = selectedGroups.value.filter((g) => g.id !== group.id);
}

function isRoleSelected(role) {
    return selectedRoles.value.includes(role.name);
}

function toggleRole(role) {
    const idx = selectedRoles.value.indexOf(role.name);

    if (idx >= 0) {
        selectedRoles.value.splice(idx, 1);
    } else {
        selectedRoles.value.push(role.name);
    }

    v$.value.selectedRoles.$reset();
}

function close() {
    selectedGroups.value = [];
    selectedRoles.value = [];
    searchQuery.value = "";
    v$.value.$reset();
    emit("update:modelValue", false);
}

async function add() {
    const isValid = await v$.value.$validate();

    if (!isValid) return;
    emit("add", {
        groupIds: selectedGroups.value.map((g) => g.id),
        roles: [...selectedRoles.value],
    });
    close();
}

onMounted(() => groupSearch.search("", 0));
</script>
