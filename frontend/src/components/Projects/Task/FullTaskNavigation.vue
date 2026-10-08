<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div class="h-full bg-white shadow-lg flex flex-col overflow-hidden min-h-0">
        <div
            class="sticky top-0 flex h-16 items-center justify-between border-b border-gray-200 z-10"
            style="height: 55px"
        >
            <div class="flex space-x-6 px-6">
                <button
                    class="text-sm font-medium py-2 focus:outline-none"
                    :class="
                        activeTab === 'details'
                            ? 'text-indigo-600 border-b-2 border-indigo-600'
                            : 'text-gray-500 hover:text-gray-700'
                    "
                    @click="activeTab = 'details'"
                >
                    {{ t("projects.full_task_navigation.title") }}
                </button>

                <button
                    class="text-sm font-medium py-2 focus:outline-none"
                    :class="
                        activeTab === 'comments'
                            ? 'text-indigo-600 border-b-2 border-indigo-600'
                            : 'text-gray-500 hover:text-gray-700'
                    "
                    @click="activeTab = 'comments'"
                >
                    {{ t("projects.full_task_navigation.activities") }}
                </button>
            </div>

            <button
                type="button"
                class="px-4 text-gray-400 hover:text-gray-500"
                @click="closeModal"
            >
                <span class="sr-only">{{ t("common.button.close") }}</span>
                <XMarkIcon class="h-5 w-5" aria-hidden="true" />
            </button>
        </div>

        <div v-if="activeTab === 'details'" class="flex min-h-0 flex-1 flex-col">
            <div
                class="flex-1 min-h-0 overflow-y-auto overscroll-y-contain px-1"
                @wheel.stop
                @touchmove.stop
            >
                <div v-if="nameIndex !== -1" class="px-5 pb-2 pt-4">
                    <input
                        v-if="isEditing[nameIndex]"
                        :id="`task-field-${nameIndex}`"
                        v-model="fieldValues.name"
                        type="text"
                        class="-mx-2 block w-[calc(100%+1rem)] rounded-md border-0 px-2 py-1 text-lg font-semibold text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 focus:ring-2 focus:ring-inset focus:ring-indigo-600"
                        @blur="saveEdit(nameIndex)"
                        @keyup.enter="saveEdit(nameIndex)"
                        @keyup.esc="cancelFieldEdit(nameIndex)"
                        @focusin="typingField = 'name'"
                        @focusout="typingField = null"
                    />
                    <button
                        v-else
                        type="button"
                        :disabled="!canEditRow"
                        class="-mx-2 block w-[calc(100%+1rem)] rounded-md px-2 py-1 text-left text-lg font-semibold text-gray-900 outline outline-1 -outline-offset-1 outline-transparent transition-[outline-color] delay-75 duration-150 hover:outline-gray-300 disabled:cursor-default disabled:hover:outline-transparent"
                        @click="enableEditing(nameIndex)"
                    >
                        {{ fieldValues.name }}
                    </button>
                </div>

                <template v-for="(header, index) in setFullViewHeaders" :key="header.name">
                    <div
                        v-if="header.name !== 'name'"
                        :class="[
                            'min-h-[37px] rounded px-4 py-1 transition-colors duration-200 hover:bg-gray-50',
                            header.name === 'description' ? 'flex flex-col gap-1' : 'flex',
                        ]"
                    >
                        <div
                            :class="[
                                'mt-0.5 flex min-w-0 flex-shrink-0 items-center gap-1 px-2 text-sm leading-normal text-gray-500',
                                header.name === 'description' ? 'mb-1 w-full' : 'w-[40%]',
                            ]"
                            :title="header.display_name || header.name"
                        >
                            <span class="truncate">{{ header.display_name || header.name }}</span>
                        </div>

                        <div
                            :class="[
                                'flex min-h-full flex-wrap items-center gap-2 px-2 text-sm text-gray-900',
                                header.name === 'description' ? 'w-full' : 'w-[60%]',
                            ]"
                            @focusin="typingField = header.name"
                            @focusout="typingField = null"
                        >
                            <div
                                v-if="
                                    header.header_usage === 'default_assignee' ||
                                    header.header_usage === 'assignee'
                                "
                                :class="{
                                    'cursor-pointer': header.name !== 'created_by',
                                }"
                                @click="
                                    header.name !== 'created_by'
                                        ? openAssignDialog(getSelectedItem, index)
                                        : null
                                "
                                class="flex items-center space-x-2"
                            >
                                <template v-if="fieldValues[header.name]">
                                    <UserAvatarWithText
                                        class="gap-2"
                                        :user-id="fieldValues[header.name]"
                                        avatar-class="h-5 w-5 shrink-0"
                                        text-class=""
                                    />
                                </template>
                                <template v-else>
                                    <UserPlusIcon
                                        class="h-4 w-4 text-gray-400"
                                        aria-hidden="true"
                                    />
                                    <span class="text-gray-400">{{
                                        t("projects.full_task_navigation.assign_user")
                                    }}</span>
                                </template>
                            </div>

                            <div v-else-if="header.header_type === 'TINYINT'">
                                <input
                                    type="checkbox"
                                    :checked="isChecked(fieldValues[header.name])"
                                    :disabled="!canEditRow"
                                    @change="toggleCheckbox(header.name, $event)"
                                    class="h-4 w-4 rounded border-gray-300 text-indigo-600 focus:ring-indigo-600 disabled:opacity-50 disabled:cursor-not-allowed"
                                />
                            </div>

                            <div v-else-if="header.header_usage === 'status'">
                                <button
                                    type="button"
                                    :class="
                                        fieldValues[header.name]?.name
                                            ? 'inline-flex max-w-full items-center gap-x-1.5 rounded-full bg-white px-2 py-0.5 text-xs font-medium text-gray-700 ring-1 ring-inset ring-gray-200 hover:bg-gray-50'
                                            : '-mx-1 inline-flex items-center gap-x-1.5 rounded px-1 py-0.5 text-gray-400 hover:bg-gray-100'
                                    "
                                    @click="
                                        openLinkedTableDialog(
                                            fieldValues[header.name],
                                            header.name,
                                            header.linked_id,
                                        )
                                    "
                                >
                                    <template v-if="fieldValues[header.name]?.name">
                                        <span
                                            class="h-2 w-2 flex-none rounded-full"
                                            :class="getStatusStyle(fieldValues[header.name]).class"
                                            :style="getStatusStyle(fieldValues[header.name]).style"
                                            aria-hidden="true"
                                        />
                                        <span class="truncate">{{
                                            fieldValues[header.name].name
                                        }}</span>
                                    </template>
                                    <template v-else>
                                        <span
                                            class="block h-3 w-3 flex-none rounded-full border border-dashed border-gray-400"
                                            aria-hidden="true"
                                        />
                                        <span>{{
                                            t("projects.full_task_navigation.set_status")
                                        }}</span>
                                    </template>
                                </button>
                            </div>

                            <div v-else-if="['description'].includes(header.name)" class="w-full">
                                <div
                                    class="mt-2 rounded-xl shadow-sm ring-1 ring-inset ring-gray-300 focus-within:ring-2 focus-within:ring-indigo-600"
                                >
                                    <textarea
                                        v-model="fieldValues[header.name]"
                                        rows="4"
                                        @focus="rememberValue(index)"
                                        :placeholder="
                                            t(
                                                'projects.full_task_navigation.description_placeholder',
                                            )
                                        "
                                        @blur="saveEdit(index)"
                                        @keydown.ctrl.enter="saveEdit(index)"
                                        @keydown.meta.enter="saveEdit(index)"
                                        class="block w-full resize-y rounded-xl border-0 bg-transparent px-3 py-1.5 text-sm text-gray-900 placeholder:text-gray-400 focus:ring-0 sm:leading-6"
                                    ></textarea>
                                </div>
                            </div>

                            <div
                                v-else-if="
                                    ['start_date', 'due_date'].includes(header.name) ||
                                    header.header_usage === 'date'
                                "
                                class="flex items-center gap-2"
                            >
                                <button
                                    type="button"
                                    class="-mx-1 inline-flex min-w-0 items-center gap-x-1.5 rounded px-1 py-0.5 hover:bg-gray-100"
                                    @click="toggleDatePickerPopover($event, getSelectedItem, index)"
                                >
                                    <CalendarIcon
                                        class="h-4 w-4 flex-none text-gray-400"
                                        aria-hidden="true"
                                    />
                                    <span v-if="!fieldValues[header.name]" class="text-gray-400">
                                        {{
                                            header.name === "start_date"
                                                ? t("projects.full_task_navigation.set_start_date")
                                                : header.name === "due_date"
                                                  ? t("projects.full_task_navigation.set_due_date")
                                                  : t("projects.grid_view.pick_a_date")
                                        }}
                                    </span>
                                    <span v-else-if="isDayField(header)" class="truncate">
                                        {{
                                            getDate(
                                                getTimestampFromDateString(
                                                    fieldValues[header.name],
                                                ),
                                            )
                                        }}
                                    </span>
                                    <span v-else class="truncate">
                                        {{ formatDateForDisplay(fieldValues[header.name], header) }}
                                    </span>
                                </button>
                            </div>

                            <div v-else-if="['created_at', 'updated_at'].includes(header.name)">
                                <div :class="!fieldValues[header.name] && 'text-gray-400'">
                                    {{
                                        fieldValues[header.name]
                                            ? getDate(fieldValues[header.name])
                                            : "–"
                                    }}
                                </div>
                            </div>

                            <div
                                v-else-if="header.header_usage === 'link'"
                                class="flex w-full min-w-0 items-center gap-x-1.5"
                            >
                                <LinkedRecordsCell
                                    v-if="fieldValues[header.name]?.length"
                                    :records="fieldValues[header.name]"
                                    class="!flex-initial"
                                    @open="showFullView($event, header, false)"
                                />
                                <span v-else class="text-gray-400">
                                    {{ t("projects.full_task_navigation.link_records") }}
                                </span>

                                <button
                                    type="button"
                                    class="flex-none rounded p-0.5 text-gray-400 hover:bg-gray-100 hover:text-gray-600"
                                    :title="t('common.button.add')"
                                    :aria-label="t('common.button.add')"
                                    @click="
                                        openLinkedTableDialog(
                                            fieldValues[header.name],
                                            header.name,
                                            header.linked_id,
                                        )
                                    "
                                >
                                    <PlusIcon class="h-4 w-4" aria-hidden="true" />
                                </button>
                                <button
                                    type="button"
                                    class="flex-none rounded p-0.5 text-gray-400 hover:bg-gray-100 hover:text-gray-600"
                                    :title="t('projects.full_task_navigation.go_to_linked_table')"
                                    :aria-label="
                                        t('projects.full_task_navigation.go_to_linked_table')
                                    "
                                    @click="redirectToLinkedTable(header)"
                                >
                                    <ArrowTopRightOnSquareIcon class="h-4 w-4" aria-hidden="true" />
                                </button>
                            </div>

                            <div
                                v-else-if="header.header_usage == 'master link'"
                                class="group flex w-full min-w-0 items-center gap-x-1.5"
                            >
                                <template v-if="fieldValues[header.name]">
                                    <button
                                        type="button"
                                        class="inline-flex min-w-0 items-center gap-x-1 rounded-full bg-gray-100 px-2 py-0.5 text-xs font-medium text-gray-700 hover:bg-gray-200"
                                        :title="
                                            t('projects.full_task_navigation.go_to_linked_table')
                                        "
                                        @click="goToTable(fieldValues[header.name])"
                                    >
                                        <ArrowTopRightOnSquareIcon
                                            class="h-3 w-3 flex-none text-gray-500"
                                            aria-hidden="true"
                                        />
                                        <span class="truncate">{{
                                            getTableName(fieldValues[header.name])
                                        }}</span>
                                    </button>

                                    <button
                                        type="button"
                                        class="flex-none rounded p-0.5 text-gray-400 opacity-0 hover:bg-gray-100 hover:text-gray-600 focus:opacity-100 group-hover:opacity-100 [@media(hover:none)]:opacity-100"
                                        :title="
                                            t('projects.full_task_navigation.edit_linked_table')
                                        "
                                        :aria-label="
                                            t('projects.full_task_navigation.edit_linked_table')
                                        "
                                        @click="openMasterLinkTableDialog(fieldValues, header.name)"
                                    >
                                        <PencilIcon class="h-3.5 w-3.5" aria-hidden="true" />
                                    </button>
                                </template>

                                <button
                                    v-else
                                    type="button"
                                    class="-mx-1 inline-flex items-center gap-x-1.5 rounded px-1 py-0.5 text-gray-400 hover:bg-gray-100"
                                    @click="openMasterLinkTableDialog(fieldValues, header.name)"
                                >
                                    <PlusIcon class="h-4 w-4 flex-none" aria-hidden="true" />
                                    <span>{{ t("projects.full_task_navigation.link_table") }}</span>
                                </button>
                            </div>

                            <div v-else-if="header.header_usage === 'file'" class="w-full">
                                <div class="flex flex-col gap-1.5">
                                    <div
                                        v-for="(file, fileIndex) in fieldValues[header.name] || []"
                                        :key="file.id || fileIndex"
                                        class="group/file flex items-center gap-1 rounded-md bg-gray-50 p-1.5 transition-colors hover:bg-gray-100"
                                    >
                                        <button
                                            type="button"
                                            class="flex min-w-0 flex-1 items-center gap-2.5 text-left"
                                            :title="getFileName(file)"
                                            @click="openFile(file, fieldValues[header.name])"
                                        >
                                            <span
                                                class="flex h-8 w-8 flex-none items-center justify-center rounded border border-gray-200 bg-white"
                                            >
                                                <component
                                                    :is="fileTypeOf(file).icon"
                                                    :class="[fileTypeOf(file).color, 'h-4 w-4']"
                                                    aria-hidden="true"
                                                />
                                            </span>
                                            <span class="min-w-0">
                                                <span
                                                    class="block truncate font-medium text-gray-900 group-hover/file:text-indigo-600"
                                                >
                                                    {{ getFileName(file) }}
                                                </span>
                                                <span class="block text-xs text-gray-500">
                                                    {{ getFileExtension(file.name).toUpperCase() }}
                                                    {{ t("projects.full_task_navigation.file") }}
                                                    ({{ convertSize(file.size) }})
                                                </span>
                                            </span>
                                        </button>

                                        <button
                                            type="button"
                                            class="flex-none rounded p-1 text-gray-400 hover:bg-gray-200 hover:text-gray-700"
                                            :title="t('projects.full_task_navigation.download')"
                                            :aria-label="
                                                t('projects.full_task_navigation.download')
                                            "
                                            @click="downloadFile(file)"
                                        >
                                            <ArrowDownTrayIcon class="h-4 w-4" aria-hidden="true" />
                                        </button>
                                        <button
                                            type="button"
                                            class="flex-none rounded p-1 text-gray-400 hover:bg-gray-200 hover:text-red-600"
                                            :title="t('projects.full_task_navigation.remove')"
                                            :aria-label="t('projects.full_task_navigation.remove')"
                                            @click="
                                                openFileDeleteConfirmation(
                                                    getSelectedItem,
                                                    index,
                                                    fileIndex,
                                                    file.name || file.original_name,
                                                )
                                            "
                                        >
                                            <TrashIcon class="h-4 w-4" aria-hidden="true" />
                                        </button>
                                    </div>

                                    <button
                                        type="button"
                                        class="-mx-1 inline-flex w-fit items-center gap-x-1.5 rounded px-1 py-0.5 text-gray-400 hover:bg-gray-100 hover:text-gray-600"
                                        @click="triggerFileInput(getSelectedItem, index)"
                                    >
                                        <PaperClipIcon
                                            class="h-4 w-4 flex-none"
                                            aria-hidden="true"
                                        />
                                        <span>
                                            {{
                                                fieldValues[header.name]?.length
                                                    ? t("projects.full_task_navigation.add_more")
                                                    : t(
                                                          "projects.full_task_navigation.upload_files",
                                                      )
                                            }}
                                        </span>
                                    </button>

                                    <input
                                        type="file"
                                        :data-file-id="`fileInput-fullView-${getSelectedItem?.id}-${index}`"
                                        multiple
                                        @change="handleFileUpload($event, getSelectedItem, index)"
                                        style="display: none"
                                    />
                                </div>
                            </div>

                            <div v-else-if="header.header_usage === 'url'" class="w-full min-w-0">
                                <div
                                    v-if="fieldValues[header.name] && !isEditing[index]"
                                    class="group/url flex min-w-0 items-center gap-1.5"
                                >
                                    <a
                                        :href="safeHref(fieldValues[header.name])"
                                        target="_blank"
                                        rel="noopener noreferrer"
                                        class="truncate text-indigo-600 underline"
                                        :title="fieldValues[header.name]"
                                        @click.stop
                                    >
                                        {{ fieldValues[header.name] }}
                                    </a>

                                    <button
                                        v-if="canEditRow"
                                        type="button"
                                        class="flex-none rounded p-0.5 text-gray-400 opacity-0 hover:bg-gray-100 hover:text-gray-600 focus:opacity-100 group-hover/url:opacity-100 [@media(hover:none)]:opacity-100"
                                        :title="t('projects.full_task_navigation.edit_url')"
                                        :aria-label="t('projects.full_task_navigation.edit_url')"
                                        @click="enableEditing(index)"
                                    >
                                        <PencilIcon class="h-3.5 w-3.5" aria-hidden="true" />
                                    </button>
                                </div>

                                <button
                                    v-else-if="!isEditing[index]"
                                    type="button"
                                    :disabled="!canEditRow"
                                    class="-mx-2 block w-[calc(100%+1rem)] truncate rounded-md px-2 py-1 text-left outline outline-1 -outline-offset-1 outline-transparent transition-[outline-color] delay-75 duration-150 hover:outline-gray-300 disabled:cursor-default disabled:hover:outline-transparent text-gray-400"
                                    @click="enableEditing(index)"
                                >
                                    {{ t("projects.full_task_navigation.add_link") }}
                                </button>

                                <input
                                    v-else
                                    :id="`task-field-${index}`"
                                    v-model="fieldValues[header.name]"
                                    type="text"
                                    :placeholder="t('projects.full_task_navigation.enter_url')"
                                    class="-mx-2 block w-[calc(100%+1rem)] rounded-md border-0 px-2 py-1 text-sm text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 focus:ring-2 focus:ring-inset focus:ring-indigo-600"
                                    @blur="saveEdit(index)"
                                    @keyup.enter="saveEdit(index)"
                                    @keyup.esc="cancelFieldEdit(index)"
                                />
                            </div>

                            <div v-else-if="header.header_usage == 'single select'">
                                <button
                                    type="button"
                                    :class="
                                        fieldValues[header.name]
                                            ? 'inline-flex max-w-full items-center gap-x-1.5 rounded-full bg-white px-2 py-0.5 text-xs font-medium text-gray-700 ring-1 ring-inset ring-gray-200 hover:bg-gray-50'
                                            : '-mx-1 inline-flex items-center rounded px-1 py-0.5 text-gray-400 hover:bg-gray-100'
                                    "
                                    :title="
                                        fieldValues[header.name]
                                            ? fieldValues[header.name].name ||
                                              t('projects.grid_view.unnamed_record')
                                            : undefined
                                    "
                                    @click="
                                        openLinkedTableDialog(
                                            fieldValues[header.name],
                                            header.name,
                                            header.linked_id,
                                        )
                                    "
                                >
                                    <template v-if="fieldValues[header.name]">
                                        <span
                                            class="h-2 w-2 flex-none rounded-full"
                                            :class="getStatusStyle(fieldValues[header.name]).class"
                                            :style="getStatusStyle(fieldValues[header.name]).style"
                                            aria-hidden="true"
                                        />
                                        <span class="truncate">
                                            {{
                                                fieldValues[header.name].name ||
                                                t("projects.grid_view.unnamed_record")
                                            }}
                                        </span>
                                    </template>
                                    <template v-else>
                                        {{ t("projects.full_task_navigation.select_option") }}
                                    </template>
                                </button>
                            </div>

                            <div
                                v-else-if="isComputedCalculations(header)"
                                class="w-full min-w-0 truncate"
                                :title="calculatedValue(header)"
                            >
                                <span v-if="isBlank(calculatedValue(header))" class="text-gray-400"
                                    >–</span
                                >
                                <template v-else>{{ calculatedValue(header) }}</template>
                            </div>

                            <div
                                v-else-if="isRegularInput(header) || isNumberInput(header)"
                                class="w-full min-w-0"
                            >
                                <button
                                    v-if="!isEditing[index]"
                                    type="button"
                                    :disabled="!canEditRow"
                                    :class="[
                                        '-mx-2 block w-[calc(100%+1rem)] truncate rounded-md px-2 py-1 text-left outline outline-1 -outline-offset-1 outline-transparent transition-[outline-color] delay-75 duration-150 hover:outline-gray-300 disabled:cursor-default disabled:hover:outline-transparent',
                                        isBlank(fieldValues[header.name]) && 'text-gray-400',
                                    ]"
                                    :title="
                                        isBlank(fieldValues[header.name])
                                            ? undefined
                                            : String(fieldValues[header.name])
                                    "
                                    @click="enableEditing(index)"
                                >
                                    {{
                                        isBlank(fieldValues[header.name])
                                            ? isNumberInput(header)
                                                ? t("projects.full_task_navigation.add_number")
                                                : t("projects.full_task_navigation.add_text")
                                            : fieldValues[header.name]
                                    }}
                                </button>
                                <input
                                    v-else
                                    :id="`task-field-${index}`"
                                    v-model="fieldValues[header.name]"
                                    :type="isNumberInput(header) ? 'number' : 'text'"
                                    :step="
                                        !isNumberInput(header)
                                            ? undefined
                                            : ['INT', 'BIGINT'].includes(header.header_type)
                                              ? '1'
                                              : 'any'
                                    "
                                    class="-mx-2 block w-[calc(100%+1rem)] rounded-md border-0 px-2 py-1 text-sm text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 focus:ring-2 focus:ring-inset focus:ring-indigo-600"
                                    @blur="saveEdit(index)"
                                    @keyup.enter="saveEdit(index)"
                                    @keyup.esc="cancelFieldEdit(index)"
                                />
                            </div>

                            <div
                                v-else
                                :class="isBlank(fieldValues[header.name]) && 'text-gray-400'"
                            >
                                {{
                                    isBlank(fieldValues[header.name])
                                        ? "–"
                                        : fieldValues[header.name]
                                }}
                            </div>
                        </div>
                    </div>
                </template>
                <div v-if="subtasks.length > 0" class="px-3 pt-4 pb-6 border-t border-gray-200">
                    <h3 class="text-sm font-medium text-gray-700 mb-2 mx-3">
                        {{ t("projects.full_task_navigation.subtasks") }}
                    </h3>

                    <ul class="space-y-2">
                        <li
                            v-for="subtask in subtasks"
                            :key="subtask.id"
                            class="bg-gray-50 rounded hover:bg-gray-100 cursor-pointer"
                            @click="showFullView(subtask, fieldValues['name'], true)"
                        >
                            <div class="w-full flex items-center min-w-0 divide-x divide-gray-200">
                                <div class="px-3 py-2 basis-[43%] min-w-0">
                                    <span
                                        class="block truncate text-sm text-gray-900"
                                        :title="
                                            subtask.name || t('projects.grid_view.unnamed_record')
                                        "
                                    >
                                        {{ subtask.name || t("projects.grid_view.unnamed_record") }}
                                    </span>
                                </div>

                                <div class="px-3 py-2 basis-[10%] min-w-0">
                                    <span
                                        v-if="subtask.assignee"
                                        class="block text-xs text-gray-500 truncate"
                                    >
                                        <div class="h-4 w-4 shrink-0">
                                            <UserAvatar :user-id="subtask.assignee" />
                                        </div>
                                    </span>
                                    <UserCircleIcon v-else class="h-4 w-4 text-gray-400" />
                                </div>

                                <div class="px-3 py-2 basis-[23%] min-w-0">
                                    <div class="min-w-0">
                                        <span
                                            v-if="subtask.status?.name"
                                            class="inline-flex max-w-full items-center gap-x-1.5 rounded-full bg-white px-2 py-0.5 text-xs font-medium text-gray-700 ring-1 ring-inset ring-gray-200"
                                            :title="subtask.status.name"
                                        >
                                            <span
                                                class="h-2 w-2 flex-none rounded-full"
                                                :class="getStatusStyle(subtask.status).class"
                                                :style="getStatusStyle(subtask.status).style"
                                                aria-hidden="true"
                                            />
                                            <span class="truncate">{{ subtask.status.name }}</span>
                                        </span>

                                        <span
                                            v-else
                                            class="ml-1.5 block h-3 w-3 rounded-full border border-dashed border-gray-400"
                                            aria-hidden="true"
                                        />
                                    </div>
                                </div>

                                <div class="px-3 py-2 basis-[23%] min-w-0">
                                    <span
                                        v-if="subtask.due_date"
                                        class="block text-xs text-gray-500 truncate"
                                    >
                                        {{ getDate(subtask.due_date).slice(0, 10) }}
                                    </span>
                                    <CalendarDaysIcon v-else class="h-4 w-4 text-gray-400" />
                                </div>
                            </div>
                        </li>
                    </ul>
                </div>
            </div>
        </div>

        <div
            v-if="activeTab === 'comments'"
            class="flex-1 min-h-0 overflow-y-auto overscroll-y-contain"
            @wheel.stop
            @touchmove.stop
        >
            <CommentsComponent />
        </div>

        <div>
            <ShareWorkspaceDialog
                :open="shareWorkspaceDialog"
                @close="shareWorkspaceDialog = false"
                :selectedWorkspace="selectedWorkspace"
                :selectedTask="selectedTask"
                :assignFieldName="assignFieldName"
                :selectedAssigneTable="selectedAssigneTable"
                :selectedUserID="selectedUserID"
            />
        </div>

        <div v-if="editLink">
            <LinkedTableDialog
                :open="linkedTableDialog"
                @close="linkedTableDialog = false"
                :linkedID="linkedID"
                :editItemLink="editItemLink"
                :singleSelect="singleSelect"
                :field="field"
                :selectedItems="selectedItems"
                :parentTableID="parentTableID"
                :currentLocalTableID="currentLocalTableID"
                :linkedItemsForCheckbox="linkedItemsForCheckbox"
                :title="linkedTitle"
            />
        </div>

        <MasterLinkDialog
            :open="masterLinkDialog"
            @close="masterLinkDialog = false"
            :field="selectedField"
            :masterLinkTask="masterLinkTask"
        />

        <DatePickerPopover
            v-model:visible="datePickerPopoverVisible"
            :anchor="datePickerPopoverAnchor"
            :value="activeDateValue"
            :activePopoverItem="activeDatePickerPopoverItem"
            @save="handleDateSave"
        />
        <ConfirmDialog
            :open="showFileDeleteConfirm"
            :title="t('projects.full_task_navigation.delete_file')"
            :message="fileDeleteMessage"
            :confirm-label="t('common.button.delete')"
            @confirm="confirmFileDelete"
            @close="showFileDeleteConfirm = false"
        />
    </div>
