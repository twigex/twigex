<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div>
        <div>
            <label class="block text-sm font-medium text-gray-900">{{
                t("projects.grid_view.formula")
            }}</label>
            <Listbox v-model="chosenFormula" as="div">
                <div class="relative mt-2">
                    <ListboxButton
                        class="relative w-full cursor-default rounded-md bg-white py-1.5 pl-3 pr-10 text-left text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 focus:outline-none focus:ring-2 focus:ring-indigo-600 sm:text-sm"
                    >
                        <span class="block truncate">
                            {{ chosenFormula?.label || t("projects.grid_view.select_a_formula") }}
                        </span>
                    </ListboxButton>

                    <ListboxOptions
                        class="absolute z-10 mt-1 max-h-60 w-full overflow-auto rounded-md bg-white py-1 text-base shadow-lg ring-1 ring-black ring-opacity-5 focus:outline-none sm:text-sm"
                    >
                        <ListboxOption
                            v-for="f in POPULAR_FORMULAS"
                            :key="f.name"
                            :value="f"
                            as="template"
                            v-slot="{ active, selected: isSelected }"
                        >
                            <li
                                :class="[
                                    active ? 'bg-indigo-600 text-white' : 'text-gray-900',
                                    'cursor-default select-none py-2 pl-3 pr-3',
                                ]"
                            >
                                <div class="flex flex-col">
                                    <span
                                        :class="[
                                            isSelected ? 'font-semibold' : 'font-normal',
                                            'truncate',
                                        ]"
                                    >
                                        {{ f.label }}
                                    </span>
                                    <span
                                        :class="[
                                            active ? 'text-indigo-100' : 'text-gray-500',
                                            'text-xs',
                                        ]"
                                        >{{ f.desc }}</span
                                    >
                                    <code
                                        :class="[
                                            active ? 'text-indigo-200' : 'text-gray-400',
                                            'text-[11px]',
                                        ]"
                                        >{{ f.sig }}</code
                                    >
                                </div>
                            </li>
                        </ListboxOption>
                    </ListboxOptions>
                </div>
            </Listbox>
        </div>

        <div v-if="chosenFormula">
            <div
                v-if="['SUM', 'AVERAGE', 'PRODUCT', 'MIN', 'MAX'].includes(chosenFormula?.name)"
                class="space-y-2"
            >
                <div class="flex items-center">
                    <input
                        v-model="useMultipleFields"
                        type="checkbox"
                        id="multipleFieldsToggle"
                        class="h-4 w-4 rounded border-gray-300 text-indigo-600 focus:ring-indigo-600"
                    />
                    <label for="multipleFieldsToggle" class="ml-2 block text-sm text-gray-900">
                        {{ t("projects.grid_view.use_multiple_columns") }}
                    </label>
                </div>

                <p class="text-xs text-gray-500">
                    <span v-if="useMultipleFields">
                        {{ t("projects.grid_view.pick") }}
                        <strong>{{ t("projects.grid_view.one_or_more_numeric_columns") }}</strong>
                        {{ t("projects.grid_view.to") }}
                        {{ chosenFormula.name.toLowerCase() }}
                        {{ t("projects.grid_view.formula_decription") }}
                    </span>
                    <span v-else>
                        {{ t("projects.grid_view.the_formula_will") }}
                        {{ chosenFormula.name.toLowerCase() }}
                        {{ t("projects.grid_view.all_values") }}
                        <strong>{{ t("projects.grid_view.within_this_same_column") }}</strong
                        >{{ t("projects.grid_view.no_field_selection_needed") }}
                    </span>
                </p>
            </div>

            <p
                v-if="
                    ['ROUND', 'ROUNDUP', 'ROUNDDOWN'].includes(
                        String(chosenFormula?.name).toUpperCase(),
                    )
                "
                class="text-xs text-gray-500"
            >
                {{ t("projects.grid_view.pick_the") }}
                <strong>{{ t("projects.grid_view.source_numeric_column") }}</strong>
                {{ t("projects.grid_view.to_round_description") }}
            </p>

            <div
                v-if="
                    chosenFormula.arity === 'variadic' &&
                    ['SUM', 'AVERAGE', 'PRODUCT', 'MIN', 'MAX'].includes(chosenFormula?.name) &&
                    useMultipleFields
                "
            >
                <label class="block text-sm font-medium text-gray-900">
                    {{ t("projects.grid_view.columns_to") }}
                    {{ chosenFormula.name.toLowerCase() }}
                    <span class="text-xs text-gray-500 ml-1">{{
                        t("projects.select_one_or_more")
                    }}</span>
                </label>

                <div
                    class="mt-2 max-h-60 overflow-y-auto border border-gray-300 rounded-md bg-white p-2"
                >
                    <div class="space-y-2">
                        <div
                            v-for="h in fieldOptionsForArg({
                                type: chosenFormula.argType,
                            })"
                            :key="h.header_name || h.name"
                            class="flex items-center"
                        >
                            <input
                                :id="`field-${h.header_name || h.name}`"
                                v-model="formulaArgs"
                                :value="h"
                                type="checkbox"
                                class="h-4 w-4 rounded border-gray-300 text-indigo-600 focus:ring-indigo-600"
                            />
                            <label
                                :for="`field-${h.header_name || h.name}`"
                                class="ml-3 flex-1 text-sm text-gray-900 cursor-pointer py-1"
                            >
                                {{ fieldLabel(h) }}
                            </label>
                        </div>

                        <div
                            v-if="
                                fieldOptionsForArg({
                                    type: chosenFormula.argType,
                                }).length === 0
                            "
                            class="text-center py-4 text-gray-500 text-sm"
                        >
                            {{ t("projects.grid_view.no_suitable_columns_found") }}
                        </div>
                    </div>
                </div>

                <div v-if="formulaArgs.length > 0" class="mt-2">
                    <p class="text-xs text-gray-600">
                        {{ t("projects.grid_view.selected") }}
                        {{ formulaArgs.length }}
                        {{ t("projects.grid_view.columns") }}
                    </p>
                </div>
            </div>

            <div v-else-if="chosenFormula.args">
                <div v-for="(arg, i) in chosenFormula.args" :key="i" class="mt-2">
                    <label class="block text-sm font-medium text-gray-900">
                        {{ arg.label }}
                    </label>

                    <div v-if="arg.kind === 'field'">
                        <Listbox v-model="formulaArgs[i]" :by="'header_name'" as="div">
                            <div class="relative mt-2">
                                <ListboxButton
                                    class="relative w-full cursor-default rounded-md bg-white py-1.5 pl-3 pr-10 text-left text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 focus:outline-none focus:ring-2 focus:ring-indigo-600 sm:text-sm"
                                >
                                    <span class="block truncate">
                                        {{
                                            (formulaArgs[i] && fieldLabel(formulaArgs[i])) ||
                                            t("projects.grid_view.pick_a_field")
                                        }}
                                    </span>
                                </ListboxButton>

                                <ListboxOptions
                                    class="absolute z-10 mt-1 max-h-60 w-full overflow-auto rounded-md bg-white py-1 text-base shadow-lg ring-1 ring-black ring-opacity-5 focus:outline-none sm:text-sm"
                                >
                                    <ListboxOption
                                        v-for="h in fieldOptionsForArg(arg)"
                                        :key="h.header_name || h.name"
                                        :value="h"
                                        as="template"
                                        v-slot="{ active, selected: isSelected }"
                                    >
                                        <li
                                            :class="[
                                                active
                                                    ? 'bg-indigo-600 text-white'
                                                    : 'text-gray-900',
                                                'cursor-default select-none py-2 pl-3 pr-3',
                                            ]"
                                        >
                                            <span
                                                :class="[
                                                    isSelected ? 'font-semibold' : 'font-normal',
                                                    'truncate',
                                                ]"
                                            >
                                                {{ fieldLabel(h) }}
                                            </span>
                                        </li>
                                    </ListboxOption>
                                </ListboxOptions>
                            </div>
                        </Listbox>
                    </div>

                    <div v-else-if="arg.kind === 'const'">
                        <input
                            v-model.number="formulaArgs[i]"
                            :type="arg.type === 'number' ? 'number' : 'text'"
                            class="block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 focus:ring-2 focus:ring-indigo-600 sm:text-sm"
                            :placeholder="arg.label"
                            :min="arg.type === 'number' ? 0 : null"
                            step="1"
                        />
                    </div>
                </div>
            </div>

            <div
                v-else-if="
                    chosenFormula.arity === 'variadic' &&
                    ['SUM', 'AVERAGE', 'PRODUCT', 'MIN', 'MAX'].includes(chosenFormula?.name) &&
                    !useMultipleFields
                "
            >
                <p class="text-sm text-gray-600">
                    {{ t("projects.grid_view.this_formula_description") }}
                    <strong>{{ t("projects.grid_view.same_folmula") }}</strong
                    >.
                </p>
            </div>
        </div>
    </div>
</template>

<script setup>
import { t } from "@/i18n/index.js";
import { Listbox, ListboxButton, ListboxOptions, ListboxOption } from "@headlessui/vue";
import { POPULAR_FORMULAS } from "@/constants/formulasCatalog.js";
import { fieldLabel } from "@/utils/projects/rows";

// editor is what useFormulaEditor returns. It stays with the field editor so
// its state is set and reset there; this only draws it.
const props = defineProps({
    editor: { type: Object, required: true },
});

const { chosenFormula, formulaArgs, useMultipleFields, fieldOptionsForArg } = props.editor;
</script>
