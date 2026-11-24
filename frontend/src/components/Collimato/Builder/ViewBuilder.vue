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
                            class="flex max-h-[85vh] w-full max-w-3xl flex-col rounded-lg bg-white shadow-xl sm:min-h-[32rem]"
                        >
                            <div class="border-b border-gray-200 px-6 py-4">
                                <DialogTitle class="text-base font-semibold text-gray-900">
                                    {{
                                        isEdit
                                            ? t("collimato.view_builder.title_edit")
                                            : t("collimato.view_builder.title_new")
                                    }}
                                </DialogTitle>
                                <p class="mt-1 text-sm text-gray-500">
                                    {{ t("collimato.view_builder.subtitle") }}
                                </p>
                            </div>

                            <div class="flex-1 space-y-6 overflow-y-auto px-6 py-5">
                                <p
                                    v-if="!datasetNames.length"
                                    class="rounded-md bg-yellow-50 px-3 py-2 text-sm text-yellow-800"
                                >
                                    {{ t("collimato.view_builder.no_datasets") }}
                                </p>

                                <template v-else>
                                    <!-- View name + base dataset -->
                                    <div class="grid grid-cols-1 gap-4 sm:grid-cols-2">
                                        <div>
                                            <label :class="labelClass">{{
                                                t("collimato.view_builder.view_name")
                                            }}</label>
                                            <input
                                                v-model="viewName"
                                                type="text"
                                                :disabled="isEdit"
                                                :class="[
                                                    inputClass,
                                                    nameError ? 'ring-red-600' : '',
                                                    isEdit ? 'bg-gray-50 text-gray-400' : '',
                                                ]"
                                            />
                                            <p v-if="nameError" class="mt-1 text-sm text-red-600">
                                                {{ t("collimato.view_builder.name_error") }}
                                            </p>
                                        </div>
                                        <div>
                                            <label :class="labelClass">{{
                                                t("collimato.view_builder.base_dataset")
                                            }}</label>
                                            <div class="mt-1">
                                                <BaseSelect
                                                    v-model="baseDataset"
                                                    :options="datasetNames"
                                                    :disabled="isEdit"
                                                    :placeholder="
                                                        t('collimato.view_builder.base_dataset')
                                                    "
                                                    @update:modelValue="onBaseChange"
                                                />
                                            </div>
                                        </div>
                                    </div>

                                    <!-- Base dataset members -->
                                    <div v-if="baseDataset">
                                        <h3 class="text-sm font-semibold text-gray-900">
                                            {{
                                                t("collimato.view_builder.include_from", {
                                                    dataset: baseDataset,
                                                })
                                            }}
                                        </h3>
                                        <p class="mb-2 text-xs text-gray-500">
                                            {{ t("collimato.view_builder.include_help") }}
                                        </p>
                                        <MemberPicker
                                            :members="memberOptions(baseDataset)"
                                            :selected="baseIncludes"
                                            @toggle="toggleInclude(baseIncludes, $event)"
                                            @all="baseIncludes = [...memberOptions(baseDataset)]"
                                            @clear="baseIncludes = []"
                                        />
                                    </div>

                                    <!-- Joined datasets -->
                                    <div v-if="baseDataset">
                                        <div class="flex items-center justify-between">
                                            <h3 class="text-sm font-semibold text-gray-900">
                                                {{ t("collimato.view_builder.joined_datasets") }}
                                            </h3>
                                            <button
                                                type="button"
                                                class="text-sm font-medium text-indigo-600 hover:text-indigo-500 disabled:text-gray-300"
                                                :disabled="!canAddDataset"
                                                @click="addJoinedDataset"
                                            >
                                                +
                                                {{ t("collimato.view_builder.add_dataset") }}
                                            </button>
                                        </div>
                                        <p class="mb-2 text-xs text-gray-500">
                                            {{ t("collimato.view_builder.joined_help") }}
                                        </p>

                                        <p
                                            v-if="!joinableDatasets.length"
                                            class="text-sm text-gray-400"
                                        >
                                            {{ t("collimato.view_builder.no_joinable") }}
                                        </p>

                                        <div
                                            v-for="(jc, i) in joinedDatasets"
                                            :key="i"
                                            class="mb-2 space-y-2 rounded-md p-3 ring-1 ring-inset ring-gray-200"
                                        >
                                            <div class="flex items-center gap-3">
                                                <div class="w-52">
                                                    <BaseSelect
                                                        v-model="jc.dataset"
                                                        :options="optionsForRow(jc)"
                                                        :placeholder="
                                                            t('collimato.view_builder.dataset')
                                                        "
                                                        @update:modelValue="
                                                            onJoinedDatasetChange(jc)
                                                        "
                                                    />
                                                </div>
                                                <span
                                                    v-if="jc.dataset"
                                                    class="rounded bg-gray-100 px-1.5 py-0.5 font-mono text-xs text-gray-500"
                                                    >{{ baseDataset }}.{{ jc.dataset }}</span
                                                >
                                                <label
                                                    class="flex items-center gap-1.5 text-sm text-gray-600"
                                                    :title="t('collimato.view_builder.prefix_help')"
                                                >
                                                    <input
                                                        v-model="jc.prefix"
                                                        type="checkbox"
                                                        class="h-4 w-4 rounded border-gray-300 text-indigo-600 focus:ring-indigo-600"
                                                    />
                                                    {{ t("collimato.view_builder.prefix") }}
                                                </label>
                                                <button
                                                    type="button"
                                                    class="ml-auto text-gray-400 hover:text-red-600"
                                                    @click="joinedDatasets.splice(i, 1)"
                                                >
                                                    <XMarkIcon class="h-5 w-5" />
                                                </button>
                                            </div>
                                            <MemberPicker
                                                v-if="jc.dataset"
                                                :members="memberOptions(jc.dataset)"
                                                :selected="jc.includes"
                                                @toggle="toggleInclude(jc.includes, $event)"
                                                @all="jc.includes = [...memberOptions(jc.dataset)]"
                                                @clear="jc.includes = []"
                                            />
                                        </div>
                                    </div>
                                </template>
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
import BaseSelect from "@/components/BaseSelect.vue";
import MemberPicker from "@/components/Collimato/Builder/MemberPicker.vue";
import collimatoService from "@/services/collimatoService";
import { useAlertStore } from "@/store/alerts";
import { extractErrorMessage } from "@/utils/errors";
import { t } from "@/i18n/index.js";

