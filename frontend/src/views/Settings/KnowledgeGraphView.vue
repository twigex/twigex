<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div class="flex flex-col h-full overflow-hidden">
        <div class="flex flex-row justify-center flex-1 overflow-y-auto p-2">
            <div class="w-full md:w-3/4 space-y-6">
                <div class="pb-6 border-b border-gray-900/10">
                    <h2 class="text-base font-semibold leading-7 text-gray-900">
                        {{ t("settings.knowledge_graph.title") }}
                    </h2>
                    <p class="mt-1 text-sm leading-6 text-gray-600">
                        {{ t("settings.knowledge_graph.description") }}
                    </p>

                    <div class="mt-6 flex items-center gap-x-4">
                        <button
                            @click="rebuild"
                            :disabled="rebuilding"
                            type="button"
                            class="rounded-md bg-white px-2.5 py-1.5 text-sm font-semibold text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
                        >
                            {{ buttonLabel }}
                        </button>
                        <p
                            v-if="message"
                            class="text-sm"
                            :class="message.ok ? 'text-green-600' : 'text-red-600'"
                        >
                            {{ message.text }}
                        </p>
                    </div>
                    <p v-if="failureMessage" class="mt-2 text-sm text-red-600">
                        {{ failureMessage }}
                    </p>
                </div>
            </div>
        </div>
    </div>
</template>

<script setup>
import { ref, computed, onMounted, onBeforeUnmount } from "vue";
import { t } from "@/i18n/index.js";
import graphService from "@/services/graphService";
import { extractErrorMessage } from "@/utils/errors";

const status = ref(null);
const rebuilding = ref(false);
const requestResult = ref(null);

let leftPage = false;

const buttonLabel = computed(() => {
    if (rebuilding.value) {
        return t.value("settings.knowledge_graph.rebuilding");
    }

    return t.value("settings.knowledge_graph.rebuild");
});

const message = computed(() => {
    if (requestResult.value) {
        return requestResult.value;
    }

    if (!status.value) {
        return null;
    }

    const result = status.value.result;

    switch (status.value.status) {
        case "never":
            if (status.value.built) {
                return { ok: true, text: t.value("settings.knowledge_graph.built") };
            }

            return { ok: true, text: t.value("settings.knowledge_graph.never") };
        case "pending":
            return { ok: true, text: t.value("settings.knowledge_graph.queued") };
        case "running":
            return { ok: true, text: t.value("settings.knowledge_graph.rebuilding") };
        case "failed":
        case "cancelled":
        case "refused":
            return {
                ok: false,
                text: t.value("settings.knowledge_graph.failed", {
                    reason: status.value.reason,
                }),
            };
        case "completed":
            return {
                ok: true,
                text: t.value("settings.knowledge_graph.finished", {
                    relations_added: result.relations_added,
                    relations_removed: result.relations_removed,
                    fields_added: result.fields_added,
                    fields_removed: result.fields_removed,
                }),
            };
        default:
            return null;
    }
});

const failureMessage = computed(() => {
    if (!status.value) {
        return null;
    }

    const result = status.value.result;

    if (result.failed_workspaces === 0) {
        return null;
    }

    return t.value("settings.knowledge_graph.finished_with_failures", {
        failed: result.failed_workspaces,
        workspaces: result.workspaces,
    });
});

function isActive() {
    if (!status.value) {
        return false;
    }

    const current = status.value.status;

    return current === "pending" || current === "running";
}

async function fetchStatus() {
    try {
        const response = await graphService.getRebuildStatus();

        status.value = response.data;

        return true;
    } catch (e) {
        requestResult.value = { ok: false, text: extractErrorMessage(e) };

        return false;
    }
}

// The rebuild runs as a background job, so its outcome is polled until it
// ends. Leaving the page early is safe: the result is stored with the job.
async function waitForRebuild() {
    while (!leftPage) {
        await new Promise((resolve) => setTimeout(resolve, 3000));

        if (leftPage) {
            return;
        }

        const read = await fetchStatus();

        if (!read) {
            return;
        }

        if (!isActive()) {
            requestResult.value = null;

            return;
        }
    }
}

async function rebuild() {
    rebuilding.value = true;
    requestResult.value = null;

    try {
        await graphService.rebuild();

        requestResult.value = {
            ok: true,
            text: t.value("settings.knowledge_graph.rebuild_started"),
        };

        await waitForRebuild();
    } catch (e) {
        requestResult.value = { ok: false, text: extractErrorMessage(e) };
    } finally {
        rebuilding.value = false;
    }
}

onMounted(async () => {
    await fetchStatus();

    if (!isActive()) {
        return;
    }

    rebuilding.value = true;
    await waitForRebuild();
    rebuilding.value = false;
});

onBeforeUnmount(() => {
    leftPage = true;
});
</script>
