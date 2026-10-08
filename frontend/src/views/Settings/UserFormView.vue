<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div v-if="loaded" class="flex flex-col h-full">
        <div class="flex-1 overflow-y-auto">
            <div class="mx-auto max-w-2xl px-6 py-6">
                <div class="space-y-6">
                    <!-- Name -->
                    <div class="grid grid-cols-1 gap-x-6 gap-y-4 sm:grid-cols-2">
                        <div>
                            <label
                                for="first-name"
                                class="block text-sm font-medium leading-6 text-gray-900"
                            >
                                {{ t("common.label.first_name") }}
                            </label>
                            <div class="mt-2">
                                <input
                                    v-model="form.name"
                                    id="first-name"
                                    type="text"
                                    autocomplete="given-name"
                                    :disabled="!canEditUser"
                                    class="block w-full rounded-md border-0 py-1.5 px-3 text-sm text-gray-900 shadow-sm ring-1 ring-inset placeholder:text-gray-400 focus:ring-2 focus:ring-inset sm:leading-6 disabled:bg-gray-50 disabled:text-gray-500 disabled:cursor-not-allowed"
                                    :class="
                                        v$.name.$error
                                            ? 'ring-red-600 focus:ring-red-600'
                                            : 'ring-gray-300 focus:ring-indigo-600'
                                    "
                                />
                                <p v-if="v$.name.$error" class="mt-1.5 text-sm text-red-600">
                                    {{ t("settings.edit_user.name_required") }}
                                </p>
                            </div>
                        </div>
                        <div>
                            <label
                                for="last-name"
                                class="block text-sm font-medium leading-6 text-gray-900"
                            >
                                {{ t("common.label.last_name") }}
                            </label>
                            <div class="mt-2">
                                <input
                                    v-model="form.lastname"
                                    id="last-name"
                                    type="text"
                                    autocomplete="family-name"
                                    :disabled="!canEditUser"
                                    class="block w-full rounded-md border-0 py-1.5 px-3 text-sm text-gray-900 shadow-sm ring-1 ring-inset placeholder:text-gray-400 focus:ring-2 focus:ring-inset sm:leading-6 disabled:bg-gray-50 disabled:text-gray-500 disabled:cursor-not-allowed"
                                    :class="
                                        v$.lastname.$error
                                            ? 'ring-red-600 focus:ring-red-600'
                                            : 'ring-gray-300 focus:ring-indigo-600'
                                    "
                                />
                                <p v-if="v$.lastname.$error" class="mt-1.5 text-sm text-red-600">
                                    {{ t("settings.edit_user.lastname_required") }}
                                </p>
                            </div>
                        </div>
                    </div>

                    <!-- Username -->
                    <div>
                        <label
                            for="username"
                            class="block text-sm font-medium leading-6 text-gray-900"
                        >
                            {{ t("common.label.username") }}
                        </label>
                        <div class="mt-2">
                            <input
                                v-model="form.username"
                                id="username"
                                type="text"
                                :disabled="usernameManagedExternally"
                                class="block w-full rounded-md border-0 py-1.5 px-3 text-sm text-gray-900 shadow-sm ring-1 ring-inset placeholder:text-gray-400 focus:ring-2 focus:ring-inset sm:leading-6 disabled:bg-gray-50 disabled:text-gray-500"
                                :class="
                                    v$.username.$error
                                        ? 'ring-red-600 focus:ring-red-600'
                                        : 'ring-gray-300 focus:ring-indigo-600'
                                "
                            />
                            <p
                                v-if="usernameManagedExternally"
                                class="mt-1.5 text-sm text-gray-500"
                            >
                                {{ t("users.hint.username_from_directory") }}
                            </p>
                            <p v-else-if="v$.username.$error" class="mt-1.5 text-sm text-red-600">
                                {{ t("settings.edit_user.username_required") }}
                            </p>
                        </div>
                    </div>

                    <!-- Email -->
                    <div>
                        <label
                            for="email"
                            class="block text-sm font-medium leading-6 text-gray-900"
                        >
                            {{ t("common.label.email") }}
                        </label>
                        <div class="mt-2">
                            <input
                                v-model="form.email"
                                id="email"
                                type="email"
                                autocomplete="email"
                                :disabled="!canEditUser"
                                class="block w-full rounded-md border-0 py-1.5 px-3 text-sm text-gray-900 shadow-sm ring-1 ring-inset placeholder:text-gray-400 focus:ring-2 focus:ring-inset sm:leading-6 disabled:bg-gray-50 disabled:text-gray-500 disabled:cursor-not-allowed"
                                :class="
                                    v$.email.$error
                                        ? 'ring-red-600 focus:ring-red-600'
                                        : 'ring-gray-300 focus:ring-indigo-600'
                                "
                            />
                            <p v-if="v$.email.$error" class="mt-1.5 text-sm text-red-600">
                                {{ t("settings.edit_user.email_required") }}
                            </p>
                        </div>
                    </div>

                    <!-- Role -->
                    <div>
                        <label class="block text-sm font-medium leading-6 text-gray-900">
                            {{ t("common.label.user_role") }}
                        </label>
                        <div class="mt-2">
                            <Listbox v-if="canChangeRole" v-model="form.selectedRole">
                                <div class="relative">
                                    <ListboxButton
                                        class="relative w-full cursor-default rounded-md bg-white py-1.5 pl-3 pr-10 text-left text-sm text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 focus:outline-none focus:ring-2 focus:ring-indigo-600 sm:leading-6"
                                    >
                                        <span class="block truncate">{{
                                            form.selectedRole?.name
                                        }}</span>
                                        <span
                                            class="pointer-events-none absolute inset-y-0 right-0 flex items-center pr-2"
                                        >
                                            <ChevronUpDownIcon
                                                class="h-5 w-5 text-gray-400"
                                                aria-hidden="true"
                                            />
                                        </span>
                                    </ListboxButton>
                                    <transition
                                        leave-active-class="transition ease-in duration-100"
                                        leave-from-class="opacity-100"
                                        leave-to-class="opacity-0"
                                    >
                                        <ListboxOptions
                                            class="absolute z-10 mt-1 max-h-60 w-full overflow-auto rounded-md bg-white py-1 text-sm shadow-lg ring-1 ring-black ring-opacity-5 focus:outline-none"
                                        >
                                            <ListboxOption
                                                v-for="role in assignableRoles"
                                                :key="role.value"
                                                :value="role"
                                                as="template"
                                                v-slot="{ active, selected }"
                                            >
                                                <li
                                                    :class="[
                                                        active
                                                            ? 'bg-indigo-600 text-white'
                                                            : 'text-gray-900',
                                                        'relative cursor-default select-none py-2 pl-3 pr-9',
                                                    ]"
                                                >
                                                    <span
                                                        :class="[
                                                            selected
                                                                ? 'font-semibold'
                                                                : 'font-normal',
                                                            'block truncate',
                                                        ]"
                                                        >{{ role.name }}</span
                                                    >
                                                    <span
                                                        v-if="selected"
                                                        :class="[
                                                            active
                                                                ? 'text-white'
                                                                : 'text-indigo-600',
                                                            'absolute inset-y-0 right-0 flex items-center pr-4',
                                                        ]"
                                                    >
                                                        <CheckIcon
                                                            class="h-5 w-5"
                                                            aria-hidden="true"
                                                        />
                                                    </span>
                                                </li>
                                            </ListboxOption>
                                        </ListboxOptions>
                                    </transition>
                                </div>
                            </Listbox>
                            <input
                                v-else
                                :value="form.selectedRole?.name"
                                disabled
                                type="text"
                                class="block w-full rounded-md border-0 py-1.5 px-3 text-sm text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 bg-gray-50 text-gray-500 cursor-not-allowed sm:leading-6"
                            />
                        </div>
                    </div>

                    <!-- Storage space -->
                    <div>
                        <label class="block text-sm font-medium leading-6 text-gray-900">
                            {{ t("common.label.storage_space") }}
                        </label>
                        <div class="mt-2">
                            <Listbox v-if="canEditUser" v-model="form.storage_limit">
                                <div class="relative">
                                    <ListboxButton
                                        class="relative w-full cursor-default rounded-md bg-white py-1.5 pl-3 pr-10 text-left text-sm text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 focus:outline-none focus:ring-2 focus:ring-indigo-600 sm:leading-6"
                                    >
                                        <span class="block truncate">{{
                                            storageLabel(form.storage_limit)
                                        }}</span>
                                        <span
                                            class="pointer-events-none absolute inset-y-0 right-0 flex items-center pr-2"
                                        >
                                            <ChevronUpDownIcon
                                                class="h-5 w-5 text-gray-400"
                                                aria-hidden="true"
                                            />
                                        </span>
                                    </ListboxButton>
                                    <transition
                                        leave-active-class="transition ease-in duration-100"
                                        leave-from-class="opacity-100"
                                        leave-to-class="opacity-0"
                                    >
                                        <ListboxOptions
                                            class="absolute z-10 mt-1 max-h-60 w-full overflow-auto rounded-md bg-white py-1 text-sm shadow-lg ring-1 ring-black ring-opacity-5 focus:outline-none"
                                        >
                                            <ListboxOption
                                                v-for="space in storageLimitOptions"
                                                :key="space.value"
                                                :value="space.value"
                                                as="template"
                                                v-slot="{ active, selected }"
                                            >
                                                <li
                                                    :class="[
                                                        active
                                                            ? 'bg-indigo-600 text-white'
                                                            : 'text-gray-900',
                                                        'relative cursor-default select-none py-2 pl-3 pr-9',
                                                    ]"
                                                >
                                                    <span
                                                        :class="[
                                                            selected
                                                                ? 'font-semibold'
                                                                : 'font-normal',
                                                            'block truncate',
                                                        ]"
                                                        >{{ space.label }}</span
                                                    >
                                                    <span
                                                        v-if="selected"
                                                        :class="[
                                                            active
                                                                ? 'text-white'
                                                                : 'text-indigo-600',
                                                            'absolute inset-y-0 right-0 flex items-center pr-4',
                                                        ]"
                                                    >
                                                        <CheckIcon
                                                            class="h-5 w-5"
                                                            aria-hidden="true"
                                                        />
                                                    </span>
                                                </li>
                                            </ListboxOption>
                                        </ListboxOptions>
                                    </transition>
                                </div>
                            </Listbox>
                            <input
                                v-else
                                :value="storageLabel(form.storage_limit)"
                                disabled
                                type="text"
                                class="block w-full rounded-md border-0 py-1.5 px-3 text-sm text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 bg-gray-50 text-gray-500 cursor-not-allowed sm:leading-6"
                            />
                        </div>
                    </div>

                    <!-- Password (create only) -->
                    <div v-if="!isEdit">
                        <label
                            for="password"
                            class="block text-sm font-medium leading-6 text-gray-900"
                        >
                            {{ t("common.label.password") }}
                        </label>
                        <div class="mt-2">
                            <input
                                v-model="form.password"
                                id="password"
                                type="password"
                                autocomplete="new-password"
                                class="block w-full rounded-md border-0 py-1.5 px-3 text-sm text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:leading-6"
                            />
                        </div>
                        <p class="mt-1.5 text-xs text-gray-500">
                            {{ t("settings.new_user.password_leave_empty") }}
                        </p>
                    </div>

                    <!-- Deactivate / Reset password (edit only) -->
                    <div v-if="isEdit" class="rounded-lg border border-gray-200 overflow-hidden">
                        <div class="px-4 py-3 bg-gray-50 border-b border-gray-200">
                            <span class="text-sm font-semibold text-gray-900">{{
                                t("settings.edit_user.account_actions")
                            }}</span>
                        </div>
                        <div class="divide-y divide-gray-100">
                            <div
                                v-if="can('edit_users') && canManageTarget"
                                class="flex items-center justify-between px-4 py-3"
                            >
                                <div>
                                    <p class="text-sm font-medium text-gray-900">
                                        {{ t("settings.edit_user.reset_password") }}
                                    </p>
                                    <p class="text-xs text-gray-500 mt-0.5">
                                        {{ t("settings.edit_user.reset_password_description") }}
                                    </p>
                                </div>
                                <button
                                    type="button"
                                    @click="resetPassword"
                                    class="rounded-md px-3 py-1.5 text-sm font-semibold text-gray-700 ring-1 ring-inset ring-gray-300 hover:bg-gray-50"
                                >
                                    {{ t("settings.edit_user.reset_password") }}
                                </button>
                            </div>
                            <div
                                v-if="can('edit_users') && canManageTarget"
                                class="flex items-center justify-between px-4 py-3"
                            >
                                <div>
                                    <p class="text-sm font-medium text-gray-900">
                                        {{ t("settings.edit_user.revoke_sessions") }}
                                    </p>
                                    <p class="text-xs text-gray-500 mt-0.5">
                                        {{ t("settings.edit_user.revoke_sessions_description") }}
                                    </p>
                                </div>
                                <button
                                    type="button"
                                    @click="revokeSessions"
                                    class="rounded-md px-3 py-1.5 text-sm font-semibold text-gray-700 ring-1 ring-inset ring-gray-300 hover:bg-gray-50"
                                >
                                    {{ t("settings.edit_user.revoke_sessions") }}
                                </button>
                            </div>
                            <div
                                v-if="can('delete_users') && canManageTarget"
                                class="flex items-center justify-between px-4 py-3"
                            >
                                <div>
                                    <p
                                        class="text-sm font-medium"
                                        :class="
                                            deactivated_at != 0 ? 'text-green-700' : 'text-red-700'
                                        "
                                    >
                                        {{
                                            deactivated_at != 0
                                                ? t("settings.edit_user.activate_user")
                                                : t("settings.edit_user.deactivate_user")
                                        }}
                                    </p>
                                    <p class="text-xs text-gray-500 mt-0.5">
                                        {{
                                            deactivated_at != 0
                                                ? t("settings.edit_user.activate_description")
                                                : t("settings.edit_user.deactivate_description")
                                        }}
                                    </p>
                                </div>
                                <button
                                    type="button"
                                    @click="deactivate"
                                    class="rounded-md px-3 py-1.5 text-sm font-semibold text-white shadow-sm"
                                    :class="
                                        deactivated_at != 0
                                            ? 'bg-green-600 hover:bg-green-500'
                                            : 'bg-red-600 hover:bg-red-500'
                                    "
                                >
                                    {{
                                        deactivated_at != 0
                                            ? t("settings.edit_user.activate_user")
                                            : t("settings.edit_user.deactivate_user")
                                    }}
                                </button>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </div>

        <!-- Footer -->
        <div class="shrink-0 flex items-center justify-end gap-x-3 border-t bg-gray-50 px-6 py-4">
            <button
                type="button"
                @click="router.push({ name: 'users' })"
                class="rounded-md px-3 py-2 text-sm font-semibold text-gray-700 hover:text-gray-900"
            >
                {{ t("common.button.cancel") }}
            </button>
            <button
                v-if="canEditUser"
                type="button"
                @click="submit"
                class="rounded-md bg-indigo-600 px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-indigo-500 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600"
            >
                {{ isEdit ? t("common.button.save") : t("common.button.create") }}
            </button>
        </div>
    </div>
