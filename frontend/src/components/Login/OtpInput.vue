<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div class="flex justify-center items-center space-x-2">
        <input
            v-for="(char, index) in otp"
            :key="index"
            type="text"
            maxlength="1"
            class="w-10 h-10 border border-gray-300 rounded text-center text-xl"
            v-model="otp[index]"
            @input="handleInput(index, $event)"
            @keyup="handleKeyup(index, $event)"
            :ref="(el) => (otpInput[index] = el)"
        />
    </div>
</template>

<script setup>
import { ref } from "vue";

// Props
const props = defineProps({
    length: {
        type: Number,
        required: true,
        validator: (value) => value > 0,
    },
});

const emit = defineEmits(["otp-complete"]);
const otp = ref(Array(props.length).fill(""));
const otpInput = ref([]);

const handleInput = (index, event) => {
    if (event.inputType === "insertText") {
        if (index < props.length - 1) {
            otpInput.value[index + 1]?.focus();
        }
    }

    if (otp.value.join("").length === props.length) {
        emitOtp();
    }
};

const handleKeyup = (index, event) => {
    if (event.key === "Backspace" && !otp.value[index] && index > 0) {
        otpInput.value[index - 1]?.focus();
    }
};

const emitOtp = () => {
    emit("otp-complete", otp.value.join(""));
};
</script>
