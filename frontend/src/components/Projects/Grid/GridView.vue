<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div class="relative">
        <div
            v-if="awaitingFilterLoad && !isDialog"
            class="absolute inset-0 z-20 flex items-center justify-center bg-white"
        >
            <div class="flex flex-col items-center gap-y-3">
                <BaseSpinner size="lg" class="text-indigo-600" :label="t('common.label.loading')" />
                <p class="text-sm text-gray-500">{{ t("common.label.loading") }}</p>
            </div>
        </div>
        <div class="table-container" :style="dialogStyle" :data-table-id="tableUniqueId">
            <GridHeaderRow
                v-model:headers="tableHeaders"
                :column-widths="columnWidths"
                :table-id="isDialog ? dialogTableID : route.params.tid"
                :is-dialog="isDialog"
                :is-data-loaded="isDataLoaded"
                :is-drag-disabled="isDragDisabled"
                :can-edit-fields="canEditFields"
                :can-manage-fields="canManageFields"
                @reorder="onHeaderDrop"
                @resize-start="startResize"
                @open-menu="(event, header) => toggleItemMenuPopover(event, header, 'field')"
                @add-field="(event) => fieldEditor?.toggle(event)"
                @update-header="patchHeader"
            />
            <div class="table-content">
                <div class="grid-container">
                    <div
                        class="grid-row task-row"
                        v-for="(item, index) in paginatedData"
                        :key="item.id"
                        v-memo="[
                            _rowVersion,
                            columnWidths,
                            userStore.knownUserCount,
                            item,
                            item._indent,
                            getSubtaskCount(item.id),
                            _rowEditVersions[item.id] || 0,
                            expandedTasks.has(item.id),
                            item.itemSelected,
                            canEditRow,
                            editableField.item?.id === item.id && lastClickedElement !== 'checkbox'
                                ? editableField.colIndex
                                : -1,
                        ]"
                        :data-id="item.id"
                        :ref="index === paginatedData.length - 1 ? setLastTaskRef : null"
                    >
                        <div
                            v-for="(header, colIndex) in rowHeaders"
                            :key="header.name"
                            :data-column-index="colIndex"
                            class="grid-item-text bg-white"
                            :style="
                                header.visible
                                    ? {
                                          width: columnWidths[colIndex] + 'px',
                                          display: 'flex',
                                      }
                                    : { width: '0px', display: 'none' }
                            "
                            @contextmenu.prevent="toggleItemMenuPopover($event, item, 'item')"
                        >
                            <div
                                v-if="header.visible"
                                @click="startCellEdit(colIndex, item)"
                                :class="[
                                    'clickable-area group/cell',
                                    !canEditRow && 'cursor-not-allowed',
                                    canEditRow &&
                                        colIndex > 1 &&
                                        !isEditing(colIndex, item) &&
                                        'rounded-lg outline outline-1 -outline-offset-4 outline-transparent transition-[outline-color] delay-75 duration-150 hover:outline-gray-300',
                                ]"
                            >
                                <div
                                    v-if="getHeaderType(colIndex).header_type === 'TINYINT'"
                                    class="text-left alignLeft flex items-center"
                                >
                                    <input
                                        type="checkbox"
                                        :checked="!!item[header.name]"
                                        :disabled="!canEditRow"
                                        @change="toggleCheckbox(item, colIndex, $event)"
                                        class="h-4 w-4 rounded border-gray-300 text-indigo-600 focus:ring-indigo-600 disabled:opacity-50 disabled:cursor-not-allowed"
                                    />
                                </div>

                                <div
                                    v-else-if="
                                        ['status_type'].includes(
                                            getHeaderType(colIndex).header_name,
                                        )
                                    "
                                    class="w-full h-full"
                                >
                                    <StatusTypeMenu
                                        :value="item[header.name] || ''"
                                        :label="statusTypeLabel(item[header.name])"
                                        :options="statusTypeOptions"
                                        :disabled="!canEditRow"
                                        @select="updateStatusType(item, colIndex, $event)"
                                        @denied="
                                            useAlertStore().showError(
                                                t('projects.grid_view.no_permission_edit_fields'),
                                            )
                                        "
                                    />
                                </div>

                                <div
                                    v-else-if="
                                        ['TEXT', 'VARCHAR', 'INT', 'FILE'].includes(
                                            getHeaderType(colIndex).header_type,
                                        ) && getHeaderType(colIndex).linked_id == ''
                                    "
                                    style="width: 100%"
                                >
                                    <div
                                        v-if="
                                            getHeaderType(colIndex).header_usage ==
                                                'default_assignee' ||
                                            getHeaderType(colIndex).header_usage == 'assignee'
                                        "
                                    >
                                        <div
                                            v-if="
                                                item[header.name] !== '' &&
                                                item[header.name] !== null &&
                                                typeof item[header.name] !== 'undefined'
                                            "
                                            style="
                                                display: flex;
                                                align-items: center;
                                                gap: 6px;
                                                cursor: pointer;
                                            "
                                            @click="
                                                getHeaderType(colIndex).header_name !== 'created_by'
                                                    ? openAssignDialog(item, colIndex)
                                                    : null
                                            "
                                        >
                                            <UserAvatarWithText
                                                class="gap-1.5"
                                                :user-id="item[header.name]"
                                                avatar-class="h-5 w-5 shrink-0"
                                                text-class="whitespace-nowrap text-[13px]"
                                            />
                                        </div>
                                        <div
                                            v-else
                                            @click="
                                                getHeaderType(colIndex).header_name !== 'created_by'
                                                    ? openAssignDialog(item, colIndex)
                                                    : null
                                            "
                                        >
                                            <button
                                                type="button"
                                                :aria-label="t('projects.grid_view.assignee')"
                                                class="flex items-center"
                                            >
                                                <UserPlusIcon
                                                    class="block h-4 w-4 text-gray-400"
                                                    aria-hidden="true"
                                                />
                                            </button>
                                        </div>
                                    </div>
                                    <div
                                        v-else-if="getHeaderType(colIndex).header_usage == 'file'"
                                        class="w-full min-w-0"
                                    >
                                        <div class="mr-1.5 flex min-w-0 flex-nowrap items-center">
                                            <div
                                                v-if="
                                                    Array.isArray(item[header.name]) &&
                                                    item[header.name].length > 0
                                                "
                                                class="flex min-w-0 flex-1 items-center"
                                            >
                                                <button
                                                    type="button"
                                                    class="mr-1.5 flex flex-none items-center"
                                                    :title="t('projects.grid_view.files_title')"
                                                    :aria-label="
                                                        t('projects.grid_view.files_title')
                                                    "
                                                    @click.stop="triggerFileInput(item, colIndex)"
                                                >
                                                    <PlusIcon
                                                        class="block h-4 w-4 text-gray-500"
                                                        aria-hidden="true"
                                                    />
                                                </button>
                                                <FilesCell
                                                    :files="item[header.name]"
                                                    @open="openFile($event, item[header.name])"
                                                />
                                            </div>

                                            <div v-else>
                                                <button
                                                    type="button"
                                                    class="flex items-center"
                                                    :aria-label="t('projects.grid_view.upload')"
                                                >
                                                    <PaperClipIcon
                                                        class="block h-4 w-4 text-gray-400"
                                                        aria-hidden="true"
                                                    />
                                                </button>
                                            </div>

                                            <input
                                                type="file"
                                                :data-file-id="`fileInput-${item.id}-${colIndex}`"
                                                multiple
                                                @click.stop
                                                @change="handleFileUpload($event, item, colIndex)"
                                                style="display: none"
                                            />
                                        </div>
                                    </div>
                                    <div
                                        v-else-if="
                                            getHeaderType(colIndex).header_usage == 'master link'
                                        "
                                    >
                                        <div
                                            v-if="
                                                item[header.name] !== '' &&
                                                item[header.name] !== null &&
                                                typeof item[header.name] !== 'undefined'
                                            "
                                            class="group flex min-w-0 items-center gap-x-1.5"
                                        >
                                            <button
                                                type="button"
                                                class="inline-flex min-w-0 items-center gap-x-1 rounded-full bg-gray-100 px-2 py-0.5 text-xs font-medium text-gray-700 hover:bg-gray-200"
                                                :title="
                                                    t(
                                                        'projects.full_task_navigation.go_to_linked_table',
                                                    )
                                                "
                                                @click.stop="goToTable(item[header.name])"
                                            >
                                                <ArrowTopRightOnSquareIcon
                                                    class="h-3 w-3 flex-none text-gray-500"
                                                    aria-hidden="true"
                                                />
                                                <span class="truncate">{{
                                                    getTableName(item[header.name])
                                                }}</span>
                                            </button>

                                            <button
                                                type="button"
                                                class="flex-none rounded p-0.5 text-gray-400 opacity-0 hover:bg-gray-100 hover:text-gray-600 focus:opacity-100 group-hover:opacity-100 [@media(hover:none)]:opacity-100"
                                                :title="
                                                    t(
                                                        'projects.full_task_navigation.edit_linked_table',
                                                    )
                                                "
                                                :aria-label="
                                                    t(
                                                        'projects.full_task_navigation.edit_linked_table',
                                                    )
                                                "
                                                @click.stop="
                                                    openMasterLinkTableDialog(item, colIndex)
                                                "
                                            >
                                                <PencilIcon
                                                    class="h-3.5 w-3.5"
                                                    aria-hidden="true"
                                                />
                                            </button>
                                        </div>

                                        <div v-else>
                                            <button
                                                type="button"
                                                class="flex items-center"
                                                :aria-label="t('common.button.add')"
                                            >
                                                <span class="text-[13px] text-gray-400">–</span>
                                            </button>
                                        </div>
                                    </div>

                                    <div v-else-if="getHeaderType(colIndex).header_usage == 'url'">
                                        <template v-if="isEditing(colIndex, item)">
                                            <div class="alignTextCenter">
                                                <InlineEditInput
                                                    :value="item[tableHeaders[colIndex].name]"
                                                    :colIndex="colIndex"
                                                    :item="item"
                                                    :headerType="
                                                        getHeaderType(colIndex).header_type
                                                    "
                                                    @save="saveValue"
                                                    @cancel="cancelEdit"
                                                />
                                            </div>
                                        </template>

                                        <template v-else>
                                            <div
                                                v-if="item[header.name]"
                                                class="group/url"
                                                style="display: flex; align-items: center; gap: 6px"
                                            >
                                                <a
                                                    :href="safeHref(item[header.name])"
                                                    target="_blank"
                                                    rel="noopener noreferrer"
                                                    :style="{
                                                        fontSize: '13px',
                                                        color: '#3b82f6',
                                                        textDecoration: 'underline',
                                                        cursor: 'pointer',
                                                        overflow: 'hidden',
                                                        textOverflow: 'ellipsis',
                                                        whiteSpace: 'nowrap',
                                                        maxWidth: '100%',
                                                        display: 'inline-block',
                                                    }"
                                                    :title="item[header.name]"
                                                    @click.stop
                                                >
                                                    {{ item[header.name] }}
                                                </a>

                                                <button
                                                    :title="
                                                        t('projects.full_task_navigation.edit_url')
                                                    "
                                                    :aria-label="
                                                        t('projects.full_task_navigation.edit_url')
                                                    "
                                                    class="hidden items-center p-1 text-gray-700 bg-gray-200 rounded shadow hover:bg-gray-300 focus:flex focus:outline-none group-hover/url:flex [@media(hover:none)]:flex focus:ring-1 focus:ring-gray-400 focus:ring-offset-0"
                                                    @pointerdown.stop.prevent="
                                                        () => {
                                                            if (canEditRow) {
                                                                editableField.colIndex = colIndex;
                                                                editableField.item = item;
                                                            }
                                                        }
                                                    "
                                                    style="width: 20px; height: 20px"
                                                >
                                                    <PencilIcon
                                                        class="h-[10px] w-[10px] text-gray-600 hover:text-gray-800"
                                                        aria-hidden="true"
                                                    />
                                                </button>
                                            </div>

                                            <div v-else>
                                                <button
                                                    type="button"
                                                    class="flex items-center"
                                                    :aria-label="t('common.button.add')"
                                                >
                                                    <span class="text-[13px] text-gray-400">–</span>
                                                </button>
                                            </div>
                                        </template>
                                    </div>

                                    <div v-else>
                                        <div
                                            class="text-left alignLeft text-sm font-medium leading-6 text-gray-600 flex items-center"
                                            style="margin-right: 6px; gap: 6px"
                                        >
                                            <input
                                                v-if="
                                                    getHeaderType(colIndex).header_name ===
                                                        'name' && isDialog
                                                "
                                                v-model="item.itemSelected"
                                                id="edit"
                                                aria-describedby="candidates-description"
                                                name="edit"
                                                type="checkbox"
                                                class="h-4 w-4 rounded border-gray-300 text-indigo-600 focus:ring-indigo-600 flex-shrink-0"
                                                @click.stop="handleCheckboxClick($event, item)"
                                                @change="updateSelectedItems"
                                            />

                                            <template v-if="isEditing(colIndex, item)">
                                                <div
                                                    v-if="isProtectedStatus(colIndex, item)"
                                                    class="text-left alignLeft text-sm font-medium leading-6 text-gray-600"
                                                >
                                                    {{ item.name }}
                                                </div>
                                                <InlineEditInput
                                                    v-else
                                                    :value="item[tableHeaders[colIndex].name]"
                                                    :colIndex="colIndex"
                                                    :item="item"
                                                    :headerType="
                                                        getHeaderType(colIndex).header_usage ===
                                                        'calculations'
                                                            ? 'INT'
                                                            : getHeaderType(colIndex).header_type
                                                    "
                                                    @save="saveValue"
                                                    @cancel="cancelEdit"
                                                />
                                            </template>

                                            <template v-else>
                                                <div
                                                    class="flex items-center w-full min-w-0"
                                                    :style="{
                                                        paddingLeft:
                                                            getHeaderType(colIndex).header_name ===
                                                                'name' &&
                                                            (hasChildren(item.id) ||
                                                                item._is_subtask ||
                                                                (item.parent_task_id &&
                                                                    item.parent_task_id !==
                                                                        item.id))
                                                                ? (item._indent || 0) * 20 + 'px'
                                                                : '0px',
                                                    }"
                                                >
                                                    <span
                                                        v-if="
                                                            getHeaderType(colIndex).header_name ===
                                                            'name'
                                                        "
                                                        class="flex-shrink-0 w-4 h-4 mr-1 flex items-center justify-center"
                                                    >
                                                        <button
                                                            v-if="hasChildren(item.id)"
                                                            class="cursor-pointer flex-shrink-0"
                                                            @click.stop="toggleExpand(item.id)"
                                                            @dblclick.stop
                                                        >
                                                            <component
                                                                :is="
                                                                    expandedTasks.has(item.id)
                                                                        ? ChevronDownIcon
                                                                        : ChevronRightIcon
                                                                "
                                                                class="w-4 h-4 text-gray-500 transition-transform flex-shrink-0"
                                                            />
                                                        </button>
                                                        <span v-else class="w-4 h-4"></span>
                                                    </span>

                                                    <span
                                                        v-if="
                                                            getHeaderType(colIndex).header_name ===
                                                                'name' &&
                                                            (item._is_subtask ||
                                                                (item.parent_task_id &&
                                                                    item.parent_task_id !==
                                                                        item.id))
                                                        "
                                                        class="flex-shrink-0 mr-1"
                                                        :title="
                                                            t('projects.grid_view.show_parent_task')
                                                        "
                                                    >
                                                        <Square2StackIcon
                                                            @click.stop="
                                                                showParentTask(
                                                                    item.parent_task_id,
                                                                    colIndex,
                                                                    item,
                                                                )
                                                            "
                                                            class="w-4 h-4 flex-shrink-0 cursor-pointer"
                                                        />
                                                    </span>

                                                    <div
                                                        class="flex items-center justify-between w-full min-w-0 flex-1"
                                                    >
                                                        <div
                                                            class="flex items-center min-w-0 flex-1"
                                                        >
                                                            <span
                                                                class="ml-1 truncate cursor-pointer min-w-0 flex-1"
                                                                :class="
                                                                    getHeaderType(colIndex)
                                                                        .header_name === 'name'
                                                                        ? 'text-sm font-medium text-gray-800'
                                                                        : 'text-[13px] font-normal text-gray-700'
                                                                "
                                                                :title="item[header.name]"
                                                                @click="
                                                                    getHeaderType(colIndex)
                                                                        .header_name === 'name' &&
                                                                    !isDialog &&
                                                                    openTaskOnClick(item, colIndex)
                                                                "
                                                                @dblclick.stop="
                                                                    getHeaderType(colIndex)
                                                                        .header_name === 'name' &&
                                                                    !isDialog &&
                                                                    renameTask(colIndex, item)
                                                                "
                                                            >
                                                                <div
                                                                    v-if="
                                                                        ['calculations'].includes(
                                                                            getHeaderType(colIndex)
                                                                                .header_usage,
                                                                        )
                                                                    "
                                                                    class="truncate"
                                                                >
                                                                    <span
                                                                        v-if="
                                                                            isBlank(
                                                                                getCellDisplayValue(
                                                                                    header,
                                                                                    item,
                                                                                ),
                                                                            )
                                                                        "
                                                                        class="text-[13px] text-gray-400"
                                                                        >–</span
                                                                    >
                                                                    <template v-else>{{
                                                                        getCellDisplayValue(
                                                                            header,
                                                                            item,
                                                                        )
                                                                    }}</template>
                                                                </div>
                                                                <div
                                                                    v-else
                                                                    class="truncate min-w-0 flex-1"
                                                                >
                                                                    <span
                                                                        v-if="
                                                                            isBlank(
                                                                                item[header.name],
                                                                            )
                                                                        "
                                                                        class="text-gray-400"
                                                                        >–</span
                                                                    >
                                                                    <template v-else>{{
                                                                        item[header.name]
                                                                    }}</template>
                                                                </div>
                                                            </span>
                                                        </div>

                                                        <div
                                                            class="flex items-center gap-1 flex-shrink-0 ml-2"
                                                        >
                                                            <button
                                                                v-if="
                                                                    getHeaderType(colIndex)
                                                                        .header_name === 'name' &&
                                                                    !isDialog &&
                                                                    canEditRow
                                                                "
                                                                type="button"
                                                                class="rounded p-0.5 text-gray-400 opacity-0 hover:bg-gray-100 hover:text-gray-600 focus:opacity-100 group-hover/cell:opacity-100 [@media(hover:none)]:opacity-100"
                                                                :title="t('common.button.rename')"
                                                                :aria-label="
                                                                    t('common.button.rename')
                                                                "
                                                                @click.stop="
                                                                    startCellEdit(colIndex, item, {
                                                                        name: true,
                                                                    })
                                                                "
                                                            >
                                                                <PencilIcon
                                                                    class="h-3.5 w-3.5"
                                                                    aria-hidden="true"
                                                                />
                                                            </button>
                                                            <span
                                                                v-if="
                                                                    getHeaderType(colIndex)
                                                                        .header_name === 'name' &&
                                                                    getSubtaskCount(item.id) > 0
                                                                "
                                                                class="flex items-center gap-1 text-gray-400 text-xs flex-shrink-0"
                                                                :title="`${getSubtaskCount(item.id)} ${t('projects.assigned_to_me.subtasks')}`"
                                                            >
                                                                <Square2StackIcon
                                                                    class="w-4 h-4 flex-shrink-0"
                                                                />
                                                                <span class="flex-shrink-0">{{
                                                                    getSubtaskCount(item.id)
                                                                }}</span>
                                                            </span>

                                                            <button
                                                                v-if="
                                                                    getHeaderType(colIndex)
                                                                        .header_name === 'name' &&
                                                                    !props.isSingleSelect
                                                                "
                                                                class="p-1 rounded-full hover:bg-gray-100 focus:outline-none focus:ring-1 focus:ring-inset focus:ring-gray-300 flex-shrink-0"
                                                                @click.stop="addSubtask(item)"
                                                                :title="
                                                                    t(
                                                                        'projects.grid_view.add_subtask',
                                                                    )
                                                                "
                                                            >
                                                                <PlusCircleIcon
                                                                    class="w-4 h-4 text-gray-500 flex-shrink-0"
                                                                />
                                                            </button>
                                                        </div>

                                                        <button
                                                            v-if="
                                                                isDialog &&
                                                                getHeaderType(colIndex)
                                                                    .header_name === 'name' &&
                                                                !isProtectedStatus(colIndex, item)
                                                            "
                                                            type="button"
                                                            class="ml-2"
                                                            @click="openDeleteConfirm(item.id)"
                                                        >
                                                            <TrashIcon
                                                                class="w-4 h-4 flex-shrink-0"
                                                            />
                                                        </button>
                                                    </div>
                                                </div>
                                            </template>
                                        </div>
                                    </div>
                                </div>

                                <div
                                    v-else-if="
                                        ['color'].includes(getHeaderType(colIndex).header_name)
                                    "
                                >
                                    <input
                                        type="color"
                                        class="p-1 h-7 w-7 block bg-white border border-gray-200 cursor-pointer rounded-lg disabled:opacity-50 disabled:pointer-events-none"
                                        id="hs-color-input"
                                        :value="item[header.name] || '#000000'"
                                        :title="t('projects.grid_view.choose_your_color')"
                                        :disabled="!canEditRow"
                                        @change="canEditRow && saveValue(colIndex, item, $event)"
                                    />
                                </div>

                                <div
                                    v-else-if="
                                        ['DATE'].includes(getHeaderType(colIndex).header_type) &&
                                        getHeaderType(colIndex).linked_id == ''
                                    "
                                    style="width: 100%"
                                >
                                    <button
                                        type="button"
                                        class="flex w-full min-w-0 items-center gap-x-1.5 text-left text-[13px] text-gray-700"
                                        :title="t('projects.grid_view.pick_a_date')"
                                        @click.stop="
                                            toggleDatePickerPopover($event, item, colIndex)
                                        "
                                    >
                                        <CalendarIcon
                                            class="h-4 w-4 flex-none text-gray-400"
                                            aria-hidden="true"
                                        />
                                        <span v-if="item[header.name]" class="truncate">{{
                                            dateCellText(header, item[header.name])
                                        }}</span>
                                    </button>
                                </div>

                                <div
                                    v-else-if="
                                        ['DECIMAL', 'BIGINT'].includes(
                                            getHeaderType(colIndex).header_type,
                                        ) && getHeaderType(colIndex).linked_id == ''
                                    "
                                    style="width: 100%"
                                >
                                    <div
                                        v-if="
                                            getHeaderType(colIndex).header_usage === 'default_date'
                                        "
                                    >
                                        <div
                                            v-if="
                                                ['created_at', 'updated_at', 'deleted_at'].includes(
                                                    getHeaderType(colIndex).header_name,
                                                )
                                            "
                                            class="w-full min-w-0 text-left leading-6 text-gray-700"
                                        >
                                            <div
                                                class="truncate"
                                                :title="
                                                    getDateAndTime(
                                                        item[header.name]
                                                            ? +item[header.name]
                                                            : null,
                                                    )
                                                "
                                            >
                                                {{
                                                    getDate(
                                                        item[header.name]
                                                            ? +item[header.name]
                                                            : null,
                                                    ) || " "
                                                }}
                                            </div>
                                        </div>

                                        <div
                                            v-else-if="
                                                ['start_date', 'due_date'].includes(
                                                    getHeaderType(colIndex).header_name,
                                                )
                                            "
                                            class="w-full min-w-0"
                                        >
                                            <button
                                                type="button"
                                                class="flex w-full min-w-0 items-center gap-x-1.5 text-left text-[13px] text-gray-700"
                                                :title="t('projects.grid_view.pick_a_date')"
                                                @click.stop="
                                                    toggleDatePickerPopover($event, item, colIndex)
                                                "
                                            >
                                                <CalendarIcon
                                                    class="h-4 w-4 flex-none text-gray-400"
                                                    aria-hidden="true"
                                                />
                                                <span v-if="item[header.name]" class="truncate">{{
                                                    dateCellText(header, item[header.name])
                                                }}</span>
                                            </button>
                                        </div>
                                    </div>

                                    <div
                                        v-else-if="getHeaderType(colIndex).header_usage === 'date'"
                                        class="w-full min-w-0"
                                    >
                                        <button
                                            type="button"
                                            class="flex w-full min-w-0 items-center gap-x-1.5 text-left text-[13px] text-gray-700"
                                            :title="t('projects.grid_view.pick_a_date')"
                                            @click.stop="
                                                toggleDatePickerPopover($event, item, colIndex)
                                            "
                                        >
                                            <CalendarIcon
                                                class="h-4 w-4 flex-none text-gray-400"
                                                aria-hidden="true"
                                            />
                                            <span v-if="item[header.name]" class="truncate">{{
                                                dateCellText(header, item[header.name])
                                            }}</span>
                                        </button>
                                    </div>

                                    <template v-else>
                                        <template v-if="isEditing(colIndex, item)">
                                            <div class="alignTextCenter">
                                                <InlineEditInput
                                                    :value="item[tableHeaders[colIndex].name]"
                                                    :colIndex="colIndex"
                                                    :item="item"
                                                    :headerType="
                                                        getHeaderType(colIndex).header_usage ===
                                                        'calculations'
                                                            ? 'INT'
                                                            : getHeaderType(colIndex).header_type
                                                    "
                                                    :step="
                                                        getHeaderType(colIndex).header_type ===
                                                        'DECIMAL'
                                                            ? 'any'
                                                            : '1'
                                                    "
                                                    @save="saveValue"
                                                    @cancel="cancelEdit"
                                                />
                                            </div>
                                        </template>
                                        <template v-else>
                                            <div
                                                class="w-full min-w-0 text-left leading-6 text-gray-700"
                                            >
                                                <div class="truncate" :title="item[header.name]">
                                                    <span
                                                        v-if="isBlank(item[header.name])"
                                                        class="text-gray-400"
                                                        >–</span
                                                    >
                                                    <template v-else>{{
                                                        item[header.name]
                                                    }}</template>
                                                </div>
                                            </div>
                                        </template>
                                    </template>
                                </div>

                                <div
                                    v-else-if="
                                        ['TEXT', 'VARCHAR', 'INT'].includes(
                                            getHeaderType(colIndex).header_type,
                                        ) && getHeaderType(colIndex).linked_id != ''
                                    "
                                    style="width: 100%"
                                >
                                    <div
                                        v-if="
                                            item[header.name] == null ||
                                            item[header.name] == '' ||
                                            (getHeaderType(colIndex).single_select &&
                                                item[header.name] == 0)
                                        "
                                    >
                                        <button
                                            type="button"
                                            :aria-label="t('common.button.add')"
                                            class="flex items-center"
                                        >
                                            <span
                                                v-if="
                                                    getHeaderType(colIndex).header_usage == 'status'
                                                "
                                                class="ml-1.5 block h-3 w-3 rounded-full border border-dashed border-gray-400"
                                                aria-hidden="true"
                                            />
                                            <span v-else class="text-[13px] text-gray-400">–</span>
                                        </button>
                                    </div>

                                    <div v-else-if="getHeaderType(colIndex).single_select">
                                        <div
                                            v-if="getHeaderType(colIndex).header_usage == 'status'"
                                        >
                                            <button
                                                type="button"
                                                class="inline-flex max-w-full items-center gap-x-1.5 rounded-full bg-white px-2 py-0.5 text-xs font-medium text-gray-700 ring-1 ring-inset ring-gray-200 hover:bg-gray-50"
                                            >
                                                <span
                                                    class="h-2 w-2 flex-none rounded-full"
                                                    :class="getStatusStyle(item.status).class"
                                                    :style="getStatusStyle(item.status).style"
                                                    aria-hidden="true"
                                                />
                                                <span class="truncate">{{
                                                    getStatusDisplayName(item.status)
                                                }}</span>
                                            </button>
                                        </div>

                                        <div
                                            v-else
                                            class="value-container"
                                            style="margin-right: 6px"
                                        >
                                            <button
                                                type="button"
                                                class="inline-flex max-w-full items-center gap-x-1.5 rounded-full bg-white px-2 py-0.5 text-xs font-medium text-gray-700 ring-1 ring-inset ring-gray-200 hover:bg-gray-50"
                                            >
                                                <span
                                                    class="h-2 w-2 flex-none rounded-full"
                                                    :class="getStatusStyle(item[header.name]).class"
                                                    :style="getStatusStyle(item[header.name]).style"
                                                    aria-hidden="true"
                                                />
                                                <span class="truncate">{{
                                                    item[header.name]?.name || " "
                                                }}</span>
                                            </button>
                                        </div>
                                    </div>
                                    <div
                                        v-else
                                        class="value-container"
                                        style="
                                            margin-right: 6px;
                                            display: flex;
                                            flex-wrap: nowrap;
                                            align-items: center;
                                        "
                                    >
                                        <button
                                            type="button"
                                            :aria-label="t('common.button.add')"
                                            class="mr-1.5 flex flex-none items-center"
                                        >
                                            <PlusIcon
                                                class="block h-4 w-4 text-gray-500"
                                                aria-hidden="true"
                                            />
                                        </button>
                                        <LinkedRecordsCell
                                            :records="item[header.name]"
                                            @open="showFullView($event, colIndex, 'subtask')"
                                        />
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>

                    <div class="new-task-row-sticky">
                        <div :class="!isAddingTask ? 'grid-row empty-row' : 'grid-row'">
                            <div
                                class="grid-item-text"
                                v-for="(header, colIndex) in tableHeaders"
                                :key="colIndex"
                                :style="{
                                    width: tableHeaders[colIndex].visible
                                        ? columnWidths[colIndex] + 'px'
                                        : '0px',
                                    display: tableHeaders[colIndex].visible ? 'flex' : 'none',
                                }"
                            >
                                <template v-if="colIndex === 1">
                                    <span v-if="!isAddingTask" class="flex items-center">
                                        <BaseButton
                                            type="button"
                                            size="small"
                                            variant="secondary"
                                            :prepend-icon="PlusIcon"
                                            :aria-label="t('projects.kanban_view.add_task')"
                                            @click="isAddingTask = true"
                                        />
                                    </span>
                                    <span v-else class="flex w-full items-center">
                                        <NewTaskInput
                                            v-if="!props.isSingleSelect"
                                            :initial-value="newTaskDraft"
                                            @submit="(value) => addNewTask(value, header)"
                                            @cancel="cancelTask"
                                        />
                                        <NewTaskInput
                                            v-else
                                            @submit="addNewSection"
                                            @cancel="cancelTask"
                                        />
                                    </span>
                                </template>
                                <template
                                    v-else-if="
                                        getHeaderType(colIndex).header_usage == 'calculations'
                                    "
                                >
                                    <span
                                        class="text-xs text-gray-600 truncate"
                                        style="padding-left: 6px"
                                    >
                                        {{ columnTotal(getHeaderType(colIndex), tableData) }}
                                    </span>
                                </template>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </div>

        <div
            class="sticky-row flex items-center justify-between bg-white py-2 pl-4 pr-4 border-t border-gray-200"
        >
            <span class="inline-flex items-center gap-x-1.5 text-sm text-gray-700 font-medium">
                {{ t("projects.grid_view.total_items") }}
                <BaseSpinner
                    v-if="topLevelTaskCount == null"
                    size="sm"
                    class="text-gray-400"
                    :label="t('common.label.loading')"
                />
                <template v-else>{{ topLevelTaskCount }}</template>
            </span>
            <BasePagination
                v-if="totalPages > 1"
                class="!flex-none !border-0 !bg-transparent !p-0"
                :current-page="currentPage"
                :total-items="totalRows || visibleData.length"
                :items-per-page="itemsPerPage"
                :show-summary="false"
                @update:current-page="(page) => fetchPage(page, { reuseCount: true })"
            />
        </div>

        <FieldEditorPopover
            ref="fieldEditor"
            :table-id="isDialog ? dialogTableID : route.params.tid"
            :table-headers="tableHeaders"
            :can-manage-fields="canManageFields"
            :view-id="selectedView?.id ?? null"
            @created="addFieldColumn"
            @update-header="patchHeader"
        />

        <ConfirmDialog
            :open="showConfirm"
            :title="deleteTitle"
            :message="deleteMessage"
            :confirm-label="t('common.button.delete')"
            @confirm="confirmDelete"
            @close="showConfirm = false"
        />

        <div v-if="editLink">
            <LinkedTableDialog
                :open="linkedTableDialog"
                @close="closeLinkedDialog"
                :linkedID="linkedID"
                :editItemLink="editItemLink"
                :singleSelect="singleSelect"
                :field="field"
                :selectedItemsForLinked="selectedItems"
                :parentTableID="parentTableID"
                :currentLocalTableID="currentLocalTableID"
                :linkedItemsForCheckbox="linkedItemsForCheckbox"
                :isAtFirstLinkedLevel="isAtFirstLinkedLevel"
                :title="linkedTitle"
                @saved="applyLinkedSave"
            />
        </div>

        <div>
            <ShareWorkspaceDialog
                :open="shareWorkspaceDialog"
                @close="shareWorkspaceDialog = false"
                :selectedWorkspace="selectedWorkspace"
                :selectedTask="selectedTask"
                :assignFieldName="assignFieldName"
                :selectedAssigneTable="selectedAssigneTable"
                :isDialog="isDialog"
                @update:assignee="updateAssignee"
                :selectedUserID="selectedUserID"
            />
        </div>

        <RowMenu
            :visible="itemPopoverVisible && parentComp === 'item'"
            :anchor="itemPopoverAnchor"
            :task="activePopoverItem"
            @close="
                itemPopoverVisible = false;
                activePopoverItem = null;
            "
            @task-deleted="onTaskDeleted"
        />
        <FieldMenu
            :visible="itemPopoverVisible && parentComp === 'field'"
            :anchor="itemPopoverAnchor"
            :field="activePopoverItem"
            @close="
                itemPopoverVisible = false;
                activePopoverItem = null;
            "
            @edit-field-panel="handleEditField"
        />

        <DatePickerPopover
            :visible="datePickerPopoverVisible"
            :anchor="datePickerPopoverAnchor"
            :value="activeDateValue"
            v-model:visible="datePickerPopoverVisible"
            @save="handleSave"
        />

        <MasterLinkDialog
            :open="masterLinkDialog"
            @close="masterLinkDialog = false"
            :field="field"
            :masterLinkTask="masterLinkTask"
        />
    </div>