</template>

<script setup>
import { t } from "@/i18n/index";
import { ref, computed, onMounted } from "vue";
import { useRouter, useRoute } from "vue-router";
import { useAlertStore } from "@/store/alerts";
import { useVuelidate } from "@vuelidate/core";
import { required } from "@vuelidate/validators";
import userService from "@/services/userService";
import roleService from "@/services/roleService";
import { extractErrorMessage } from "@/utils/errors";
import { storageLimitOptions, storageLabel } from "@/constants/storage";
import { usePermissions } from "@/composables/usePermissions";
import { useUserStore } from "@/store/user";
import { Listbox, ListboxButton, ListboxOption, ListboxOptions } from "@headlessui/vue";
import { CheckIcon, ChevronUpDownIcon } from "@heroicons/vue/20/solid";
import { roleLabel } from "@/utils/roleLabels";

const router = useRouter();
const route = useRoute();
const { can } = usePermissions();
const userStore = useUserStore();
const targetRole = ref("");

const actorIsAdmin = computed(() => userStore.user?.role === "system_admin");
const canManageTarget = computed(
    () => actorIsAdmin.value || !targetRole.value.split(" ").includes("system_admin"),
);

const canEditUser = computed(
    () => (isEdit.value ? can("edit_users") : can("create_users")) && canManageTarget.value,
);

