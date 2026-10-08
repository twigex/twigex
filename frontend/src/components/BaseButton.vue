<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <button
        :type="type"
        :class="[
            'font-medium shadow-sm focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 flex items-center justify-center',
            buttonSizeClass,
            roundedClass,
            stateClass,
        ]"
        :disabled="isDisabled || isLoading"
        @click="handleClick"
    >
        <span
            v-if="isLoading"
            :class="[
                'loader border-2 border-t-transparent rounded-full animate-spin',
                variant === 'primary' ? 'border-white' : 'border-indigo-600',
                size === 'small' ? 'h-3 w-3' : 'h-4 w-4',
                iconGapClass,
            ]"
        ></span>
        <span v-if="prependIcon && !isLoading" :class="iconGapClass">
            <component :is="prependIcon" :class="iconClass" />
        </span>
        <slot></slot>
        <span v-if="appendIcon && !isLoading" :class="appendGapClass">
            <component :is="appendIcon" :class="iconClass" />
        </span>
    </button>
</template>

<script setup>
import { computed, useSlots } from "vue";

const props = defineProps({
    prependIcon: {
        type: [Object, Function],
        default: null,
    },
    appendIcon: {
        type: [Object, Function],
        default: null,
    },
    size: {
        type: String,
        default: "medium",
        validator: (value) => ["small", "medium", "large"].includes(value),
    },
    variant: {
        type: String,
        default: "primary",
        validator: (value) => ["primary", "secondary", "soft"].includes(value),
    },
    // Overrides the variant's colours.
    color: {
        type: String,
        default: "",
    },
    type: {
        type: String,
        default: undefined,
    },
    isDisabled: {
        type: Boolean,
        default: false,
    },
    isLoading: {
        type: Boolean,
        default: false,
    },
    rounded: {
        type: String,
        default: "default",
        validator: (value) => ["default", "full"].includes(value),
    },
});

const emit = defineEmits(["click"]);
const slots = useSlots();

const VARIANT_COLORS = {
    primary: "bg-indigo-600 hover:bg-indigo-500 text-white",
    secondary:
        "bg-white text-gray-700 ring-1 ring-inset ring-gray-300 hover:bg-gray-50 hover:text-gray-900",
    soft: "bg-indigo-50 text-indigo-600 hover:bg-indigo-100",
};

const buttonSizeClass = computed(() => {
    switch (props.size) {
        case "small":
            return slots.default ? "px-2.5 py-1.5 text-sm" : "p-1.5 text-sm";
        case "large":
            return "px-4 py-3 text-lg";
        default:
            return slots.default ? "px-3 py-2 text-sm" : "p-2 text-sm";
    }
});

const iconClass = computed(() => (props.size === "small" ? "h-4 w-4" : "h-5 w-5"));
// An icon on its own, with the label in aria-label, keeps no gap beside it.
const iconGapClass = computed(() =>
    !slots.default ? "" : props.size === "small" ? "mr-1.5" : "mr-2",
);
const appendGapClass = computed(() =>
    !slots.default ? "" : props.size === "small" ? "ml-1.5" : "ml-2",
);

// A primary button greys out when it cannot be used; the others fade, so
// they still look like the same button.
const stateClass = computed(() => {
    const colors = props.color || VARIANT_COLORS[props.variant];

    if (!props.isDisabled && !props.isLoading) return colors;
    if (props.variant === "primary") return "bg-gray-400 text-white";

    return `${colors} opacity-50 cursor-not-allowed`;
});

const roundedClass = computed(() => {
    if (props.rounded === "full") return "rounded-full";

    return props.size === "small" ? "rounded" : "rounded-md";
});

// The event goes on, for a Headless UI button rendered as this one.
const handleClick = (event) => {
    if (!props.isLoading && !props.isDisabled) {
        emit("click", event);
    }
};
</script>

<style scoped>
.loader {
    display: inline-block;
}
</style>
