<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div class="flex h-full flex-col overflow-hidden">
        <!-- Loading -->
        <div v-if="!loaded" class="flex h-full items-center justify-center">
            <div
                class="loader h-8 w-8 rounded-full border-4 border-indigo-500 border-t-transparent animate-spin"
            ></div>
        </div>

        <!-- Form -->
        <template v-else>
            <div class="flex-1 overflow-y-auto">
                <div class="mx-auto max-w-lg px-4 py-8">
                    <div>
                        <h2 class="text-base font-semibold leading-7 text-gray-900">
                            {{ t("collimato.new_connection.new_database_connection") }}
                        </h2>
                        <p class="mt-1 text-sm leading-6 text-gray-600">
                            {{ t("collimato.new_connection.new_database_connection_description") }}
                        </p>
                    </div>

                    <div class="mt-8 space-y-5">
                        <BaseSelect
                            v-model="selected"
                            :label="t('collimato.new_connection.database_type')"
                            :options="DATABASE_TYPES"
                        />

                        <!-- Host + Port -->
                        <div class="grid grid-cols-3 gap-x-4">
                            <div class="col-span-2">
                                <label
                                    for="host"
                                    class="block text-sm font-medium leading-6 text-gray-900"
                                >
                                    {{ t("collimato.new_connection.host") }}
                                </label>
                                <div class="mt-2">
                                    <input
                                        v-model="host"
                                        type="text"
                                        id="host"
                                        class="block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset placeholder:text-gray-400 focus:ring-2 focus:ring-inset sm:text-sm sm:leading-6"
                                        :class="
                                            v$.host.$error
                                                ? 'ring-red-600 focus:ring-red-600'
                                                : 'ring-gray-300 focus:ring-indigo-600'
                                        "
                                    />
                                    <p v-if="v$.host.$error" class="mt-1.5 text-sm text-red-600">
                                        {{ t("common.error.required_field") }}
                                    </p>
                                </div>
                            </div>
                            <div>
                                <label
                                    for="port"
                                    class="block text-sm font-medium leading-6 text-gray-900"
                                >
                                    {{ t("collimato.new_connection.port") }}
                                </label>
                                <div class="mt-2">
                                    <input
                                        v-model="port"
                                        type="number"
                                        id="port"
                                        class="block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset placeholder:text-gray-400 focus:ring-2 focus:ring-inset sm:text-sm sm:leading-6"
                                        :class="
                                            v$.port.$error
                                                ? 'ring-red-600 focus:ring-red-600'
                                                : 'ring-gray-300 focus:ring-indigo-600'
                                        "
                                    />
                                    <p v-if="v$.port.$error" class="mt-1.5 text-sm text-red-600">
                                        {{ t("common.error.required_field") }}
                                    </p>
                                </div>
                            </div>
                        </div>

                        <!-- Database name -->
                        <div>
                            <label
                                for="database-name"
                                class="block text-sm font-medium leading-6 text-gray-900"
                            >
                                {{ t("collimato.new_connection.database_name") }}
                            </label>
                            <div class="mt-2">
                                <input
                                    v-model="database"
                                    type="text"
                                    id="database-name"
                                    class="block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6"
                                    :class="
                                        v$.database.$error
                                            ? 'ring-red-600 focus:ring-red-600'
                                            : 'ring-gray-300 focus:ring-indigo-600'
                                    "
                                />
                                <p v-if="v$.database.$error" class="mt-1.5 text-sm text-red-600">
                                    {{ t("common.error.required_field") }}
                                </p>
                            </div>
                        </div>

                        <!-- Username -->
                        <div>
                            <label
                                for="username"
                                class="block text-sm font-medium leading-6 text-gray-900"
                            >
                                {{ t("collimato.new_connection.username") }}
                            </label>
                            <div class="mt-2">
                                <input
                                    v-model="user"
                                    type="text"
                                    id="username"
                                    autocomplete="username"
                                    class="block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset placeholder:text-gray-400 focus:ring-2 focus:ring-inset sm:text-sm sm:leading-6"
                                    :class="
                                        v$.user.$error
                                            ? 'ring-red-600 focus:ring-red-600'
                                            : 'ring-gray-300 focus:ring-indigo-600'
                                    "
                                />
                                <p v-if="v$.user.$error" class="mt-1.5 text-sm text-red-600">
                                    {{ t("common.error.required_field") }}
                                </p>
                            </div>
                        </div>

                        <!-- Password -->
                        <div>
                            <label
                                for="password"
                                class="block text-sm font-medium leading-6 text-gray-900"
                            >
                                {{ t("collimato.new_connection.password") }}
                            </label>
                            <div class="mt-2">
                                <input
                                    v-model="password"
                                    type="password"
                                    id="password"
                                    autocomplete="current-password"
                                    :placeholder="
                                        isEdit
                                            ? t('collimato.new_connection.password_unchanged')
                                            : ''
                                    "
                                    class="block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset placeholder:text-gray-400 focus:ring-2 focus:ring-inset sm:text-sm sm:leading-6"
                                    :class="
                                        v$.password.$error
                                            ? 'ring-red-600 focus:ring-red-600'
                                            : 'ring-gray-300 focus:ring-indigo-600'
                                    "
                                />
                                <p v-if="v$.password.$error" class="mt-1.5 text-sm text-red-600">
                                    {{ t("common.error.required_field") }}
                                </p>
                            </div>
                        </div>

                        <!-- Display name -->
                        <div>
                            <label
                                for="display-name"
                                class="block text-sm font-medium leading-6 text-gray-900"
                            >
                                {{ t("collimato.new_connection.display_name") }}
                            </label>
                            <div class="mt-2">
                                <input
                                    v-model="displayName"
                                    type="text"
                                    id="display-name"
                                    class="block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset placeholder:text-gray-400 focus:ring-2 focus:ring-inset sm:text-sm sm:leading-6"
                                    :class="
                                        v$.displayName.$error
                                            ? 'ring-red-600 focus:ring-red-600'
                                            : 'ring-gray-300 focus:ring-indigo-600'
                                    "
                                />
                                <p v-if="v$.displayName.$error" class="mt-1.5 text-sm text-red-600">
                                    {{ t("common.error.required_field") }}
                                </p>
                            </div>
                        </div>

                        <ConnectionSecurity v-model:ssl-mode="sslMode" v-model:ca-cert="caCert" />

                        <div class="flex items-center gap-x-3">
                            <button
                                type="button"
                                :disabled="testing || loading"
                                @click="test"
                                class="inline-flex items-center gap-x-1.5 rounded-md bg-white px-3 py-2 text-sm font-semibold text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 hover:bg-gray-50 disabled:cursor-not-allowed disabled:text-gray-400"
                            >
                                <div
                                    v-if="testing"
                                    class="loader h-3 w-3 rounded-full border-2 border-gray-500 border-t-transparent animate-spin"
                                ></div>
                                {{ t("collimato.new_connection.test") }}
                            </button>

                            <p
                                v-if="testResult"
                                class="text-sm"
                                :class="testResult.ok ? 'text-green-700' : 'text-red-600'"
                            >
                                {{ testResult.message }}
                            </p>
                        </div>

                        <!-- Error banner -->
                        <div v-if="error !== null" class="rounded-md bg-red-50 p-4">
                            <div class="flex">
                                <XCircleIcon
                                    class="h-5 w-5 shrink-0 text-red-400"
                                    aria-hidden="true"
                                />
                                <div class="ml-3">
                                    <h3 class="text-sm font-medium text-red-800">
                                        {{
                                            t(
                                                "collimato.new_connection.error.error_creating_connection",
                                            )
                                        }}
                                    </h3>
                                    <p class="mt-1 text-sm text-red-700">
                                        {{ error }}
                                    </p>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
            </div>

            <!-- Footer -->
            <div
                class="shrink-0 flex items-center justify-end gap-x-3 border-t bg-gray-50 px-4 py-4"
            >
                <button
                    type="button"
                    class="text-sm font-semibold leading-6 text-gray-900 hover:text-gray-700"
                    @click="router.push({ name: 'connections' })"
                >
                    {{ t("common.button.cancel") }}
                </button>
                <button
                    type="button"
                    :disabled="loading"
                    @click="create"
                    class="inline-flex items-center gap-x-1.5 rounded-md px-3 py-2 text-sm font-semibold text-white shadow-sm focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600"
                    :class="
                        loading
                            ? 'bg-gray-300 cursor-not-allowed'
                            : 'bg-indigo-600 hover:bg-indigo-500'
                    "
                >
                    <div
                        v-if="loading"
                        class="loader h-3 w-3 rounded-full border-2 border-white border-t-transparent animate-spin"
                    ></div>
                    {{ isEdit ? t("common.button.update") : t("common.button.create") }}
                </button>
            </div>
        </template>
    </div>
