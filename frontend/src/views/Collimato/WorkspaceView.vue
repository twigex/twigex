<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <SidebarLayout :title="workspaceName">
        <template #sidebar="{ mini }">
            <WorkspaceNavigation :mini="mini" />
        </template>

        <div class="mx-auto lg:px-0 h-full w-full">
            <div class="mx-auto w-full grow lg:flex xl:px-0 lg:inset-y-0 h-full">
                <div class="flex-1 xl:flex lg:inset-y-0 h-full w-full">
                    <div class="px-0 py-0 lg:pl-0 lg:pr-0 xl:flex-1 xl:pl-0 xl:pr-0 h-full w-full">
                        <RouterView v-if="loaded" />
                    </div>
                    <div></div>
                </div>
            </div>
        </div>
    </SidebarLayout>
</template>

<script setup>
import { ref, onMounted } from "vue";
import { useRoute, useRouter, RouterView, onBeforeRouteUpdate } from "vue-router";
import { useCollimatoStore } from "@/store/collimato";
import { useWorkspaceAccess } from "@/composables/collimato/useWorkspaceAccess";
import collimatoService from "@/services/collimatoService";
import SidebarLayout from "@/components/Navigation/SidebarLayout.vue";
import WorkspaceNavigation from "@/components/Collimato/WorkspaceNavigation.vue";

const collimatoStore = useCollimatoStore();
const router = useRouter();
const route = useRoute();
const { canEnter, firstSection } = useWorkspaceAccess();
const loaded = ref(false);
const workspaceName = ref("");

onBeforeRouteUpdate((to) => {
    if (!loaded.value || canEnter(to)) return;

    return firstSection(to.params.workspaceId);
});

onMounted(async () => {
    collimatoStore.setUser(null);

    await collimatoService
        .getWorkspaceById(route.params.workspaceId)
        .then((response) => {
            workspaceName.value = response.data?.name ?? "";
        })
        .catch((error) => {
            console.error(error);
            router.push({ name: "workspaces" });
        });

    await collimatoService.me(route.params.workspaceId).then((response) => {
        collimatoStore.setUser(response.data);
    });

    if (!canEnter(route)) {
        await router.replace(firstSection(route.params.workspaceId));
    }

    loaded.value = true;
});
</script>
