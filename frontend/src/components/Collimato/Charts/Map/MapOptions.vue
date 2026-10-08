<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <Disclosure v-slot="{ open }" :defaultOpen="true">
        <DisclosureButton
            :class="[
                'flex w-full justify-between px-4 py-2 text-left text-sm font-medium text-gray-700',
                open ? 'bg-indigo-50 hover:bg-indigo-100' : 'bg-gray-50 hover:bg-gray-100',
            ]"
        >
            <span>{{ t("collimato.charts.new_chart.map.map_options") }}</span>
            <ChevronUpIcon
                :class="open ? 'rotate-180 transform' : ''"
                class="h-5 w-5 text-indigo-500"
                aria-hidden="true"
            />
        </DisclosureButton>
        <DisclosurePanel class="px-4 pb-4 pt-4 text-sm text-gray-500">
            <h2 class="text-base font-semibold leading-7 text-gray-900">
                {{ t("collimato.charts.new_chart.map.map_style") }}
            </h2>
            <p class="text-sm leading-6 text-gray-500">
                {{ t("collimato.charts.new_chart.map.customize_map_appearance") }}
            </p>

            <!-- Coordinates -->
            <div class="mt-4">
                <label class="block text-sm font-medium leading-6 text-gray-900">
                    {{ t("collimato.charts.new_chart.map.longitude_and_latitude") }}
                </label>
                <p class="text-sm leading-6 text-gray-500">
                    {{ t("collimato.charts.new_chart.map.starting_point_map") }}
                </p>
                <div class="mt-2 flex -space-x-px">
                    <div class="min-w-0 flex-1">
                        <input
                            :value="chartOptions.longitude"
                            @input="setCenter({ longitude: $event.target.value })"
                            type="number"
                            min="-180"
                            max="180"
                            step="0.0001"
                            class="relative block w-full rounded-none rounded-bl rounded-tl border-0 bg-transparent py-1.5 text-gray-900 ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:z-10 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6"
                            :placeholder="t('collimato.charts.new_chart.map.longitude')"
                            @blur="update()"
                        />
                    </div>
                    <div class="min-w-0 flex-1">
                        <input
                            :value="chartOptions.latitude"
                            @input="setCenter({ latitude: $event.target.value })"
                            type="number"
                            min="-90"
                            max="90"
                            step="0.0001"
                            class="relative block w-full rounded-none rounded-br rounded-tr border-0 bg-transparent py-1.5 text-gray-900 ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:z-10 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6"
                            :placeholder="t('collimato.charts.new_chart.map.latitude')"
                            @blur="update()"
                        />
                    </div>
                </div>
            </div>

            <!-- Map Style Selector -->
            <div class="mt-4">
                <label class="block text-sm font-medium leading-6 text-gray-900">
                    {{ t("collimato.charts.new_chart.map.map_style") }}
                </label>
                <Listbox as="div" v-model="mapStyle" class="mt-2">
                    <div class="relative">
                        <ListboxButton
                            class="relative w-full cursor-default rounded-md bg-white py-1.5 pl-3 pr-10 text-left text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 focus:outline-none focus:ring-2 focus:ring-indigo-600 sm:text-sm sm:leading-6"
                        >
                            <span class="block truncate">{{ mapStyle.name }}</span>
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
                                class="absolute z-10 mt-1 max-h-60 w-full overflow-auto rounded-md bg-white py-1 text-base shadow-lg ring-1 ring-black/5 focus:outline-none sm:text-sm"
                            >
                                <ListboxOption
                                    as="template"
                                    v-for="style in styles"
                                    :key="style.value"
                                    :value="style"
                                    v-slot="{ active, selected }"
                                >
                                    <li
                                        :class="[
                                            active ? 'bg-indigo-600 text-white' : 'text-gray-900',
                                            'relative cursor-default select-none py-2 pl-3 pr-9',
                                        ]"
                                    >
                                        <span
                                            :class="[
                                                selected ? 'font-semibold' : 'font-normal',
                                                'block truncate',
                                            ]"
                                            >{{ style.name }}</span
                                        >
                                        <span
                                            v-if="selected"
                                            :class="[
                                                active ? 'text-white' : 'text-indigo-600',
                                                'absolute inset-y-0 right-0 flex items-center pr-4',
                                            ]"
                                        >
                                            <CheckIcon class="h-5 w-5" aria-hidden="true" />
                                        </span>
                                    </li>
                                </ListboxOption>
                            </ListboxOptions>
                        </transition>
                    </div>
                </Listbox>
            </div>

            <!-- Zoom -->
            <div class="mt-4">
                <div class="mb-2 flex items-center justify-between">
                    <label class="block text-sm font-medium leading-6 text-gray-900">
                        {{ t("collimato.charts.new_chart.map.zoom_level") }}
                    </label>
                    <span
                        class="inline-flex items-center rounded-md bg-indigo-50 px-2 py-0.5 text-xs font-medium text-indigo-700 ring-1 ring-inset ring-indigo-700/10"
                    >
                        {{ chartOptions.zoom }}
                    </span>
                </div>
                <BaseSlider :model-value="chartOptions.zoom" :min="1" :max="20" @input="setZoom" />
            </div>

            <!-- Layers header -->
            <div class="mt-5 flex items-center justify-between border-t pt-4">
                <h2 class="text-base font-semibold leading-7 text-gray-900">
                    {{ t("collimato.charts.new_chart.map.layers") }}
                </h2>
                <button
                    @click="
                        $emit('add-layer', {
                            name: t('collimato.charts.new_chart.map.new_layer'),
                            type: MAP_LAYER_TYPE_IDS[0],
                        })
                    "
                    type="button"
                    class="inline-flex items-center gap-x-1.5 rounded-md bg-indigo-600 px-2.5 py-1.5 text-sm font-semibold text-white shadow-sm hover:bg-indigo-500 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600"
                >
                    <PlusIcon class="-ml-0.5 h-4 w-4" aria-hidden="true" />
                    {{ t("collimato.charts.new_chart.map.add_layer") }}
                </button>
            </div>

            <!-- Layer list -->
            <div v-for="layer in layers" :key="layer.id" class="mt-2">
                <Disclosure v-slot="{ open: sectionOpen }">
                    <DisclosureButton
                        class="flex w-full items-center justify-between rounded-lg bg-indigo-50 px-4 py-3 text-left text-sm font-medium text-indigo-900 hover:bg-indigo-100 focus:outline-none focus-visible:ring focus-visible:ring-indigo-500/75"
                    >
                        <input
                            :value="layer.name"
                            type="text"
                            class="min-w-0 flex-1 rounded border-0 bg-transparent py-0 text-sm text-gray-900 ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-1 focus:ring-inset focus:ring-indigo-600"
                            :placeholder="t('collimato.charts.new_chart.map.layer_name')"
                            @input="$emit('update-layer-name', layer.id, $event.target.value)"
                            @click.stop
                            @mousedown.stop
                        />
                        <div class="ml-3 flex shrink-0 items-center gap-x-1.5">
                            <span
                                class="hidden rounded-full bg-indigo-100 px-2 py-0.5 text-xs font-medium text-indigo-700 sm:inline"
                            >
                                {{ typeLabel(layer.type) }}
                            </span>
                            <span
                                v-if="statusOf(layer) === 'loading'"
                                class="h-4 w-4 shrink-0 animate-spin rounded-full border-2 border-indigo-600 border-t-transparent"
                                :title="t('collimato.charts.new_chart.map.loading')"
                            ></span>
                            <button
                                v-else-if="layer.model"
                                @click.stop="$emit('reload-layer', layer.id)"
                                type="button"
                                class="rounded p-0.5 text-indigo-600 hover:bg-indigo-200 hover:text-indigo-900"
                                :title="t('collimato.charts.new_chart.map.reload_layer')"
                            >
                                <ArrowPathIcon class="h-4 w-4" aria-hidden="true" />
                            </button>
                            <button
                                @click.stop="$emit('toggle-layer', layer.id, !layer.show)"
                                type="button"
                                class="rounded p-0.5 text-indigo-600 hover:bg-indigo-200 hover:text-indigo-900"
                            >
                                <EyeIcon v-if="layer.show" class="h-4 w-4" aria-hidden="true" />
                                <EyeSlashIcon v-else class="h-4 w-4" aria-hidden="true" />
                            </button>
                            <button
                                @click.stop="$emit('remove-layer', layer.id)"
                                type="button"
                                class="rounded p-0.5 text-red-400 hover:bg-red-100 hover:text-red-600"
                            >
                                <TrashIcon class="h-4 w-4" aria-hidden="true" />
                            </button>
                            <ChevronUpIcon
                                :class="sectionOpen ? 'rotate-180 transform' : ''"
                                class="h-4 w-4 text-indigo-500"
                                aria-hidden="true"
                            />
                        </div>
                    </DisclosureButton>
                    <DisclosurePanel class="px-4 pb-3 pt-4 text-sm text-gray-500">
                        <!-- Model selector -->
                        <div class="mb-3">
                            <label class="mb-1 block text-xs font-medium text-gray-700">
                                {{ t("collimato.charts.new_chart.map.select_model") }}
                            </label>
                            <Listbox
                                as="div"
                                :model-value="layer.model"
                                @update:model-value="$emit('update-layer-model', layer.id, $event)"
                            >
                                <div class="relative">
                                    <ListboxButton
                                        class="relative w-full cursor-default rounded-md bg-white py-1.5 pl-3 pr-10 text-left text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 focus:outline-none focus:ring-2 focus:ring-indigo-500 sm:text-sm sm:leading-6"
                                    >
                                        <span class="block truncate">
                                            {{
                                                layer.model ||
                                                t("collimato.charts.new_chart.map.select_model")
                                            }}
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
                                            class="absolute z-10 mt-1 max-h-56 w-full overflow-auto rounded-md bg-white py-1 text-base shadow-lg ring-1 ring-black/5 focus:outline-none sm:text-sm"
                                        >
                                            <ListboxOption
                                                as="template"
                                                v-for="c in cubes"
                                                :key="c.table"
                                                :value="c.table"
                                                v-slot="{ active, selected }"
                                                @click="update()"
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
                                                            selected
                                                                ? 'font-semibold'
                                                                : 'font-normal',
                                                            'block truncate',
                                                        ]"
                                                        >{{ c.table }}</span
                                                    >
                                                    <span
                                                        v-if="selected"
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

                        <!-- Layer filters -->
                        <div v-if="layer.model" class="mt-3">
                            <h2 class="text-base font-semibold leading-7 text-gray-900">
                                {{ t("collimato.charts.new_chart.map.filters") }}
                            </h2>
                            <p class="text-sm leading-6 text-gray-500">
                                {{ t("collimato.charts.new_chart.map.filters_description") }}
                            </p>
                            <FilterList
                                :filters="layer.filters ?? []"
                                :cubes="modelsFor(layer)"
                                @add-filter="$emit('add-layer-filter', layer.id, $event)"
                                @update-filter="
                                    (member, filter) =>
                                        $emit('update-layer-filter', layer.id, member, filter)
                                "
                                @remove-filter="$emit('remove-layer-filter', layer.id, $event)"
                            />
                        </div>

                        <div v-if="layer.model" class="mt-3">
                            <label class="mb-1 block text-xs font-medium text-gray-700">
                                {{ t("collimato.charts.new_chart.map.row_limit") }}
                            </label>
                            <input
                                type="number"
                                min="1"
                                :max="MAX_QUERY_LIMIT"
                                :placeholder="String(MAX_QUERY_LIMIT)"
                                :value="layer.config.limit ?? ''"
                                class="block w-full rounded-md border-0 py-1.5 text-gray-900 ring-1 ring-inset ring-gray-300 sm:text-sm"
                                @change="
                                    $emit('update-layer-fields', layer.id, {
                                        limit: $event.target.value
                                            ? Number($event.target.value)
                                            : null,
                                    })
                                "
                            />
                            <p v-if="isTruncated(layer)" class="mt-1 text-xs text-amber-700">
                                {{
                                    t("collimato.charts.new_chart.map.row_limit_reached", {
                                        rows: layer.config.limit ?? MAX_QUERY_LIMIT,
                                    })
                                }}
                            </p>
                            <p v-else class="mt-1 text-xs text-gray-500">
                                {{
                                    t("collimato.charts.new_chart.map.row_limit_description", {
                                        max: MAX_QUERY_LIMIT,
                                    })
                                }}
                            </p>
                        </div>

                        <!-- Layer type selector -->
                        <div>
                            <label class="mb-1 block text-xs font-medium text-gray-700">
                                {{ t("collimato.charts.new_chart.map.layer_type") }}
                            </label>
                            <Listbox
                                as="div"
                                :model-value="layer.type"
                                @update:model-value="$emit('update-layer-type', layer.id, $event)"
                            >
                                <div class="relative">
                                    <ListboxButton
                                        class="relative w-full cursor-default rounded-md bg-white py-1.5 pl-3 pr-10 text-left text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 focus:outline-none focus:ring-2 focus:ring-indigo-600 sm:text-sm sm:leading-6"
                                    >
                                        <span class="block truncate">{{
                                            typeLabel(layer.type)
                                        }}</span>
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
                                            class="absolute z-10 mt-1 max-h-60 w-full overflow-auto rounded-md bg-white py-1 text-base shadow-lg ring-1 ring-black/5 focus:outline-none sm:text-sm"
                                        >
                                            <ListboxOption
                                                as="template"
                                                v-for="type in types"
                                                :key="type.value"
                                                :value="type.value"
                                                v-slot="{ active, selected }"
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
                                                            selected
                                                                ? 'font-semibold'
                                                                : 'font-normal',
                                                            'block truncate',
                                                        ]"
                                                        >{{ type.name }}</span
                                                    >
                                                    <span
                                                        v-if="selected"
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

                        <!-- Layer content by type -->
                        <div class="mt-3">
                            <component
                                :is="editorFor(layer.type)"
                                :cubes="cubes"
                                :config="layer.config"
                                :layer="layer"
                                :selectedCube="layer.model"
                                @update:config="$emit('update-layer-config', layer.id, $event)"
                                @update:coordinates="
                                    $emit('update-layer-coordinates', layer.id, $event)
                                "
                                @update:fields="$emit('update-layer-fields', layer.id, $event)"
                            />
                        </div>
                    </DisclosurePanel>
                </Disclosure>
            </div>
        </DisclosurePanel>
    </Disclosure>