</template>

<script setup>
import { ref, computed, onMounted, watch } from "vue";
import ConnectionSecurity from "@/components/Collimato/ConnectionSecurity.vue";
import BaseSelect from "@/components/BaseSelect.vue";
import { DATABASE_TYPES, DEFAULT_DATABASE_TYPE } from "@/utils/collimato/databaseTypes";
import { extractErrorMessage } from "@/utils/errors";
import {
    SSL_MODE_REQUIRE,
    securityOf,
    securityPayload,
} from "@/utils/collimato/connectionSecurity";
import { useRouter, useRoute } from "vue-router";
import { useCollimatoStore } from "@/store/collimato";
import { useAlertStore } from "@/store/alerts";
import { t } from "@/i18n/index.js";
import collimatoService from "@/services/collimatoService";
import { useVuelidate } from "@vuelidate/core";
import { required } from "@vuelidate/validators";
import { XCircleIcon } from "@heroicons/vue/20/solid";

const loaded = ref(false);
const loading = ref(false);
const error = ref(null);
const collimatoStore = useCollimatoStore();
const alertStore = useAlertStore();
const route = useRoute();
const router = useRouter();

const selected = ref(DEFAULT_DATABASE_TYPE);
const host = ref("");
const port = ref("");
const database = ref("");
const user = ref("");
const password = ref("");
const displayName = ref("");
const sslMode = ref(SSL_MODE_REQUIRE);
const caCert = ref("");