</template>

<script setup>
import { t } from "@/i18n/index.js";
import { computed, reactive, watch, ref, onMounted, nextTick } from "vue";
import { useWorkspaceStore } from "@/store/workspaces";
import { useWorkspaceRolesStore } from "@/store/workspaceRoles";
import workspaceService from "@/services/workspaceService";
import { useGoToTable } from "@/composables/projects/useGoToTable";
import { useDatePickerPopover } from "@/composables/projects/useDatePickerPopover";
import ShareWorkspaceDialog from "@/components/Projects/Members/ShareWorkspaceDialog.vue";
import LinkedTableDialog from "@/components/Projects/Dialogs/LinkedTableDialog.vue";
import LinkedRecordsCell from "@/components/Projects/Grid/LinkedRecordsCell.vue";
import { useRoute } from "vue-router";
import CommentsComponent from "@/components/Projects/Task/TaskComments.vue";
import { convertSize } from "@/utils/utils";
import { extractErrorMessage } from "@/utils/errors";
import UserAvatar from "@/components/UserAvatar.vue";
import UserAvatarWithText from "@/components/UserAvatarWithText.vue";
import { useUserStore } from "@/store/user";
import MasterLinkDialog from "@/components/Projects/Dialogs/MasterLinkDialog.vue";
import { fileTypeOf } from "@/utils/projects/files";
import { useTaskFiles } from "@/composables/projects/useTaskFiles";
import { useTaskSubtasks } from "@/composables/projects/useTaskSubtasks";
import {
    calculateRowValueForRow,
    formatCellValue,
    isComputedCalculations,
} from "@/utils/projects/formulas";
import DatePickerPopover from "@/components/Projects/Menu/DatePickerPopover.vue";
import useDateOperations from "@/composables/useDateOperations.js";
import {
    PlusIcon,
    ArrowTopRightOnSquareIcon,
    UserPlusIcon,
    PencilIcon,
    PaperClipIcon,
    ArrowDownTrayIcon,
    TrashIcon,
    XMarkIcon,
    CalendarIcon,
} from "@heroicons/vue/24/outline";
import ConfirmDialog from "@/components/ConfirmDialog.vue";
const masterLinkDialog = ref(false);
const currentLocalTableID = ref(null);
const parentTableID = ref(null);
const selectedItems = ref([]);
const route = useRoute();
const workspaceStore = useWorkspaceStore();
const rolesStore = useWorkspaceRolesStore();
const canEditRow = computed(() =>
    rolesStore.canRowAction(setTableID.value || route.params.tid, "update_task"),
);
const userStore = useUserStore();
const linkedTableDialog = ref(false);
const linkedTitle = ref("");
const masterLinkTask = ref(null);

