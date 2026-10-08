<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <TransitionRoot as="template" :show="modelValue">
        <Dialog as="div" class="relative z-50" @close="emit('update:modelValue', false)">
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
                            <!-- Header -->
                            <div class="px-6 pt-6 pb-4">
                                <DialogTitle
                                    as="h3"
                                    class="text-base font-semibold leading-6 text-gray-900"
                                >
                                    {{ t("collimato.users.role_dialog.title") }}
                                </DialogTitle>
                                <p class="mt-1 text-sm text-gray-500">
                                    {{ t("collimato.users.role_dialog.description") }}
                                </p>

                                <!-- User info -->
                                <div
                                    v-if="user?.user_info"
                                    class="mt-3 flex items-center gap-x-3 rounded-lg bg-gray-50 px-3 py-2"
                                >
                                    <div class="h-8 w-8 shrink-0">
                                        <UserAvatar
                                            :user="user.user_info"
                                            :name="user.user_info.name"
                                        />
                                    </div>
                                    <div class="min-w-0">
                                        <p class="truncate text-sm font-medium text-gray-900">
                                            {{ user.user_info.name }}
                                        </p>
                                        <p class="truncate text-xs text-gray-500">
                                            {{ user.user_info.email }}
                                        </p>
                                    </div>
                                </div>
                            </div>

                            <!-- Role list -->
                            <div class="px-6 pb-2 max-h-72 overflow-y-auto">
                                <div
                                    v-if="roles.length === 0"
                                    class="py-6 text-center text-sm text-gray-400"
                                >
                                    No roles available in this workspace.
                                </div>
                                <ul v-else role="list" class="space-y-2">
                                    <li v-for="role in roles" :key="role.id">
                                        <label
                                            class="flex items-start gap-x-3 rounded-lg border p-3 cursor-pointer transition-colors"
                                            :class="
                                                isSelected(role)
                                                    ? 'border-indigo-300 bg-indigo-50'
                                                    : 'border-gray-200 hover:bg-gray-50'
                                            "
                                        >
                                            <input
                                                type="checkbox"
                                                :checked="isSelected(role)"
                                                @change="toggleRole(role)"
                                                class="mt-0.5 h-4 w-4 shrink-0 rounded border-gray-300 text-indigo-600 focus:ring-indigo-600"
                                            />
                                            <span class="flex-1 min-w-0">
                                                <span
                                                    class="block text-sm font-medium text-gray-900"
                                                    >{{ roleLabel(role, scope) }}</span
                                                >
                                                <span
                                                    v-if="roleDescription(role, scope)"
                                                    class="block text-xs text-gray-500 mt-0.5"
                                                    >{{ roleDescription(role, scope) }}</span
                                                >
                                            </span>
                                        </label>
                                    </li>
                                </ul>
                            </div>

                            <!-- Footer -->
                            <div
                                class="flex flex-row-reverse gap-x-3 border-t bg-gray-50 px-6 py-4 mt-4"
                            >
                                <button
                                    type="button"
                                    @click="onSave"
                                    class="inline-flex items-center rounded-md bg-indigo-600 px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-indigo-500 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600"
                                >
                                    {{ t("common.button.save") }}
                                </button>
                                <button
                                    type="button"
                                    @click="emit('update:modelValue', false)"
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
import { ref, watch } from "vue";
import { t } from "@/i18n/index.js";
import UserAvatar from "@/components/UserAvatar.vue";
import { roleLabel, roleDescription } from "@/utils/roleLabels";
import { Dialog, DialogPanel, DialogTitle, TransitionChild, TransitionRoot } from "@headlessui/vue";

const props = defineProps({
    modelValue: {
        type: Boolean,
        required: true,
    },
    roles: {
        type: Array,
        required: true,
    },
    user: {
        type: Object,
        required: true,
    },
    scope: {
        type: String,
        default: "collimato",
    },
});

const emit = defineEmits(["add", "update:modelValue"]);

const selectedRoles = ref([]);

watch(
    () => [props.modelValue, props.user],
    ([open, user]) => {
        if (!open || !user) return;
        selectedRoles.value = [];
        if (!user.role) return;
        const roleNames = user.role
            .split(",")
            .map((r) => r.trim())
            .filter(Boolean);

        roleNames.forEach((roleName) => {
            const found = props.roles.find((r) => r.name === roleName);

            if (found) selectedRoles.value.push(found);
        });
    },
    { immediate: true },
);

function isSelected(role) {
    return selectedRoles.value.some((r) => r.id === role.id);
}

function toggleRole(role) {
    const idx = selectedRoles.value.findIndex((r) => r.id === role.id);

    if (idx >= 0) {
        selectedRoles.value.splice(idx, 1);
    } else {
        selectedRoles.value.push(role);
    }
}

function onSave() {
    emit(
        "add",
        selectedRoles.value.map((r) => r.name),
    );
    emit("update:modelValue", false);
}
</script>
