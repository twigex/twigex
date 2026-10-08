<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <SidebarLayout :title="t('main.sections.projects')">
        <template #sidebar="{ mini }">
            <ProjectNavigation :mini="mini" />
        </template>

        <div class="outer-container">
            <div
                v-if="!can('view_projects') && !isLoading"
                class="flex flex-col items-center justify-center h-full text-center px-6"
            >
                <LockClosedIcon class="h-12 w-12 text-gray-300" />
                <h3 class="mt-4 text-base font-semibold text-gray-900">
                    {{ t("projects.access_restricted.title") }}
                </h3>
                <p class="mt-1 text-sm text-gray-500 max-w-sm">
                    {{ t("projects.access_restricted.description") }}
                </p>
            </div>
            <template v-else>
                <div v-if="!isLoading && !hasWorkspaces" class="empty-state-wrapper">
                    <div
                        class="flex h-full w-full flex-col items-center justify-center text-center"
                    >
                        <CubeIcon class="mx-auto h-16 w-16 text-indigo-400" aria-hidden="true" />
                        <h2 class="mt-4 text-lg font-semibold text-gray-900">
                            {{ t("projects.projecs_view.title") }}
                        </h2>
                        <p class="mt-2 max-w-sm text-sm text-gray-500">
                            {{ t("projects.projecs_view.description") }}
                        </p>
                        <BaseButton
                            type="button"
                            class="mt-6 gap-x-1.5 !px-4 !font-semibold"
                            @click="workspaceCreateDialogOpen = true"
                        >
                            <PlusIcon class="-ml-0.5 h-4 w-4" aria-hidden="true" />
                            {{ t("projects.projecs_view.button.create_workspace") }}
                        </BaseButton>
                    </div>
                </div>

                <div v-else-if="!isLoading" class="projects-main">
                    <div v-if="showsTopNavigation" class="top-project-navigation">
                        <div
                            v-if="
                                route.name != 'assigned-to-me' &&
                                route.name != 'detailed-task-report'
                            "
                            style="flex: 1; min-width: 0"
                        >
                            <TopNavigation />
                        </div>
                        <div
                            v-else-if="route.name == 'detailed-task-report'"
                            style="flex: 1; min-width: 0"
                        >
                            <TopNavTaskReport />
                        </div>
                    </div>
                    <div v-if="route.name != 'project-folder'" class="projects-page">
                        <div
                            v-if="route.params.tid === undefined || userLoaded"
                            style="height: 100%"
                        >
                            <RouterView />
                        </div>
                    </div>
                    <div class="projects-page-folder" v-else>
                        <RouterView />
                    </div>
                </div>

                <div
                    v-if="rightSideNavigation && hasWorkspaces"
                    style="
                        width: 450px;
                        border-left: 1px solid lightgray;
                        border-right: 1px solid lightgray;
                    "
                >
                    <CustomizeCards />
                </div>

                <div
                    v-if="showFullTask && hasWorkspaces"
                    style="
                        width: 450px;
                        border-left: 1px solid lightgray;
                        border-right: 1px solid lightgray;
                    "
                >
                    <FullTaskNavigation />
                </div>

                <NewTaskDialog @created="onTaskCreated" />
                <TaskCompletionDialog />
            </template>
        </div>
    </SidebarLayout>
</template>

<script setup>
import { t } from "@/i18n/index.js";
import SidebarLayout from "@/components/Navigation/SidebarLayout.vue";
import ProjectNavigation from "@/components/Projects/ProjectNavigation.vue";
import { ref, onMounted, watch, computed } from "vue";
import { useRoute } from "vue-router";
import { RouterView } from "vue-router";
import TopNavigation from "../../components/Projects/TopNavigation.vue";
import TopNavTaskReport from "../../components/Projects/Reports/TopNavTaskReport.vue";
import { useWorkspaceStore } from "@/store/workspaces";
import workspaceService from "@/services/workspaceService";
import { useProjectRealtime } from "@/composables/projects/useProjectRealtime";
import CustomizeCards from "../../components/Projects/Navigation/CustomCardsNavigation.vue";
import FullTaskNavigation from "../../components/Projects/Task/FullTaskNavigation.vue";
import BaseButton from "@/components/BaseButton.vue";
import NewTaskDialog from "@/components/Projects/Dialogs/NewTaskDialog.vue";
import TaskCompletionDialog from "@/components/Projects/Dialogs/TaskCompletionDialog.vue";
import { PlusIcon, CubeIcon, LockClosedIcon } from "@heroicons/vue/24/outline";
import { useWorkspaceRolesStore } from "@/store/workspaceRoles";
import { usePermissions } from "@/composables/usePermissions";

const { can } = usePermissions();

const route = useRoute();