</template>

<script setup>
import { t } from "@/i18n/index.js";
import {
    ref,
    computed,
    onMounted,
    nextTick,
    onBeforeUnmount,
    toRaw,
    reactive,
    watch,
    watchEffect,
} from "vue";
import { PlusIcon, PaperClipIcon, UserPlusIcon, CalendarIcon } from "@heroicons/vue/24/outline";
import LinkedTableDialog from "@/components/Projects/Dialogs/LinkedTableDialog.vue";
import ShareWorkspaceDialog from "@/components/Projects/Members/ShareWorkspaceDialog.vue";
import { useRouter, useRoute } from "vue-router";
import UserAvatarWithText from "@/components/UserAvatarWithText.vue";
import BaseSpinner from "@/components/BaseSpinner.vue";
import BaseButton from "@/components/BaseButton.vue";
import BasePagination from "@/components/BasePagination.vue";
import { useUserStore } from "@/store/user";
import { useWorkspaceRolesStore } from "@/store/workspaceRoles";
import workspaceService from "@/services/workspaceService";
import { useShownFilter } from "@/composables/projects/useShownFilter";
import { useGoToTable } from "@/composables/projects/useGoToTable";
import { useLinkedTableDialog } from "@/composables/projects/useLinkedTableDialog";
import { useDatePickerPopover } from "@/composables/projects/useDatePickerPopover";
import { useItemMenu } from "@/composables/projects/useItemMenu";
import { onReconnect } from "@/js/websocket";
import { useWorkspaceStore } from "@/store/workspaces";
import RowMenu from "@/components/Projects/Menu/RowMenu.vue";
import FieldMenu from "@/components/Projects/Menu/FieldMenu.vue";
import MasterLinkDialog from "@/components/Projects/Dialogs/MasterLinkDialog.vue";
import FieldEditorPopover from "@/components/Projects/Grid/FieldEditorPopover.vue";
import GridHeaderRow from "@/components/Projects/Grid/GridHeaderRow.vue";
import FilesCell from "@/components/Projects/Grid/FilesCell.vue";
import { useTaskFiles } from "@/composables/projects/useTaskFiles";
import {
    calculateRowValueForRow,
    formatCellValue,
    isComputedCalculations,
    columnTotal,
} from "@/utils/projects/formulas";
import { useAlertStore } from "@/store/alerts";
import { extractErrorMessage } from "@/utils/errors";
import { isBlank, newFieldColumn, normalizeTaskRow, withoutTask } from "@/utils/projects/rows";
import {
    parseViewSort,
    pickView,
    tableDisplayNames,
    viewFilterState,
    viewHeaders,
} from "@/utils/projects/gridView";
import { safeHref } from "@/utils/links";
import { withoutFilterOptions } from "@/utils/projects/filterCombine";
import DatePickerPopover from "@/components/Projects/Menu/DatePickerPopover.vue";
import NewTaskInput from "@/components/Projects/Grid/NewTaskInput.vue";
import LinkedRecordsCell from "@/components/Projects/Grid/LinkedRecordsCell.vue";
import StatusTypeMenu from "@/components/Projects/Menu/StatusTypeMenu.vue";
import InlineEditInput from "@/components/Projects/Grid/InlineEditInput.vue";
import useDateOperations from "@/composables/useDateOperations.js";
import { useTableFilters } from "@/composables/projects/useTableFilters";
import { useColumnResize } from "@/composables/projects/useColumnResize";
import { useProgressiveRender } from "@/composables/useProgressiveRender";
import { useLatestRequest } from "@/composables/useLatestRequest";
import { flattenTaskTree } from "@/utils/projects/tree";
import { PRESET_LABELS } from "@/composables/projects/useTaskReportFilters";
import ConfirmDialog from "@/components/ConfirmDialog.vue";
import { getStatusDisplayName, getStatusStyle } from "@/utils/projects/cells";
import { useSubtaskCounts } from "@/composables/projects/useSubtaskCounts";
import {
    ArrowTopRightOnSquareIcon,
    PencilIcon,
    ChevronRightIcon,
    Square2StackIcon,
    ChevronDownIcon,
    PlusCircleIcon,
    TrashIcon,
} from "@heroicons/vue/20/solid";

