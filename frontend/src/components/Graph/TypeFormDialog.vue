<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <FormDialog
        :open="open"
        :title="editing ? t('graph.form.type.edit_title') : t('graph.form.type.new_title')"
        :confirm-label="editing ? t('common.button.save') : t('common.button.create')"
        :confirm-disabled="!canSave"
        :loading="saving"
        width="sm:max-w-2xl"
        scrollable
        @confirm="save"
        @close="emit('close')"
    >
        <div class="grid grid-cols-1 gap-x-4 gap-y-5 sm:grid-cols-2">
            <div>
                <label for="type-name" :class="labelClass">{{ t("graph.form.type.name") }}</label>
                <input
                    id="type-name"
                    v-model="form.name"
                    type="text"
                    :class="inputClass"
                    @input="followName"
                />
            </div>
            <div>
                <label for="type-plural" :class="labelClass">{{
                    t("graph.form.type.plural")
                }}</label>
                <input
                    id="type-plural"
                    v-model="form.plural_name"
                    type="text"
                    :placeholder="t('graph.form.optional')"
                    :class="inputClass"
                />
            </div>

            <div>
                <label for="type-key" :class="labelClass">{{ t("graph.form.key") }}</label>
                <input
                    id="type-key"
                    v-model="form.key"
                    type="text"
                    :disabled="editing"
                    :class="[inputClass, 'font-mono']"
                    @input="keyEdited = true"
                />
                <p class="mt-1 text-xs text-gray-500">{{ t("graph.form.key_help") }}</p>
            </div>
            <BaseSelect
                v-model="form.status"
                :label="t('graph.form.type.status')"
                :options="statusOptions"
            />

            <div class="sm:col-span-2">
                <label for="type-description" :class="labelClass">
                    {{ t("graph.form.description") }}
                </label>
                <textarea
                    id="type-description"
                    v-model="form.description"
                    rows="2"
                    :placeholder="t('graph.form.optional')"
                    :class="inputClass"
                ></textarea>
            </div>

            <div class="sm:col-span-2">
                <p id="type-icon-label" :class="labelClass">{{ t("graph.form.type.icon") }}</p>
                <div
                    class="mt-2 flex flex-wrap gap-2"
                    role="radiogroup"
                    aria-labelledby="type-icon-label"
                >
                    <button
                        v-for="(icon, name) in TYPE_ICONS"
                        :key="name"
                        type="button"
                        role="radio"
                        :aria-checked="form.icon === name"
                        :title="name"
                        :class="[
                            form.icon === name
                                ? 'bg-indigo-50 text-indigo-600 ring-2 ring-indigo-600'
                                : 'bg-white text-gray-500 ring-1 ring-gray-300 hover:bg-gray-50',
                            'flex size-9 items-center justify-center rounded-md',
                        ]"
                        @click="form.icon = name"
                    >
                        <component :is="icon" class="size-5" aria-hidden="true" />
                        <span class="sr-only">{{ name }}</span>
                    </button>
                </div>
            </div>

            <div class="sm:col-span-2">
                <p id="type-color-label" :class="labelClass">{{ t("graph.form.type.color") }}</p>
                <div
                    class="mt-2 flex flex-wrap items-center gap-2"
                    role="radiogroup"
                    aria-labelledby="type-color-label"
                >
                    <button
                        type="button"
                        role="radio"
                        :aria-checked="form.color === ''"
                        :class="[
                            form.color === '' ? 'ring-2 ring-indigo-600' : 'ring-1 ring-gray-300',
                            'rounded-full px-3 py-1 text-xs font-medium text-gray-700',
                        ]"
                        @click="form.color = ''"
                    >
                        {{ t("graph.form.type.color_auto") }}
                    </button>
                    <button
                        v-for="(style, name) in COLORS"
                        :key="name"
                        type="button"
                        role="radio"
                        :aria-checked="form.color === name"
                        :title="name"
                        :class="[
                            form.color === name ? 'ring-2 ring-offset-2 ring-gray-900' : '',
                            'size-7 rounded-full',
                        ]"
                        :style="{ background: style.color }"
                        @click="form.color = name"
                    >
                        <span class="sr-only">{{ name }}</span>
                    </button>
                </div>
            </div>

            <div class="sm:col-span-2 border-t border-gray-100 pt-5">
                <p class="text-sm font-semibold text-gray-900">
                    {{ t("graph.form.type.source_title") }}
                </p>
                <p class="mt-1 text-xs text-gray-500">{{ t("graph.form.type.source_help") }}</p>
            </div>

            <div class="sm:col-span-2">
                <BaseSelect
                    v-model="form.source_id"
                    :label="t('graph.form.type.dataset')"
                    :options="datasetOptions"
                    :placeholder="loadingChoices ? t('graph.form.loading') : t('graph.form.choose')"
                    @update:model-value="datasetChanged"
                />
            </div>
            <BaseSelect
                v-model="form.key_property"
                :label="t('graph.form.type.key_field')"
                :options="fieldOptions"
                :disabled="!form.source_id"
                :placeholder="t('graph.form.choose')"
            />
            <BaseSelect
                v-model="form.title_property"
                :label="t('graph.form.type.name_field')"
                :options="fieldOptions"
                :disabled="!form.source_id"
                :placeholder="t('graph.form.choose')"
            />

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
import { COLORS } from "@/utils/graph/palette.js";
import { TYPE_ICONS } from "@/utils/graph/icons.js";
import { suggestKey, fallbackKey, KEY_PATTERN } from "@/utils/graph/keys.js";
import FormDialog from "@/components/FormDialog.vue";
import BaseSelect from "@/components/BaseSelect.vue";