const testing = ref(false);
const testResult = ref(null);

watch([selected, host, port, database, user, password, sslMode, caCert], () => {
    testResult.value = null;
});

const isEdit = computed(() => route.name === "edit-connection");

const rules = computed(() => ({
    host: { required },
    port: { required },
    database: { required },
    user: { required },
    password: isEdit.value ? {} : { required },
    displayName: { required },
}));

const v$ = useVuelidate(rules, {
    host,
    port,
    database,
    user,
    password,
    displayName,
});

onMounted(() => {
    if (isEdit.value) {
        if (!collimatoStore.hasPermissionToEditConnections) {
            alertStore.showError(t.value("collimato.new_connection.error.no_permission_edit"));
            router.push({ name: "connections" });

            return;
        }

        collimatoService
            .getConnectionById(route.params.workspaceId, route.params.id)
            .then((response) => {
                const conn = response.data;

                selected.value = conn.type;
                host.value = conn.host;
                port.value = conn.port;
                database.value = conn.database;
                user.value = conn.username;
                password.value = conn.password;
                displayName.value = conn.displayname;
                const security = securityOf(conn);

                sslMode.value = security.sslMode;
                caCert.value = security.caCert;
                loaded.value = true;
            });

        return;
    }

    if (!collimatoStore.hasPermissionToCreateConnections) {
        alertStore.showError(t.value("collimato.new_connection.error.no_permission_create"));
        router.push({ name: "connections" });

        return;
    }

    loaded.value = true;
});

function connectionPayload() {
    return {
        type: selected.value,
        port: String(port.value),
        host: host.value,
        username: user.value,
        password: password.value,
        displayname: displayName.value,
        database: database.value,
        ...securityPayload(sslMode.value, caCert.value),
    };
}

async function test() {
    testResult.value = null;

    const isValid = await v$.value.$validate();

    if (!isValid) return;

    testing.value = true;

    try {
        await collimatoService.testConnection(
            route.params.workspaceId,
            route.params.id,
            connectionPayload(),
        );
        testResult.value = {
            ok: true,
            message: t.value("collimato.new_connection.test_ok"),
        };
    } catch (err) {
        testResult.value = { ok: false, message: extractErrorMessage(err) };
    } finally {
        testing.value = false;
    }
}

async function create() {
    error.value = null;
    const isValid = await v$.value.$validate();

    if (!isValid) return;

    loading.value = true;

    const payload = connectionPayload();

    if (isEdit.value) {
        collimatoService
            .updateConnection(route.params.workspaceId, route.params.id, payload)
            .then(() => {
                loading.value = false;
                router.push({ name: "connections" });
            })
            .catch((err) => {
                error.value = extractErrorMessage(err);
                loading.value = false;
            });

        return;
    }

    collimatoService
        .createConnection(route.params.workspaceId, payload)
        .then(() => {
            loading.value = false;
            router.push({ name: "connections" });
        })
        .catch((err) => {
            error.value = extractErrorMessage(err);
            loading.value = false;
        });
}
</script>