const showFileDeleteConfirm = ref(false);
const pendingFileDelete = ref({
    item: null,
    colIndex: null,
    fileIndex: null,
    fileName: null,
});
const fileDeleteMessage = computed(() => {
    const message = t.value("projects.full_task_navigation.confirm_delete_file");
    const fileName = pendingFileDelete.value.fileName;

    return fileName
        ? `${message} ${t.value("projects.full_task_navigation.file_in_dialog")} ${fileName}`
        : message;
});

import { useAlertStore } from "@/store/alerts";
import { isBlank, linkedDialogTitle } from "@/utils/projects/rows";
import { safeHref } from "@/utils/links";
import { getStatusStyle } from "@/utils/projects/cells";

import { UserCircleIcon, CalendarDaysIcon } from "@heroicons/vue/20/solid";

const { getFileExtension, getFileName, openFile, downloadFile, chooseFiles, uploadTaskFiles } =
    useTaskFiles();

const triggerFileInput = (item, colIndex) =>
    chooseFiles(`[data-file-id="fileInput-fullView-${item.id}-${colIndex}"]`, canEditRow.value);

// showTaskFiles puts a file field's new list in the panel, the open task and
// the grid's row.
const showTaskFiles = (item, headerName, files) => {
    fieldValues[headerName] = files;
    workspaceStore.setSelectedItem({ ...getSelectedItem.value, [headerName]: files });
    tableData.value = tableData.value.map((row) =>
        row.id === item.id ? { ...row, [headerName]: files } : row,
    );
};

