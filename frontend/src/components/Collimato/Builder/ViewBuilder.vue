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
                                    v-if="!cubeNames.length"
                                    class="rounded-md bg-yellow-50 px-3 py-2 text-sm text-yellow-800"
                                >
                                    {{ t("collimato.view_builder.no_cubes") }}
                                </p>

                                <template v-else>
                                    <!-- View name + base cube -->
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
                                                t("collimato.view_builder.base_cube")
                                            }}</label>
                                            <div class="mt-1">
                                                <BaseSelect
                                                    v-model="baseCube"
                                                    :options="cubeNames"
                                                    :disabled="isEdit"
                                                    :placeholder="
                                                        t('collimato.view_builder.base_cube')
                                                    "
                                                    @update:modelValue="onBaseChange"
                                                />
                                            </div>
                                        </div>
                                    </div>

                                    <!-- Base cube members -->
                                    <div v-if="baseCube">
                                        <h3 class="text-sm font-semibold text-gray-900">
                                            {{
                                                t("collimato.view_builder.include_from", {
                                                    cube: baseCube,
                                                })
                                            }}
                                        </h3>
                                        <p class="mb-2 text-xs text-gray-500">
                                            {{ t("collimato.view_builder.include_help") }}
                                        </p>
                                        <MemberPicker
                                            :members="memberOptions(baseCube)"
                                            :selected="baseIncludes"
                                            @toggle="toggleInclude(baseIncludes, $event)"
                                            @all="baseIncludes = [...memberOptions(baseCube)]"
                                            @clear="baseIncludes = []"
                                        />
                                    </div>

                                    <!-- Joined cubes -->
                                    <div v-if="baseCube">
                                        <div class="flex items-center justify-between">
                                            <h3 class="text-sm font-semibold text-gray-900">
                                                {{ t("collimato.view_builder.joined_cubes") }}
                                            </h3>
                                            <button
                                                type="button"
                                                class="text-sm font-medium text-indigo-600 hover:text-indigo-500 disabled:text-gray-300"
                                                :disabled="!canAddCube"
                                                @click="addJoinedCube"
                                            >
                                                +
                                                {{ t("collimato.view_builder.add_cube") }}
                                            </button>
                                        </div>
                                        <p class="mb-2 text-xs text-gray-500">
                                            {{ t("collimato.view_builder.joined_help") }}
                                        </p>

                                        <p
                                            v-if="!joinableCubes.length"
                                            class="text-sm text-gray-400"
                                        >
                                            {{ t("collimato.view_builder.no_joinable") }}
                                        </p>

                                        <div
                                            v-for="(jc, i) in joinedCubes"
                                            :key="i"
                                            class="mb-2 space-y-2 rounded-md p-3 ring-1 ring-inset ring-gray-200"
                                        >
                                            <div class="flex items-center gap-3">
                                                <div class="w-52">
                                                    <BaseSelect
                                                        v-model="jc.cube"
                                                        :options="optionsForRow(jc)"
                                                        :placeholder="
                                                            t('collimato.view_builder.cube')
                                                        "
                                                        @update:modelValue="onJoinedCubeChange(jc)"
                                                    />
                                                </div>
                                                <span
                                                    v-if="jc.cube"
                                                    class="rounded bg-gray-100 px-1.5 py-0.5 font-mono text-xs text-gray-500"
                                                    >{{ baseCube }}.{{ jc.cube }}</span
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
                                                    @click="joinedCubes.splice(i, 1)"
                                                >
                                                    <XMarkIcon class="h-5 w-5" />
                                                </button>
                                            </div>
                                            <MemberPicker
                                                v-if="jc.cube"
                                                :members="memberOptions(jc.cube)"
                                                :selected="jc.includes"
                                                @toggle="toggleInclude(jc.includes, $event)"
                                                @all="jc.includes = [...memberOptions(jc.cube)]"
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
    // Cubes available as view members, each { name, model } (model may be null
    // for hand-written cubes, in which case all members are included via "*").
    existingCubes: { type: Array, default: () => [] },
    editFile: { type: Object, default: null },
});
const emit = defineEmits(["close", "saved"]);

const alertStore = useAlertStore();

const labelClass = "block text-sm font-medium text-gray-700";
const inputClass =
    "mt-1 block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm";

const viewName = ref("");
const baseCube = ref("");
const baseIncludes = ref([]);
const joinedCubes = ref([]); // [{ cube, includes: [], prefix }]
const saving = ref(false);

const isEdit = computed(() => !!props.editFile);
const cubeNames = computed(() => props.existingCubes.map((c) => c.name));