const props = defineProps({
    open: {
        type: Boolean,
        default: false,
    },
    // The type to edit; empty for a new one.
    entityType: {
        type: Object,
        default: null,
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
const choices = datasetChoices.choices;
const loadingChoices = datasetChoices.loading;

const editing = computed(() => Boolean(props.entityType));

const statusOptions = computed(() => [
    { value: "experimental", label: t.value("graph.ontology.status.experimental") },
    { value: "active", label: t.value("graph.ontology.status.active") },
    { value: "deprecated", label: t.value("graph.ontology.status.deprecated") },
]);

const datasetOptions = computed(() => {
    return choices.value.map((choice) => {
        return { value: choice.id, label: `${choice.workspace_name} · ${choice.name}` };
    });
});

const fieldOptions = computed(() => datasetChoices.fieldOptions(form.value.source_id));

const canSave = computed(() => {
    const value = form.value;

    return Boolean(
        value.name.trim() &&
        KEY_PATTERN.test(value.key) &&
        value.source_id &&
        value.key_property &&
        value.title_property,
    );
});

function emptyForm() {
    return {
        name: "",
        plural_name: "",
        key: "",
        status: "active",
        description: "",
        icon: "",
        color: "",
        source_id: "",
        key_property: "",
        title_property: "",
    };
}

function formFor(entityType) {
    return {
        name: entityType.name,
        plural_name: entityType.plural_name ?? "",
        key: entityType.key,
        status: entityType.status || "active",
        description: entityType.description ?? "",
        icon: entityType.icon ?? "",
        color: entityType.color ?? "",
        source_id: entityType.source_id ?? "",
        key_property: entityType.key_property ?? "",
        title_property: entityType.title_property ?? "",
    };
}

// The key follows the name until the admin types a key of their own.
function followName() {
    if (editing.value || keyEdited.value) {
        return;
    }

    const suggested = suggestKey(form.value.name);

    if (suggested) {
        form.value.key = suggested;

        return;
    }

    if (form.value.name.trim() === "") {
        form.value.key = "";

        return;
    }

    if (!KEY_PATTERN.test(form.value.key)) {
        form.value.key = fallbackKey("type");
    }
}

// A field of another dataset can't be the key or name.
function datasetChanged() {
    form.value.key_property = "";
    form.value.title_property = "";
}

async function save() {
    saving.value = true;
    errorMessage.value = "";

    try {
        let response;

        if (editing.value) {
            response = await graphService.updateEntityType(props.entityType.id, form.value);
        } else {
            response = await graphService.createEntityType(form.value);
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

        if (props.entityType) {
            form.value = formFor(props.entityType);
        } else {
            form.value = emptyForm();
        }

        datasetChoices.load();
    },
    { immediate: true },
);
</script>