const handleFileUpload = async (event, item, colIndex) => {
    const files = Array.from(event.target.files);

    if (files.length === 0) return;

    const header = setFullViewHeaders.value[colIndex];

    if (!header) return;

    try {
        const uploaded = await uploadTaskFiles(files, {
            tableId: setTableID.value,
            taskId: item.id,
            header,
        });

        if (uploaded) {
            const current = Array.isArray(fieldValues[header.name]) ? fieldValues[header.name] : [];

            showTaskFiles(item, header.name, [...current, ...uploaded]);
        }
    } catch (error) {
        console.error("Upload error:", error);
        useAlertStore().showError(
            extractErrorMessage(error, t.value("projects.full_task_navigation.file_upload_failed")),
        );
    }

    event.target.value = "";
};

const removeFile = async (item, colIndex, fileIndex) => {
    const header = setFullViewHeaders.value[colIndex];

    if (!header) return;

    const headerName = header.name;
    const files = fieldValues[headerName];

    if (!files || !Array.isArray(files) || fileIndex >= files.length) return;

    const fileToRemove = files[fileIndex];
    const fileId = fileToRemove.id || fileToRemove.name;

    if (!fileId) {
        useAlertStore().showError(t.value("projects.full_task_navigation.file_id_not_found"));

        return;
    }

    try {
        await workspaceService.deleteFile({
            workspace_id: route.params.id,
            file_id: fileId,
        });

        showTaskFiles(
            item,
            headerName,
            files.filter((_, i) => i !== fileIndex),
        );
    } catch (error) {
        useAlertStore().showError(
            extractErrorMessage(
                error,
                t.value("projects.full_task_navigation.file_removal_failed"),
            ),
        );
    }
};

