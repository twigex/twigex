<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <TransitionRoot :show="show" as="template" @after-leave="reset">
        <Dialog as="div" class="relative z-40" @close="$emit('close')">
            <TransitionChild
                as="template"
                enter="ease-out duration-200"
                enter-from="opacity-0"
                enter-to="opacity-100"
                leave="ease-in duration-150"
                leave-from="opacity-100"
                leave-to="opacity-0"
            >
                <div class="fixed inset-0 bg-gray-500/75" />
            </TransitionChild>

            <div class="fixed inset-0 z-10 overflow-y-auto">
                <div class="flex min-h-full items-center justify-center p-4">
                    <TransitionChild
                        as="template"
                        enter="ease-out duration-200"
                        enter-from="opacity-0 translate-y-2"
                        enter-to="opacity-100 translate-y-0"
                        leave="ease-in duration-150"
                        leave-from="opacity-100 translate-y-0"
                        leave-to="opacity-0 translate-y-2"
                    >
                        <DialogPanel
                            class="flex max-h-[85vh] w-full max-w-4xl flex-col rounded-lg bg-white shadow-xl"
                        >
                            <div class="border-b border-gray-200 px-6 py-4">
                                <DialogTitle class="text-base font-semibold text-gray-900">
                                    {{
                                        isEdit
                                            ? t("collimato.builder.title_edit")
                                            : t("collimato.builder.title_new")
                                    }}
                                </DialogTitle>
                                <p class="mt-1 text-sm text-gray-500">
                                    {{ t("collimato.builder.subtitle") }}
                                </p>
                            </div>

                            <div class="relative flex-1 space-y-6 overflow-y-auto px-6 py-5">
                                <div
                                    v-if="loading"
                                    class="absolute inset-0 z-10 flex flex-col items-center justify-center gap-3 bg-white/80"
                                >
                                    <BaseSpinner size="lg" class="text-indigo-600" />
                                    <p class="text-sm text-gray-500">
                                        {{ t("collimato.builder.loading_schema") }}
                                    </p>
                                </div>

                                <!-- Source -->
                                <div class="grid grid-cols-1 gap-4 sm:grid-cols-3">
                                    <div>
                                        <label :class="labelClass">{{
                                            t("collimato.builder.connection")
                                        }}</label>
                                        <div class="mt-1">
                                            <BaseSelect
                                                v-model="connectionId"
                                                :options="connectionOptions"
                                                :disabled="isEdit"
                                                :placeholder="t('collimato.builder.connection')"
                                                @update:modelValue="onConnectionChange"
                                            />
                                        </div>
                                    </div>
                                    <div>
                                        <label :class="labelClass">{{
                                            t("collimato.builder.table")
                                        }}</label>
                                        <div class="mt-1">
                                            <BaseSelect
                                                v-model="tableName"
                                                :options="tableOptions"
                                                :disabled="isEdit || !connectionId"
                                                :placeholder="t('collimato.builder.table')"
                                                @update:modelValue="onTableChange"
                                            />
                                        </div>
                                    </div>
                                    <div>
                                        <label :class="labelClass">{{
                                            t("collimato.builder.cube_name")
                                        }}</label>
                                        <input
                                            v-model="cubeName"
                                            type="text"
                                            :disabled="isEdit"
                                            :class="[
                                                inputClass,
                                                nameError ? 'ring-red-600' : '',
                                                isEdit ? 'bg-gray-50 text-gray-400' : '',
                                            ]"
                                        />
                                        <p v-if="nameError" class="mt-1 text-sm text-red-600">
                                            {{ t("collimato.builder.name_error") }}
                                        </p>
                                    </div>
                                </div>

                                <!-- Dimensions -->
                                <div>
                                    <h3 class="text-sm font-semibold text-gray-900">
                                        {{ t("collimato.builder.dimensions") }}
                                    </h3>
                                    <p class="mb-2 text-xs text-gray-500">
                                        {{ t("collimato.builder.dimensions_help") }}
                                    </p>
                                    <div
                                        v-if="columns.length"
                                        class="overflow-hidden rounded-md ring-1 ring-inset ring-gray-200"
                                    >
                                        <table class="min-w-full divide-y divide-gray-200 text-sm">
                                            <thead
                                                class="bg-gray-50 text-left text-xs font-medium text-gray-500"
                                            >
                                                <tr>
                                                    <th class="w-16 px-3 py-2">
                                                        {{ t("collimato.builder.col_include") }}
                                                    </th>
                                                    <th class="px-3 py-2">
                                                        {{ t("collimato.builder.col_column") }}
                                                    </th>
                                                    <th class="w-32 px-3 py-2">
                                                        {{ t("collimato.builder.col_type") }}
                                                    </th>
                                                    <th class="w-20 px-3 py-2">
                                                        {{ t("collimato.builder.col_pk") }}
                                                    </th>
                                                    <th class="px-3 py-2">
                                                        {{ t("collimato.builder.col_title") }}
                                                    </th>
                                                </tr>
                                            </thead>
                                            <tbody class="divide-y divide-gray-100 bg-white">
                                                <tr
                                                    v-for="col in columns"
                                                    :key="col.name"
                                                    :class="col.include ? '' : 'opacity-50'"
                                                >
                                                    <td class="px-3 py-2">
                                                        <input
                                                            v-model="col.include"
                                                            type="checkbox"
                                                            class="h-4 w-4 rounded border-gray-300 text-indigo-600 focus:ring-indigo-600"
                                                        />
                                                    </td>
                                                    <td class="px-3 py-2 font-mono text-gray-900">
                                                        {{ col.name }}
                                                    </td>
                                                    <td class="px-3 py-2">
                                                        <BaseSelect
                                                            v-model="col.type"
                                                            :options="dimensionTypes"
                                                            :disabled="!col.include"
                                                        />
                                                    </td>
                                                    <td class="px-3 py-2">
                                                        <input
                                                            v-model="col.primaryKey"
                                                            type="checkbox"
                                                            :disabled="!col.include"
                                                            class="h-4 w-4 rounded border-gray-300 text-indigo-600 focus:ring-indigo-600"
                                                        />
                                                    </td>
                                                    <td class="px-3 py-2">
                                                        <input
                                                            v-model="col.title"
                                                            type="text"
                                                            :disabled="!col.include"
                                                            :placeholder="
                                                                t('collimato.builder.optional')
                                                            "
                                                            :class="cellInputClass"
                                                        />
                                                    </td>
                                                </tr>
                                            </tbody>
                                        </table>
                                    </div>
                                    <p v-else class="text-sm text-gray-400">
                                        {{ t("collimato.builder.pick_table") }}
                                    </p>
                                </div>

                                <!-- Custom dimensions -->
                                <div>
                                    <div class="flex items-center justify-between">
                                        <h3 class="text-sm font-semibold text-gray-900">
                                            {{ t("collimato.builder.custom_dimensions") }}
                                        </h3>
                                        <button
                                            type="button"
                                            class="text-sm font-medium text-indigo-600 hover:text-indigo-500"
                                            @click="addCustomDim"
                                        >
                                            +
                                            {{ t("collimato.builder.add_custom_dim") }}
                                        </button>
                                    </div>
                                    <p class="mb-2 text-xs text-gray-500">
                                        {{ t("collimato.builder.custom_dimensions_help") }}
                                    </p>

                                    <div
                                        v-for="(d, i) in customDimensions"
                                        :key="i"
                                        class="mb-2 space-y-2 rounded-md p-3 ring-1 ring-inset ring-gray-200"
                                    >
                                        <div class="grid grid-cols-12 items-center gap-2">
                                            <input
                                                v-model="d.name"
                                                type="text"
                                                :placeholder="t('collimato.builder.dim_name')"
                                                :class="[cellInputClass, 'col-span-4 font-mono']"
                                            />
                                            <div class="col-span-3">
                                                <BaseSelect
                                                    v-model="d.type"
                                                    :options="dimensionTypes"
                                                />
                                            </div>
                                            <input
                                                v-model="d.title"
                                                type="text"
                                                :placeholder="t('collimato.builder.optional')"
                                                :class="[cellInputClass, 'col-span-4']"
                                            />
                                            <button
                                                type="button"
                                                class="col-span-1 flex justify-center text-gray-400 hover:text-red-600"
                                                @click="customDimensions.splice(i, 1)"
                                            >
                                                <XMarkIcon class="h-5 w-5" />
                                            </button>
                                        </div>
                                        <textarea
                                            v-model="d.sql"
                                            rows="4"
                                            spellcheck="false"
                                            :placeholder="t('collimato.builder.sql_expression_ph')"
                                            class="block w-full rounded-md border-0 py-1.5 font-mono text-xs text-gray-900 ring-1 ring-inset ring-gray-300 focus:ring-2 focus:ring-inset focus:ring-indigo-600"
                                        />
                                    </div>
                                </div>

                                <!-- Measures -->
                                <div>
                                    <div class="flex items-center justify-between">
                                        <h3 class="text-sm font-semibold text-gray-900">
                                            {{ t("collimato.builder.measures") }}
                                        </h3>
                                        <button
                                            type="button"
                                            class="text-sm font-medium text-indigo-600 hover:text-indigo-500"
                                            @click="addMeasure"
                                        >
                                            +
                                            {{ t("collimato.builder.add_measure") }}
                                        </button>
                                    </div>
                                    <p class="mb-2 text-xs text-gray-500">
                                        {{ t("collimato.builder.measures_help") }}
                                    </p>

                                    <div
                                        class="mb-2 flex items-center gap-2 rounded-md bg-gray-50 px-3 py-2 text-sm text-gray-600"
                                    >
                                        <span class="font-mono">count</span>
                                        <span class="text-xs text-gray-400">{{
                                            t("collimato.builder.count_note")
                                        }}</span>
                                    </div>

                                    <div v-if="measures.length">
                                        <div
                                            class="grid grid-cols-12 gap-2 px-1 pb-1 text-xs font-medium text-gray-500"
                                        >
                                            <div class="col-span-3">
                                                {{ t("collimato.builder.measure_name") }}
                                            </div>
                                            <div class="col-span-2">
                                                {{ t("collimato.builder.col_type") }}
                                            </div>
                                            <div class="col-span-3">
                                                {{ t("collimato.builder.column") }}
                                            </div>
                                            <div class="col-span-3">
                                                {{ t("collimato.builder.col_title") }}
                                            </div>
                                            <div class="col-span-1"></div>
                                        </div>
                                        <div
                                            v-for="(m, i) in measures"
                                            :key="i"
                                            class="mb-2 grid grid-cols-12 items-center gap-2"
                                        >
                                            <input
                                                v-model="m.name"
                                                type="text"
                                                :placeholder="t('collimato.builder.measure_name')"
                                                :class="[cellInputClass, 'col-span-3']"
                                            />
                                            <div class="col-span-2">
                                                <BaseSelect
                                                    :model-value="m.type"
                                                    :options="measureTypes"
                                                    @update:model-value="setMeasureType(m, $event)"
                                                />
                                            </div>
                                            <div class="col-span-3">
                                                <BaseSelect
                                                    v-model="m.column"
                                                    :options="columnOptions"
                                                    :disabled="m.type === 'count'"
                                                    :placeholder="t('collimato.builder.column')"
                                                />
                                            </div>
                                            <input
                                                v-model="m.title"
                                                type="text"
                                                :placeholder="t('collimato.builder.optional')"
                                                :class="[cellInputClass, 'col-span-3']"
                                            />
                                            <button
                                                type="button"
                                                class="col-span-1 flex justify-center text-gray-400 hover:text-red-600"
                                                @click="measures.splice(i, 1)"
                                            >
                                                <XMarkIcon class="h-5 w-5" />
                                            </button>
                                        </div>
                                    </div>
                                </div>

                                <!-- Custom measures -->
                                <div>
                                    <div class="flex items-center justify-between">
                                        <h3 class="text-sm font-semibold text-gray-900">
                                            {{ t("collimato.builder.custom_measures") }}
                                        </h3>
                                        <button
                                            type="button"
                                            class="text-sm font-medium text-indigo-600 hover:text-indigo-500"
                                            @click="addCustomMeasure"
                                        >
                                            +
                                            {{ t("collimato.builder.add_custom_measure") }}
                                        </button>
                                    </div>
                                    <p class="mb-2 text-xs text-gray-500">
                                        {{ t("collimato.builder.custom_measures_help") }}
                                    </p>

                                    <div
                                        v-for="(m, i) in customMeasures"
                                        :key="i"
                                        class="mb-2 space-y-2 rounded-md p-3 ring-1 ring-inset ring-gray-200"
                                    >
                                        <div class="grid grid-cols-12 items-center gap-2">
                                            <input
                                                v-model="m.name"
                                                type="text"
                                                :placeholder="t('collimato.builder.measure_name')"
                                                :class="[cellInputClass, 'col-span-4 font-mono']"
                                            />
                                            <div class="col-span-3">
                                                <BaseSelect
                                                    v-model="m.type"
                                                    :options="customMeasureTypes"
                                                />
                                            </div>
                                            <input
                                                v-model="m.title"
                                                type="text"
                                                :placeholder="t('collimato.builder.optional')"
                                                :class="[cellInputClass, 'col-span-4']"
                                            />
                                            <button
                                                type="button"
                                                class="col-span-1 flex justify-center text-gray-400 hover:text-red-600"
                                                @click="customMeasures.splice(i, 1)"
                                            >
                                                <XMarkIcon class="h-5 w-5" />
                                            </button>
                                        </div>
                                        <textarea
                                            v-model="m.sql"
                                            rows="3"
                                            spellcheck="false"
                                            :placeholder="t('collimato.builder.measure_sql_ph')"
                                            class="block w-full rounded-md border-0 py-1.5 font-mono text-xs text-gray-900 ring-1 ring-inset ring-gray-300 focus:ring-2 focus:ring-inset focus:ring-indigo-600"
                                        />
                                    </div>
                                </div>

                                <!-- Joins -->
                                <div>
                                    <div class="flex items-center justify-between">
                                        <h3 class="text-sm font-semibold text-gray-900">
                                            {{ t("collimato.builder.joins") }}
                                        </h3>
                                        <button
                                            type="button"
                                            class="text-sm font-medium text-indigo-600 hover:text-indigo-500 disabled:text-gray-300"
                                            :disabled="!availableJoinCubes.length"
                                            @click="addJoin"
                                        >
                                            +
                                            {{ t("collimato.builder.add_join") }}
                                        </button>
                                    </div>
                                    <p class="mb-2 text-xs text-gray-500">
                                        {{ t("collimato.builder.joins_help") }}
                                    </p>

                                    <p
                                        v-if="!availableJoinCubes.length"
                                        class="text-sm text-gray-400"
                                    >
                                        {{ t("collimato.builder.no_cubes_to_join") }}
                                    </p>

                                    <div v-else-if="joins.length">
                                        <div
                                            class="grid grid-cols-12 gap-2 px-1 pb-1 text-xs font-medium text-gray-500"
                                        >
                                            <div class="col-span-3">
                                                {{ t("collimato.builder.target_cube") }}
                                            </div>
                                            <div class="col-span-3">
                                                {{ t("collimato.builder.relationship") }}
                                            </div>
                                            <div class="col-span-3">
                                                {{ t("collimato.builder.this_column") }}
                                            </div>
                                            <div class="col-span-2">
                                                {{ t("collimato.builder.other_column") }}
                                            </div>
                                            <div class="col-span-1"></div>
                                        </div>
                                        <div
                                            v-for="(j, i) in joins"
                                            :key="i"
                                            class="mb-2 grid grid-cols-12 items-center gap-2"
                                        >
                                            <div class="col-span-3">
                                                <BaseSelect
                                                    v-model="j.target"
                                                    :options="availableJoinCubes"
                                                    :placeholder="
                                                        t('collimato.builder.target_cube')
                                                    "
                                                    @update:modelValue="onJoinTargetChange(j)"
                                                />
                                            </div>
                                            <div class="col-span-3">
                                                <BaseSelect
                                                    v-model="j.relationship"
                                                    :options="relationshipOptions"
                                                />
                                            </div>
                                            <div class="col-span-3">
                                                <BaseSelect
                                                    v-model="j.thisColumn"
                                                    :options="columnOptions"
                                                    :placeholder="
                                                        t('collimato.builder.this_column')
                                                    "
                                                />
                                            </div>
                                            <div class="col-span-2">
                                                <BaseSelect
                                                    v-if="targetColumns(j.target)"
                                                    v-model="j.otherColumn"
                                                    :options="targetColumns(j.target)"
                                                    :placeholder="
                                                        t('collimato.builder.other_column')
                                                    "
                                                />
                                                <input
                                                    v-else
                                                    v-model="j.otherColumn"
                                                    type="text"
                                                    :placeholder="
                                                        t('collimato.builder.other_column')
                                                    "
                                                    :class="cellInputClass"
                                                />
                                            </div>
                                            <button
                                                type="button"
                                                class="col-span-1 flex justify-center text-gray-400 hover:text-red-600"
                                                @click="joins.splice(i, 1)"
                                            >
                                                <XMarkIcon class="h-5 w-5" />
                                            </button>
                                        </div>
                                    </div>
                                </div>
                            </div>

                            <div
                                class="flex items-center justify-end gap-3 border-t border-gray-200 px-6 py-4"
                            >
                                <button
                                    type="button"
                                    class="rounded-md bg-white px-3 py-2 text-sm font-semibold text-gray-900 ring-1 ring-inset ring-gray-300 hover:bg-gray-50"
                                    @click="$emit('close')"
                                >
                                    {{ t("common.button.cancel") }}
                                </button>
                                <BaseButton
                                    color="bg-indigo-600 hover:bg-indigo-500 text-white"
                                    :isDisabled="!canSave"
                                    :isLoading="saving"
                                    @click="save"
                                >
                                    {{ t("common.button.save") }}
                                </BaseButton>
                            </div>
                        </DialogPanel>
                    </TransitionChild>
                </div>
            </div>
        </Dialog>
    </TransitionRoot>
