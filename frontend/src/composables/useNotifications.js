// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import notificationService from "@/services/notificationService";
import { useAlertStore } from "@/store/alerts";
import { extractErrorMessage } from "@/utils/errors";
import { app } from "@/constants/system";
import { useRouter } from "vue-router";
import workspaceService from "@/services/workspaceService";
import { useWorkspaceStore } from "@/store/workspaces";
import { computed, nextTick } from "vue";

export default function useNotificationOperations() {
    const router = useRouter();

    function handleFilesRedirect(data) {
        if (data.status == "ok") {
            switch (data.redirect) {
                case "shared":
                    router.push({ name: "shared", query: { file: data.item } });
                    break;
                case "files":
                    router.push({
                        name: "file",
                        params: { id: data.parent_id },
                        query: { file: data.item },
                    });
                    break;
                default:
                    break;
            }
        }
    }

    const workspaceStore = useWorkspaceStore();

    const selectedItem = computed({
        get() {
            return workspaceStore.getSelectedItem;
        },
        set(value) {
            workspaceStore.setSelectedItem(value);
        },
    });

    const setFullViewHeaders = computed({
        get() {
            return workspaceStore.getFullViewHeaders;
        },
        set(value) {
            workspaceStore.setFullViewHeaders(value);
        },
    });

    const showFullTask = computed({
        get() {
            return workspaceStore.getFullTask;
        },
        set(value) {
            workspaceStore.setFullTask(value);
        },
    });

    const notificationRedirect = computed({
        get() {
            return workspaceStore.getNotificationRedirect;
        },
        set(value) {
            workspaceStore.setNotificationRedirect(value);
        },
    });

    function handleProjectRedirect(notification) {
        const { details } = notification;
        const notificationType = notification.type || details.notificationType;

        notificationRedirect.value = true;

        switch (notificationType) {
            case "task_status_changed":
            case "task_deleted":
            case "task_created":
            case "task_assigned":
            case "task_comment": {
                const workspaceId = details.workspace_id;
                const tableId = details.tableID;
                const viewId = details.viewID;
                const taskId = details.task_id || details.taskID;

                if (!workspaceId || !tableId || !viewId) {
                    useAlertStore().showError("Cannot redirect: Missing project view information");
                    notificationRedirect.value = false;

                    return;
                }

                if (taskId) {
                    workspaceService
                        .getItemForTableByID({
                            workspace_id: workspaceId,
                            table_id: tableId,
                            task_id: taskId,
                        })
                        .then((response) => {
                            selectedItem.value = response.data.item;
                            setFullViewHeaders.value = response.data.headers;
                            workspaceStore.setTableID(tableId);

                            router
                                .push({
                                    name: "grid-view",
                                    params: {
                                        id: workspaceId,
                                        tid: tableId,
                                        fid: viewId,
                                    },
                                    query: { task: taskId },
                                })
                                .then(() => {
                                    showFullTask.value = true;
                                })
                                .finally(() => {
                                    nextTick(() => {
                                        notificationRedirect.value = false;
                                    });
                                });
                        })
                        .catch((error) => {
                            const errorMessage = extractErrorMessage(
                                error,
                                "Could not load task details",
                            );

                            useAlertStore().showError(errorMessage);

                            router
                                .push({
                                    name: "grid-view",
                                    params: {
                                        id: workspaceId,
                                        tid: tableId,
                                        fid: viewId,
                                    },
                                })
                                .finally(() => {
                                    notificationRedirect.value = false;
                                });
                        });
                } else {
                    router
                        .push({
                            name: "grid-view",
                            params: {
                                id: workspaceId,
                                tid: tableId,
                                fid: viewId,
                            },
                        })
                        .finally(() => {
                            notificationRedirect.value = false;
                        });
                }

                break;
            }

            case "project_invite": {
                const workspaceId = details.workspaceID;

                notificationRedirect.value = true;
                router
                    .push(
                        workspaceId
                            ? {
                                  name: "project-folder",
                                  params: { id: workspaceId },
                              }
                            : { name: "assigned-to-me" },
                    )
                    .finally(() => {
                        notificationRedirect.value = false;
                    });
                break;
            }

            case "project_deleted":
                notificationRedirect.value = true;
                router
                    .push({
                        name: "assigned-to-me",
                    })
                    .finally(() => {
                        notificationRedirect.value = false;
                    });
                break;

            default:
                notificationRedirect.value = false;
                break;
        }
    }

    const redirect = (notification) => {
        notificationService.resolve(notification.id).then((response) => {
            if (response.data.status != "ok") {
                useAlertStore().showError("You have no permission to view this item.");

                return;
            }

            notification.read_at = Math.floor(Date.now() / 1000);

            switch (response.data.app) {
                case app.Files:
                    handleFilesRedirect(response.data);
                    break;
                case app.Projects:
                    handleProjectRedirect(notification);
                    break;
                case app.Chat:
                    // Meeting invites deep-link to the meetings list, which a non-member can open.
                    if (response.data.redirect === "meeting") {
                        router.push({
                            name: "meeting",
                            params: { id: response.data.item },
                        });
                    } else {
                        router.push({
                            name: "chat",
                            params: { chatId: response.data.item },
                        });
                    }

                    break;
                default:
                    break;
            }
        });
    };

    return {
        redirect,
    };
}