const openFileDeleteConfirmation = (item, colIndex, fileIndex, fileName) => {
    pendingFileDelete.value = {
        item,
        colIndex,
        fileIndex,
        fileName,
    };
    showFileDeleteConfirm.value = true;
};

const confirmFileDelete = async () => {
    if (!pendingFileDelete.value.item) return;

    const { item, colIndex, fileIndex } = pendingFileDelete.value;

    await removeFile(item, colIndex, fileIndex);
    showFileDeleteConfirm.value = false;
    pendingFileDelete.value = {
        item: null,
        colIndex: null,
        fileIndex: null,
        fileName: null,
    };
};

const { getDate, getTimestampFromDateString, getDateInputValue, isDayField } = useDateOperations();

const formatDateForDisplay = (value) => {
    if (!value) {
        return "";
    }

    let timestamp;

    if (typeof value === "string" && /^\d{4}-\d{2}-\d{2}$/.test(value)) {
        timestamp = getTimestampFromDateString(value);
    } else if (isValidTimestamp(value)) {
        timestamp = Number(value);
    } else {
        return t.value("projects.full_task_navigation.invalid_date");
    }

    return getDate(timestamp);
};

const isValidTimestamp = (value) => {
    const num = Number(value);

    return !isNaN(num) && num >= 0 && num <= 32503680000; // UNIX timestamps up to year 3000
};