const showsTopNavigation = computed(
    () =>
        ![
            "projects-home",
            "workspace-settings",
            "create-role",
            "update-role",
            "assigned-to-me",
        ].includes(route.name),
);

const workspaceStore = useWorkspaceStore();

const workspaceCreateDialogOpen = computed({
    get() {
        return workspaceStore.getWorkspaceCreateDialogOpen;
    },
    set(value) {
        workspaceStore.setWorkspaceCreateDialogOpen(value);
    },
});

const userLoaded = ref(false);

function generateUniqueClientId() {
    return "client-" + Math.random().toString(36).substr(2, 16);
}

const hasWorkspaces = computed({
    get() {
        return workspaceStore.getHasWorkspaces || workspaceStore.getWorkspaces?.length > 0;
    },
    set(value) {
        workspaceStore.setHasWorkspaces(value);
    },
});

const isLoading = ref(true);

const checkWorkspaces = async () => {
    isLoading.value = true;
    try {
        const response = await workspaceService.getWorkspaces();

        hasWorkspaces.value = response.data && response.data.length > 0;
    } catch (error) {
        console.error("Error checking workspaces:", error);
        hasWorkspaces.value = false;
    } finally {
        isLoading.value = false;
    }
};

const showFullTask = computed({
    get() {
        return workspaceStore.getFullTask;
    },
    set(value) {
        workspaceStore.setFullTask(value);
    },
});

const rightSideNavigation = computed({
    get() {
        return workspaceStore.getRightNavigation;
    },
    set(value) {
        workspaceStore.setRightNavigation(value);
    },
});

const clientId = generateUniqueClientId();

workspaceStore.setConnectionID(clientId);

const rolesStore = useWorkspaceRolesStore();

const { onTaskCreated } = useProjectRealtime();

const previousRoute = ref(null);

watch(
    () => route.fullPath,
    (newRoute, oldRoute) => {
        previousRoute.value = oldRoute;
    },
    { immediate: true },
);

watch(
    () => route.params.tid,
    (newTid, oldTid) => {
        if (newTid !== oldTid) {
            if (notificationRedirect.value) {
                showFullTask.value = true;
            } else {
                if (
                    !previousRoute.value ||
                    (!previousRoute.value.includes("/projects/assigned-to-me") &&
                        !previousRoute.value.includes("/detailed-task-report"))
                ) {
                    showFullTask.value = false;
                }
            }
        }
    },
);

// The Fields panel and the task panel belong to a table, so they close on a
// page without one; Gantt and Calendar have no Fields panel either.
watch(
    () => [route.name, route.params.tid],
    ([name, tid]) => {
        if (!tid) {
            rightSideNavigation.value = false;
            showFullTask.value = false;
        } else if (name === "gantt-view" || name === "calendar-view") {
            rightSideNavigation.value = false;
        }
    },
);

const notificationRedirect = computed({
    get() {
        return workspaceStore.getNotificationRedirect;
    },
    set(value) {
        workspaceStore.setNotificationRedirect(value);
    },
});

const setUser = async () => {
    const workspaceId = route.params?.id;

    userLoaded.value = false;

    try {
        if (workspaceId) {
            const response = await workspaceService.me(workspaceId);

            if (response.data) {
                rolesStore.setUser(response.data);
            }
        }
    } catch (error) {
        console.error("Failed to fetch user data for workspace:", workspaceId, error);
    } finally {
        userLoaded.value = true;
        checkWorkspaces();
    }
};

onMounted(() => {
    setUser();
});

watch(
    () => route.params.id,
    (newId, oldId) => {
        if (newId !== oldId) {
            if (newId && workspaceStore.getWorkspace?.id !== newId) setUser();
        }
    },
);
</script>

<style scoped>
.outer-container {
    display: flex;
    height: 100%; /* Ensure it respects the height of its children */
    background-color: #fff;
}

.projects-main {
    flex: 1; /* Let the main column take the remaining space */
    display: flex;
    flex-direction: column;
    overflow-x: hidden; /* Hide any horizontal overflow */
    overflow-y: hidden; /* Hide any vertical overflow */
}

.projects-page {
    flex: 1; /* Allow it to grow and take up remaining space */
    min-width: 0; /* Prevent the page from causing overflow */
    overflow-x: auto; /* Enable horizontal scrolling within the page */
    height: 100%; /* Ensure it stretches within the parent */
}

.projects-page-folder {
    flex: 1;
    display: flex;
    flex-direction: column;
    height: 100%; /* Ensure it stretches */
    overflow: hidden; /* Prevents the whole section from scrolling */
}

.top-project-navigation {
    background-color: #ffffff;
    border-bottom: 1px solid #ccc;
    display: flex;
    align-items: stretch;
}

.empty-state-wrapper {
    width: 100%;
    height: 100%;
    display: flex;
    align-items: center;
    justify-content: center;
}
</style>
