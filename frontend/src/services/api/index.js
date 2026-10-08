// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import axios from "axios";
import { getCsrfFromCookie } from "@/utils/utils";
import { useUserStore } from "@/store/user";
import { useSettingsStore } from "@/store/settings";
import { useVersionStore } from "@/store/version";
import { useWorkspaceStore } from "@/store/workspaces";
import { resetAllStores } from "@/store/reset";

const PUBLIC_ENDPOINTS = [
    "/auth/login",
    "/auth/session",
    "/password/reset", // this checks /password/reset/{token}
    "/password/policy",
    "/guest", // a guest has no session to expire, so a 401 must not sign them out
];

function isPublicAuthCall(url = "") {
    const path = url.split("?")[0];

    return PUBLIC_ENDPOINTS.some((e) => path === e || path.startsWith(e + "/"));
}

function getUserLocale() {
    const userStore = useUserStore();
    const settingsStore = useSettingsStore();

    const userLocale = userStore.getLanguage;

    if (userLocale) return userLocale;

    const serverLocale = settingsStore.config?.DefaultLocale;

    if (serverLocale) return serverLocale;

    return navigator.language || "en";
}

const axiosInstance = axios.create({
    baseURL: "/api/",
    // add more default settings like headers and so on
});

axiosInstance.interceptors.request.use(function (config) {
    config.headers["X-CSRF-TOKEN"] = getCsrfFromCookie();
    config.headers["Accept-Language"] = getUserLocale();

    const clientId = useWorkspaceStore().getConnectionID;

    if (clientId) config.headers["X-Client-ID"] = clientId;

    return config;
});

// The router is passed in from main.js after Pinia and the router are created,
// rather than imported here. A top-level import would create a cycle
// (router -> views -> services -> this module).
export function setupInterceptors(router) {
    axiosInstance.interceptors.response.use(
        (response) => {
            const deploy = response.headers?.["x-app-deploy"];

            if (deploy !== undefined) {
                useVersionStore().check({
                    deploy,
                    version: response.headers["x-app-version"],
                    minVersion: response.headers["x-app-min-version"],
                });
            }

            return response;
        },
        (error) => {
            // A background 401 (session expired while idle) isn't caught by the
            // navigation guard, since no navigation happens. Redirect from here.
            if (
                error.response?.status === 401 &&
                !isPublicAuthCall(error.config?.url) &&
                router.currentRoute.value.name !== "login"
            ) {
                resetAllStores();
                router.replace({
                    name: "login",
                    query: { redirect: router.currentRoute.value.fullPath },
                });
            }

            return Promise.reject(error);
        },
    );
}

export default axiosInstance;