const selectedField = ref("");

const openMasterLinkTableDialog = (item, fieldName) => {
    masterLinkTask.value = getSelectedItem.value;
    selectedField.value = fieldName;
    masterLinkDialog.value = true;
};

const selectedItem = computed({
    get() {
        return workspaceStore.getSelectedItem;
    },
    set(value) {
        workspaceStore.setSelectedItem(value);
    },
});

const setterFullViewHeaders = computed({
    get() {
        return workspaceStore.getFullViewHeaders;
    },
    set(value) {
        workspaceStore.setFullViewHeaders(value);
    },
});

function showFullView(item, tableOrName, subtask) {
    var tableID = tableOrName?.parent_table_id;

    if (subtask) {
        tableID = route.params.tid;
    }

    setTableID.value = subtask ? route.params.tid : tableOrName?.parent_table_id;

    workspaceService
        .getItemForTableByID({
            workspace_id: route.params.id,
            table_id: tableID,
            task_id: item.id,
        })
        .then((response) => {
            selectedItem.value = response.data.item;

            setterFullViewHeaders.value = response.data.headers;

            showFullTask.value = true;
        })
        .catch((error) => {
            useAlertStore().showError(extractErrorMessage(error));
        });
}

const {
    visible: datePickerPopoverVisible,
    anchor: datePickerPopoverAnchor,
    item: activeDatePickerPopoverItem,
    colIndex: activeDatePickerColIndex,
    toggle: toggleDatePickerPopover,
} = useDatePickerPopover({ canOpen: () => canEditRow.value });

const activeDateValue = computed(() => {
    const header = setFullViewHeaders.value?.[activeDatePickerColIndex.value];

    return getDateInputValue(fieldValues[header?.name], isDayField(header));
});

function toYMD(dateLike) {
    if (!dateLike) return "";
    if (typeof dateLike === "string") return dateLike.slice(0, 10);
    const d = new Date(dateLike);

    if (isNaN(d.getTime())) return "";

    return d.toISOString().slice(0, 10);
}

function handleDateSave(picked) {
    const colIndex = activeDatePickerColIndex.value;
    const item = getSelectedItem.value;

    if (colIndex == null || !item) {
        datePickerPopoverVisible.value = false;

        return;
    }

    const fieldName = setFullViewHeaders.value[colIndex]?.name;

    if (!fieldName) {
        datePickerPopoverVisible.value = false;

        return;
    }

    const ymd = toYMD(picked);

    fieldValues[fieldName] = ymd;
    workspaceStore.setSelectedItem({
        ...getSelectedItem.value,
        [fieldName]: ymd,
    });

    saveValue(fieldName, item, ymd);

    datePickerPopoverVisible.value = false;
    activeDatePickerPopoverItem.value = null;
    activeDatePickerColIndex.value = null;
}

