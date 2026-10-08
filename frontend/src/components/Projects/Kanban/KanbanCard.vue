<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div class="task" :data-task-id="task.id" @contextmenu.prevent="emit('menu', $event, task)">
        <div class="task-content text-sm leading-6 text-gray-600">
            <div class="task-line task-name text-gray-600" style="font-weight: 450">
                <div @click.stop="emit('open', task)" style="cursor: pointer; font-size: 15px">
                    {{ task.name }}
                </div>

                <div
                    v-if="subtasks"
                    class="text-xs text-gray-500 flex items-center mt-1"
                    style="font-size: 13px"
                >
                    <Square2StackIcon class="w-4 h-4" />
                    {{ subtasks }}
                    {{ t("projects.kanban_view.subtask") }}<span v-if="subtasks > 1">s</span>
                </div>
            </div>

            <div
                v-if="shown.has('start_date') || shown.has('due_date')"
                class="task-line text-sm text-gray-600"
                style="font-size: 13px"
            >
                <template
                    v-if="
                        shown.has('start_date') &&
                        task.start_date &&
                        shown.has('due_date') &&
                        task.due_date
                    "
                >
                    <span class="font-medium">{{ t("projects.kanban_view.start") }} </span>
                    <span style="font-size: 13px">
                        {{ formatDate(task.start_date) }}
                    </span>
                    &nbsp;–&nbsp;
                    <span class="font-medium">{{ t("projects.kanban_view.due") }} </span>
                    <span style="font-size: 13px">
                        {{ formatDate(task.due_date) }}
                    </span>
                </template>

                <template v-else-if="shown.has('start_date') && task.start_date">
                    <span class="font-medium">{{ t("projects.kanban_view.start") }}</span>
                    {{ formatDate(task.start_date) }}
                </template>

                <template v-else-if="shown.has('due_date') && task.due_date">
                    <span class="font-medium">{{ t("projects.kanban_view.due") }}</span>
                    {{ formatDate(task.due_date) }}
                </template>
            </div>

            <div
                v-if="shown.has('assignee')"
                class="task-line task-assignee flex items-center space-x-2"
            >
                <div
                    v-if="task.assignee"
                    style="display: flex; align-items: center; gap: 6px; cursor: pointer"
                    @click="emit('assign', task, 'assignee')"
                >
                    <UserAvatarWithText
                        class="gap-1.5"
                        :user-id="task.assignee"
                        avatar-class="h-5 w-5 shrink-0"
                        text-class="text-[13px]/5 text-gray-600"
                        @dragstart.prevent
                    />
                </div>

                <div
                    v-else
                    style="margin-left: 2px; cursor: pointer"
                    @click="emit('assign', task, 'assignee')"
                >
                    <UserPlusIcon class="h-4 w-4" aria-hidden="true" />
                </div>
            </div>

            <div
                v-if="shown.has('status') && task.status && (task.status.id || task.status.name)"
                class="task-line task-status flex items-center space-x-2"
            >
                <button
                    class="flex items-center px-1.5 py-0.5 text-white text-xs rounded-full shadow focus:outline-none focus:ring-1 focus:ring-offset-0"
                    :class="getStatusStyle(task.status)?.class"
                    :style="getStatusStyle(task.status)?.style"
                    type="button"
                >
                    <span style="font-size: 12px">
                        {{ getStatusDisplayName(task.status) }}
                    </span>
                </button>
            </div>

            <div v-for="header in fields" :key="header.name" class="task-line task-dynamic">
                <template v-if="task[header.name]">
                    <div
                        v-if="['default_assignee', 'assignee'].includes(header.header_usage)"
                        class="task-assignee flex items-center space-x-2"
                        style="cursor: pointer"
                        @click="emit('assign', task, header.name)"
                    >
                        <span class="font-medium mr-1 text-sm text-gray-600" style="font-size: 13px"
                            >{{ header.display_name || header.name }}:</span
                        >
                        <div v-if="task[header.name]">
                            <UserAvatarWithText
                                class="gap-2"
                                :user-id="task[header.name]"
                                avatar-class="h-5 w-5 shrink-0"
                                text-class="text-[13px] text-gray-600"
                            />
                        </div>
                    </div>

                    <div v-else-if="header.header_usage === 'date'">
                        <span class="font-medium text-gray-600" style="font-size: 13px">
                            {{ header.display_name || header.name }}:
                        </span>
                        <span class="text-xs text-gray-600" style="font-size: 13px">
                            {{ formatDateOnly(task[header.name]) }}
                        </span>
                    </div>
                    <div
                        v-else-if="header.header_usage === 'default_date'"
                        class="inline-flex items-baseline flex-wrap gap-1"
                    >
                        <span class="font-medium" style="font-size: 13px">
                            {{ header.display_name || header.name }}:
                        </span>
                        <div class="text-gray-600" style="font-size: 13px">
                            {{ formatUnixTimestamp(task[header.name]) }}
                        </div>
                    </div>

                    <div
                        v-else-if="header.header_usage === 'bool'"
                        class="inline-flex items-baseline flex-wrap gap-1"
                    >
                        <span class="font-medium" style="font-size: 13px">
                            {{ header.display_name || header.name }}:
                        </span>
                        <span class="text-xs text-gray-600" style="font-size: 13px">
                            {{
                                task[header.name]
                                    ? `✔️ ${t("projects.kanban_view.assignee")}`
                                    : `❌ ${t("projects.kanban_view.no")}`
                            }}
                        </span>
                    </div>

                    <div
                        v-else-if="header.header_usage === 'url'"
                        class="inline-flex items-baseline flex-wrap gap-1"
                    >
                        <span class="font-medium" style="font-size: 13px">
                            {{ header.display_name || header.name }}:
                        </span>
                        <a
                            :href="safeHref(task[header.name])"
                            target="_blank"
                            rel="noopener"
                            class="text-xs text-indigo-600 underline hover:text-indigo-500"
                            style="font-size: 13px"
                        >
                            {{ task[header.name] }}
                        </a>
                    </div>

                    <div
                        v-else-if="
                            header.header_usage === 'link' && typeof task[header.name] === 'object'
                        "
                        class="inline-flex items-baseline flex-wrap gap-1"
                    >
                        <span class="font-medium" style="font-size: 13px">
                            {{ header.display_name || header.name }}:
                        </span>

                        <template v-if="Array.isArray(task[header.name])">
                            <button
                                v-for="(item, i) in task[header.name]"
                                :key="i"
                                type="button"
                                class="link-chip mt-1 flex items-center px-1.5 py-0.5 text-xs rounded-full shadow"
                                :class="
                                    item.restricted
                                        ? 'italic text-gray-400'
                                        : 'cursor-pointer text-gray-700'
                                "
                                :title="
                                    item.restricted
                                        ? t('projects.grid_view.restricted_record_hint')
                                        : item.name || t('projects.grid_view.unnamed_record')
                                "
                            >
                                <span class="mx-1">
                                    {{
                                        item.restricted
                                            ? t("projects.grid_view.restricted_record")
                                            : item.name || t("projects.grid_view.unnamed_record")
                                    }}
                                </span>
                            </button>
                        </template>
                    </div>

                    <div
                        v-else-if="header.single_select && typeof task[header.name] === 'object'"
                        class="inline-flex items-baseline flex-wrap gap-1"
                    >
                        <span class="font-medium" style="font-size: 13px">
                            {{ header.display_name || header.name }}:
                        </span>

                        <button
                            class="flex items-center px-1.5 py-0.5 text-white text-xs rounded-full shadow focus:outline-none focus:ring-1 focus:ring-offset-0"
                            :class="getStatusStyle(task[header.name])?.class"
                            :style="getStatusStyle(task[header.name])?.style"
                            type="button"
                        >
                            <span style="font-size: 12px">
                                {{
                                    header.name === "status"
                                        ? getStatusDisplayName(task[header.name])
                                        : task[header.name]?.name || " "
                                }}
                            </span>
                        </button>
                    </div>

                    <div
                        v-else-if="header.header_usage === 'number'"
                        class="inline-flex items-baseline flex-wrap gap-1"
                    >
                        <span class="font-medium" style="font-size: 13px">
                            {{ header.display_name || header.name }}:
                        </span>
                        <span class="text-xs text-gray-600" style="font-size: 13px">
                            {{ task[header.name] }}
                        </span>
                    </div>

                    <div
                        v-else-if="header.header_usage?.toLowerCase() === 'master link'"
                        class="inline-flex items-baseline flex-wrap gap-1 task-line task-master-link"
                    >
                        <span class="font-medium" style="font-size: 13px">
                            {{ header.display_name || header.name }}:
                        </span>

                        <button
                            v-if="task[header.name]"
                            @click="emit('goToTable', task[header.name])"
                            class="inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium bg-gray-200 text-gray-700 hover:bg-gray-300 transition"
                            style="font-size: 13px"
                        >
                            {{ t("projects.kanban_view.go_to_table") }}
                        </button>

                        <span v-else class="text-gray-400 text-xs" style="font-size: 13px">
                            {{ t("projects.kanban_view.no_link") }}
                        </span>
                    </div>

                    <div v-else class="inline-flex items-baseline flex-wrap gap-1">
                        <span class="font-medium" style="font-size: 13px">
                            {{ header.display_name || header.name }}:
                        </span>
                        <span class="text-xs text-gray-600" style="font-size: 13px">
                            {{ task[header.name] }}
                        </span>
                    </div>
                </template>
            </div>
        </div>
    </div>