const canChangeRole = computed(
    () => canEditUser.value && (actorIsAdmin.value || route.params.id !== userStore.user?.id),
);

const assignableRoles = computed(() => {
    if (actorIsAdmin.value) return roles.value;

    return roles.value.filter((r) => r.value !== "system_admin");
});

const isEdit = computed(() => !!route.params.id);
const loaded = ref(false);
const roles = ref([]);
const deactivated_at = ref(0);
const authService = ref("");
const usernameManagedExternally = computed(() => isEdit.value && authService.value === "ldap");

const DEFAULT_ROLE = "system_user";

const form = ref({
    name: "",
    lastname: "",
    username: "",
    email: "",
    password: "",
    selectedRole: null,
    storage_limit: storageLimitOptions[0].value,
});

const rules = computed(() => ({
    name: { required },
    lastname: { required },
    email: { required },
    username: { required },
}));

const v$ = useVuelidate(rules, form);

function toRoleOption(name) {
    if (!name) return null;

    return roles.value.find((r) => r.value === name) ?? { name, value: name };
}

onMounted(async () => {
    try {
        if (canEditUser.value) {
            const rolesRes = await roleService.getRoles();

            roles.value = rolesRes.data.map((r) => ({
                name: roleLabel(r, "system"),
                value: r.name,
            }));
        }

        if (isEdit.value) {
            const res = await userService.byIds([route.params.id]);
            const user = res.data?.[0];

            if (!user) {
                useAlertStore().showError(t.value("settings.users.error.load_failed"));
                router.push({ name: "users" });

                return;
            }

            form.value.name = user.name;
            form.value.lastname = user.lastname;
            form.value.username = user.username;
            form.value.email = user.email;
            authService.value = user.auth_service ?? "";
            form.value.storage_limit = storageLimitOptions.some(
                (o) => o.value === user.storage_limit,
            )
                ? user.storage_limit
                : storageLimitOptions[0].value;
            form.value.selectedRole = toRoleOption(user.role);
            targetRole.value = user.role ?? "";
            deactivated_at.value = user.deactivated_at ?? 0;
        } else {
            form.value.selectedRole = toRoleOption(DEFAULT_ROLE);
        }
    } catch {
        useAlertStore().showError(t.value("settings.users.error.load_failed"));
        router.push({ name: "users" });

        return;
    }

    loaded.value = true;
});

