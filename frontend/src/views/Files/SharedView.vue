<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <FileBrowser :view="view" />
</template>

<script setup>
import { ref } from "vue";
import FileBrowser from "@/components/Files/FileBrowser.vue";

const view = ref("shared");
</script>
