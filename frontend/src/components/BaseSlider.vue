<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <input
        v-model="value"
        type="range"
        :min="props.min"
        :max="props.max"
        :step="props.step"
        @input="updateValue"
        class="w-full h-2 rounded-lg appearance-none cursor-pointer"
        :style="{
            background: `linear-gradient(to right, #6366F1 ${percentage}%, #e5e7eb ${percentage}%)`,
            accentColor: '#6366F1',
        }"
    />
</template>

<script setup>
import { ref, computed } from "vue";

const props = defineProps({
    modelValue: {
        type: Number,
        default: 0,
    },
    min: {
        type: Number,
        default: 0,
    },
    max: {
        type: Number,
        default: 100,
    },
    step: {
        type: Number,
        default: 1,
    },
});

const emit = defineEmits(["update:modelValue", "input"]);

const percentage = computed(() => {
    const range = props.max - props.min;

    if (range === 0) return 0;

    return ((value.value - props.min) / range) * 100;
});

const value = ref(props.modelValue);

function updateValue(event) {
    const newValue = Number(event.target.value);

    emit("update:modelValue", newValue);
    emit("input", newValue);
}
</script>