const props = defineProps({
    show: { type: Boolean, default: false },
    workspaceId: { type: String, required: true },
    // Datasets available as view members, each { name, model } (model may be null
    // for hand-written datasets, in which case all members are included via "*").
    existingDatasets: { type: Array, default: () => [] },
    editFile: { type: Object, default: null },
});
const emit = defineEmits(["close", "saved"]);

const alertStore = useAlertStore();

const labelClass = "block text-sm font-medium text-gray-700";
const inputClass =
    "mt-1 block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm";

const viewName = ref("");
const baseDataset = ref("");
const baseIncludes = ref([]);
const joinedDatasets = ref([]); // [{ dataset, includes: [], prefix }]
const saving = ref(false);

const isEdit = computed(() => !!props.editFile);
const datasetNames = computed(() => props.existingDatasets.map((c) => c.name));

// memberOptions returns a dataset's selectable members (dimension + measure names)
// from its stored model, or null when the dataset was hand-written (unknown members).
function memberOptions(name) {
    const dataset = props.existingDatasets.find((c) => c.name === name);

    if (!dataset || !dataset.model) return null;
    const dims = (dataset.model.dimensions || []).map((d) => d.name);
    const meas = (dataset.model.measures || []).map((m) => m.name);

    return [...dims, ...meas];
}

// joinableDatasets are datasets directly joined to the base (in either direction), so a
// one-hop join path base.<dataset> resolves.
const joinableDatasets = computed(() => {
    if (!baseDataset.value) return [];
    const set = new Set();
    const base = props.existingDatasets.find((c) => c.name === baseDataset.value);

    (base?.model?.joins || []).forEach((j) => set.add(j.name));
    props.existingDatasets.forEach((c) => {
        if (c.name === baseDataset.value) return;
        if ((c.model?.joins || []).some((j) => j.name === baseDataset.value)) {
            set.add(c.name);
        }
    });

    return [...set];
});