const getTableName = (tableId) => workspaceStore.tableNameOf(tableId);

const workspaceTables = computed({
    get() {
        return workspaceStore.getWorkspaceTables;
    },
    set(value) {
        workspaceStore.setWorkspaceTables(value);
    },
});

const workspaceTablesForLinked = computed(() => workspaceStore.getWorkspaceTablesForLinked);

const { goToTable } = useGoToTable();

const setFullViewHeaders = computed(() => {
    const headers = workspaceStore.getFullViewHeaders || [];

    return headers.filter((header) => header.name !== "id");
});

const activeTab = ref("details");

const closeModal = () => {
    showFullTask.value = false;
};

const showFullTask = computed({
    get() {
        return workspaceStore.getFullTask;
    },
    set(value) {
        workspaceStore.setFullTask(value);
    },
});

const editableField = reactive({
    colIndex: null,
    item: null,
});

const resetEditState = () => {
    editableField.colIndex = null;
    editableField.item = null;
    editableValue.value = "";
};
const editableValue = ref("");

const tableData = computed({
    get() {
        return workspaceStore.getTableData;
    },
    set(value) {
        workspaceStore.setTableData(value);
    },
});

const redirectToLinkedTable = (header) => {
    if (header?.parent_table_id) goToTable(header.parent_table_id);
};

const taskLists = computed({
    get() {
        return workspaceStore.getKanbanTaskList || [];
    },
    set(value) {
        workspaceStore.setKanbanTaskList(value);
    },
});

const saveValue = (fieldName, item, value) => {
    if (!canEditRow.value) return;
    if (route.params.fid) {
        for (let i = 0; i < taskLists.value.length; i++) {
            const column = taskLists.value[i];

            if (Array.isArray(column.order) || Array.isArray(column.tasks)) {
                let updated = false;

                const updatedOrder =
                    column.order?.map((orderedItem) => {
                        if (orderedItem.id === item.id) {
                            updated = true;

                            return {
                                ...orderedItem,
                                [fieldName]: value,
                            };
                        }

                        return orderedItem;
                    }) || column.order;

                const updatedTasks =
                    column.tasks?.map((task) => {
                        if (task.id === item.id) {
                            updated = true;

                            return {
                                ...task,
                                [fieldName]: value,
                            };
                        }

                        return task;
                    }) || column.tasks;

                if (updated) {
                    taskLists.value[i] = {
                        ...column,
                        order: updatedOrder,
                        tasks: updatedTasks,
                    };
                }
            }
        }
    }

    let newValue = typeof value === "string" ? value : String(value);

    // Convert date fields to timestamp and ensure value is a string before sending to the backend
    if (fieldName === "start_date" || fieldName === "due_date") {
        newValue = String(getTimestampFromDateString(newValue));
        getSelectedItem.value[fieldName] = newValue;
    }

    workspaceService
        .updateTask({
            workspace_id: route.params.id,
            table_id: setTableID.value || route.params.tid,
            task_id: item.id,
            field: fieldName,
            value: newValue,
        })
        .then((response) => {
            let saved = Object.hasOwn(response.data, fieldName)
                ? response.data[fieldName]
                : newValue;

            if (
                ["start_date", "due_date"].includes(fieldName) &&
                saved != null &&
                !isNaN(Number(saved))
            ) {
                saved = parseInt(saved, 10);
            }

            const nowMs = Date.now();

            item[fieldName] = saved;
            item.updated_at = nowMs;

            // The grid draws a row again only when it is a new object.
            tableData.value = tableData.value.map((row) =>
                row.id == item.id ? { ...row, [fieldName]: saved, updated_at: nowMs } : row,
            );
            workspaceStore.announceSavedTaskField(item.id, fieldName, saved);
            resetEditState();
        })
        .catch((error) => {
            useAlertStore().showError(extractErrorMessage(error));
            reloadAfterFailedSave(item.id);
        });
};

// The panel and the board show a value before it is saved, so one the
// server refuses is undone by reading the task, and the view, again.
function reloadAfterFailedSave(taskId) {
    workspaceStore.bumpGridReloadToken();
    workspaceService
        .getItemForTableByID({
            workspace_id: route.params.id,
            table_id: setTableID.value || route.params.tid,
            task_id: taskId,
        })
        .then((response) => {
            if (workspaceStore.getSelectedItem?.id === taskId) {
                workspaceStore.setSelectedItem(response.data.item);
            }
        })
        .catch(() => {});
}

const setTableID = computed({
    get() {
        return workspaceStore.getTableID;
    },
    set(value) {
        workspaceStore.setTableID(value);
    },
});

// A checkbox is saved as 1 from the grid and read back from the server as 1
// or "1", and as true once ticked here.
const isChecked = (value) => value === true || value === 1 || value === "1" || value === "true";

const toggleCheckbox = (fieldName, event) => {
    if (!canEditRow.value) {
        event.preventDefault();

        return;
    }

    const checked = event.target.checked;

    const updatedItem = { ...getSelectedItem.value };

    updatedItem[fieldName] = checked;

    workspaceStore.setSelectedItem(updatedItem);

    saveValue(fieldName, getSelectedItem.value, checked);
};

const { subtasks } = useTaskSubtasks(
    () => route.params.id,
    () => workspaceStore.getTableID || route.params.tid,
    () => workspaceStore.getSelectedItem?.id,
);

const field = ref("");
const linkedID = ref(null);
const editLink = ref(false);
const singleSelect = ref(false);
let editItemLink = ref({});

