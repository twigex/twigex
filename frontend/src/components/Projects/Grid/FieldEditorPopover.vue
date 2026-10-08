<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <Popover v-if="popoverVisible" class="popover">
        <PopoverButton ref="popoverButton" style="display: none">{{
            t("projects.grid_view.trigger")
        }}</PopoverButton>
        <Teleport to="body">
            <div
                ref="popoverContent"
                :style="{ ...fieldPopupStyles, maxWidth: '250px' }"
                class="z-[60] rounded-lg bg-white p-4 shadow-xl ring-1 ring-black/5"
                @click.stop
            >
                <PopoverPanel static class="bg-white">
                    <p v-if="isEditMode" class="text-sm font-semibold text-gray-800 mb-2">
                        {{ t("projects.grid_view.edit_field") }}
                    </p>
                    <div class="mt-3">
                        <label
                            for="name"
                            class="block text-sm text-left font-medium leading-6 text-gray-900"
                            >{{ t("projects.grid_view.field_type") }}</label
                        >
                        <div class="mt-2">
                            <Listbox as="div" v-model="selected" :disabled="isEditMode">
                                <div class="relative mt-2">
                                    <ListboxButton
                                        :class="[
                                            'relative w-full cursor-default rounded-md py-1.5 pl-3 pr-10 text-left shadow-sm ring-1 ring-inset ring-gray-300 focus:outline-none sm:text-sm sm:leading-6',
                                            isEditMode
                                                ? 'bg-gray-50 text-gray-500 cursor-not-allowed'
                                                : 'bg-white text-gray-900 focus:ring-2 focus:ring-indigo-600',
                                        ]"
                                    >
                                        <span class="block truncate">{{ selected.name }}</span>
                                        <span
                                            class="pointer-events-none absolute inset-y-0 right-0 flex items-center pr-2"
                                        >
                                            <ChevronUpDownIcon
                                                class="h-5 w-5 text-gray-400"
                                                aria-hidden="true"
                                            />
                                        </span>
                                    </ListboxButton>

                                    <transition
                                        leave-active-class="transition ease-in duration-100"
                                        leave-from-class="opacity-100"
                                        leave-to-class="opacity-0"
                                    >
                                        <ListboxOptions
                                            class="absolute z-10 mt-1 max-h-60 w-full overflow-auto rounded-md bg-white py-1 text-base shadow-lg ring-1 ring-black ring-opacity-5 focus:outline-none sm:text-sm"
                                        >
                                            <ListboxOption
                                                as="template"
                                                v-for="type in types"
                                                :key="type.value"
                                                :value="type"
                                                v-slot="{ active, selected: isSelected }"
                                                @click="loadTables(selected)"
                                            >
                                                <li
                                                    :class="[
                                                        active
                                                            ? 'bg-indigo-600 text-white'
                                                            : 'text-gray-900',
                                                        'relative cursor-default select-none py-2 pl-3 pr-9',
                                                    ]"
                                                >
                                                    <span
                                                        :class="[
                                                            isSelected
                                                                ? 'font-semibold'
                                                                : 'font-normal',
                                                            'block truncate',
                                                        ]"
                                                    >
                                                        {{
                                                            type.display_name?.Valid &&
                                                            type.display_name?.String?.trim()
                                                                ? type.display_name.String
                                                                : type.name
                                                        }}
                                                    </span>

                                                    <span
                                                        v-if="isSelected"
                                                        :class="[
                                                            active
                                                                ? 'text-white'
                                                                : 'text-indigo-600',
                                                            'absolute inset-y-0 right-0 flex items-center pr-4',
                                                        ]"
                                                    >
                                                        <CheckIcon
                                                            class="h-5 w-5"
                                                            aria-hidden="true"
                                                        />
                                                    </span>
                                                </li>
                                            </ListboxOption>
                                        </ListboxOptions>
                                    </transition>
                                </div>
                            </Listbox>
                        </div>
                    </div>

                    <FormulaFieldEditor
                        v-if="
                            selected.value === 'calculations' ||
                            (isEditMode && editingFieldIsCalculations)
                        "
                        class="mt-3 space-y-3"
                        :editor="formulaEditor"
                    />

                    <div v-if="selectedType === 'link'" class="mt-3">
                        <label
                            for="name"
                            class="block text-sm text-left font-medium leading-6 text-gray-900"
                        >
                            {{ t("projects.grid_view.linked_table") }}
                        </label>
                        <div class="mt-2">
                            <Listbox as="div" v-model="selectedTable" :disabled="isEditMode">
                                <div class="relative mt-2">
                                    <ListboxButton
                                        :class="[
                                            'relative w-full cursor-default rounded-md py-1.5 pl-3 pr-10 text-left shadow-sm ring-1 ring-inset ring-gray-300 focus:outline-none sm:text-sm sm:leading-6',
                                            isEditMode
                                                ? 'bg-gray-50 text-gray-500 cursor-not-allowed'
                                                : 'bg-white text-gray-900 focus:ring-2 focus:ring-indigo-600',
                                        ]"
                                    >
                                        <span
                                            class="block truncate"
                                            style="
                                                max-width: 250px;
                                                overflow: hidden;
                                                text-overflow: ellipsis;
                                                white-space: nowrap;
                                            "
                                        >
                                            {{ selectedTableName }}
                                        </span>
                                        <span
                                            class="pointer-events-none absolute inset-y-0 right-0 flex items-center pr-2"
                                        >
                                            <ChevronUpDownIcon
                                                class="h-5 w-5 text-gray-400"
                                                aria-hidden="true"
                                            />
                                        </span>
                                    </ListboxButton>

                                    <transition
                                        leave-active-class="transition ease-in duration-100"
                                        leave-from-class="opacity-100"
                                        leave-to-class="opacity-0"
                                    >
                                        <ListboxOptions
                                            class="absolute z-10 mt-1 max-h-60 w-full overflow-auto rounded-md bg-white py-1 text-base shadow-lg ring-1 ring-black ring-opacity-5 focus:outline-none sm:text-sm"
                                            style="right: 0; left: auto"
                                        >
                                            <ListboxOption
                                                as="template"
                                                v-for="table in linkTables"
                                                :key="table.id"
                                                :value="table"
                                                v-slot="{ active, selected: isSelected }"
                                            >
                                                <li
                                                    :class="[
                                                        active
                                                            ? 'bg-indigo-600 text-white'
                                                            : 'text-gray-900',
                                                        'relative cursor-default select-none py-2 pl-3 pr-9',
                                                    ]"
                                                >
                                                    <span
                                                        :class="[
                                                            isSelected
                                                                ? 'font-semibold'
                                                                : 'font-normal',
                                                            'block truncate',
                                                        ]"
                                                    >
                                                        {{
                                                            table.display_name?.Valid &&
                                                            table.display_name?.String?.trim()
                                                                ? table.display_name.String
                                                                : table.name
                                                        }}
                                                    </span>

                                                    <span
                                                        v-if="isSelected"
                                                        :class="[
                                                            active
                                                                ? 'text-white'
                                                                : 'text-indigo-600',
                                                            'absolute inset-y-0 right-0 flex items-center pr-4',
                                                        ]"
                                                    >
                                                        <CheckIcon
                                                            class="h-5 w-5"
                                                            aria-hidden="true"
                                                        />
                                                    </span>
                                                </li>
                                            </ListboxOption>
                                        </ListboxOptions>
                                    </transition>
                                </div>
                            </Listbox>
                        </div>
                    </div>

                    <div class="mt-2">
                        <input
                            v-model="newField"
                            type="text"
                            name="value"
                            id="value"
                            maxlength="50"
                            class="block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6"
                            :placeholder="t('projects.grid_view.new_field_name')"
                        />
                    </div>

                    <div
                        v-if="selected.value === 'link' && !isEditMode"
                        class="mt-2 flex items-center"
                    >
                        <input
                            type="checkbox"
                            v-model="linkBothDirections"
                            :name="inputName"
                            class="h-4 w-4 rounded border-gray-300 text-indigo-600 focus:ring-indigo-600"
                        />
                        <span class="ml-2 text-gray-900 sm:text-sm sm:leading-6">
                            {{ inputName }}
                        </span>
                    </div>

                    <div
                        v-if="selected.value === 'link' && linkBothDirections"
                        class="mt-2 flex items-center"
                    >
                        <input
                            v-model="fieldNameInSecondTable"
                            type="text"
                            name="value"
                            maxlength="50"
                            id="value"
                            :disabled="isEditMode"
                            :class="[
                                'block w-full rounded-md border-0 py-1.5 shadow-sm ring-1 ring-inset ring-gray-300 sm:text-sm sm:leading-6',
                                isEditMode
                                    ? 'bg-gray-50 text-gray-500 cursor-not-allowed'
                                    : 'bg-white text-gray-900 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600',
                            ]"
                            :placeholder="t('projects.grid_view.new_field_name_in_secound_table')"
                        />
                    </div>

                    <div class="mt-4 flex justify-end gap-x-2">
                        <BaseButton
                            type="button"
                            size="small"
                            variant="secondary"
                            @click="popoverVisible = false"
                        >
                            {{ t("common.button.cancel") }}
                        </BaseButton>
                        <BaseButton
                            v-if="isEditMode && editingFieldIsCalculations"
                            type="button"
                            size="small"
                            @click="saveFieldFormula()"
                        >
                            {{ t("common.button.save") }}
                        </BaseButton>
                        <BaseButton
                            v-else-if="isEditMode"
                            type="button"
                            size="small"
                            @click="saveFieldEdit()"
                        >
                            {{ t("common.button.save") }}
                        </BaseButton>
                        <BaseButton v-else type="button" size="small" @click="createNewField()">
                            {{ t("common.button.create") }}
                        </BaseButton>
                    </div>
                </PopoverPanel>
            </div>
        </Teleport>
    </Popover>
