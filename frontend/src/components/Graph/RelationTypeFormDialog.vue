<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <FormDialog
        :open="open"
        :title="editing ? t('graph.form.relation.edit_title') : t('graph.form.relation.new_title')"
        :confirm-label="editing ? t('common.button.save') : t('common.button.create')"
        :confirm-disabled="!canSave"
        :loading="saving"
        width="sm:max-w-2xl"
        scrollable
        @confirm="save"
        @close="emit('close')"
    >
        <div class="grid grid-cols-1 gap-x-4 gap-y-5 sm:grid-cols-2">
            <BaseSelect
                v-model="form.from_type_id"
                :label="t('graph.form.relation.from')"
                :options="typeOptions"
                :disabled="editing"
                :placeholder="t('graph.form.choose')"
                @update:model-value="typesChanged"
            />
            <BaseSelect
                v-model="form.to_type_id"
                :label="t('graph.form.relation.to')"
                :options="typeOptions"
                :disabled="editing"
                :placeholder="t('graph.form.choose')"
                @update:model-value="typesChanged"
            />

            <div>
                <label for="relation-name" :class="labelClass">{{
                    t("graph.form.relation.name")
                }}</label>
                <input
                    id="relation-name"
                    v-model="form.name"
                    type="text"
                    :placeholder="t('graph.form.relation.name_example')"
                    :class="inputClass"
                    @input="followNames"
                />
            </div>
            <div>
                <label for="relation-reverse" :class="labelClass">
                    {{ t("graph.form.relation.reverse_name") }}
                </label>
                <input
                    id="relation-reverse"
                    v-model="form.reverse_name"
                    type="text"
                    :placeholder="t('graph.form.relation.reverse_example')"
                    :class="inputClass"
                />
            </div>

            <div>
                <label for="relation-key" :class="labelClass">{{ t("graph.form.key") }}</label>
                <input
                    id="relation-key"
                    v-model="form.key"
                    type="text"
                    :disabled="editing"
                    :class="[inputClass, 'font-mono']"
                    @input="keyEdited = true"
                />
                <p class="mt-1 text-xs text-gray-500">{{ t("graph.form.key_help") }}</p>
            </div>
            <BaseSelect
                v-model="form.cardinality"
                :label="t('graph.ontology.column.cardinality')"
                :options="cardinalityOptions"
            />

            <div class="sm:col-span-2 border-t border-gray-100 pt-5">
                <p class="text-sm font-semibold text-gray-900">
                    {{ t("graph.form.relation.source_title") }}
                </p>
                <p class="mt-1 text-xs text-gray-500">{{ t("graph.form.relation.source_help") }}</p>
            </div>

            <BaseSelect
                v-model="form.property"
                :label="propertyLabel"
                :options="toFieldOptions"
                :disabled="!form.to_type_id"
                :placeholder="loadingChoices ? t('graph.form.loading') : t('graph.form.choose')"
            />
            <BaseSelect
                v-model="form.from_property"
                :label="fromPropertyLabel"
                :options="fromFieldOptions"
                :disabled="!form.from_type_id"
            />

            <div class="sm:col-span-2">
                <label for="relation-description" :class="labelClass">
                    {{ t("graph.form.description") }}
                </label>
                <textarea
                    id="relation-description"
                    v-model="form.description"
                    rows="2"
                    :placeholder="t('graph.form.optional')"
                    :class="inputClass"
                ></textarea>
            </div>

            <div
                v-if="preview"
                class="sm:col-span-2 rounded-md bg-gray-50 p-3 text-sm text-gray-700"
            >
                <span class="font-medium">{{ preview.from }}</span>
                {{ " " }}
                <span class="text-indigo-600">{{ preview.name }}</span>
                {{ " " }}
                <span class="font-medium">{{ preview.to }}</span>
                <span v-if="preview.field" class="text-gray-500">
                    · {{ t("graph.form.relation.read_from", { field: preview.field }) }}
                </span>
            </div>

            <p
                v-if="errorMessage"
                ref="errorBox"
                class="sm:col-span-2 rounded-md bg-red-50 p-3 text-sm text-red-700"
            >
                {{ errorMessage }}
            </p>
        </div>
    </FormDialog>
</template>

<script setup>
import { ref, computed, watch, nextTick } from "vue";
import { t } from "@/i18n/index.js";
import graphService from "@/services/graphService";
import { extractErrorMessage } from "@/utils/errors";
import { useDatasetChoices } from "@/composables/graph/useDatasetChoices.js";
import { suggestKey, KEY_PATTERN } from "@/utils/graph/keys.js";
import FormDialog from "@/components/FormDialog.vue";
import BaseSelect from "@/components/BaseSelect.vue";

const props = defineProps({
    open: {
        type: Boolean,
        default: false,
    },
    // The relation type to edit; empty for a new one.
    relationType: {
        type: Object,
        default: null,
    },
    entityTypes: {
        type: Array,
        required: true,
    },
    // A new relation type from a type's own page starts from that type.
    fromTypeId: {
        type: String,
        default: "",
    },
});

const emit = defineEmits(["saved", "close"]);