function openLinkedTableDialog(passed, headerName, linked_id) {
    if (!canEditRow.value) {
        useAlertStore().showError(t.value("projects.grid_view.no_permission_edit_fields"));

        return;
    }

    const fieldName = headerName;

    editItemLink.value = getSelectedItem.value;

    const header = setFullViewHeaders.value.find((h) => h.name === headerName);

    if (!header) {
        return;
    }

    if (!header.single_select) {
        // parent_table_id is empty for one-directional links; fall back to the junction table's parent
        let targetTableId = header.parent_table_id;

        if (!targetTableId) {
            const junction = workspaceTablesForLinked.value.find((t) => t.id === header.linked_id);

            targetTableId = junction?.parent_table_id || "";
        }

        if (!rolesStore.hasTablePermission(targetTableId, "view")) {
            useAlertStore().showError(t.value("projects.grid_view.no_permission_view_table"));

            return;
        }
    }

    parentTableID.value = header.single_select ? header.linked_id : header.parent_table_id;
    singleSelect.value = header.single_select || header.value || null;

    linkedID.value = linked_id;
    field.value = fieldName;
    linkedTitle.value = linkedDialogTitle(header, [
        ...(workspaceTablesForLinked.value || []),
        ...(workspaceTables.value || []),
    ]);

    const raw = passed ?? getSelectedItem.value?.[fieldName];

    linkedItemsForCheckbox.value = Array.isArray(raw)
        ? raw.map((i) => i?.id).filter(Boolean)
        : raw && raw.id
          ? [raw.id]
          : [];

    editLink.value = true;
    linkedTableDialog.value = true;
}

const linkedItemsForCheckbox = ref([]);

const getSelectedItem = computed({
    get() {
        return workspaceStore.getSelectedItem;
    },
    set(value) {
        workspaceStore.setSelectedItem(value);
    },
});

const fieldValues = reactive({});

// The field being typed in keeps what was typed when someone else changes the
// task, and is saved when it is left.
const typingField = ref(null);
let shownTaskId = null;

watch(
    getSelectedItem,
    (newItem) => {
        if (newItem) {
            const sameTask = newItem.id === shownTaskId;

            shownTaskId = newItem.id;
            setFullViewHeaders.value.forEach((header) => {
                if (sameTask && header.name === typingField.value) return;
                fieldValues[header.name] = newItem[header.name] || "";
            });
            const ids = [newItem.assignee, newItem.created_by].filter(Boolean);

            if (ids.length) userStore.ensureUsers(ids);
        }
    },
    { immediate: true },
);

watch(
    fieldValues,
    (newValues) => {
        const updatedItem = { ...getSelectedItem.value, ...newValues };

        workspaceStore.setSelectedItem(updatedItem);
    },
    { deep: true },
);

onMounted(() => {
    currentLocalTableID.value = setTableID.value;
    // Ensure the current task's assignee and creator are in cache
    const item = workspaceStore.getSelectedItem;
    const ids = [item?.assignee, item?.created_by].filter(Boolean);

    if (ids.length) userStore.ensureUsers(ids);
});

const isEditing = reactive({});

const nameIndex = computed(() =>
    setFullViewHeaders.value.findIndex((header) => header.name === "name"),
);

// What each field held when its edit began. What is typed reaches the
// selected task as it is typed, so the task cannot tell what changed.
const valuesBeforeEdit = {};

const rememberValue = (index) => {
    const fieldName = setFullViewHeaders.value[index]?.name;

    if (fieldName) valuesBeforeEdit[fieldName] = fieldValues[fieldName];
};

const enableEditing = (index) => {
    if (!canEditRow.value) return;
    rememberValue(index);
    isEditing[index] = true;
    nextTick(() => document.getElementById(`task-field-${index}`)?.focus());
};

const cancelFieldEdit = (index) => {
    const fieldName = setFullViewHeaders.value[index]?.name;

    if (fieldName && fieldName in valuesBeforeEdit)
        fieldValues[fieldName] = valuesBeforeEdit[fieldName];
    isEditing[index] = false;
};

const saveEdit = (index) => {
    if (!canEditRow.value) return;

    // An input closed by Enter or Esc still blurs as it goes.
    if (isEditing[index] === false) return;

    const header = setFullViewHeaders.value[index];
    const fieldName = header.name;

    if (
        fieldName in valuesBeforeEdit &&
        String(fieldValues[fieldName] ?? "") === String(valuesBeforeEdit[fieldName] ?? "")
    ) {
        isEditing[index] = false;

        return;
    }

    valuesBeforeEdit[fieldName] = fieldValues[fieldName];

    const updatedItem = { ...getSelectedItem.value };

    updatedItem[fieldName] = fieldValues[fieldName];

    workspaceStore.setSelectedItem(updatedItem);

    isEditing[index] = false;

    saveValue(fieldName, getSelectedItem.value, fieldValues[fieldName]);
};

const isRegularInput = (header) => {
    return (
        ["TEXT", "VARCHAR"].includes(header.header_type) &&
        !["assignee", "status"].includes(header.header_usage)
    );
};

// A calculation over no fields holds a number typed in by hand.
const isNumberInput = (header) =>
    ["DECIMAL", "INT", "BIGINT"].includes(header.header_type) ||
    (header.header_usage || "").toLowerCase() === "calculations";

const calculatedValue = (header) =>
    formatCellValue(calculateRowValueForRow(header, fieldValues), header).trim();

const selectedTask = ref(null);
const selectedAssigneTable = ref(null);
const shareWorkspaceDialog = ref(false);
const assignFieldName = ref("");
const selectedWorkspace = ref(null);

const workspaceDetails = computed({
    get() {
        return workspaceStore.getWorkspaceDetails;
    },
    set(value) {
        workspaceStore.setWorkspaceDetails(value);
    },
});

const openAssignDialog = (item, colIndex) => {
    if (!canEditRow.value) return;
    try {
        const workspace = workspaceDetails.value;

        if (!workspace) {
            useAlertStore().showError(
                t.value("projects.full_task_navigation.workspace_not_found_in_store"),
            );

            return;
        }

        selectedWorkspace.value = workspace;

        selectedAssigneTable.value =
            workspace.tables?.find((table) => table.id === setTableID.value) || null;

        assignFieldName.value = setFullViewHeaders.value?.[colIndex]?.name;

        selectedTask.value = item;

        selectedUserID.value = selectedTask.value?.[assignFieldName.value] ?? null;

        shareWorkspaceDialog.value = true;
    } catch (error) {
        useAlertStore().showError(
            extractErrorMessage(
                error,
                t.value("projects.full_task_navigation.failed_to_open_assign_dialog"),
            ),
        );
    }
};

const selectedUserID = ref(null);
</script>
