<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div class="rounded-md bg-yellow-50 p-4">
        <div class="flex">
            <ExclamationTriangleIcon class="size-5 shrink-0 text-yellow-400" aria-hidden="true" />
            <p class="ml-3 text-sm text-yellow-800">
                {{ t("settings.env_managed_notice") }}
            </p>
        </div>
    </div>
</template>

<script setup>
import { t } from "@/i18n/index";

import { ExclamationTriangleIcon } from "@heroicons/vue/20/solid";
</script>