</template>

<script setup>
import { ref, computed, watch } from "vue";
import { t } from "@/i18n/index.js";
import { joinableDataModels } from "@/utils/collimato/chartQuery.js";
import { MAX_QUERY_LIMIT } from "@/utils/collimato/chartUtils.js";
import { MAP_LAYER_TYPES, MAP_LAYER_TYPE_IDS } from "@/utils/collimato/mapLayerTypes.js";
import { editorFor } from "@/components/Collimato/Charts/Map/mapLayerEditors.js";
import BaseSlider from "@/components/BaseSlider.vue";
import {
    Disclosure,
    DisclosureButton,
    DisclosurePanel,
    Listbox,
    ListboxButton,
    ListboxOption,
    ListboxOptions,
} from "@headlessui/vue";
import {
    ChevronUpIcon,
    ChevronUpDownIcon,
    CheckIcon,
    PlusIcon,
    TrashIcon,
    EyeIcon,
    EyeSlashIcon,
    ArrowPathIcon,
} from "@heroicons/vue/20/solid";

import FilterList from "@/components/Collimato/Charts/Sidebar/FilterList.vue";

const props = defineProps({
    chartOptions: {
        type: Object,
        required: true,
    },
    cubes: {
        type: Array,
        default: () => [],
    },
    statuses: {
        type: Array,
        default: () => [],
    },
});