</template>

<script setup>
import { ref, computed, watch } from "vue";
import { TransitionRoot, TransitionChild, Dialog, DialogPanel, DialogTitle } from "@headlessui/vue";
import { XMarkIcon } from "@heroicons/vue/20/solid";
import BaseButton from "@/components/BaseButton.vue";
import BaseSpinner from "@/components/BaseSpinner.vue";
import BaseSelect from "@/components/BaseSelect.vue";
import collimatoService from "@/services/collimatoService";
import { useAlertStore } from "@/store/alerts";
import { extractErrorMessage } from "@/utils/errors";
import { t } from "@/i18n/index.js";

const props = defineProps({
    show: { type: Boolean, default: false },
    workspaceId: { type: String, required: true },
    // Existing cube names (no extension) available as join targets.
    existingCubes: { type: Array, default: () => [] },
    // Existing file to edit (must carry builder_model); null for a new cube.
    editFile: { type: Object, default: null },
});
const emit = defineEmits(["close", "saved"]);

const alertStore = useAlertStore();

const labelClass = "block text-sm font-medium text-gray-700";
const inputClass =
    "mt-1 block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm";
const cellInputClass =
    "block w-full rounded-md border-0 py-1.5 text-gray-900 ring-1 ring-inset ring-gray-300 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm disabled:bg-gray-50";

