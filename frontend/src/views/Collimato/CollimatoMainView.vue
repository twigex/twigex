<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div class="mx-auto lg:px-0 h-full w-full">
        <div
            v-if="!can('view_collimato') && loaded"
            class="flex flex-col items-center justify-center h-full text-center px-6"
        >
            <LockClosedIcon class="h-12 w-12 text-gray-300" />
            <h3 class="mt-4 text-base font-semibold text-gray-900">Access restricted</h3>
            <p class="mt-1 text-sm text-gray-500 max-w-sm">
                You don't have permission to access Collimato. Contact your administrator to request
                access.
            </p>
        </div>
        <template v-else>
            <div class="mx-auto w-full grow lg:flex xl:px-0 lg:inset-y-0 h-full">
                <div class="flex-1 xl:flex lg:inset-y-0 h-full w-full">
                    <div class="px-0 py-0 lg:pl-0 lg:pr-0 xl:flex-1 xl:pl-0 xl:pr-0 h-full w-full">
                        <RouterView />
                    </div>
                </div>
            </div>
        </template>
    </div>
</template>

<script setup>
import { ref, onMounted } from "vue";
import collimatoService from "@/services/collimatoService";
import { usePermissions } from "@/composables/usePermissions";
import { LockClosedIcon } from "@heroicons/vue/24/outline";

const { can } = usePermissions();
const workspaces = ref([]);
const loaded = ref(false);

onMounted(() => {
    collimatoService
        .getWorkspaces()
        .then((response) => {
            workspaces.value = response.data;
            loaded.value = true;
        })
        .catch(() => {
            loaded.value = true;
        });
});
</script>