</template>

<script setup>
import { t } from "@/i18n/index.js";
import { ref, computed, watch, nextTick, onBeforeUnmount, toRef } from "vue";
import { useRoute } from "vue-router";
import {
    Popover,
    PopoverButton,
    PopoverPanel,
    Listbox,
    ListboxButton,
    ListboxOptions,
    ListboxOption,
} from "@headlessui/vue";
import { CheckIcon, ChevronUpDownIcon } from "@heroicons/vue/20/solid";
import BaseButton from "@/components/BaseButton.vue";
import { POPULAR_FORMULAS } from "@/constants/formulasCatalog.js";
import { useAnchoredPopup } from "@/composables/useAnchoredPopup";
import { useFormulaEditor } from "@/composables/projects/useFormulaEditor";
import FormulaFieldEditor from "@/components/Projects/Grid/FormulaFieldEditor.vue";
import workspaceService from "@/services/workspaceService";
import { useWorkspaceStore } from "@/store/workspaces";
import { useAlertStore } from "@/store/alerts";
import { extractErrorMessage } from "@/utils/errors";
import { fieldLabel } from "@/utils/projects/rows";
import { formulaArgsFor, normalizeFormulaSpec } from "@/utils/projects/formulas";
import { workspaceTables } from "@/utils/projects/tree";

