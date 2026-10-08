<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div class="ml-7">
        <div class="rounded-md border border-gray-200 overflow-hidden">
            <!-- Provider rows -->
            <div v-if="providers.length > 0">
                <div
                    v-for="(provider, index) in providers"
                    :key="provider.id"
                    class="flex items-center justify-between px-4 py-3 bg-white"
                    :class="index < providers.length - 1 ? 'border-b border-gray-200' : ''"
                >
                    <div class="flex items-center gap-x-3 min-w-0">
                        <div
                            class="h-8 w-8 rounded-full flex items-center justify-center text-sm font-semibold text-white shrink-0"
                            :style="{
                                backgroundColor: provider.button_color || '#4f46e5',
                            }"
                        >
                            {{ provider.name.charAt(0).toUpperCase() }}
                        </div>
                        <div class="min-w-0">
                            <p class="text-sm font-medium text-gray-900">
                                {{ provider.name }}
                            </p>
                            <p class="text-xs text-gray-500 truncate">
                                {{ provider.discovery_url }}
                            </p>
                        </div>
                    </div>

                    <div class="flex items-center gap-x-2 ml-4 shrink-0">
                        <span
                            class="inline-flex items-center rounded-full px-2 py-1 text-xs font-medium"
                            :class="
                                provider.enabled
                                    ? 'bg-green-100 text-green-700'
                                    : 'bg-gray-100 text-gray-600'
                            "
                        >
                            {{
                                provider.enabled
                                    ? t("settings.authorization.oidc_provider_active")
                                    : t("settings.authorization.oidc_provider_disabled")
                            }}
                        </span>
                        <button
                            @click="emit('toggle', provider)"
                            type="button"
                            class="rounded-md bg-white px-2 py-1 text-xs font-semibold text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 hover:bg-gray-50"
                        >
                            {{
                                provider.enabled
                                    ? t("settings.authorization.oidc_provider_disable")
                                    : t("settings.authorization.oidc_provider_enable")
                            }}
                        </button>
                        <button
                            @click="emit('edit', provider)"
                            type="button"
                            class="rounded-md bg-white px-2 py-1 text-xs font-semibold text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 hover:bg-gray-50"
                        >
                            {{ t("common.button.edit") }}
                        </button>
                        <button
                            @click="emit('delete', provider)"
                            type="button"
                            class="rounded-md bg-white px-2 py-1 text-xs font-semibold text-red-600 shadow-sm ring-1 ring-inset ring-gray-300 hover:bg-gray-50"
                        >
                            {{ t("common.button.delete") }}
                        </button>
                    </div>
                </div>
            </div>

            <!-- Empty state -->
            <div v-else class="px-4 py-8 text-center bg-white">
                <p class="text-sm text-gray-500">
                    {{ t("settings.authorization.oidc_no_providers") }}
                </p>
                <p class="text-xs text-gray-400 mt-1">
                    {{ t("settings.authorization.oidc_no_providers_hint") }}
                </p>
            </div>

            <!-- Add provider -->
            <div class="px-4 py-3 bg-gray-50 border-t border-gray-200">
                <button
                    @click="emit('add')"
                    type="button"
                    class="rounded-md bg-white px-2.5 py-1.5 text-sm font-semibold text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 hover:bg-gray-50"
                >
                    {{ t("settings.authorization.oidc_add_provider") }}
                </button>
            </div>
        </div>
    </div>
</template>

<script setup>
import { t } from "@/i18n/index.js";

defineProps({
    providers: {
        type: Array,
        default: () => [],
    },
});

const emit = defineEmits(["toggle", "edit", "delete", "add"]);
</script>