// memberOptions returns a cube's selectable members (dimension + measure names)
// from its stored model, or null when the cube was hand-written (unknown members).
function memberOptions(name) {
    const cube = props.existingCubes.find((c) => c.name === name);

    if (!cube || !cube.model) return null;
    const dims = (cube.model.dimensions || []).map((d) => d.name);
    const meas = (cube.model.measures || []).map((m) => m.name);

    return [...dims, ...meas];
}

// joinableCubes are cubes directly joined to the base (in either direction), so a
// one-hop join path base.<cube> resolves.
const joinableCubes = computed(() => {
    if (!baseCube.value) return [];
    const set = new Set();
    const base = props.existingCubes.find((c) => c.name === baseCube.value);

    (base?.model?.joins || []).forEach((j) => set.add(j.name));
    props.existingCubes.forEach((c) => {
        if (c.name === baseCube.value) return;
        if ((c.model?.joins || []).some((j) => j.name === baseCube.value)) {
            set.add(c.name);
        }
    });

    return [...set];
});

// A cube can only be joined into the view once; offer each row the cubes not
// already taken by another row (keeping the row's own current pick).
const usedCubes = computed(() => new Set(joinedCubes.value.map((j) => j.cube).filter(Boolean)));

function optionsForRow(jc) {
    return joinableCubes.value.filter((n) => n === jc.cube || !usedCubes.value.has(n));
}

const canAddCube = computed(() => joinableCubes.value.some((n) => !usedCubes.value.has(n)));

const nameError = computed(
    () => viewName.value !== "" && !/^[A-Za-z_][A-Za-z0-9_]*$/.test(viewName.value),
);

const canSave = computed(
    () =>
        !!viewName.value &&
        !nameError.value &&
        !!baseCube.value &&
        (baseIncludes.value.length > 0 ||
            memberOptions(baseCube.value) === null ||
            joinedCubes.value.some(
                (jc) => jc.cube && (jc.includes.length > 0 || memberOptions(jc.cube) === null),
            )),
);

function toggleInclude(arr, member) {
    const i = arr.indexOf(member);

    if (i >= 0) arr.splice(i, 1);
    else arr.push(member);
}

function onBaseChange() {
    baseIncludes.value = [];
    joinedCubes.value = [];
}

function onJoinedCubeChange(jc) {
    jc.includes = [];
}

function addJoinedCube() {
    joinedCubes.value.push({ cube: "", includes: [], prefix: false });
}

function reset() {
    viewName.value = "";
    baseCube.value = "";
    baseIncludes.value = [];
    joinedCubes.value = [];
    saving.value = false;
}

function expandIncludes(cubeName, includes) {
    if (Array.isArray(includes) && includes.length === 1 && includes[0] === "*") {
        return memberOptions(cubeName) || ["*"];
    }

    return includes || [];
}

function hydrateFromModel(model) {
    viewName.value = model.name;
    const refs = model.cubes || [];
    const baseRef = refs.find((r) => !r.join_path.includes("."));

    baseCube.value = baseRef ? baseRef.join_path : refs[0]?.join_path.split(".")[0] || "";
    baseIncludes.value = baseRef ? expandIncludes(baseCube.value, baseRef.includes) : [];
    joinedCubes.value = refs
        .filter((r) => r.join_path.includes("."))
        .map((r) => {
            const seg = r.join_path.split(".");
            const cube = seg[seg.length - 1];

            return {
                cube,
                includes: expandIncludes(cube, r.includes),
                prefix: !!r.prefix,
            };
        });
}

async function save() {
    if (!canSave.value) return;
    saving.value = true;

    // An unknown (hand-written) cube has no member list, so it is included as "*".
    const refFor = (joinPath, cubeName, includes, prefix) => {
        const inc = memberOptions(cubeName) === null ? ["*"] : includes;

        return {
            join_path: joinPath,
            includes: inc,
            prefix: prefix || undefined,
        };
    };

    const cubes = [];

    if (baseIncludes.value.length > 0 || memberOptions(baseCube.value) === null) {
        cubes.push(refFor(baseCube.value, baseCube.value, baseIncludes.value, false));
    }

    joinedCubes.value
        .filter((jc) => jc.cube && (jc.includes.length > 0 || memberOptions(jc.cube) === null))
        .forEach((jc) => {
            cubes.push(refFor(`${baseCube.value}.${jc.cube}`, jc.cube, jc.includes, jc.prefix));
        });

    try {
        const res = await collimatoService.buildView(props.workspaceId, {
            model: { name: viewName.value, cubes },
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