const props = defineProps({
    tableId: { type: String, default: "" },
    tableHeaders: { type: Array, default: () => [] },
    canManageFields: { type: Boolean, default: false },
    viewId: { type: String, default: null },
});

const emit = defineEmits(["created", "update-header"]);

const route = useRoute();
const workspaceStore = useWorkspaceStore();

const linkBothDirections = ref(false);
const inputName = ref(t.value("projects.grid_view.link_table_both_directions"));
const fieldNameInSecondTable = ref("");
let newField = ref("");

const types = [
    { name: t.value("projects.grid_view.types_text"), value: "text" },
    { name: t.value("projects.grid_view.types_number"), value: "number" },
    { name: t.value("projects.grid_view.types_decimal"), value: "decimal" },
    { name: t.value("projects.grid_view.types_checkbox"), value: "bool" },
    {
        name: t.value("projects.grid_view.types_link_to_records"),
        value: "link",
    },
    {
        name: t.value("projects.grid_view.types_single_select"),
        value: "single select",
    },
    { name: t.value("projects.grid_view.types_person"), value: "assignee" },
    { name: t.value("projects.grid_view.types_date"), value: "date" },
    {
        name: t.value("projects.grid_view.types_calculations"),
        value: "calculations",
    },
    {
        name: t.value("projects.grid_view.types_link_to_table"),
        value: "master link",
    },
    { name: t.value("projects.grid_view.types_website_url"), value: "url" },
    { name: t.value("projects.grid_view.types_file"), value: "file" },
];