const dimensionTypes = ["string", "number", "time", "boolean"];
const measureTypes = ["sum", "avg", "min", "max", "countDistinct", "count"];
const customMeasureTypes = ["number", "sum", "avg", "min", "max", "countDistinct"];
const relationshipOptions = computed(() => [
    {
        value: "many_to_one",
        label: t.value("collimato.builder.rel_many_to_one"),
    },
    {
        value: "one_to_many",
        label: t.value("collimato.builder.rel_one_to_many"),
    },
    { value: "one_to_one", label: t.value("collimato.builder.rel_one_to_one") },
]);

const databases = ref([]); // [{ database: connId, name: label, tables: [...] }]
const connectionId = ref("");
const tableName = ref("");
const cubeName = ref("");
const columns = ref([]); // [{ name, type, include, primaryKey, title }]
const customDimensions = ref([]); // raw-SQL dimensions [{ name, sql, type, title }]
const measures = ref([]); // column-based measures (count is implicit)
const customMeasures = ref([]); // raw-SQL measures [{ name, type, sql, title }]
const joins = ref([]); // [{ target, relationship, thisColumn, otherColumn }]
const saving = ref(false);
const loading = ref(false); // introspection (tables/columns) can be slow

const isEdit = computed(() => !!props.editFile);

function setMeasureType(measure, type) {
    measure.type = type;

    if (type === "count") {
        measure.column = "";
    }
}

