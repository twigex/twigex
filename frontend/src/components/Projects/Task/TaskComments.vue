<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div class="h-[85vh] flex flex-col">
        <div ref="commentList" class="flex-1 overflow-y-auto px-4">
            <TaskActivityList :activity="activity" />
        </div>

        <TaskCommentComposer @added="scrollToBottom" />
    </div>
</template>

<script setup>
import { t } from "@/i18n/index.js";
import { ref, computed, watch, nextTick } from "vue";
import { useUserStore } from "@/store/user";
import { storeToRefs } from "pinia";
import { useWorkspaceStore } from "@/store/workspaces";
import workspaceService from "@/services/workspaceService";
import { useRoute } from "vue-router";
import { useAlertStore } from "@/store/alerts";
import { extractErrorMessage } from "@/utils/errors";
import { toActivityRow } from "@/utils/projects/taskComment";
import TaskActivityList from "./TaskActivityList.vue";
import TaskCommentComposer from "./TaskCommentComposer.vue";

const route = useRoute();
const workspaceStore = useWorkspaceStore();
const userStore = useUserStore();

const { activity } = storeToRefs(workspaceStore);

const commentList = ref(null);

const scrollToBottom = () => {
    nextTick(() => {
        const container = commentList.value;

        if (container) {
            container.scrollTop = container.scrollHeight;
        }
    });
};

const getSelectedItem = computed(() => workspaceStore.getSelectedItem);

const commentsTableId = computed(() => workspaceStore.getTableID || route.params.tid);

const loadComments = async () => {
    try {
        if (!getSelectedItem.value?.id || !route.params.id || !commentsTableId.value) return;

        const response = await workspaceService.getTaskComments({
            workspace_id: route.params.id,
            table_id: commentsTableId.value,
            task_id: getSelectedItem.value.id,
        });

        if (response.status !== 200) {
            throw new Error("Network response was not ok");
        }

        if (!response.data || !Array.isArray(response.data)) {
            workspaceStore.setActivity([]);

            return;
        }

        // Ensure all users referenced in activity are in cache
        const activityUserIds = [
            ...new Set(response.data.flatMap((e) => [e.user_id, e.affected_user].filter(Boolean))),
        ];

        if (activityUserIds.length) await userStore.ensureUsers(activityUserIds);

        const deps = { getUser: userStore.getUserById, translate: (key) => t.value(key) };
        const rows = response.data
            .map((entry) => toActivityRow(entry, deps))
            .sort((a, b) => new Date(a.dateTime) - new Date(b.dateTime));

        workspaceStore.setActivity(rows);

        nextTick(() => {
            scrollToBottom();
        });
    } catch (error) {
        useAlertStore().showError(extractErrorMessage(error));
        workspaceStore.setActivity([]);
    }
};

watch(
    () => [getSelectedItem.value?.id, route.params.id, commentsTableId.value],
    ([newId], [oldId, oldWorkspace, oldTable] = []) => {
        if (!newId) return;
        if (
            newId !== oldId ||
            route.params.id !== oldWorkspace ||
            commentsTableId.value !== oldTable
        ) {
            loadComments();
        }
    },
    { immediate: true },
);
</script>