const linkTables = ref([]);
const selected = ref({
    name: t.value("projects.grid_view.types_text"),
    value: "text",
});
const selectedTable = ref(linkTables.value[0]);
const popoverVisible = ref(false);
const popoverContent = ref(null);
const { floatingStyles: fieldPopupStyles, anchorTo: anchorFieldPopup } = useAnchoredPopup({
    floating: popoverContent,
});
const handleClickOutside = (event) => {
    if (!popoverContent.value) return;

    const popoverContentElement =
        popoverContent.value instanceof Element
            ? popoverContent.value
            : (popoverContent.value.$el ?? null);

    if (popoverContentElement && !popoverContentElement.contains(event.target)) {
        popoverVisible.value = false;
        newField.value = "";
        selected.value = types[0];
        selectedTable.value = linkTables.value[0];
        document.removeEventListener("mousedown", handleClickOutside);
    }
};

const togglePopover = async (event) => {
    anchorFieldPopup(event.currentTarget);
    popoverVisible.value = !popoverVisible.value;

    if (!popoverVisible.value) {
        document.removeEventListener("mousedown", handleClickOutside);

        return;
    }

    await nextTick();
    document.addEventListener("mousedown", handleClickOutside);
};

async function loadTables(selected) {
    if (selected.value === "link") {
        try {
            const tableID = props.tableId;

            const response = await workspaceStore.fetchWorkspaceNav(route.params.id);
            const workspace = response.data;

            if (!workspace) {
                return;
            }

            const allTables = [...(workspace.tables || []), ...(workspace.children || [])];

            const remainingTables = allTables.filter((table) => {
                return (
                    table.id !== tableID &&
                    (table.single_select === false || table.single_select == null) &&
                    (!table.parent_table_id || table.parent_table_id === "")
                );
            });

            const uniqueTables = Array.from(new Set(remainingTables.map((table) => table.id))).map(
                (id) => remainingTables.find((table) => table.id === id),
            );

            const filteredTables = uniqueTables.filter((table) => {
                if (
                    typeof table.single_select === "undefined" ||
                    typeof table.parent_table_id === "undefined"
                ) {
                    return false;
                }

                const isSingleSelect = table.single_select === true;
                const isLinkedTable =
                    table.parent_table_id !== null && table.parent_table_id !== "";

                return !isSingleSelect && !isLinkedTable;
            });

            linkTables.value.splice(0, linkTables.value.length, ...filteredTables);
        } catch (error) {
            console.warn("loadTables failed:", error);
        }
    }
}

const formulaEditor = useFormulaEditor(toRef(props, "tableHeaders"));
const {
    chosenFormula,
    formulaArgs,
    useMultipleFields,
    allHeaders,
    payload: getFormulaPayload,
} = formulaEditor;