const showConfirm = ref(false);
const deleteTitle = ref("");
const deleteMessage = ref("");
const deleteID = ref(null);
const openDeleteConfirm = (itemId) => {
    deleteID.value = itemId;
    deleteTitle.value = t.value("projects.grid_view.delete_confirmantion");
    const hasSubtasks = tableData.value.some(
        (row) =>
            String(row.parent_task_id ?? "") === String(itemId) &&
            String(row.id) !== String(itemId),
    );

    deleteMessage.value = hasSubtasks
        ? t.value("projects.popover_menu.delete_message_item_with_subtasks")
        : t.value("projects.grid_view.delete_confirmation_description");
    showConfirm.value = true;
};

const onTaskDeleted = () => {
    if (totalRows.value > 0) totalRows.value--;
    if (rootTotalRows.value > 0) rootTotalRows.value--;
};

const confirmDelete = async () => {
    if (!deleteID.value) {
        showConfirm.value = false;

        return;
    }

    try {
        const isRoot = !props.isDialog;
        let currentTableId;

        if (isRoot) {
            currentTableId = route.params.tid;
        } else {
            currentTableId = dialogTableID.value;
        }

        if (!currentTableId || !route.params.id) {
            throw new Error("Could not determine table or workspace ID for deletion");
        }

        await workspaceService.deleteWorkspaceItem({
            workspace_id: route.params.id,
            table_id: currentTableId,
            item_id: deleteID.value,
        });

        if (isRoot) {
            tableData.value = withoutTask(tableData.value, deleteID.value);
            if (totalRows.value > 0) totalRows.value--;
            if (rootTotalRows.value > 0) rootTotalRows.value--;
        } else if (tableData.value) {
            tableData.value = withoutTask(tableData.value, deleteID.value);
        }
    } catch (error) {
        useAlertStore().showError(
            extractErrorMessage(error, t.value("projects.grid_view.failed_to_delete_item")),
        );
    } finally {
        showConfirm.value = false;
        deleteID.value = null;
    }
};

