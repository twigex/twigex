<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div class="flex flex-col h-full overflow-hidden">
        <!-- Access restricted -->
        <div
            v-if="!can('manage_roles')"
            class="flex flex-col items-center justify-center w-full h-full text-center px-6"
        >
            <LockClosedIcon class="h-12 w-12 text-gray-300" />
            <h3 class="mt-4 text-base font-semibold text-gray-900">
                {{ t("settings.access_restricted.title") }}
            </h3>
            <p class="mt-1 text-sm text-gray-500 max-w-sm">
                {{ t("settings.access_restricted.roles") }}
            </p>
        </div>

        <!-- Loading -->
        <div v-else-if="!loaded" class="flex h-full items-center justify-center">
            <div
                class="loader h-8 w-8 rounded-full border-4 border-indigo-500 border-t-transparent animate-spin"
            ></div>
        </div>

        <template v-else>
            <!-- Header -->
            <div class="shrink-0 flex items-center gap-x-3 border-b px-4 py-3">
                <button
                    type="button"
                    @click="router.push({ name: 'roles' })"
                    class="shrink-0 rounded p-1 text-gray-400 hover:bg-gray-100 hover:text-gray-600"
                >
                    <ChevronLeftIcon class="h-5 w-5" aria-hidden="true" />
                </button>
                <div>
                    <h1 class="text-base font-semibold text-gray-900">
                        {{
                            isEdit ? t("settings.roles.edit_title") : t("settings.roles.new_title")
                        }}
                    </h1>
                    <p class="text-xs text-gray-500">{{ t("settings.roles.form_subtitle") }}</p>
                </div>
            </div>

            <!-- Scrollable content -->
            <div class="flex-1 overflow-y-auto">
                <div class="mx-auto max-w-3xl px-6 py-6 space-y-8">
                    <!-- Name & Description -->
                    <div class="space-y-4">
                        <div>
                            <label
                                for="role-display-name"
                                class="block text-sm font-medium leading-6 text-gray-900"
                            >
                                {{ t("settings.roles.display_name") }}
                            </label>
                            <div class="mt-2">
                                <input
                                    v-model="form.display_name"
                                    id="role-display-name"
                                    type="text"
                                    :disabled="isBuiltIn || readOnly"
                                    class="block w-full rounded-md border-0 py-1.5 px-3 text-sm text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:leading-6 disabled:bg-gray-50 disabled:text-gray-500"
                                    :placeholder="t('settings.roles.display_name_placeholder')"
                                />
                            </div>
                        </div>
                        <div>
                            <label
                                for="role-name"
                                class="block text-sm font-medium leading-6 text-gray-900"
                            >
                                {{ t("settings.roles.name_id") }}
                            </label>
                            <div class="mt-2">
                                <input
                                    :value="form.name"
                                    @input="
                                        form.name = $event.target.value
                                            .toLowerCase()
                                            .replace(/[^a-z0-9_-]/g, '')
                                    "
                                    id="role-name"
                                    type="text"
                                    :disabled="isBuiltIn || isEdit"
                                    class="block w-full rounded-md border-0 py-1.5 px-3 text-sm text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:leading-6 disabled:bg-gray-50 disabled:text-gray-500 font-mono"
                                    :placeholder="t('settings.roles.name_id_placeholder')"
                                />
                            </div>
                            <p v-if="!isBuiltIn && !isEdit" class="mt-1 text-xs text-gray-400">
                                {{ t("settings.roles.name_hint") }}
                            </p>
                        </div>
                        <div>
                            <label
                                for="role-description"
                                class="block text-sm font-medium leading-6 text-gray-900"
                            >
                                {{ t("settings.roles.description") }}
                            </label>
                            <div class="mt-2">
                                <textarea
                                    v-model="form.description"
                                    id="role-description"
                                    rows="2"
                                    :disabled="isBuiltIn || readOnly"
                                    class="block w-full rounded-md border-0 py-1.5 px-3 text-sm text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:leading-6 disabled:bg-gray-50 disabled:text-gray-500"
                                    :placeholder="t('settings.roles.description_placeholder')"
                                />
                            </div>
                        </div>
                        <p v-if="isBuiltIn" class="text-sm text-gray-500">
                            {{ t("settings.roles.built_in_notice") }}
                        </p>
                    </div>

                    <!-- Permission groups -->
                    <div class="space-y-6">
                        <div
                            v-for="section in permissionSections"
                            :key="section.key"
                            class="rounded-lg border border-gray-200 overflow-hidden"
                        >
                            <!-- Section header -->
                            <div
                                class="flex items-center justify-between px-4 py-3 bg-gray-50 border-b border-gray-200"
                            >
                                <span class="text-sm font-semibold text-gray-900">{{
                                    section.label
                                }}</span>
                                <button
                                    v-if="!readOnly"
                                    type="button"
                                    @click="toggleSection(section)"
                                    class="text-xs text-indigo-600 hover:text-indigo-500"
                                >
                                    {{
                                        isSectionFullySelected(section)
                                            ? t("settings.roles.deselect_all")
                                            : t("settings.roles.select_all")
                                    }}
                                </button>
                            </div>

                            <!-- Permission items -->
                            <div
                                class="grid grid-cols-1 sm:grid-cols-2 divide-y sm:divide-y-0 sm:divide-x divide-gray-100"
                            >
                                <label
                                    v-for="perm in section.items"
                                    :key="perm.value"
                                    class="flex items-start gap-x-3 px-4 py-3 hover:bg-gray-50 cursor-pointer border-b border-gray-100 last:border-b-0 sm:[&:nth-last-child(-n+2)]:border-b-0"
                                >
                                    <input
                                        type="checkbox"
                                        v-model="form.permissions"
                                        :value="perm.value"
                                        :disabled="readOnly"
                                        class="mt-0.5 h-4 w-4 shrink-0 rounded border-gray-300 text-indigo-600 focus:ring-indigo-600"
                                    />
                                    <span class="min-w-0">
                                        <span class="block text-sm font-medium text-gray-900">{{
                                            perm.label
                                        }}</span>
                                        <span class="block text-xs text-gray-500 mt-0.5">{{
                                            perm.desc
                                        }}</span>
                                    </span>
                                </label>
                            </div>
                        </div>
                    </div>
                </div>
            </div>

            <!-- Footer -->
            <div
                class="shrink-0 flex items-center justify-end gap-x-3 border-t bg-gray-50 px-6 py-4"
            >
                <button
                    type="button"
                    @click="router.push({ name: 'roles' })"
                    class="rounded-md px-3 py-2 text-sm font-semibold text-gray-700 hover:text-gray-900"
                >
                    {{ t("common.button.cancel") }}
                </button>
                <button
                    v-if="!readOnly"
                    type="button"
                    @click="save"
                    class="rounded-md bg-indigo-600 px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-indigo-500 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600"
                >
                    {{ isEdit ? t("common.button.update") : t("common.button.create") }}
                </button>
            </div>
        </template>
    </div>