const connectionOptions = computed(() =>
    databases.value.map((db) => ({ value: db.database, label: db.name })),
);
const tablesForConnection = computed(() => {
    const db = databases.value.find((d) => d.database === connectionId.value);

    return db ? db.tables : [];
});
const tableOptions = computed(() => tablesForConnection.value.map((tbl) => tbl.name));
const columnOptions = computed(() => columns.value.map((c) => c.name));
const availableJoinCubes = computed(() =>
    props.existingCubes.filter((c) => c.name !== cubeName.value).map((c) => c.name),
);

// targetColumns resolves the joined cube's table columns (via its stored model)
// so the "other column" can be a dropdown. Returns null when unknown (e.g. a
// hand-written cube with no builder model), in which case we fall back to text.
function targetColumns(targetName) {
    const cube = props.existingCubes.find((c) => c.name === targetName);

    if (!cube || !cube.model) return null;
    const db = databases.value.find((d) => d.database === cube.model.connection_id);
    const tbl = db?.tables.find((t) => t.name === cube.model.sql_table);

    return tbl ? tbl.columns.map((c) => c.name) : null;
}

function onJoinTargetChange(j) {
    j.otherColumn = "";
}

const nameError = computed(
    () => cubeName.value !== "" && !/^[A-Za-z_][A-Za-z0-9_]*$/.test(cubeName.value),
);

