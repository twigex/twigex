<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div>
        <!-- Header (collapses the pane) -->
        <div
            class="flex shrink-0 cursor-pointer select-none items-center justify-between border-b bg-gray-50 px-3 py-2 hover:bg-gray-100"
            @click="$emit('update:open', !open)"
        >
            <div class="flex min-w-0 items-center gap-x-2">
                <ChevronRightIcon
                    class="h-4 w-4 shrink-0 text-gray-400 transition-transform duration-200"
                    :class="open ? 'rotate-90' : ''"
                    aria-hidden="true"
                />
                <span
                    class="rounded px-1.5 py-0.5 text-[10px] font-semibold uppercase tracking-wide"
                    :class="
                        isView ? 'bg-emerald-100 text-emerald-700' : 'bg-indigo-100 text-indigo-700'
                    "
                >
                    {{
                        isView
                            ? t("collimato.model_summary.view")
                            : t("collimato.model_summary.cube")
                    }}
                </span>
                <span class="truncate text-sm font-medium text-gray-900">
                    {{ model?.name || file.name }}
                </span>
            </div>
            <button
                v-if="editable"
                type="button"
                class="inline-flex items-center gap-x-1.5 rounded-md bg-indigo-600 px-2.5 py-1 text-xs font-semibold text-white shadow-sm transition-colors hover:bg-indigo-500"
                @click.stop="$emit('edit')"
            >
                <PencilSquareIcon class="h-3.5 w-3.5" aria-hidden="true" />
                {{ t("collimato.data_view.edit_in_builder") }}
            </button>
        </div>

        <!-- Body -->
        <div v-show="open" class="flex min-h-0 flex-1 flex-col">
            <!-- Structured chips: fills the pane, or yields room when code is open -->
            <div
                class="space-y-4 p-3"
                :class="
                    showCode ? 'max-h-40 shrink-0 overflow-auto' : 'min-h-0 flex-1 overflow-auto'
                "
            >
                <p v-if="!model" class="text-sm text-gray-400">
                    {{ t("collimato.model_summary.unparsable") }}
                </p>

                <!-- Cube -->
                <template v-else-if="!isView">
                    <p class="text-xs text-gray-500">
                        {{ t("collimato.model_summary.table") }} ·
                        <span class="font-mono text-gray-700">{{ model.sql_table }}</span>
                    </p>

                    <section v-if="model.dimensions?.length">
                        <h4 :class="groupClass">
                            {{ t("collimato.builder.dimensions") }}
                            <span class="text-gray-400">({{ model.dimensions.length }})</span>
                        </h4>
                        <div class="flex flex-wrap gap-1.5">
                            <span v-for="d in model.dimensions" :key="d.name" :class="chipClass">
                                <span class="font-mono">{{ d.name }}</span>
                                <span v-if="d.primary_key" :class="tagClass">PK</span>
                                <span v-if="d.custom" :class="tagClass">SQL</span>
                            </span>
                        </div>
                    </section>

                    <section v-if="model.measures?.length">
                        <h4 :class="groupClass">
                            {{ t("collimato.builder.measures") }}
                            <span class="text-gray-400">({{ model.measures.length }})</span>
                        </h4>
                        <div class="flex flex-wrap gap-1.5">
                            <span v-for="m in model.measures" :key="m.name" :class="chipClass">
                                <span class="font-mono">{{ m.name }}</span>
                                <span :class="mutedTagClass">{{ m.type }}</span>
                            </span>
                        </div>
                    </section>

                    <section v-if="model.joins?.length">
                        <h4 :class="groupClass">
                            {{ t("collimato.builder.joins") }}
                            <span class="text-gray-400">({{ model.joins.length }})</span>
                        </h4>
                        <div class="flex flex-wrap gap-1.5">
                            <span v-for="j in model.joins" :key="j.name" :class="chipClass">
                                <span class="text-gray-400">→</span>
                                <span class="font-mono">{{ j.name }}</span>
                                <span :class="mutedTagClass">{{
                                    t("collimato.builder.rel_" + j.relationship)
                                }}</span>
                            </span>
                        </div>
                    </section>
                </template>

                <!-- View -->
                <template v-else>
                    <section v-if="model.cubes?.length">
                        <h4 :class="groupClass">
                            {{ t("collimato.model_summary.cubes") }}
                            <span class="text-gray-400">({{ model.cubes.length }})</span>
                        </h4>
                        <div class="space-y-2">
                            <div
                                v-for="(c, i) in model.cubes"
                                :key="i"
                                class="flex flex-wrap items-center gap-1.5"
                            >
                                <span class="font-mono text-xs text-gray-600">{{
                                    c.join_path
                                }}</span>
                                <span v-if="c.prefix" :class="tagClass">{{
                                    t("collimato.model_summary.prefix")
                                }}</span>
                                <span v-for="m in c.includes" :key="m" :class="chipClass">
                                    <span class="font-mono">{{
                                        m === "*"
                                            ? t("collimato.model_summary.all_members_short")
                                            : m
                                    }}</span>
                                </span>
                            </div>
                        </div>
                    </section>
                </template>
            </div>

            <!-- Advanced: read-only generated YAML that expands to fill -->
            <div class="flex flex-col border-t" :class="showCode ? 'min-h-0 flex-1' : 'shrink-0'">
                <button
                    type="button"
                    class="flex shrink-0 items-center gap-1 px-3 py-2 text-xs font-medium text-gray-500 hover:text-gray-700"
                    @click="showCode = !showCode"
                >
                    <ChevronRightIcon
                        class="h-4 w-4 transition-transform duration-200"
                        :class="showCode ? 'rotate-90' : ''"
                        aria-hidden="true"
                    />
                    {{ t("collimato.model_summary.advanced") }}
                </button>
                <pre
                    v-show="showCode"
                    class="mx-3 mb-3 min-h-[10rem] flex-1 overflow-auto rounded bg-gray-900 p-3 font-mono text-xs text-gray-100"
                    >{{ file.content }}</pre
                >
            </div>
        </div>
    </div>
</template>

<script setup>
import { ref, computed } from "vue";
import { PencilSquareIcon, ChevronRightIcon } from "@heroicons/vue/20/solid";
import { t } from "@/i18n/index.js";

const props = defineProps({
    file: { type: Object, required: true },
    open: { type: Boolean, default: true },
    editable: { type: Boolean, default: true },
});

defineEmits(["edit", "update:open"]);

const showCode = ref(false);

const isView = computed(() => props.file.type === "view");
const model = computed(() => {
    try {
        return props.file.builder_model ? JSON.parse(props.file.builder_model) : null;
    } catch {
        return null;
    }
});

const groupClass = "mb-1.5 text-xs font-semibold uppercase tracking-wide text-gray-500";
const chipClass =
    "inline-flex items-center gap-1 rounded-md bg-gray-50 px-2 py-1 text-xs text-gray-700 ring-1 ring-inset ring-gray-200";
const tagClass = "rounded bg-indigo-100 px-1 text-[10px] font-semibold uppercase text-indigo-700";
const mutedTagClass = "rounded bg-gray-200 px-1 text-[10px] font-medium text-gray-600";
</script>