function createNewField() {
    if (!props.canManageFields) {
        useAlertStore().showError(t.value("projects.grid_view.no_permission_edit_fields"));

        return;
    }

    const isCalculations = (selected.value?.value || "").toUpperCase() === "CALCULATIONS";
    const isFile = selected.value.value === "file";
    // A table picked for a link and then left is still selected; only a link
    // field has one.
    const linkedTableID = selected.value.value === "link" ? selectedTable.value?.id || null : null;
    const formulaSpec = isCalculations ? getFormulaPayload() : null;

    if (isCalculations && (!formulaSpec || !formulaSpec.name)) {
        useAlertStore().showError(t.value("projects.grid_view.pick_formula_alert"));

        return;
    }

    workspaceService
        .createNewField({
            fieldNameDisplay: (newField.value || "").trim(),
            fieldType: selected.value.value,
            workspace_id: route.params.id,
            table_id: route.params.tid,
            linkedTableID,
            selectedType: selected.value.value,
            linkBothDirections: linkBothDirections.value,
            fieldNameInSecondTableDisplay: (fieldNameInSecondTable.value || "").trim(),
            view_id: props.viewId,
            formula: isCalculations ? formulaSpec : null,
            header_usage: isCalculations ? "calculations" : isFile ? "file" : undefined,
        })
        .then((response) => {
            const newFieldData = response.data;

            if (newFieldData.header_usage == "single select") {
                newFieldData.single_select = true;
            }

            newFieldData.parent_table_id = linkedTableID;

            newFieldData.header_type = newFieldData.header_type.toUpperCase();

            if (isFile && !newFieldData.header_usage) {
                newFieldData.header_usage = "file";
            }

            if (isCalculations && formulaSpec) {
                newFieldData.formula = formulaSpec;
            }

            emit("created", newFieldData);

            linkBothDirections.value = false;
            popoverVisible.value = false;

            nextTick(() => {
                newField.value = "";
                selected.value = types[0];
                selectedTable.value = null;
            });
        })
        .catch((error) => {
            popoverVisible.value = false;
            useAlertStore().showError(
                error.response?.data || t.value("projects.grid_view.failed_to_add_field"),
            );
        });
}

const selectedTableName = computed(() => {
    const table = selectedTable.value;

    if (!table) return t.value("projects.dialogs.master_link_dialog.select_table");

    return table.display_name?.Valid && table.display_name?.String?.trim()
        ? table.display_name.String
        : table.name;
});

const selectedType = computed(() => selected.value.value);

watch(popoverVisible, (visible) => {
    if (!visible) {
        newField.value = "";
        fieldNameInSecondTable.value = "";

        editingField.value = null;
        formulaEditor.reset();
    }
});

// Clear stale type-specific state whenever the field type selector changes.
// Without this, switching from "link" to "calculations" leaves linkBothDirections=true
// and fieldNameInSecondTable set, causing a backend validation error on save.
// Skip during edit mode. handleEditField sets selected then populates formula/link state.
watch(selected, (newType) => {
    if (isEditMode.value) return;
    linkBothDirections.value = false;
    fieldNameInSecondTable.value = "";
    formulaEditor.reset();
    if (newType?.value === "link") {
        loadTables(newType);
    } else {
        selectedTable.value = null;
    }
});

const editingField = ref(null);

const isEditMode = computed(() => editingField.value !== null);

const editingFieldIsCalculations = computed(
    () => (editingField.value?.header_usage || "").toLowerCase() === "calculations",
);