const canSave = computed(
    () =>
        !!connectionId.value &&
        !!tableName.value &&
        !!cubeName.value &&
        !nameError.value &&
        (columns.value.some((c) => c.include) ||
            customDimensions.value.some((d) => d.name && d.sql)),
);

function cubeType(sqlType) {
    const s = (sqlType || "").toLowerCase();

    if (/^(varchar|text|char|tinytext|mediumtext|longtext|character)/.test(s)) return "string";
    if (/^(int|tinyint|smallint|mediumint|bigint|float|double|numeric|decimal|real)/.test(s))
        return "number";
    if (/^(timestamp|datetime|date|time)/.test(s)) return "time";
    if (/^(bool|boolean)/.test(s)) return "boolean";

    return "string";
}

function loadColumnsForTable() {
    const tbl = tablesForConnection.value.find((t) => t.name === tableName.value);

    columns.value = (tbl?.columns || []).map((c) => ({
        name: c.name,
        type: cubeType(c.type),
        include: true,
        primaryKey: c.name === "id",
        title: "",
    }));
}

function onConnectionChange() {
    tableName.value = "";
    columns.value = [];
}

function onTableChange() {
    if (!isEdit.value) cubeName.value = tableName.value;
    loadColumnsForTable();
}

function addMeasure() {
    measures.value.push({ name: "", type: "sum", column: "", title: "" });
}

