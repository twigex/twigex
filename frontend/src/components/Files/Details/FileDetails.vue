<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <!-- lg matches MobileDetails' lg:hidden: this inline panel only shows at lg+, the mobile dialog takes over below it -->
    <div
        class="relative hidden lg:block h-full flex-shrink-0 overflow-hidden border-l border-gray-200 bg-white transition-[width] duration-200 ease-in-out"
        :style="{ width: detailsStore.open ? '20rem' : '0px' }"
    >
        <!-- Fixed width so content never squishes during the animation; pinned right so collapsing extends it inward, not past the viewport edge -->
        <div class="absolute inset-y-0 right-0 w-80">
            <FileDetailsContent />
        </div>
    </div>
</template>

<script setup>
import FileDetailsContent from "@/components/Files/Details/FileDetailsContent.vue";
import { useDetailsStore } from "@/store/details";

const detailsStore = useDetailsStore();
</script>
