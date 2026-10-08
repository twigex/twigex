<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div class="w-full">
        <input
            ref="inputRef"
            v-model="inputValue"
            type="text"
            class="block w-full rounded-md border-0 bg-white px-2 py-1 text-sm leading-5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600"
            :placeholder="placeholder"
            @blur="handleBlur"
            @keydown="handleKeyDown"
        />
    </div>
</template>

<script setup>
import { ref, onMounted, nextTick, watch } from "vue";
import { t } from "@/i18n/index.js";

const props = defineProps({
    placeholder: {
        type: String,
        default: t.value("projects.new_task_input.placeholder"),
    },
    // What was typed before a save that failed, so it can be tried again.
    initialValue: {
        type: String,
        default: "",
    },
});

const emit = defineEmits(["submit", "cancel"]);

const inputRef = ref(null);
const inputValue = ref(props.initialValue);

watch(
    () => props.initialValue,
    (value) => {
        if (value) inputValue.value = value;
    },
);

onMounted(() => {
    nextTick(() => {
        if (inputRef.value) {
            inputRef.value.focus();
        }
    });
});

const handleBlur = () => {
    if (inputValue.value.trim() === "") {
        emit("cancel");
    } else {
        emit("submit", inputValue.value.trim());
        inputValue.value = "";
    }
};

const handleKeyDown = (event) => {
    if (event.key === "Enter") {
        event.preventDefault();
        if (inputValue.value.trim() !== "") {
            emit("submit", inputValue.value.trim());
            inputValue.value = "";
        }
    } else if (event.key === "Escape") {
        emit("cancel");
        inputValue.value = "";
    }
};
</script>