function addCustomDim() {
    customDimensions.value.push({
        name: "",
        sql: "",
        type: "string",
        title: "",
    });
}

function addCustomMeasure() {
    customMeasures.value.push({
        name: "",
        type: "number",
        sql: "",
        title: "",
    });
}

function addJoin() {
    joins.value.push({
        target: "",
        relationship: "many_to_one",
        thisColumn: "",
        otherColumn: "",
    });
}

function reset() {
    connectionId.value = "";
    tableName.value = "";
    cubeName.value = "";
    columns.value = [];
    customDimensions.value = [];
    measures.value = [];
    customMeasures.value = [];
    joins.value = [];
    saving.value = false;
    loading.value = false;
}

async function initialize() {
    reset();
    loading.value = true;
    try {
        const res = await collimatoService.tables(props.workspaceId);

        databases.value = res.data || [];
        if (isEdit.value && props.editFile.builder_model) {
            hydrateFromModel(JSON.parse(props.editFile.builder_model));
        }
    } catch (error) {
        alertStore.showError(extractErrorMessage(error));
    } finally {
        loading.value = false;
    }
}

function hydrateFromModel(model) {
    connectionId.value = model.connection_id;
    tableName.value = model.sql_table;
    cubeName.value = model.name;
    loadColumnsForTable();

    const dims = model.dimensions || [];
    const byName = Object.fromEntries(dims.filter((d) => !d.custom).map((d) => [d.name, d]));

    columns.value.forEach((col) => {
        const d = byName[col.name];

        col.include = !!d;
        if (d) {
            col.type = d.type;
            col.primaryKey = !!d.primary_key;
            col.title = d.title || "";
        }
    });

    customDimensions.value = dims
        .filter((d) => d.custom)
        .map((d) => ({
            name: d.name,
            sql: d.sql || "",
            type: d.type || "string",
            title: d.title || "",
        }));

    const ms = model.measures || [];

    measures.value = ms
        .filter((m) => m.name !== "count" && !m.custom)
        .map((m) => ({
            name: m.name,
            type: m.type,
            column: m.sql || "",
            title: m.title || "",
        }));
    customMeasures.value = ms
        .filter((m) => m.custom)
        .map((m) => ({
            name: m.name,
            type: m.type,
            sql: m.sql || "",
            title: m.title || "",
        }));

    joins.value = (model.joins || []).map((j) => ({
        target: j.name,
        relationship: j.relationship || "many_to_one",
        thisColumn: j.this_column || "",
        otherColumn: j.other_column || "",
    }));
}