const labelClass = "block text-sm font-medium leading-6 text-gray-900";
const inputClass =
    "mt-2 block w-full rounded-md border-0 py-1.5 text-sm text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 disabled:bg-gray-50 disabled:text-gray-500";

const form = ref(emptyForm());
const saving = ref(false);
const errorMessage = ref("");
const errorBox = ref(null);
const keyEdited = ref(false);

const datasetChoices = useDatasetChoices(errorMessage);
const loadingChoices = datasetChoices.loading;

const editing = computed(() => Boolean(props.relationType));

// Both ends must read their entities from a dataset, as the server checks.
const datasetTypes = computed(() => {
    return props.entityTypes.filter((entityType) => entityType.source === "dataset");
});

const typeOptions = computed(() => {
    return datasetTypes.value.map((entityType) => {
        return { value: entityType.id, label: entityType.name };
    });
});

const cardinalityOptions = computed(() => {
    return ["one_to_one", "one_to_many", "many_to_one", "many_to_many"].map((value) => {
        return { value: value, label: t.value(`graph.ontology.cardinality.${value}`) };
    });
});

function typeByID(id) {
    return props.entityTypes.find((entityType) => entityType.id === id);
}

function fieldOptionsOf(typeID) {
    const entityType = typeByID(typeID);

    if (!entityType) {
        return [];
    }

    return datasetChoices.fieldOptions(entityType.source_id);
}

const toFieldOptions = computed(() => fieldOptionsOf(form.value.to_type_id));

// Empty matches on the "from" type's key, which is what most relations do.
const fromFieldOptions = computed(() => {
    const keyOption = { value: "", label: t.value("graph.form.relation.match_key") };

    return [keyOption, ...fieldOptionsOf(form.value.from_type_id)];
});

const propertyLabel = computed(() => {
    const to = typeByID(form.value.to_type_id);

    if (!to) {
        return t.value("graph.form.relation.property");
    }

    return t.value("graph.form.relation.property_of", { type: to.name });
});

const fromPropertyLabel = computed(() => {
    const from = typeByID(form.value.from_type_id);

    if (!from) {
        return t.value("graph.form.relation.match_on");
    }

    return t.value("graph.form.relation.match_on_of", { type: from.name });
});

const preview = computed(() => {
    const from = typeByID(form.value.from_type_id);
    const to = typeByID(form.value.to_type_id);

    if (!from || !to || !form.value.name.trim()) {
        return null;
    }

    const field = toFieldOptions.value.find((option) => option.value === form.value.property);

    return {
        from: from.name,
        to: to.name,
        name: form.value.name.trim(),
        field: field ? field.name : "",
    };
});

const canSave = computed(() => {
    const value = form.value;

    return Boolean(
        value.from_type_id &&
        value.to_type_id &&
        value.name.trim() &&
        value.reverse_name.trim() &&
        KEY_PATTERN.test(value.key) &&
        value.property,
    );
});

function emptyForm() {
    return {
        from_type_id: props.fromTypeId,
        to_type_id: "",
        name: "",
        reverse_name: "",
        key: "",
        cardinality: "one_to_many",
        property: "",
        from_property: "",
        description: "",
    };
}

function formFor(relationType) {
    return {
        from_type_id: relationType.from_type_id,
        to_type_id: relationType.to_type_id,
        name: relationType.name,
        reverse_name: relationType.reverse_name,
        key: relationType.key,
        cardinality: relationType.cardinality || "one_to_many",
        property: relationType.kind_config?.property ?? "",
        from_property: relationType.kind_config?.from_property ?? "",
        description: relationType.description ?? "",
    };
}

// The key follows the two types and the name until the admin types a key of
// their own: "company_places_order".
function followNames() {
    if (editing.value || keyEdited.value) {
        return;
    }

    const from = typeByID(form.value.from_type_id);
    const to = typeByID(form.value.to_type_id);

    form.value.key = suggestKey(from?.key ?? "", form.value.name, to?.key ?? "");
}

// The fields belong to the chosen types' datasets, so changing a type clears
// them.
function typesChanged() {
    form.value.property = "";
    form.value.from_property = "";

    followNames();
}

async function save() {
    saving.value = true;
    errorMessage.value = "";

    try {
        let response;

        if (editing.value) {
            response = await graphService.updateRelationType(props.relationType.id, form.value);
        } else {
            response = await graphService.createRelationType(form.value);
        }

        emit("saved", response.data);
    } catch (e) {
        errorMessage.value = extractErrorMessage(e);

        await showError();
    } finally {
        saving.value = false;
    }
}

// The form scrolls inside the dialog, so a save error at its end is brought
// into view.
async function showError() {
    await nextTick();

    errorBox.value?.scrollIntoView?.({ block: "nearest" });
}

watch(
    () => props.open,
    (open) => {
        if (!open) {
            return;
        }

        errorMessage.value = "";
        keyEdited.value = false;

        if (props.relationType) {
            form.value = formFor(props.relationType);
        } else {
            form.value = emptyForm();
        }

        datasetChoices.load();
    },
    { immediate: true },
);
</script>