</template>

<script setup>
import { ref, computed, onMounted } from "vue";
import { useRouter, useRoute } from "vue-router";
import { useRoleStore } from "@/store/roles";
import { useUserStore } from "@/store/user";
import { useAlertStore } from "@/store/alerts";
import roleService from "@/services/roleService";
import { usePermissions } from "@/composables/usePermissions";
import { ChevronLeftIcon, LockClosedIcon } from "@heroicons/vue/24/outline";
import { t } from "@/i18n/index";
import { roleLabel, roleDescription } from "@/utils/roleLabels";

const router = useRouter();
const route = useRoute();
const roleStore = useRoleStore();
const { can } = usePermissions();

const loaded = ref(false);
const isEdit = computed(() => !!route.params.id);
const isBuiltIn = ref(false);
const userStore = useUserStore();

const readOnly = computed(() => {
    const myRoles = (userStore.user?.role ?? "").split(" ");

    if (!isEdit.value || myRoles.includes("system_admin")) return false;

    return form.value.name === "system_admin" || myRoles.includes(form.value.name);
});
const roleId = ref(null);

const form = ref({
    name: "",
    display_name: "",
    description: "",
    permissions: [],
});

const permissionSections = ref([]);

onMounted(async () => {
    const [permsRes] = await Promise.all([
        roleService.getPermissions(),
        !roleStore.roles.length
            ? roleService.getRoles().then((r) => roleStore.setRoles(r.data))
            : Promise.resolve(),
    ]);

    permissionSections.value = (permsRes.data ?? []).map((section) => ({
        key: section.id,
        label: t.value(`permissions.sections.${section.id}`) || section.name,
        items: (section.permissions ?? []).map((p) => ({
            value: p.id,
            label: t.value(`permissions.${p.id}.name`) || p.name,
            desc: t.value(`permissions.${p.id}.description`) || p.description,
        })),
    }));

    if (isEdit.value) {
        const role = roleStore.getRoleById(route.params.id);

        if (role) {
            roleId.value = role.id;
            isBuiltIn.value = role.built_in;
            form.value = {
                name: role.name,
                display_name: roleLabel(role, "system"),
                description: roleDescription(role, "system"),
                permissions: [...(role.permissions ?? [])],
            };
        }
    }

    loaded.value = true;
});

function isSectionFullySelected(section) {
    return section.items.every((item) => form.value.permissions.includes(item.value));
}

function toggleSection(section) {
    if (isSectionFullySelected(section)) {
        form.value.permissions = form.value.permissions.filter(
            (p) => !section.items.some((item) => item.value === p),
        );
    } else {
        section.items.forEach((item) => {
            if (!form.value.permissions.includes(item.value)) {
                form.value.permissions.push(item.value);
            }
        });
    }
}

async function save() {
    if (isEdit.value) {
        const patch = { permissions: form.value.permissions };

        if (!isBuiltIn.value) {
            patch.display_name = form.value.display_name;
            patch.description = form.value.description;
        }

        roleService.updateRole(roleId.value, patch).then((res) => {
            roleStore.updateRole(res.data);
            useAlertStore().showSuccess(t.value("settings.roles.role_updated"));
            router.push({ name: "roles" });
        });
    } else {
        roleService.createRole(form.value).then((res) => {
            roleStore.addRole(res.data);
            useAlertStore().showSuccess(t.value("settings.roles.role_created"));
            router.push({ name: "roles" });
        });
    }
}
</script>
