<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div class="h-full overflow-y-auto bg-white">
        <div class="px-4 pt-6 sm:px-6">
            <h1 class="text-base font-semibold text-gray-900">
                {{ workspace?.title }}
            </h1>
            <p class="mt-1 text-sm text-gray-500">
                {{ t("projects.workspace_settings.description") }}
            </p>

            <nav class="mt-4 -mb-px flex gap-x-6 border-b border-gray-200" aria-label="Tabs">
                <button
                    v-for="item in tabs"
                    :key="item.id"
                    type="button"
                    :class="[
                        tab === item.id
                            ? 'border-indigo-500 text-indigo-600'
                            : 'border-transparent text-gray-500 hover:border-gray-300 hover:text-gray-700',
                        'whitespace-nowrap border-b-2 px-1 pb-3 text-sm font-medium',
                    ]"
                    :aria-current="tab === item.id ? 'page' : undefined"
                    @click="selectTab(item.id)"
                >
                    {{ item.label }}
                </button>
            </nav>
        </div>

        <form
            v-if="tab === 'general'"
            class="max-w-xl space-y-6 px-4 py-6 sm:px-6"
            @submit.prevent="save"
        >
            <div>
                <label
                    for="workspace-name"
                    class="block text-sm font-medium leading-6 text-gray-900"
                >
                    {{ t("projects.workspace_settings.name") }}
                </label>
                <input
                    id="workspace-name"
                    v-model="name"
                    type="text"
                    class="mt-2 block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6"
                />
            </div>
            <div>
                <label
                    for="workspace-description"
                    class="block text-sm font-medium leading-6 text-gray-900"
                >
                    {{ t("projects.workspace_settings.description_label") }}
                </label>
                <textarea
                    id="workspace-description"
                    v-model="description"
                    rows="3"
                    class="mt-2 block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6"
                    :placeholder="t('projects.workspace_settings.description_placeholder')"
                />
            </div>
            <div class="flex justify-end">
                <button
                    type="submit"
                    :disabled="!canSave"
                    :class="[
                        canSave
                            ? 'bg-indigo-600 hover:bg-indigo-500'
                            : 'cursor-not-allowed bg-gray-400',
                        'rounded-md px-3 py-2 text-sm font-semibold text-white shadow-sm focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600',
                    ]"
                >
                    {{ t("common.button.save") }}
                </button>
            </div>
        </form>

        <div v-else class="px-4 py-6 sm:px-6">
            <WorkspaceMembers :workspace-id="route.params.id" />
        </div>
    </div>
</template>

<script setup>
import { computed, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import { t } from "@/i18n/index.js";
import WorkspaceMembers from "@/components/Projects/Members/WorkspaceMembers.vue";
import workspaceService from "@/services/workspaceService";
import { useAlertStore } from "@/store/alerts";
import { useUserStore } from "@/store/user";
import { useWorkspaceStore } from "@/store/workspaces";
import { extractErrorMessage } from "@/utils/errors";

const route = useRoute();
const router = useRouter();
const workspaceStore = useWorkspaceStore();
const userStore = useUserStore();

const name = ref("");
const description = ref("");
const saving = ref(false);

const workspace = computed(() =>
    (workspaceStore.getWorkspaces || []).find((ws) => ws.id === route.params.id),
);

const canEdit = computed(
    () =>
        userStore.user?.role === "system_admin" ||
        (workspace.value?.user_permissions || []).includes("update_workspace"),
);

const tabs = computed(() => [
    ...(canEdit.value
        ? [{ id: "general", label: t.value("projects.workspace_settings.general") }]
        : []),
    { id: "members", label: t.value("projects.workspace_settings.members") },
]);

const tab = computed(() =>
    tabs.value.some((item) => item.id === route.params.tab) ? route.params.tab : tabs.value[0].id,
);

const canSave = computed(
    () =>
        !saving.value &&
        name.value.trim() !== "" &&
        (name.value.trim() !== workspace.value?.title ||
            description.value !== (workspace.value?.description || "")),
);

watch(
    workspace,
    (ws) => {
        name.value = ws?.title || "";
        description.value = ws?.description || "";
    },
    { immediate: true },
);

function selectTab(id) {
    router.replace({ name: "workspace-settings", params: { id: route.params.id, tab: id } });
}

async function save() {
    if (!canSave.value) return;
    saving.value = true;

    try {
        await workspaceService.updateWorkspace({
            workspace_id: route.params.id,
            name: name.value.trim(),
            description: description.value,
        });
        workspace.value.title = name.value.trim();
        workspace.value.description = description.value;
        useAlertStore().showSuccess(t.value("projects.workspace_settings.saved"));
    } catch (error) {
        useAlertStore().showError(extractErrorMessage(error));
    } finally {
        saving.value = false;
    }
}
</script>