</template>

<script setup>
import { t } from "@/i18n/index.js";
import { UserPlusIcon, Square2StackIcon } from "@heroicons/vue/20/solid";
import UserAvatarWithText from "@/components/UserAvatarWithText.vue";
import { safeHref } from "@/utils/links";
import { getStatusDisplayName } from "@/utils/projects/cells";

// A card on the board. It is drawn once per task, so it reads everything it
// needs from its props and asks the board to act for it.
defineProps({
    task: { type: Object, required: true },
    // The fields shown beneath the name, dates, assignee and status.
    fields: { type: Array, required: true },
    // The names of the fields the board shows.
    shown: { type: Set, required: true },
    subtasks: { type: Number, default: 0 },
});

const emit = defineEmits(["open", "assign", "menu", "goToTable"]);

const formatDate = (timestamp) => {
    if (!timestamp) return "-";
    const date = new Date(timestamp * 1000);

    return date
        .toLocaleString("de-DE", {
            day: "2-digit",
            month: "2-digit",
            year: "numeric",
            hour: "2-digit",
            minute: "2-digit",
        })
        .replace(",", "");
};

const getStatusStyle = (status) => {
    if (!status || (!status.id && !status.name)) return { class: "bg-transparent" };
    const val = status.color || "";

    if (typeof val === "string" && val.startsWith("bg-")) return { class: val };
    if (val)
        return {
            class: "focus:outline-none focus:ring-1 focus:ring-offset-0",
            style: { backgroundColor: val, "--tw-ring-color": val },
        };

    return { class: "bg-blue-600 hover:bg-blue-700 focus:ring-blue-500" };
};

const formatUnixTimestamp = (timestamp) => {
    if (!timestamp || isNaN(timestamp)) return "-";

    try {
        const date = new Date(timestamp * 1000);
        const day = String(date.getDate()).padStart(2, "0");
        const month = String(date.getMonth() + 1).padStart(2, "0");
        const year = date.getFullYear();

        return `${day}.${month}.${year}`;
    } catch {
        return "-";
    }
};

const formatDateOnly = (dateStr) => {
    if (!dateStr) return "-";
    const [year, month, day] = dateStr.split("-");

    return `${day}.${month}.${year}`;
};
</script>

<style scoped>
.task-content {
    display: flex;
    justify-content: space-between;
    align-items: center;
}

button:hover:not(.task-status button) {
    background-color: rgba(0, 0, 0, 0.05);
}

.task-status button {
    transition: filter 0.15s ease;
}
.task-status button:hover {
    filter: brightness(0.95);
}

button:active {
    background-color: rgba(0, 0, 0, 0.1);
}

.task-line {
    display: block;
    width: 100%;
    margin-bottom: 8px;
    white-space: normal;
}

.task-content {
    display: block !important;
}

.task-name {
    font-size: 1.1em;
    font-weight: bold;
}
</style>