async function submit() {
    const isValid = await v$.value.$validate();

    if (!isValid) return;

    if (isEdit.value) {
        userService
            .updateUser({
                id: route.params.id,
                name: form.value.name,
                lastname: form.value.lastname,
                username: form.value.username,
                email: form.value.email,
                role: form.value.selectedRole?.value,
                storage_limit: form.value.storage_limit,
            })
            .then(() => {
                useAlertStore().showSuccess(t.value("settings.edit_user.user_updated"));
                router.push({ name: "users" });
            })
            .catch((error) => {
                useAlertStore().showError(
                    extractErrorMessage(error, t.value("settings.edit_user.update_failed")),
                );
            });
    } else {
        userService
            .createUser({
                name: form.value.name,
                lastname: form.value.lastname,
                username: form.value.username,
                email: form.value.email,
                password: form.value.password,
                role: form.value.selectedRole?.value,
                storage_limit: form.value.storage_limit,
            })
            .then(() => {
                useAlertStore().showSuccess(t.value("settings.edit_user.user_created"));
                router.push({ name: "users" });
            })
            .catch((error) => {
                useAlertStore().showError(
                    extractErrorMessage(error, t.value("settings.edit_user.create_failed")),
                );
            });
    }
}

function deactivate() {
    userService.deactivate(route.params.id).then(() => {
        useAlertStore().showSuccess(t.value("settings.edit_user.user_deactivated"));
        router.push({ name: "users" });
    });
}

function resetPassword() {
    userService.resetPassword(route.params.id).then(() => {
        useAlertStore().showSuccess(t.value("settings.edit_user.password_reset_sent"));
        router.push({ name: "users" });
    });
}

function revokeSessions() {
    userService.revokeSessions(route.params.id).then(() => {
        useAlertStore().showSuccess(t.value("settings.edit_user.sessions_revoked"));
    });
}
</script>