const emit = defineEmits([
    "update",
    "add-layer",
    "remove-layer",
    "toggle-layer",
    "update-layer-name",
    "update-layer-model",
    "update-layer-type",
    "update-layer-config",
    "update-layer-coordinates",
    "update-layer-fields",
    "reload-layer",
    "add-layer-filter",
    "update-layer-filter",
    "remove-layer-filter",
]);

function isTruncated(layer) {
    return Boolean(props.statuses.find((entry) => entry.id === layer.id)?.truncated);
}

function statusOf(layer) {
    return props.statuses.find((entry) => entry.id === layer.id)?.status ?? "";
}

function modelsFor(layer) {
    return joinableDataModels(
        props.cubes,
        props.cubes.find((cube) => cube.table == layer.model) ?? null,
    );
}

const styles = [
    {
        name: t.value("collimato.charts.new_chart.map.styles.dark_with_labels"),
        value: "https://basemaps.cartocdn.com/gl/dark-matter-gl-style/style.json",
    },
    {
        name: t.value("collimato.charts.new_chart.map.styles.dark_without_labels"),
        value: "https://basemaps.cartocdn.com/gl/dark-matter-nolabels-gl-style/style.json",
    },
    {
        name: t.value("collimato.charts.new_chart.map.styles.light"),
        value: "https://basemaps.cartocdn.com/gl/positron-gl-style/style.json",
    },
    {
        name: t.value("collimato.charts.new_chart.map.styles.light_without_labels"),
        value: "https://basemaps.cartocdn.com/gl/positron-nolabels-gl-style/style.json",
    },
];

function typeLabel(value) {
    return types.value.find((type) => type.value === value)?.name ?? value;
}

const types = computed(() =>
    MAP_LAYER_TYPE_IDS.map((id) => ({
        value: id,
        name: t.value(MAP_LAYER_TYPES[id].labelKey),
    })),
);

const layers = computed(() => props.chartOptions.layers ?? []);
const mapStyle = ref(styles.find((s) => s.value === props.chartOptions.map_style) || styles[0]);

watch(mapStyle, () => update());

function update(changes = {}) {
    emit("update", {
        map_style: mapStyle.value.value,
        longitude: props.chartOptions.longitude,
        latitude: props.chartOptions.latitude,
        zoom: props.chartOptions.zoom,
        layers: layers.value,
        ...changes,
    });
}

function setCenter(patch) {
    const next = {};

    Object.entries(patch).forEach(([key, value]) => {
        next[key] = value === "" ? "" : Number(value);
    });

    update(next);
}

function setZoom(value) {
    update({ zoom: value });
}
</script>
