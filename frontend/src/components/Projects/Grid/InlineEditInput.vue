<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <input
        ref="inputRef"
        :value="value"
        :type="inputType"
        :class="inputClass"
        :style="inputStyle"
        :step="step"
        @blur="handleBlur"
        @input="updateItemValue"
        @keyup.enter="handleEnter"
        @keyup.esc="handleEscape"
        autofocus
    />
</template>

<script setup>
import { ref, nextTick, onMounted, computed, onUnmounted } from "vue";

const props = defineProps({
    value: {
        type: [String, Number, Boolean],
        default: "",
    },
    colIndex: {
        type: Number,
        required: true,
    },
    item: {
        type: Object,
        required: true,
    },
    inputClass: {
        type: String,
        // The size and place of the grid's hover outline: 4px inside the cell,
        // which is 5px above and 4px beside the cell's content.
        default:
            "-mx-1 -my-[5px] block h-[34px] min-w-0 flex-1 rounded-lg border-0 bg-white px-2 text-[13px] text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600",
    },
    inputStyle: {
        type: Object,
        default: () => ({}),
    },
    headerType: {
        type: String,
        default: "text",
    },
    step: {
        type: String,
        default: "any",
    },
});

const emit = defineEmits(["save", "cancel"]);

const inputRef = ref(null);
const isActive = ref(true);

// Determine input type based on header type
const inputType = computed(() => {
    const type = props.headerType?.toUpperCase();

    if (["DECIMAL", "BIGINT", "INT", "NUMBER"].includes(type)) {
        return "number";
    }

    return "text";
});

// Focus input when component mounts
onMounted(() => {
    nextTick(() => {
        if (inputRef.value && isActive.value) {
            inputRef.value.focus();
            inputRef.value.select();
        }
    });
});

const updateItemValue = (event) => {
    // This mimics the original v-model behavior
    // Update the item property directly
    const fieldName = props.item.tableHeaders?.[props.colIndex]?.name;

    if (fieldName && props.item) {
        props.item[fieldName] = event.target.value;
    }
};

const handleBlur = (event) => {
    if (!isActive.value) return;
    emit("save", props.colIndex, props.item, event);
};

const handleEnter = (event) => {
    if (!isActive.value) return;
    emit("save", props.colIndex, props.item, event);
};

const handleEscape = () => {
    if (!isActive.value) return;
    emit("cancel");
};

// Deactivate on unmount
onUnmounted(() => {
    isActive.value = false;
});
</script>