async function save() {
    if (!canSave.value) return;
    saving.value = true;

    const model = {
        name: cubeName.value,
        sql_table: tableName.value,
        connection_id: connectionId.value,
        dimensions: [
            ...columns.value
                .filter((c) => c.include)
                .map((c) => ({
                    name: c.name,
                    sql: c.name,
                    type: c.type,
                    primary_key: c.primaryKey,
                    title: c.title || undefined,
                })),
            ...customDimensions.value
                .filter((d) => d.name && d.sql)
                .map((d) => ({
                    name: d.name,
                    sql: d.sql,
                    type: d.type,
                    title: d.title || undefined,
                    custom: true,
                })),
        ],
        measures: [
            { name: "count", type: "count" },
            ...measures.value
                .filter((m) => m.name)
                .map((m) => ({
                    name: m.name,
                    type: m.type,
                    sql: m.type === "count" ? undefined : m.column || undefined,
                    title: m.title || undefined,
                })),
            ...customMeasures.value
                .filter((m) => m.name && m.sql)
                .map((m) => ({
                    name: m.name,
                    type: m.type,
                    sql: m.sql,
                    title: m.title || undefined,
                    custom: true,
                })),
        ],
        joins: joins.value
            .filter((j) => j.target && j.thisColumn && j.otherColumn)
            .map((j) => ({
                name: j.target,
                relationship: j.relationship,
                this_column: j.thisColumn,
                other_column: j.otherColumn,
            })),
    };

    try {
        const res = await collimatoService.build(props.workspaceId, {
            type: "cube",
            model,
            overwrite: isEdit.value,
        });

        emit("saved", res.data);
        emit("close");
    } catch (error) {
        alertStore.showError(extractErrorMessage(error));
    } finally {
        saving.value = false;
    }
}

watch(
    () => props.show,
    (open) => {
        if (open) initialize();
    },
);
</script>