async function handleEditField(field, anchor) {
    editingField.value = null; // clear stale state before re-populating
    editingField.value = field;

    const usage = (field.header_usage || "").toLowerCase();
    const matchedType = types.find((t) => t.value === usage) || types[0];

    selected.value = matchedType;

    newField.value = field.display_name?.String?.trim() || field.display_name || field.name || "";

    // For link fields: load available tables then pre-select the linked one.
    // field.linked_id is the junction table ID (created for the link).
    // The junction table has parent_table_id = the actual target table ID.
    if (usage === "link") {
        await loadTables(matchedType);
        if (field.linked_id) {
            try {
                const response = await workspaceStore.fetchWorkspaceNav(route.params.id);
                const allTables = workspaceTables(response.data);
                // Resolve junction table → actual target table via parent_table_id
                const junctionTable = allTables.find((t) => t.id === field.linked_id);
                const targetId = junctionTable?.parent_table_id;

                if (targetId) {
                    const target = allTables.find((t) => t.id === targetId);

                    if (target) selectedTable.value = target;
                }
            } catch (_) {
                /* ignore */
            }
        }

        linkBothDirections.value = !!field.link_both_directions;
        // linked_name is the other side's column, so its name is looked up.
        const otherSide = selectedTable.value?.headers?.find((h) => h.name === field.linked_name);

        fieldNameInSecondTable.value = otherSide ? fieldLabel(otherSide) : "";
    }

    // Calculate formula + args before opening. Args are applied after nextTick so
    // that watch(chosenFormula) fires first and doesn't wipe what we set.
    let pendingUseMultipleFields = false;
    let pendingFormulaArgs = [];

    const spec = usage === "calculations" ? normalizeFormulaSpec(field.formula) : null;
    const formula = spec?.name ? POPULAR_FORMULAS.find((f) => f.name === spec.name) : null;

    if (formula) {
        chosenFormula.value = formula;
        if (spec.args?.length) {
            pendingUseMultipleFields = formula.arity === "variadic";
            pendingFormulaArgs = formulaArgsFor(formula, spec, allHeaders.value);
        }
    }

    anchorFieldPopup(anchor);
    popoverVisible.value = true;

    // Wait for watch(chosenFormula) to fire (it resets useMultipleFields/formulaArgs),
    // then apply our pre-filled values on top.
    await nextTick();
    useMultipleFields.value = pendingUseMultipleFields;
    formulaArgs.value = pendingFormulaArgs;

    document.addEventListener("mousedown", handleClickOutside);
}

// Saves display name if it differs from the stored value. Returns false on error.
async function saveDisplayNameIfChanged() {
    const original = editingField.value.display_name?.String?.trim() || editingField.value.name;
    const updated = newField.value.trim();

    if (updated === original || !updated) return true;

    try {
        const response = await workspaceService.updateDisplayName({
            workspace_id: route.params.id,
            table_id: route.params.tid,
            view_id: route.params.fid,
            header_name: editingField.value.name,
            display_name: updated,
            visible: editingField.value.visible,
        });

        emit("update-header", editingField.value.name, {
            display_name: response.data.display_name,
        });

        return true;
    } catch (error) {
        useAlertStore().showError(extractErrorMessage(error));

        return false;
    }
}

async function saveFieldFormula() {
    const formulaSpec = getFormulaPayload();

    if (!formulaSpec || !formulaSpec.name) {
        useAlertStore().showError(t.value("projects.grid_view.pick_formula_alert"));

        return;
    }

    try {
        await workspaceService.updateFieldFormula({
            workspace_id: route.params.id,
            table_id: route.params.tid,
            field_name: editingField.value.name,
            formula: formulaSpec,
        });

        // Patch local header so cell values recompute
        emit("update-header", editingField.value.name, { formula: formulaSpec });
    } catch (error) {
        useAlertStore().showError(extractErrorMessage(error));

        return;
    }

    await saveDisplayNameIfChanged();
    popoverVisible.value = false;
}

async function saveFieldEdit() {
    const ok = await saveDisplayNameIfChanged();

    if (ok) popoverVisible.value = false;
}

onBeforeUnmount(() => {
    document.removeEventListener("mousedown", handleClickOutside);
});

defineExpose({ toggle: togglePopover, edit: handleEditField });
</script>

<style scoped>
.popover {
    position: absolute;
}
</style>