const triggerFileInput = (item, colIndex) =>
    chooseFiles(`[data-file-id="fileInput-${item.id}-${colIndex}"]`, canEditRow.value);

const handleFileUpload = async (event, item, colIndex) => {
    const files = Array.from(event.target.files);

    if (files.length === 0) return;

    const header = tableHeaders.value[colIndex];
    const headerName = header.name;

    try {
        const uploaded = await uploadTaskFiles(files, {
            tableId: !props.isDialog ? route.params.tid : dialogTableID.value,
            taskId: item.id,
            header,
        });

        if (uploaded) {
            const currentFiles = item[headerName] || [];
            const newFilesArray = [...currentFiles, ...uploaded];

            // Updating the item directly is already reactive if item is from tableData.value
            item[headerName] = newFilesArray;
            bumpRowEdit(item.id);

            if (showFullTask.value && getSelectedItem.value?.id === item.id) {
                const updatedSelectedItem = { ...getSelectedItem.value };

                updatedSelectedItem[headerName] = newFilesArray;
                workspaceStore.setSelectedItem(updatedSelectedItem);
            }

            tableData.value = [...tableData.value];
        }
    } catch (error) {
        console.error("Upload error:", error);
        useAlertStore().showError(
            extractErrorMessage(error, t.value("projects.grid_view.file_upload_failed")),
        );
    }

    event.target.value = "";
};

const { openFile, chooseFiles, uploadTaskFiles } = useTaskFiles();

const expandedTasks = ref(new Set());
const { getDate, getDateAndTime, getTimestampFromDateString, getDateInputValue, isDayField } =
    useDateOperations();

const toggleExpand = (id) => {
    if (expandedTasks.value.has(id)) expandedTasks.value.delete(id);
    else expandedTasks.value.add(id);
};

const visibleData = computed(() => flattenTaskTree(tableData.value, expandedTasks.value));

function handleSave(dateStr) {
    const item = activeDatePickerPopoverItem.value;
    const colIndex = activeDatePickerColIndex.value;

    if (item == null || colIndex == null) {
        datePickerPopoverVisible.value = false;

        return;
    }

    saveValue(colIndex, item, { target: { value: dateStr } });

    datePickerPopoverVisible.value = false;
    activeDatePickerPopoverItem.value = null;
    activeDatePickerColIndex.value = null;
}

function onHeaderDrop(evt) {
    const from = evt.oldIndex;
    const to = evt.newIndex;

    if (to === 0 || to === 1) {
        const movedItem = tableHeaders.value.splice(to, 1)[0];

        tableHeaders.value.splice(from, 0, movedItem);

        return;
    }

    if (typeof from !== "number" || typeof to !== "number" || from === to) return;

    const newWidths = [...columnWidths.value];
    const movedWidth = newWidths.splice(from, 1)[0];

    newWidths.splice(to, 0, movedWidth);
    columnWidths.value = newWidths;

    workspaceService
        .updateGridHeaderOrder({
            workspace_id: route.params.id,
            table_id: route.params.tid,
            headers: tableHeaders.value.map((header) => ({
                id: header.id,
                name: header.name,
            })),
            view_id: selectedView.value?.id || null,
        })
        .catch(() => {
            useAlertStore().showError(
                t.value("projects.grid_view.failed_update_grid_header_order"),
            );

            // The column goes back to where it was.
            const headers = [...tableHeaders.value];

            headers.splice(from, 0, headers.splice(to, 1)[0]);
            tableHeaders.value = headers;
            const widths = [...columnWidths.value];

            widths.splice(from, 0, widths.splice(to, 1)[0]);
            columnWidths.value = widths;
        });
}

const router = useRouter();
const props = defineProps({
    isDialog: Boolean,
    tableHeaders: Array,
    tableData: Array,
    isSingleSelect: Boolean,
    columnWidths: Array,
    totalCount: Number, // total rows from backend (dialog mode)
    dialogWorkspaceId: String, // workspace id for fetchPage in dialog mode
    dialogViewId: String, // view id for fetchPage in dialog mode
    dialogTableId: String, // table id for fetchPage in dialog mode (matches loadData's props.tableId)
    dialogSelectedIds: { type: Array, default: () => [] },
});

const loaded = computed({
    get() {
        return workspaceStore.getIsLoaded;
    },
    set(value) {
        workspaceStore.setIsLoaded(value);
    },
});

onMounted(() => {
    if (route.query.task) {
        handleTaskQuery(route.query.task);
    }
});

const currentLocalTableID = ref(null);
const masterLinkDialog = ref(false);
const {
    visible: datePickerPopoverVisible,
    anchor: datePickerPopoverAnchor,
    item: activeDatePickerPopoverItem,
    colIndex: activeDatePickerColIndex,
    toggle: toggleDatePickerPopover,
} = useDatePickerPopover({
    canOpen: () => canEditRow.value,
    onOpen: (item) => bumpRowEdit(item.id),
});
const selectedUserID = ref(null);