// A dataset can only be joined into the view once; offer each row the datasets not
// already taken by another row (keeping the row's own current pick).
const usedDatasets = computed(
    () => new Set(joinedDatasets.value.map((j) => j.dataset).filter(Boolean)),
);

function optionsForRow(jc) {
    return joinableDatasets.value.filter((n) => n === jc.dataset || !usedDatasets.value.has(n));
}

const canAddDataset = computed(() =>
    joinableDatasets.value.some((n) => !usedDatasets.value.has(n)),
);

const nameError = computed(
    () => viewName.value !== "" && !/^[A-Za-z_][A-Za-z0-9_]*$/.test(viewName.value),
);

const canSave = computed(
    () =>
        !!viewName.value &&
        !nameError.value &&
        !!baseDataset.value &&
        (baseIncludes.value.length > 0 ||
            memberOptions(baseDataset.value) === null ||
            joinedDatasets.value.some(
                (jc) =>
                    jc.dataset && (jc.includes.length > 0 || memberOptions(jc.dataset) === null),
            )),
);

function toggleInclude(arr, member) {
    const i = arr.indexOf(member);

    if (i >= 0) arr.splice(i, 1);
    else arr.push(member);
}

function onBaseChange() {
    baseIncludes.value = [];
    joinedDatasets.value = [];
}

function onJoinedDatasetChange(jc) {
    jc.includes = [];
}

function addJoinedDataset() {
    joinedDatasets.value.push({ dataset: "", includes: [], prefix: false });
}

function reset() {
    viewName.value = "";
    baseDataset.value = "";
    baseIncludes.value = [];
    joinedDatasets.value = [];
    saving.value = false;
}

function expandIncludes(datasetName, includes) {
    if (Array.isArray(includes) && includes.length === 1 && includes[0] === "*") {
        return memberOptions(datasetName) || ["*"];
    }

    return includes || [];
}

function hydrateFromModel(model) {
    viewName.value = model.name;
    const refs = model.parts || [];
    const baseRef = refs.find((r) => !r.join_path.includes("."));

    baseDataset.value = baseRef ? baseRef.join_path : refs[0]?.join_path.split(".")[0] || "";
    baseIncludes.value = baseRef ? expandIncludes(baseDataset.value, baseRef.includes) : [];
    joinedDatasets.value = refs
        .filter((r) => r.join_path.includes("."))
        .map((r) => {
            const seg = r.join_path.split(".");
            const dataset = seg[seg.length - 1];

            return {
                dataset,
                includes: expandIncludes(dataset, r.includes),
                prefix: !!r.prefix,
            };
        });
}

async function save() {
    if (!canSave.value) return;
    saving.value = true;

    // An unknown (hand-written) dataset has no member list, so it is included as "*".
    const refFor = (joinPath, datasetName, includes, prefix) => {
        const inc = memberOptions(datasetName) === null ? ["*"] : includes;

        return {
            join_path: joinPath,
            includes: inc,
            prefix: prefix || undefined,
        };
    };

    const parts = [];

    if (baseIncludes.value.length > 0 || memberOptions(baseDataset.value) === null) {
        parts.push(refFor(baseDataset.value, baseDataset.value, baseIncludes.value, false));
    }

    joinedDatasets.value
        .filter(
            (jc) => jc.dataset && (jc.includes.length > 0 || memberOptions(jc.dataset) === null),
        )
        .forEach((jc) => {
            parts.push(
                refFor(`${baseDataset.value}.${jc.dataset}`, jc.dataset, jc.includes, jc.prefix),
            );
        });

    try {
        const res = await collimatoService.buildView(props.workspaceId, {
            model: { name: viewName.value, parts },
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
        if (open) {
            reset();
            if (isEdit.value && props.editFile.builder_model) {
                hydrateFromModel(JSON.parse(props.editFile.builder_model));
            }
        }
    },
);
</script>