const dialogTableID = computed({
    get() {
        return workspaceStore.getDialogTableID;
    },
    set(value) {
        workspaceStore.setDialogTableID(value);
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

const activeDateValue = computed(() => {
    const header = tableHeaders.value[activeDatePickerColIndex.value];

    return getDateInputValue(activeDatePickerPopoverItem.value?.[header?.name], isDayField(header));
});

// Start and due dates are stored as timestamps, date fields as their day.
const dateCellText = (header, value) => {
    if (!value) return "";

    return (
        (isDayField(header) ? getDate(getTimestampFromDateString(value)) : getDate(+value)) ||
        String(value)
    );
};

const {
    visible: itemPopoverVisible,
    anchor: itemPopoverAnchor,
    item: activePopoverItem,
    context: parentComp,
    toggle: toggleItemMenuPopover,
} = useItemMenu({ canOpen: () => !props.isDialog });

const editableField = reactive({
    colIndex: null,
    item: null,
    fieldName: null, // captured at edit-start so save doesn't depend on tableHeaders index
});

const editableValue = ref("");
const isAddingTask = ref(false);
const newTaskValues = {};
const selectedWorkspace = ref(null);
const selectedItems = ref([]);
const userStore = useUserStore();
const workspaceStore = useWorkspaceStore();
const { loadTablePreset, storeTablePreset } = useTableFilters();
const rolesStore = useWorkspaceRolesStore();
const canEditRow = computed(() =>
    rolesStore.canRowAction(props.isDialog ? dialogTableID.value : route.params.tid, "update_task"),
);
const canEditFields = computed(() => {
    const permissions = rolesStore.userPermissions;

    if (permissions.length === 0) return true;
    const tableId = props.isDialog ? dialogTableID.value : route.params.tid;

    if (rolesStore.perTableMode) return rolesStore.hasTablePermission(tableId, "edit_fields");

    return permissions.includes("update_workspace_table") || permissions.includes("edit_fields");
});
const canManageFields = computed(() => {
    const permissions = rolesStore.userPermissions;

    if (permissions.length === 0) return true;
    const tableId = props.isDialog ? dialogTableID.value : route.params.tid;

    if (rolesStore.perTableMode) return rolesStore.hasTablePermission(tableId, "create_fields");

    return permissions.includes("update_workspace_table") || permissions.includes("create_fields");
});
const tableUniqueId = ref(`table-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`);

const setFullViewHeaders = computed({
    get() {
        return workspaceStore.getFullViewHeaders;
    },
    set(value) {
        workspaceStore.setFullViewHeaders(value);
    },
});

const tableHeaders = computed({
    get() {
        return props.isDialog ? props.tableHeaders : workspaceStore.getTableHeaders;
    },
    set(value) {
        workspaceStore.setTableHeaders(value);
    },
});

function updateTableHeaders(value) {
    if (!props.isDialog) {
        workspaceStore.setTableHeaders(value);
    }
}

// Local column widths for dialog mode. Initialized from prop, updated on resize,
// never saved to backend, discarded when dialog closes.
const localDialogWidths = ref([]);

watch(
    () => props.columnWidths,
    (val) => {
        if (props.isDialog && val?.length) localDialogWidths.value = [...val];
    },
    { immediate: true },
);

const columnWidths = computed({
    get() {
        return props.isDialog ? localDialogWidths.value : workspaceStore.getColumnsWidths;
    },
    set(value) {
        if (props.isDialog) {
            localDialogWidths.value = value;
        } else {
            workspaceStore.setColumnsWidths(value);
        }
    },
});

const selectedView = computed({
    get() {
        return workspaceStore.getSelectedViewData;
    },
    set(value) {
        workspaceStore.setSelectedViewData(value);
    },
});

const rowHeaders = computed(() => tableHeaders.value.filter((h) => h.name !== "itemSelected"));

const route = useRoute();

function handleTaskQuery(taskId) {
    if (!taskId) return;

    const workspaceId = route.params.id;
    const tableId = route.params.tid;

    // Clear query immediately, pass full route params so the optional fid param
    // is preserved and doesn't trigger a data reload via watch(route.params).
    router.replace({ name: route.name, params: { ...route.params }, query: {} });

    // Delay watcher setup by one tick so that watch(route.params) has already
    // run loadData() + resetState(), flushing any stale tableData from a
    // previous table before we start watching.
    nextTick(() => {
        // Expand ALL ancestors of the task so it becomes visible in visibleData.
        // Do NOT stop on miss. doPageNav may load a different page later.
        const stopTableWatch = watch(
            tableData,
            (data) => {
                if (data.length === 0) return;
                const task = data.find((item) => item.id == taskId);

                if (!task) return;
                // Walk up the ancestor chain and expand every level.
                const idMap = new Map(data.map((r) => [String(r.id), r]));
                let cur = task;

                while (cur && cur.parent_task_id && cur.parent_task_id !== cur.id) {
                    expandedTasks.value.add(String(cur.parent_task_id));
                    cur = idMap.get(String(cur.parent_task_id));
                }

                stopTableWatch();
            },
            { immediate: true },
        );

        // Highlight once the task actually appears in visibleData. Watching
        // visibleData (not tableData) means a second loadData triggered by the
        // filter-deactivation watcher is handled naturally, we just wait for
        // the task to appear after whichever load cycle completes last.
        const stopVisibleWatch = watch(
            visibleData,
            (data) => {
                const task = data.find((item) => item.id == taskId);

                if (!task) return;
                stopVisibleWatch();
                stopTableWatch();
                findAndHighlightTask(taskId);
            },
            { immediate: true },
        );

        // Find which page the task is on and navigate there if needed.
        // Must wait for the filter to be applied first (awaitingFilterLoad cycle)
        // so that getTaskPage uses the same filter the grid is showing.
        let filterLoadStarted = false;
        let pageDone = false;
        const doPageNav = () => {
            if (pageDone) return;
            pageDone = true;
            const alreadyHere = tableData.value.find((item) => item.id == taskId);

            if (alreadyHere) return;
            const sortOpts = toRaw(workspaceStore.getSortOptions) || [];
            const flatFilters = toRaw(workspaceStore.getSavedFlatFilters) || [];
            const groups = toRaw(workspaceStore.getSavedGroups) || [];

            workspaceService
                .getTaskPage({
                    workspace_id: workspaceId,
                    table_id: tableId,
                    task_id: taskId,
                    limit: 100,
                    sort: sortOpts,
                    flat_filters: flatFilters,
                    groups,
                })
                .then((res) => {
                    const targetPage = res.data?.page || 1;

                    fetchPage(targetPage);
                })
                .catch(() => {
                    fetchPage(1);
                });
        };

        // Case A: full navigation. awaitingFilterLoad cycles true→false after fetchPage(1) with filter.
        // Case B: same-route (loadData not called). awaitingFilterLoad stays false, isLoading is false.
        // Use let so each watcher closure can call the other's stop handle.
        let stopFilterCheck, stopLoadingCheck;

        stopFilterCheck = watch(
            awaitingFilterLoad,
            (awaiting) => {
                if (awaiting) {
                    filterLoadStarted = true;

                    return;
                }

                if (!filterLoadStarted) return;
                stopFilterCheck();
                stopLoadingCheck?.();
                doPageNav();
            },
            { immediate: true },
        );

        stopLoadingCheck = watch(
            isLoading,
            (loading) => {
                if (loading) return;
                if (awaitingFilterLoad.value || filterLoadStarted) return;
                stopLoadingCheck();
                stopFilterCheck?.();
                doPageNav();
            },
            { immediate: true },
        );
    });

    // DTR and AssignedToMe pre-set the store before navigating, skip the call
    if (showFullTask.value && selectedItem.value?.id == taskId) {
        return;
    }

    workspaceService
        .getItemForTableByID({
            workspace_id: workspaceId,
            table_id: tableId,
            task_id: taskId,
        })
        .then((response) => {
            selectedItem.value = response.data.item;
            setFullViewHeaders.value = response.data.headers;
            showFullTask.value = true;
        })
        .catch((error) => {
            useAlertStore().showError(extractErrorMessage(error));
            console.error("failed to load task:", error);
        });
}

watch(
    () => route.query.task,
    (taskId) => handleTaskQuery(taskId),
);

const itemsPerPage = 100;
const RENDER_BATCH = 25;
const currentPage = ref(1);
const totalRows = ref(null);

const lastTaskRef = ref(null);

function setLastTaskRef(el) {
    lastTaskRef.value = el;
}

function findAndHighlightTask(taskId) {
    const index = visibleData.value.findIndex((item) => item.id == taskId);

    if (index === -1) return;

    // With server-side pagination tableData holds exactly the current page's rows.
    // Don't recalculate page from index. currentPage is already set correctly by fetchPage.

    const doHighlight = (remaining) => {
        nextTick().then(() => {
            const el = document.querySelector(`.grid-row[data-id="${taskId}"]`);

            if (!el) {
                // RouterView may not be mounted yet (e.g. wsOpen guard in ProjectsView).
                // Retry until the element appears or we give up.
                if (remaining > 0) setTimeout(() => doHighlight(remaining - 1), 150);

                return;
            }

            el.scrollIntoView({ behavior: "smooth", block: "center" });
            el.querySelectorAll(".grid-item-text").forEach((cell) => {
                cell.classList.add("task-highlight");
            });
            setTimeout(() => {
                el.querySelectorAll(".grid-item-text").forEach((cell) => {
                    cell.classList.remove("task-highlight");
                });
            }, 3000);
        });
    };

    doHighlight(20); // retry up to 20×150ms = 3 seconds
}

const users = ref([]);

const shareWorkspaceDialog = ref(false);

const totalPages = computed(() =>
    Math.max(1, Math.ceil((totalRows.value || visibleData.value.length) / itemsPerPage)),
);

const {
    open: linkedTableDialog,
    shown: editLink,
    field,
    parentTableID,
    title: linkedTitle,
    singleSelect,
    linkedID,
    editItemLink,
    linkedItemsForCheckbox,
    isAtFirstLinkedLevel,
    openFor: openLinkedTableDialog,
    close: closeLinkedDialog,
} = useLinkedTableDialog({
    props,
    tableHeaders,
    canEditRow,
    onLinkedIds: (ids) => emit("selected", ids),
});

const workspaces = ref([]);

const singleField = computed({
    get() {
        return workspaceStore.getFieldName;
    },
    set(value) {
        workspaceStore.setFieldName(value);
    },
});

const tableData = computed({
    get() {
        return props.isDialog ? props.tableData : workspaceStore.getTableData;
    },
    set(value) {
        if (props.isDialog) {
            emit("update:tableData", value);
        } else {
            workspaceStore.setTableData(value);
        }
    },
});
const { hasChildren, getSubtaskCount } = useSubtaskCounts(tableData);

const updateAssignee = (assignee) => {
    emit("update:assigneeDialog", assignee);
    // Update the row in tableData so the grid reflects the new assignee immediately
    if (assignee?.task_id && assignee?.field_name !== undefined) {
        const row = tableData.value.find((r) => r.id === assignee.task_id);

        if (row) {
            row[assignee.field_name] = assignee.user_id || "";
            row.updated_at = Date.now();
            bumpRowEdit(assignee.task_id);
        }
    }
};

// The option or link dialog writes the store's rows, which in a dialog are not
// this grid's, so the grid puts the saved value in its own row too.
const applyLinkedSave = ({ taskId, field, value }) => {
    if (!field) return;

    tableData.value = tableData.value.map((row) =>
        row.id === taskId ? { ...row, [field]: value, updated_at: Date.now() } : row,
    );
    bumpRowEdit(taskId);
};

// _rowVersion increments on every full data reload so v-memo invalidates all rows on page flip.
const _rowVersion = ref(0);
// _rowEditVersions tracks per-row edits so v-memo re-renders only the edited row after a save.
const _rowEditVersions = reactive({});
const bumpRowEdit = (id) => {
    _rowEditVersions[id] = (_rowEditVersions[id] || 0) + 1;
};

// Invalidate v-memo when columns are shown, hidden or moved, or put back after
// a refused move, so rows re-render with the columns as the headers have them.
watch(
    () => tableHeaders.value.map((h) => `${h.name}:${h.visible}`).join(),
    () => {
        _rowVersion.value++;
    },
);
// A calculation's cells come from its header's formula, not from the row.
watch(
    () => JSON.stringify(tableHeaders.value.map((h) => h.formula ?? null)),
    () => {
        _rowVersion.value++;
    },
);

// Progressive rendering of a newly loaded page: the first rows at once, the
// rest a batch per frame. Changes to a page already shown do not restart it;
// v-memo keys each row on its object, so only a replaced row renders again.
const {
    renderedCount,
    shown: paginatedData,
    start: startBatches,
} = useProgressiveRender(visibleData, RENDER_BATCH);

function startRender() {
    _rowVersion.value++;
    startBatches();
}

const selectedItem = computed({
    get() {
        return workspaceStore.getSelectedItem;
    },
    set(value) {
        workspaceStore.setSelectedItem(value);
    },
});

const rootTotalRows = ref(null);

const topLevelTaskCount = computed(() => rootTotalRows.value || totalRows.value);

let skipNextGlobalClick = false;

const addSubtask = (parentItem, name = "") => {
    skipNextGlobalClick = true;

    const parentIndex = tableData.value.findIndex((t) => t.id === parentItem.id);

    if (parentIndex === -1) return;

    const newId = `temp-${Date.now()}`;
    const nameHeader = tableHeaders.value.find((h) => h.name === "name");

    const newSubtask = {
        id: newId,
        parent_task_id: parentItem.id,
        _indent: (parentItem._indent || 0) + 1,
        isNew: true,
    };

    tableHeaders.value.forEach((header) => {
        if (header.name !== "id") {
            newSubtask[header.name] = "";
        }
    });

    if (nameHeader) {
        newSubtask[nameHeader.name] = name;
    }

    // Use assignment (not splice) so the computed setter fires in dialog mode
    const updated = [...tableData.value];

    updated.splice(parentIndex + 1, 0, newSubtask);
    tableData.value = updated;
    expandedTasks.value.add(parentItem.id);

    // Two ticks: first flush processes tableData emit + dialog watcher (resets expandedTasks),
    // second flush processes the expandedTasks change so visibleData is fully settled.
    nextTick(async () => {
        const nameColIndex = tableHeaders.value.findIndex((h) => h.name === "name");

        if (nameColIndex === -1) return;
        editableField.colIndex = nameColIndex;
        editableField.item = newSubtask;
        editableField.fieldName = tableHeaders.value[nameColIndex]?.name || null;

        await nextTick(); // wait for expandedTasks watcher to settle

        const tempIndex = visibleData.value.findIndex((r) => r.id === newId);

        if (tempIndex !== -1 && renderedCount.value <= tempIndex) {
            renderedCount.value = tempIndex + 1;
        }

        bumpRowEdit(newId);
    });
};

const PICKED_FIELDS = [
    "assignee",
    "default_assignee",
    "file",
    "master link",
    "status",
    "url",
    "date",
    "default_date",
    "link",
];
const SERVER_KEPT_COLUMNS = [
    "id",
    "created_at",
    "updated_at",
    "deleted_at",
    "created_by",
    "parent_task_id",
];

// Whether a click edits a cell in place. Fields with a picker of their own
// open it instead, and in the grid a task's name opens the task; its name
// is edited by a double click or the pencil beside it. In a dialog the name
// is the value.
const isTypedInPlace = (header) => {
    if (!header || header.single_select || (header.linked_id && header.linked_id !== ""))
        return false;
    if (SERVER_KEPT_COLUMNS.includes(header.name) || header.name === "status_type") return false;

    const usage = (header.header_usage || "").toLowerCase();

    if (PICKED_FIELDS.includes(usage)) return false;
    if (usage === "calculations") return !isComputedCalculations(header);
    if (header.name === "name" && !props.isDialog) return false;

    return ["TEXT", "VARCHAR", "DECIMAL", "INT", "BIGINT", "NUMBER"].includes(
        (header.header_type || "").toUpperCase(),
    );
};

// A double click renames the task, so its first click waits to see whether
// a second follows before opening the task.
let nameClickTimer = 0;
const openTaskOnClick = (item, colIndex) => {
    clearTimeout(nameClickTimer);
    nameClickTimer = setTimeout(() => showFullView(item, colIndex, "task"), 250);
};

const renameTask = (colIndex, item) => {
    clearTimeout(nameClickTimer);
    startCellEdit(colIndex, item, { name: true });
};

onBeforeUnmount(() => clearTimeout(nameClickTimer));

const startCellEdit = (colIndex, item, { name = false } = {}) => {
    if (!canEditRow.value || isEditing(colIndex, item)) return;

    const header = tableHeaders.value[colIndex];

    // An option or linked records cell opens its picker from anywhere in it,
    // its empty space as well as its chip or add button.
    const usage = (header?.header_usage || "").toLowerCase();

    if (
        !name &&
        header?.linked_id &&
        !["assignee", "default_assignee", "master link", "file", "url"].includes(usage) &&
        ["TEXT", "VARCHAR", "INT"].includes((header.header_type || "").toUpperCase())
    ) {
        openLinkedTableDialog(item, colIndex);

        return;
    }

    // A link to a table, a website or a file opens from anywhere in its cell
    // too. What the cell shows, a link or a file, stops its own clicks.
    if (!name && usage === "master link") {
        openMasterLinkTableDialog(item, colIndex);

        return;
    }

    if (!name && usage === "url") {
        editableField.colIndex = colIndex;
        editableField.item = item;

        return;
    }

    if (!name && usage === "file") {
        triggerFileInput(item, colIndex);

        return;
    }

    if (!name && !isTypedInPlace(header)) return;

    const PROTECTED_STATUSES = ["Completed", "Cancelled"];

    if (
        props.isDialog &&
        header?.name === "name" &&
        PROTECTED_STATUSES.includes(item._original_name || item.name)
    ) {
        return;
    }

    editableField.colIndex = colIndex;
    editableField.item = item;
    editableField.fieldName = header?.name || null;
    editableValue.value = item[header?.name];
    bumpRowEdit(item.id);
};

const toggleCheckbox = async (item, colIndex, event) => {
    if (!canEditRow.value) {
        event.preventDefault();

        return;
    }

    const checkbox = event.target;
    const fieldName = tableHeaders.value[colIndex].name;
    const newValue = checkbox.checked ? "1" : "0";

    if (item[fieldName] == null) {
        item[fieldName] = "0";
        checkbox.value = "0";
    }

    checkbox.value = newValue;

    // The box ticks itself before the save, and the row it is drawn from
    // has not changed, so a refused save unticks it here.
    if ((await saveValue(colIndex, item, event)) === false) {
        checkbox.checked = !checkbox.checked;
    }
};

const lastClickedElement = ref(null);

const emit = defineEmits(["selected", "update:assigneeDialog", "update:tableData"]);

const isTouchedCheckbox = computed({
    get() {
        return workspaceStore.getIsTouchedCheckbox;
    },
    set(value) {
        workspaceStore.setIsTouchedCheckbox(value);
    },
});

const updateSelectedItems = (item) => {
    const list = tableData.value.filter((i) => i.itemSelected);

    emit("selected", list, item);

    isTouchedCheckbox.value = true;
};

const handleCheckboxClick = (event, item) => {
    event.stopPropagation();
    lastClickedElement.value = "checkbox";

    if (props.isSingleSelect) {
        const selected = !item.itemSelected;

        tableData.value.forEach((i) => (i.itemSelected = false));
        item.itemSelected = selected;
    } else {
        item.itemSelected = !item.itemSelected;
    }

    updateSelectedItems(item);
};

const isEditing = (colIndex, item) => {
    if (lastClickedElement.value === "checkbox") {
        return false;
    }

    if (isProtectedStatus(colIndex, item)) {
        return false;
    }

    const editingState = editableField.colIndex === colIndex && editableField.item?.id === item?.id;

    return editingState;
};

const getSelectedItem = computed({
    get() {
        return workspaceStore.getSelectedItem;
    },
    set(value) {
        workspaceStore.setSelectedItem(value);
    },
});

const setTableID = computed({
    get() {
        return workspaceStore.getTableID;
    },
    set(value) {
        workspaceStore.setTableID(value);
    },
});

const isProcessingSave = ref(false);

const saveValue = async (colIndex, item, event) => {
    if (!canEditRow.value) return;

    // Guard: if editableField is set but to a DIFFERENT item, this is a stale
    // re-mounted InlineEditInput. Close it without saving to avoid writing wrong data.
    // Allow saves when editableField.item is null (date picker, assignee, etc.).
    if (editableField.item && editableField.item.id !== item.id) {
        resetEditState();

        return;
    }

    // fieldName captured at edit-start; tableHeaders may have changed if loadData ran.
    const fieldName = editableField.fieldName || tableHeaders.value[colIndex]?.name;

    if (!fieldName) {
        resetEditState();

        return;
    }

    if (item.isNew && fieldName === "name") {
        handleNewTaskSave(colIndex, item, event);

        return;
    }

    const colHeader = tableHeaders.value[colIndex];
    const fieldID = colHeader?.id;

    let newValue = (event?.target?.value || "").trim();

    let tableID = !props.isDialog ? route.params.tid : dialogTableID.value;

    if (fieldName === "start_date" || fieldName === "due_date") {
        newValue = String(getTimestampFromDateString(newValue));
    }

    try {
        const response = await workspaceService.updateTask({
            workspace_id: route.params.id,
            table_id: tableID,
            task_id: item.id,
            field: fieldName,
            value: newValue,
            field_id: fieldID,
        });

        const appliedValue = Object.hasOwn(response.data, fieldName)
            ? response.data[fieldName]
            : newValue;

        const isDateField =
            ["start_date", "due_date"].includes(fieldName) ||
            colHeader?.header_usage === "default_date";

        item[fieldName] = isDateField && appliedValue ? +appliedValue : appliedValue;
        item.updated_at = Date.now();
        bumpRowEdit(item.id);

        resetEditState();

        // If this is an assignee field, ensure the user is in cache for immediate display
        const hdr = tableHeaders.value[colIndex];

        if (
            appliedValue &&
            (hdr?.header_usage === "default_assignee" || hdr?.header_usage === "assignee")
        ) {
            userStore.ensureUsers([appliedValue]);
        }

        const selected = workspaceStore.getSelectedItem;

        if (selected && selected.id === item.id) {
            workspaceStore.setSelectedItem({
                ...selected,
                [fieldName]: appliedValue,
                updated_at: Date.now(),
            });
        }

        if (workspaceStore.filterActive && !props.isDialog) {
            const shouldBeVisible = await checkIfTaskShouldBeVisible(item.id);

            if (!shouldBeVisible) {
                const taskIndex = tableData.value.findIndex((t) => t.id === item.id);

                if (taskIndex !== -1) {
                    tableData.value.splice(taskIndex, 1);
                }
            }
        }

        return true;
    } catch (error) {
        useAlertStore().showError(
            extractErrorMessage(error, t.value("projects.grid_view.error_updating")),
        );

        return false;
    }
};

const handleNewTaskSave = (colIndex, item, event) => {
    let newValue = (event?.target?.value || "").trim();

    if (newValue === "") {
        const index = tableData.value.findIndex((t) => t.id === item.id);

        if (index !== -1) tableData.value.splice(index, 1);
        resetEditState();

        return;
    }

    if (newValue !== "") {
        const parentId = item.parent_task_id || "";

        if (isProcessingSave.value) {
            return;
        }

        isProcessingSave.value = true;

        addNewTask(newValue, tableHeaders.value[colIndex], parentId);

        const index = tableData.value.findIndex((t) => t.id === item.id);

        if (index !== -1) tableData.value.splice(index, 1);

        resetEditState();

        setTimeout(() => {
            isProcessingSave.value = false;
        }, 500);
    }
};

const cancelEdit = () => {
    resetEditState();
};

const resetEditState = () => {
    const prevItem = editableField.item;

    editableField.colIndex = null;
    editableField.item = null;
    editableField.fieldName = null;
    editableValue.value = "";
    if (prevItem?.id) bumpRowEdit(prevItem.id);
};

const isProtectedStatus = (colIndex, item) => {
    const isInDialog = props.isDialog;

    if (!isInDialog) return false;

    const PROTECTED_STATUSES = ["Completed", "Cancelled"];

    const itemName = item._original_name || item.name;

    if (PROTECTED_STATUSES.includes(itemName)) {
        return true;
    }

    return false;
};

const headerTypesCache = ref(null);

watchEffect(() => {
    headerTypesCache.value = tableHeaders.value.map((header) => ({
        header_name: header.name,
        header_type: header.header_type,
        linked_id: header.linked_id,
        single_select: header.single_select,
        header_usage: header.header_usage,
        header_parent_table_id: header.parent_table_id,
        formula: header.formula,
    }));
});

const getHeaderType = (colIndex) => {
    return headerTypesCache.value?.[colIndex] || {};
};

const cancelTask = () => {
    isAddingTask.value = false;
    newTaskDraft.value = "";
};

const selectedTask = ref(null);
const selectedAssigneTable = ref(null);
const masterLinkTask = ref(null);
const assignFieldName = ref("");

const openMasterLinkTableDialog = (item, colIndex) => {
    const fieldName = tableHeaders.value[colIndex].name;

    field.value = fieldName;

    masterLinkTask.value = item;

    masterLinkDialog.value = true;
};

const openAssignDialog = (item, colIndex) => {
    if (!canEditRow.value) {
        useAlertStore().showError(t.value("projects.grid_view.no_permission_edit_fields"));

        return;
    }

    let tableID = null;

    if (!props.isDialog) {
        tableID = route.params.tid;
    } else {
        tableID = dialogTableID.value;
    }

    const workspace = workspaceDetails.value;

    if (!workspace) {
        return;
    }

    selectedWorkspace.value = workspace;

    selectedAssigneTable.value = workspace.tables.find((table) => table.id === tableID);

    assignFieldName.value = tableHeaders.value[colIndex].name;

    selectedTask.value = item;

    selectedUserID.value = selectedTask.value[assignFieldName.value] || null;

    shareWorkspaceDialog.value = true;
};

// The grid saves a column's width for its view; a dialog only shows it.
// Resolves to false when the width was not saved.
function saveColumnWidth(columnName, newWidth) {
    if (props.isDialog) return;

    return workspaceService
        .updateColumnWidth({
            workspace_id: route.params.id,
            table_id: route.params.tid,
            view_id: selectedView.value?.id || route.params.fid,
            column_name: columnName,
            width: newWidth,
        })
        .then(() => true)
        .catch((error) => {
            useAlertStore().showError(
                extractErrorMessage(
                    error,
                    t.value("projects.grid_view.failed_to_save_column_width"),
                ),
            );

            return false;
        });
}

const { startResize } = useColumnResize({
    headers: tableHeaders,
    widths: columnWidths,
    tableId: tableUniqueId,
    onResized: saveColumnWidth,
});

// The latest page asked for wins: a page answered after a newer request, or
// after the view changed, is dropped instead of shown. The rows land in the
// store the next page shares, so one still in flight when the grid goes away
// must not write them.
const pageLoad = useLatestRequest();

// The same for the view a load sets up: a load overtaken by another one
// stops at its next step, and never writes the filter or asks for rows.
const viewLoad = useLatestRequest();

// The page itself reloads with every other view. The tables, which hold the
// options cells show, may have changed while the socket was down too.
const stopReconnect = onReconnect(async () => {
    if (props.isDialog || !route.params.id) return;

    try {
        const res = await workspaceService.getWorkspaceTables(route.params.id, "nav");

        workspaceTablesForLinked.value = res?.data || [];
        workspaceTables.value = res?.data || [];
    } catch {
        // The page's own reload reports a server that is still unreachable.
    }
});

onBeforeUnmount(stopReconnect);

// The total only fills in the pager. Counting reads the whole table, so it is
// asked for beside the page and the rows are shown without waiting for it.
const countLoad = useLatestRequest();

async function loadCount(request) {
    const req = countLoad.start();

    totalRows.value = null;
    rootTotalRows.value = null;
    try {
        const { data } = await workspaceService.getFilteredTableData(
            { ...request, count_only: true },
            req.signal,
        );

        if (!req.isCurrent()) return;
        totalRows.value = data?.total ?? 0;
        rootTotalRows.value = data?.root_total ?? 0;
    } catch {
        // The page's own load reports a server that cannot be reached.
    }
}

// Moving to another page keeps the total, since only a new view, filter or
// reload can change it.
const fetchPage = async (page, { reuseCount = false } = {}) => {
    const req = pageLoad.start();
    const { signal } = req;

    isLoading.value = true;
    try {
        const groups = props.isDialog ? [] : toRaw(defaultGroupFilter.value) || [];
        const flatFilters = props.isDialog ? [] : toRaw(defaultFlatFilter.value) || [];
        const workspaceId = props.isDialog
            ? props.dialogWorkspaceId || route.params.id
            : route.params.id;
        const tableId = props.isDialog
            ? props.dialogTableId || dialogTableID.value || route.params.tid
            : route.params.tid;
        const viewId = props.isDialog ? props.dialogViewId || "" : route.params.fid;
        const request = {
            workspace_id: workspaceId,
            table_id: tableId,
            view_id: viewId,
            filters: {
                ...withoutFilterOptions({ groups, flatFilters }),
                page,
                limit: itemsPerPage,
                sort: toRaw(workspaceStore.getSortOptions) || [],
                timezone: userStore.getTimezone,
            },
        };

        if (!reuseCount || totalRows.value == null) loadCount(request);
        const resp = await workspaceService.getFilteredTableData(
            { ...request, skip_count: true },
            signal,
        );

        if (!req.isCurrent()) return;
        const base = resp?.data?.data_base || [];

        currentPage.value = page;
        if (!props.isDialog) workspaceStore.gridPage = page;
        // In dialog mode use prop headers; otherwise find the table from workspaces
        const effectiveHeaders = props.isDialog
            ? props.tableHeaders || []
            : workspaces.value.find((t) => t.id === route.params.tid)?.headers || [];

        if (effectiveHeaders.length > 0 && base.length > 0) {
            tableData.value = base.map((item) => normalizeTaskRow(item, effectiveHeaders));
            if (props.isDialog) {
                const selected = new Set(props.dialogSelectedIds);

                tableData.value.forEach((row) => (row.itemSelected = selected.has(String(row.id))));
            }

            ensurePageUsers(tableData.value, effectiveHeaders);
        } else if (!props.isDialog) {
            tableData.value = [];
        }

        startRender();
        if (awaitingFilterLoad.value) {
            awaitingFilterLoad.value = false;
            loaded.value = true;
        }
    } catch (err) {
        if (!req.isCurrent()) return;
        useAlertStore().showError(err?.response?.data || t.value("projects.grid_view.load_failed"));
        if (awaitingFilterLoad.value) {
            awaitingFilterLoad.value = false;
            loaded.value = true;
        }
    } finally {
        if (req.isCurrent()) isLoading.value = false;
    }
};

// Loads the users a page names, so its user cells show names and photos.
function ensurePageUsers(rows, headers) {
    const userFields = headers
        .filter(
            (h) =>
                ["assignee", "default_assignee"].includes(h.header_usage) ||
                h.name === "created_by",
        )
        .map((h) => h.name);
    const ids = [
        ...new Set(
            rows
                .flatMap((row) => userFields.map((f) => row[f]))
                .filter((id) => typeof id === "string" && id),
        ),
    ];

    if (ids.length) userStore.ensureUsers(ids);
}

const workspaceDetails = computed({
    get() {
        return workspaceStore.getWorkspaceDetails;
    },
    set(value) {
        workspaceStore.setWorkspaceDetails(value);
    },
});

const { goToTable } = useGoToTable();

function showFullView(item, colIndex, type) {
    if (type === "subtask") {
        var test = getHeaderType(colIndex).header_parent_table_id;

        setTableID.value = test;

        workspaceService
            .getItemForTableByID({
                workspace_id: route.params.id,
                table_id: test,
                task_id: item.id,
            })
            .then((response) => {
                selectedItem.value = response.data.item;
                setFullViewHeaders.value = response.data.headers;
                showFullTask.value = true;
            })
            .catch((error) => {
                useAlertStore().showError(extractErrorMessage(error));
            });
    } else {
        setFullViewHeaders.value = tableHeaders.value;
        setTableID.value = route.params.tid;
        selectedItem.value = item;
        showFullTask.value = true;
    }
}

function handleGlobalClick() {
    if (skipNextGlobalClick) {
        skipNextGlobalClick = false;

        return;
    }

    // Skip entirely if no temp rows exist (avoids full grid re-render on every click)
    const hasTempRows = tableData.value.some(
        (item) => item.isNew && item.id?.toString().startsWith("temp"),
    );

    if (!hasTempRows) return;

    setTimeout(() => {
        const active = document.activeElement;
        const filtered = tableData.value.filter((item) => {
            const isTemp = item.isNew && item.id?.toString().startsWith("temp");
            const isEmpty = (item.name || "").trim() === "";

            if (isTemp && isEmpty) {
                if (active && active.tagName === "INPUT" && active.value.trim() === "") {
                    return true;
                }

                return false;
            }

            return true;
        });

        if (filtered.length < tableData.value.length) {
            tableData.value = filtered;
        }
    }, 0);
}

onMounted(() => {
    document.addEventListener("click", handleGlobalClick);
});

onBeforeUnmount(() => {
    document.removeEventListener("click", handleGlobalClick);
});

const addNewSection = function (item) {
    let tableID;

    if (!props.isDialog) {
        tableID = route.params.tid;
    } else {
        tableID = setTableID.value;
    }

    workspaceService
        .addFieldValue({
            workspace_id: route.params.id,
            table_id: tableID,
            name: item,
            field: singleField.value,
        })
        .then((response) => {
            isAddingTask.value = false;

            props.tableData.push({
                id: response.data.id,
                name: response.data.name,
                itemSelected: false,
            });

            item = "";
        })
        .catch((error) => {
            useAlertStore().showError(
                extractErrorMessage(error, t.value("projects.grid_view.failed_to_add_new_item")),
            );
        });
};

const addNewTask = async function (itemName, header, parentTaskId = "") {
    const tableID = !props.isDialog ? route.params.tid : props.dialogTableId || dialogTableID.value;

    try {
        const response = await workspaceService.createNewTask({
            workspace_id: route.params.id,
            table_id: tableID,
            name: itemName,
            parent_task_id: parentTaskId,
        });

        const newData = { id: response.data.id };

        tableHeaders.value.forEach((header) => {
            if (header.name !== "id") {
                newData[header.name] = response.data[header.name];
            }
        });

        if (parentTaskId) {
            newData.parent_task_id = parentTaskId;
            newData._indent = getIndentByParentId(parentTaskId);
            expandedTasks.value.add(parentTaskId);

            // Single assignment so the dialog watcher fires once with the new task already present,
            // ensuring expandAllParentsForDialog sees parentTaskId as a parent and keeps it expanded.
            const filtered = tableData.value.filter(
                (t) => !(t.isNew && t.parent_task_id === parentTaskId),
            );

            tableData.value = [...filtered, newData];
        } else {
            newData._indent = 0;
            // Add to top of current page so the new task is immediately visible
            tableData.value = [newData, ...tableData.value];
            if (totalRows.value != null) totalRows.value += 1;
            if (rootTotalRows.value != null) rootTotalRows.value += 1;
        }

        bumpRowEdit(newData.id);

        newData.isNew = false;
        resetEditState();

        if (workspaceStore.filterActive && !props.isDialog) {
            const shouldBeVisible = await checkIfTaskShouldBeVisible(newData.id);

            if (!shouldBeVisible) {
                const taskIndex = tableData.value.findIndex((t) => t.id === newData.id);

                if (taskIndex !== -1) {
                    tableData.value.splice(taskIndex, 1);
                }
            }
        }

        // Always show full view for new tasks, even if filtered out
        showFullView(newData, 0, "task");

        isAddingTask.value = false;
        newTaskValues["name"] = "";
        newTaskDraft.value = "";
    } catch (error) {
        useAlertStore().showError(extractErrorMessage(error));
        isProcessingSave.value = false;

        // What was typed stays, so the task can be added again.
        const parent = parentTaskId && tableData.value.find((t) => t.id === parentTaskId);

        if (parent) {
            addSubtask(parent, itemName);
        } else {
            newTaskDraft.value = itemName;
            isAddingTask.value = true;
        }
    }
};

const newTaskDraft = ref("");

const { matchesShownFilter } = useShownFilter();
const checkIfTaskShouldBeVisible = (taskId) =>
    matchesShownFilter(route.params.id, route.params.tid, taskId);

const getCellDisplayValue = (header, row) => {
    const usage = (header?.header_usage || "").toLowerCase();

    if (usage === "calculations") {
        return formatCellValue(calculateRowValueForRow(header, row), header);
    }

    const key = headerKey(header);

    return formatCellValue(row?.[key], header);
};

const showParentTask = (parentTaskId, colIndex, item) => {
    let targetParentId = parentTaskId;

    if (!targetParentId && item && item._original_parent_task_id) {
        targetParentId = item._original_parent_task_id;
    }

    if (!targetParentId && item && item.parent_task_id) {
        targetParentId = item.parent_task_id;
    }

    if (!targetParentId) {
        return;
    }

    const parentTask = tableData.value.find((task) => task.id === targetParentId);

    if (parentTask) {
        showFullView(parentTask, colIndex, "task");
    } else {
        const headerConfig = tableHeaders.value[colIndex];
        const targetTableId = headerConfig?.linked_table_id || route.params.tid;

        workspaceService
            .getItemForTableByID({
                workspace_id: route.params.id,
                table_id: targetTableId,
                task_id: targetParentId,
            })
            .then((response) => {
                selectedItem.value = response.data.item;
                setFullViewHeaders.value = response.data.headers;
                setTableID.value = targetTableId;
                showFullTask.value = true;
            })
            .catch((error) => {
                useAlertStore().showError(extractErrorMessage(error));
            });
    }
};

const headerKey = (h) => h?.header_name || h?.name || "";

const fieldEditor = ref(null);

// The field editor reports its changes; the grid keeps its own headers,
// column widths and rows.
function addFieldColumn(newFieldData) {
    columnWidths.value.push(150);

    tableHeaders.value = [...tableHeaders.value, newFieldData];

    if (setFullViewHeaders.value?.length > 0) {
        setFullViewHeaders.value = [...setFullViewHeaders.value, newFieldData];
    }

    const column = newFieldColumn(newFieldData.header_type);

    newFieldData.header_type = column.headerType;

    tableData.value = tableData.value.map((item) => ({
        ...item,
        [newFieldData.name]: column.empty(),
    }));
}

function patchHeader(name, patch) {
    const idx = tableHeaders.value.findIndex((h) => h.name === name);

    if (idx !== -1) {
        tableHeaders.value[idx] = { ...tableHeaders.value[idx], ...patch };
    }
}

function handleEditField(field) {
    itemPopoverVisible.value = false;
    activePopoverItem.value = null;
    fieldEditor.value?.edit(field, itemPopoverAnchor.value);
}

const getIndentByParentId = (parentId) => {
    const parent = tableData.value.find((t) => t.id === parentId);

    return parent?._indent + 1 || 1;
};

const resetState = () => {
    workspaceStore.clearTableData();
    workspaceStore.setTableHeaders([]);
    workspaceStore.setColumnsWidths([]);
    workspaces.value = [];
    tableHeaders.value = [];
    columnWidths.value = [];
    tableData.value = [];
    selectedView.value = null;
};

const STATUS_TYPE_KEYS = {
    "Not started": "not_started",
    Active: "active",
    Done: "done",
    Closed: "closed",
};

const statusTypeLabel = (value) =>
    STATUS_TYPE_KEYS[value]
        ? t.value(`projects.grid_view.status_type.${STATUS_TYPE_KEYS[value]}`)
        : value;

const statusTypeOptions = computed(() =>
    ["Not started", "Active", "Done", "Closed"].map((value) => ({
        value,
        label: statusTypeLabel(value),
    })),
);

const updateStatusType = async (item, colIndex, value) => {
    await saveValue(colIndex, item, { target: { value } });
};

const isLoading = ref(false);
// Set true after filtred-mode load so fetchPage(1) shows data (prevents flash of unfiltered data)
const awaitingFilterLoad = ref(false);

const tableNameLookup = computed({
    get() {
        return workspaceStore.getTableNameLookup;
    },
    set(value) {
        workspaceStore.setTableNameLookup(value);
    },
});

const getTableName = (tableId) => workspaceStore.tableNameOf(tableId);

const workspaceTablesForLinked = computed({
    get() {
        return workspaceStore.getWorkspaceTablesForLinked;
    },
    set(value) {
        workspaceStore.setWorkspaceTablesForLinked(value);
    },
});

const workspaceTables = computed({
    get() {
        return workspaceStore.getWorkspaceTables;
    },
    set(value) {
        workspaceStore.setWorkspaceTables(value);
    },
});

const defaultFlatFilter = computed({
    get() {
        return workspaceStore.getDefaultFlatFilters;
    },
    set(value) {
        workspaceStore.setDefaultFlatFilters(value);
    },
});

const defaultGroupFilter = computed({
    get() {
        return workspaceStore.getDefaultGroupFilters;
    },
    set(value) {
        workspaceStore.setDefaultGroupFilters(value);
    },
});

const viewFilters = computed({
    get() {
        return workspaceStore.getFilters;
    },
    set(value) {
        workspaceStore.setFilters(value);
    },
});

const isDragDisabled = ref(true);

const loadData = async () => {
    users.value = userStore.users;

    if (props.isDialog) {
        isDragDisabled.value = false;

        return;
    }

    const load = viewLoad.start();

    pageLoad.cancel();
    isLoading.value = true;
    // The spinner covers the grid from the start, or its empty frame and
    // footer show until the view's filter is set up.
    awaitingFilterLoad.value = true;
    let ready = false;

    try {
        const req = workspaceService.getWorkspaceTables(route.params.id, "nav");

        resetState();
        workspaceStore.setFilterActive(false);
        isDragDisabled.value = false;

        const res = await req;

        if (!load.isCurrent()) return;
        const tables = res?.data || [];

        viewFilters.value = tables.flatMap((table) => table.filters || []);

        workspaceTablesForLinked.value = tables;
        workspaces.value = tables;
        workspaceTables.value = tables;

        tableNameLookup.value = tableDisplayNames(tables);

        const table = tables.find((t) => t.id === route.params.tid);

        if (!table) {
            tableData.value = [];
            awaitingFilterLoad.value = false;

            return;
        }

        currentLocalTableID.value = table.id;

        const selectedViewData = pickView(table, route.params.fid);

        selectedView.value = selectedViewData;

        const { headers, widths } = viewHeaders(table, selectedViewData);

        if (headers) {
            updateTableHeaders(headers);
            if (headers.some((h) => h.header_usage === "status")) loadStatusOptions();
        }

        columnWidths.value = widths;

        const groups = toRaw(selectedViewData?.filter?.groups || []);
        const flatFilters = toRaw(selectedViewData?.filter?.flatFilters || []);

        workspaceStore.setSavedGroups(groups);
        workspaceStore.setSavedFlatFilters(flatFilters);
        workspaceStore.setFilterActive(groups.length > 0 || flatFilters.length > 0);

        const filterState = viewFilterState(viewFilters.value, groups, flatFilters);

        workspaceStore.setFilter(filterState?.filter ?? null);
        if (filterState?.built) {
            workspaceStore.setSavedFlatFilters(filterState.built.flatFilters);
            workspaceStore.setSavedGroups(filterState.built.groups);
        }

        defaultFlatFilter.value = flatFilters;
        defaultGroupFilter.value = groups;

        currentPage.value = 1;
        restoreSortFromView(selectedViewData);

        awaitingFilterLoad.value = true;
        if (workspaceStore.getFilterActive) {
            const saved = viewFilters.value.find(
                (f) =>
                    f.is_active &&
                    f.table_id === route.params.tid &&
                    f.view_id === route.params.fid,
            );

            workspaceStore.setFastFilter(
                saved ? "saved_filter" : "all_tasks",
                saved ? saved.name : t.value(PRESET_LABELS.all_tasks),
            );
        } else {
            const preset = await loadTablePreset("open_tasks", route.params.id, route.params.tid);

            if (!load.isCurrent()) return;
            storeTablePreset(preset);
        }

        ready = true;
    } catch (error) {
        if (!load.isCurrent()) return;
        useAlertStore().showError(
            error?.response?.data || error?.message || t.value("projects.grid_view.load_failed"),
        );
        tableData.value = [];
        awaitingFilterLoad.value = false;
    } finally {
        if (load.isCurrent()) isLoading.value = false;
    }

    if (ready) await fetchPage(1);
};

// Each view has its own sort, or none; without resetting, the last table's
// sort, fields it may not have included, would be sent for this one.
function restoreSortFromView(selectedViewData) {
    const parsedSort = parseViewSort(toRaw(selectedViewData?.sort));

    sortOptions.value = parsedSort.length ? parsedSort : [{ field: "", direction: "asc" }];
    sortActive.value = parsedSort.length > 0;
}

// The filter rows read the status options from the store, so they are loaded
// once here rather than by each row.
function loadStatusOptions() {
    workspaceService
        .getTableStatusTypes(route.params.id, route.params.tid)
        .then((r) => {
            const opts = (r?.data?.options || []).map((o) => ({ id: o.id, name: o.name }));

            if (opts.length) workspaceStore.setStatusOptions(opts);
        })
        .catch(() => {});
}

const sortActive = computed({
    get() {
        return workspaceStore.getSortActive;
    },
    set(value) {
        workspaceStore.setSortActive(value);
    },
});

const sortOptions = computed({
    get() {
        return workspaceStore.getSortOptions;
    },
    set(value) {
        workspaceStore.setSortOptions(value);
    },
});

const dialogStyle = computed(() => {
    if (!props.isDialog) return {};

    // height:auto overrides the .table-container CSS height (calc(100vh-182px)) so the
    // dialog shrinks to fit its content. max-height caps it when there are many rows.
    return { height: "auto", maxHeight: "65vh", overflowY: "auto", overflowX: "auto" };
});

watch(
    () => workspaceStore.getGridReloadToken,
    () => {
        if (!props.isDialog) {
            fetchPage(1);
        }
    },
);

// filterActive watcher intentionally removed. In the paginated system,
// bumpGridReloadToken → fetchPage(1) already handles data reload when
// filters are cleared. Calling loadData() here races with fetchPage and
// leaves awaitingFilterLoad=true (spinner stuck).

watch(
    () => [route.params.id, route.params.tid, route.params.fid],
    async (newParams, oldParams) => {
        if (!props.isDialog && JSON.stringify(newParams) !== JSON.stringify(oldParams)) {
            await loadData();
        }
    },
    { immediate: true },
);

const isDataLoaded = ref(false);

watch(
    () => tableHeaders.value,
    (newHeaders) => {
        if (newHeaders && newHeaders.length > 0) {
            isDataLoaded.value = true;

            tableHeaders.value.forEach((header) => {
                if (header.visible === undefined) {
                    header.visible = true;
                }
            });
        }
    },
    { immediate: true },
);

function expandAllParentsForDialog(rows) {
    expandedTasks.value = new Set(
        rows.filter((r) => r.parent_task_id != null).map((r) => String(r.parent_task_id).trim()),
    );
}

watch(
    () => [props.isDialog, tableData.value, tableData.value?.length],
    () => {
        if (!props.isDialog) return;
        const rows = Array.isArray(tableData.value) ? tableData.value : [];

        if (!rows.length) return;
        expandAllParentsForDialog(rows);
        // Sync count: prefer backend total prop, fall back to row count
        totalRows.value = props.totalCount || rows.length;
        // On first mount in dialog mode, watch(tableData) doesn't fire (no change from
        // initial value), so renderedCount stays at RENDER_BATCH. Force it to cover all rows.
        renderedCount.value = Math.max(renderedCount.value, visibleData.value.length);
    },
    { immediate: true },
);
</script>

<style scoped src="../projectsTable.css"></style>

<style scoped>
.table-container {
    height: calc(100vh - 182px);
    overflow-y: auto;
    overflow-x: auto;
    background-color: rgb(255, 255, 255);
}

.bg-white {
    background-color: #ffffff;
}

.grid-item-text {
    transition: width 0.1s ease;
}

.grid-item-text:nth-child(odd) {
    /* background-color: #f9f9f9; */
}

.grid-item-text:nth-child(2) {
    /* background-color: white; */
}

.empty-row .grid-item-text {
    text-align: center;
}

.sticky-row {
    height: 42px;
    position: sticky;
    bottom: 0;
    background-color: white;
    border-top: 1px solid #ddd;
    z-index: 1; /* Ensure it appears above the scrollable content */
    display: flex;
    justify-content: space-between;
    align-items: center;
}

.empty-row {
    cursor: pointer;
}

.clickable-area {
    height: calc(100% + 18px);
    width: calc(100% + 12px);
    margin: -9px -6px;
    padding: 9px 8px;
    cursor: pointer;
    display: flex;
    align-items: center;
}

.task-row > .grid-item-text {
    font-size: 13px;
}

.task-row:hover > .grid-item-text {
    background-color: #f9fafb;
}

.alignTextCenter {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 100%;
}

.value-container {
    display: flex;
    flex-wrap: nowrap;
    overflow: hidden;
    width: 100%;
}

.new-task-row-sticky .grid-item-text {
    padding-top: 7px;
    padding-bottom: 7px;
}

.new-task-row-sticky {
    margin-top: -1px;
    position: sticky;
    bottom: 0;
    z-index: 1; /* Ensures it stays above other content */
    background-color: white;
    border-top: 1px solid #c5c5c5;

    border-bottom: none !important;
}

.new-task-row-sticky .grid-row {
    border-bottom: none !important;
}

.new-task-row-sticky .grid-item-text {
    border-bottom: none !important;
    height: 32px;
    padding-top: 0;
    padding-bottom: 0;
    align-items: center;
}

.new-task-row-sticky .grid-item-text span {
    display: flex;
    align-items: center;
    height: 100%;
}

.new-task-row-sticky .grid-item-text button {
    display: flex;
    align-items: center;
}

.new-task-row-sticky .grid-row .grid-item-text:nth-child(2) {
    z-index: 5;
    background: #fff;
    box-shadow: none;
    border-right: none;
}

.grid-item-text.task-highlight {
    background-color: #e0e7ff !important;
    transition: background-color 1s ease;
}
</style>
