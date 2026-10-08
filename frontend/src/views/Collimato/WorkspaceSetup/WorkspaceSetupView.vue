<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div class="flex flex-col h-full w-full items-center overflow-hidden">
        <!-- Step indicator -->
        <div class="shrink-0 pt-10 pb-6 w-full max-w-2xl px-4">
            <nav aria-label="Progress">
                <ol role="list" class="flex items-start">
                    <li
                        v-for="(step, idx) in steps"
                        :key="step.id"
                        class="relative flex flex-1 flex-col items-center"
                    >
                        <!-- Connecting line from previous step -->
                        <div
                            v-if="idx > 0"
                            class="absolute left-0 top-4 h-0.5 w-1/2 -translate-x-full"
                            :class="
                                steps[idx - 1].status === 'complete'
                                    ? 'bg-indigo-600'
                                    : 'bg-gray-200'
                            "
                            aria-hidden="true"
                        />
                        <!-- Connecting line to next step -->
                        <div
                            v-if="idx < steps.length - 1"
                            class="absolute right-0 top-4 h-0.5 w-1/2 translate-x-full"
                            :class="step.status === 'complete' ? 'bg-indigo-600' : 'bg-gray-200'"
                            aria-hidden="true"
                        />
                        <!-- Circle -->
                        <div
                            class="relative z-10 flex h-9 w-9 items-center justify-center rounded-full"
                            :class="{
                                'bg-indigo-600': step.status === 'complete',
                                'border-2 border-indigo-600 bg-white': step.status === 'current',
                                'border-2 border-gray-300 bg-white': step.status === 'upcoming',
                            }"
                        >
                            <CheckIcon
                                v-if="step.status === 'complete'"
                                class="h-5 w-5 text-white"
                            />
                            <span
                                v-else
                                class="text-sm font-semibold"
                                :class="
                                    step.status === 'current' ? 'text-indigo-600' : 'text-gray-400'
                                "
                            >
                                {{ step.id }}
                            </span>
                        </div>
                        <!-- Step name -->
                        <span
                            class="mt-2 text-center text-xs font-medium leading-tight max-w-[80px]"
                            :class="{
                                'text-indigo-600':
                                    step.status === 'complete' || step.status === 'current',
                                'text-gray-400': step.status === 'upcoming',
                            }"
                        >
                            {{ step.name }}
                        </span>
                    </li>
                </ol>
            </nav>
        </div>

        <!-- Step content -->
        <div class="flex-1 min-h-0 w-full flex flex-col">
            <!-- Step 1: Name & Description -->
            <div v-if="currentStepIndex === 0" class="flex-1 overflow-y-auto">
                <div class="max-w-xl mx-auto w-full px-4 pb-10 pt-2">
                    <div class="text-center mb-8">
                        <h2 class="text-2xl font-bold tracking-tight text-gray-900">
                            {{ t("collimato.workspace_setup.step1.title") }}
                        </h2>
                        <p class="mt-2 text-sm text-gray-500">
                            {{ t("collimato.workspace_setup.step1.description") }}
                        </p>
                    </div>

                    <div class="space-y-5">
                        <div>
                            <label
                                for="name"
                                class="block text-sm font-medium leading-6 text-gray-900"
                            >
                                {{ t("collimato.workspace_setup.step1.name_label") }}
                            </label>
                            <div class="mt-2">
                                <input
                                    v-model="name"
                                    type="text"
                                    id="name"
                                    name="name"
                                    class="block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset placeholder:text-gray-400 focus:ring-2 focus:ring-inset sm:text-sm sm:leading-6"
                                    :class="
                                        v$.name.$error
                                            ? 'ring-red-500 focus:ring-red-600'
                                            : 'ring-gray-300 focus:ring-indigo-600'
                                    "
                                    :placeholder="
                                        t('collimato.workspace_setup.step1.name_placeholder')
                                    "
                                />
                                <p v-if="v$.name.$error" class="mt-1.5 text-sm text-red-600">
                                    {{ t("common.error.required_field") }}
                                </p>
                            </div>
                        </div>

                        <div>
                            <label
                                for="ws-description"
                                class="block text-sm font-medium leading-6 text-gray-900"
                            >
                                {{ t("collimato.workspace_setup.step1.description_label") }}
                            </label>
                            <div class="mt-2">
                                <textarea
                                    v-model="description"
                                    id="ws-description"
                                    name="description"
                                    rows="3"
                                    class="block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6"
                                    :placeholder="
                                        t('collimato.workspace_setup.step1.description_placeholder')
                                    "
                                />
                            </div>
                        </div>
                    </div>

                    <div class="mt-8 flex items-center justify-between">
                        <button
                            type="button"
                            class="text-sm font-medium text-gray-600 hover:text-gray-900"
                            @click="router.push({ name: 'workspaces' })"
                        >
                            {{ t("common.button.cancel") }}
                        </button>
                        <BaseButton @click="verifyName">
                            {{ t("common.button.next") }}
                        </BaseButton>
                    </div>
                </div>
            </div>

            <!-- Step 2: Database connection -->
            <div v-if="currentStepIndex === 1" class="flex-1 overflow-y-auto">
                <div class="max-w-xl mx-auto w-full px-4 pb-10 pt-2">
                    <div class="text-center mb-8">
                        <h2 class="text-2xl font-bold tracking-tight text-gray-900">
                            {{ t("collimato.workspace_setup.step2.title") }}
                        </h2>
                        <p class="mt-2 text-sm text-gray-500">
                            {{ t("collimato.workspace_setup.step2.description") }}
                        </p>
                    </div>

                    <div class="space-y-5">
                        <BaseSelect
                            v-model="selected"
                            :label="t('collimato.workspace_setup.step2.db_type_label')"
                            :options="DATABASE_TYPES"
                        />

                        <div>
                            <label
                                for="display_name"
                                class="block text-sm font-medium leading-6 text-gray-900"
                            >
                                {{ t("collimato.workspace_setup.step2.displayname_label") }}
                            </label>
                            <div class="mt-2">
                                <input
                                    v-model="displayname"
                                    type="text"
                                    id="display_name"
                                    name="display_name"
                                    class="block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset placeholder:text-gray-400 focus:ring-2 focus:ring-inset sm:text-sm sm:leading-6"
                                    :class="
                                        v$.displayname.$error
                                            ? 'ring-red-500 focus:ring-red-600'
                                            : 'ring-gray-300 focus:ring-indigo-600'
                                    "
                                />
                                <p v-if="v$.displayname.$error" class="mt-1.5 text-sm text-red-600">
                                    {{ t("common.error.required_field") }}
                                </p>
                            </div>
                        </div>

                        <div class="grid grid-cols-3 gap-4">
                            <div class="col-span-2">
                                <label
                                    for="host"
                                    class="block text-sm font-medium leading-6 text-gray-900"
                                >
                                    {{ t("collimato.workspace_setup.step2.host_label") }}
                                </label>
                                <div class="mt-2">
                                    <input
                                        v-model="host"
                                        type="text"
                                        id="host"
                                        name="host"
                                        class="block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset placeholder:text-gray-400 focus:ring-2 focus:ring-inset sm:text-sm sm:leading-6"
                                        :class="
                                            v$.host.$error
                                                ? 'ring-red-500 focus:ring-red-600'
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
                                    {{ t("collimato.workspace_setup.step2.port_label") }}
                                </label>
                                <div class="mt-2">
                                    <input
                                        v-model="port"
                                        type="text"
                                        id="port"
                                        name="port"
                                        class="block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset placeholder:text-gray-400 focus:ring-2 focus:ring-inset sm:text-sm sm:leading-6"
                                        :class="
                                            v$.port.$error
                                                ? 'ring-red-500 focus:ring-red-600'
                                                : 'ring-gray-300 focus:ring-indigo-600'
                                        "
                                    />
                                    <p v-if="v$.port.$error" class="mt-1.5 text-sm text-red-600">
                                        {{ t("common.error.required_field") }}
                                    </p>
                                </div>
                            </div>
                        </div>

                        <div>
                            <label
                                for="database"
                                class="block text-sm font-medium leading-6 text-gray-900"
                            >
                                {{ t("collimato.workspace_setup.step2.database_label") }}
                            </label>
                            <div class="mt-2">
                                <input
                                    v-model="database"
                                    type="text"
                                    id="database"
                                    name="database"
                                    class="block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset placeholder:text-gray-400 focus:ring-2 focus:ring-inset sm:text-sm sm:leading-6"
                                    :class="
                                        v$.database.$error
                                            ? 'ring-red-500 focus:ring-red-600'
                                            : 'ring-gray-300 focus:ring-indigo-600'
                                    "
                                />
                                <p v-if="v$.database.$error" class="mt-1.5 text-sm text-red-600">
                                    {{ t("common.error.required_field") }}
                                </p>
                            </div>
                        </div>

                        <div class="grid grid-cols-2 gap-4">
                            <div>
                                <label
                                    for="user"
                                    class="block text-sm font-medium leading-6 text-gray-900"
                                >
                                    {{ t("collimato.workspace_setup.step2.username_label") }}
                                </label>
                                <div class="mt-2">
                                    <input
                                        v-model="username"
                                        type="text"
                                        id="user"
                                        name="user"
                                        class="block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset placeholder:text-gray-400 focus:ring-2 focus:ring-inset sm:text-sm sm:leading-6"
                                        :class="
                                            v$.username.$error
                                                ? 'ring-red-500 focus:ring-red-600'
                                                : 'ring-gray-300 focus:ring-indigo-600'
                                        "
                                    />
                                    <p
                                        v-if="v$.username.$error"
                                        class="mt-1.5 text-sm text-red-600"
                                    >
                                        {{ t("common.error.required_field") }}
                                    </p>
                                </div>
                            </div>
                            <div>
                                <label
                                    for="password"
                                    class="block text-sm font-medium leading-6 text-gray-900"
                                >
                                    {{ t("collimato.workspace_setup.step2.password_label") }}
                                </label>
                                <div class="mt-2">
                                    <input
                                        v-model="password"
                                        type="password"
                                        id="password"
                                        name="password"
                                        class="block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset placeholder:text-gray-400 focus:ring-2 focus:ring-inset sm:text-sm sm:leading-6"
                                        :class="
                                            v$.password.$error
                                                ? 'ring-red-500 focus:ring-red-600'
                                                : 'ring-gray-300 focus:ring-indigo-600'
                                        "
                                    />
                                    <p
                                        v-if="v$.password.$error"
                                        class="mt-1.5 text-sm text-red-600"
                                    >
                                        {{ t("common.error.required_field") }}
                                    </p>
                                </div>
                            </div>
                        </div>

                        <ConnectionSecurity v-model:ssl-mode="sslMode" v-model:ca-cert="caCert" />

                        <div class="flex items-center gap-x-3">
                            <button
                                type="button"
                                :disabled="testing || connecting"
                                @click="testConnection"
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

                        <div v-if="error" class="rounded-md bg-red-50 p-4">
                            <div class="flex">
                                <XCircleIcon
                                    class="h-5 w-5 shrink-0 text-red-400"
                                    aria-hidden="true"
                                />
                                <div class="ml-3">
                                    <h3 class="text-sm font-medium text-red-800">
                                        {{ t("collimato.workspace_setup.step2.connection_error") }}
                                    </h3>
                                    <p class="mt-1 text-sm text-red-700">
                                        {{ error }}
                                    </p>
                                </div>
                            </div>
                        </div>
                    </div>

                    <div class="mt-8 flex items-center justify-between">
                        <button
                            type="button"
                            class="text-sm font-medium text-gray-600 hover:text-gray-900"
                            @click="previousStep"
                        >
                            ← {{ t("common.button.back") }}
                        </button>
                        <BaseButton @click="addConnection" :isLoading="connecting">
                            {{ t("common.button.next") }}
                        </BaseButton>
                    </div>
                </div>
            </div>

            <!-- Step 3: Generate data model -->
            <div
                v-if="currentStepIndex === 2"
                class="flex flex-col flex-1 min-h-0 pt-2 pb-4 w-full max-w-xl mx-auto px-4"
            >
                <!-- Header: always visible -->
                <div class="shrink-0 text-center mb-4">
                    <h2 class="text-2xl font-bold tracking-tight text-gray-900">
                        {{ t("collimato.workspace_setup.step3.title") }}
                    </h2>
                    <p class="mt-2 text-sm text-gray-500">
                        {{ t("collimato.workspace_setup.step3.description") }}
                    </p>
                </div>

                <!-- Scrollable table selector -->
                <div class="flex-1 min-h-0 overflow-y-auto rounded-md border border-gray-200">
                    <div
                        v-if="tables.length === 0"
                        class="flex flex-col items-center justify-center gap-3 h-full py-16"
                    >
                        <div
                            class="h-10 w-10 animate-spin rounded-full border-4 border-indigo-600 border-t-transparent"
                            aria-hidden="true"
                        />
                        <p class="text-sm text-gray-500">
                            {{ t("collimato.workspace_setup.step3.loading_tables") }}
                        </p>
                    </div>

                    <DatabaseTablePicker v-else :tables="tables" @add="add" />
                </div>

                <!-- Footer: always visible -->
                <div
                    class="shrink-0 mt-4 flex items-center justify-between border-t border-gray-200 pt-4"
                >
                    <button
                        type="button"
                        class="text-sm font-medium text-gray-600 hover:text-gray-900"
                        @click="previousStep"
                    >
                        ← {{ t("common.button.back") }}
                    </button>
                    <BaseButton @click="generateDataModel" :isDisabled="tables.length === 0">
                        {{ t("collimato.workspace_setup.step3.generate_button") }}
                    </BaseButton>
                </div>
            </div>

            <!-- Step 4: Finishing -->
            <div
                v-if="currentStepIndex === 3"
                class="flex-1 flex flex-col items-center justify-center gap-4 w-full max-w-xl mx-auto px-4"
            >
                <div
                    class="h-12 w-12 animate-spin rounded-full border-4 border-indigo-600 border-t-transparent"
                    aria-hidden="true"
                />
                <h3 class="text-lg font-semibold text-gray-900">
                    {{ t("collimato.workspace_setup.step4.title") }}
                </h3>
                <p class="text-sm text-gray-500 text-center max-w-sm">
                    {{ t("collimato.workspace_setup.step4.description") }}
                </p>
            </div>
        </div>
    </div>
</template>

<script setup>
import { t } from "@/i18n/index.js";
import { ref, computed, onMounted, watch } from "vue";
import ConnectionSecurity from "@/components/Collimato/ConnectionSecurity.vue";
import BaseSelect from "@/components/BaseSelect.vue";
import { DATABASE_TYPES, DEFAULT_DATABASE_TYPE } from "@/utils/collimato/databaseTypes";
import { extractErrorMessage } from "@/utils/errors";
import { SSL_MODE_REQUIRE, securityPayload } from "@/utils/collimato/connectionSecurity";
import { useRouter, useRoute } from "vue-router";
import { CheckIcon, XCircleIcon } from "@heroicons/vue/24/outline";
import collimatoService from "@/services/collimatoService";
import { useVuelidate } from "@vuelidate/core";
import { required } from "@vuelidate/validators";
import { usePermissionsStore } from "@/store/permissions";
import { useAlertStore } from "@/store/alerts";
import DatabaseTablePicker from "@/components/Collimato/DatabaseTablePicker.vue";
import BaseButton from "@/components/BaseButton.vue";

const router = useRouter();
const route = useRoute();
const alertStore = useAlertStore();

const workspaceId = ref("");
const name = ref("");
const description = ref("");
const selected = ref(DEFAULT_DATABASE_TYPE);
const host = ref("");
const port = ref("");
const database = ref("");
const username = ref("");
const password = ref("");
const displayname = ref("");
const sslMode = ref(SSL_MODE_REQUIRE);
const caCert = ref("");

const tables = ref([]);
const dataModel = ref([]);
const error = ref(null);
const connecting = ref(false);
const testing = ref(false);
const testResult = ref(null);

watch([selected, host, port, database, username, password, sslMode, caCert], () => {
    testResult.value = null;
});

function connectionPayload() {
    return {
        host: host.value,
        port: port.value,
        database: database.value,
        username: username.value,
        password: password.value,
        displayname: displayname.value,
        type: selected.value,
        ...securityPayload(sslMode.value, caCert.value),
    };
}

async function testConnection() {
    testResult.value = null;

    const isValid = await v$.value.$validate();

    if (!isValid) return;

    testing.value = true;

    try {
        await collimatoService.testConnection(workspaceId.value, null, connectionPayload());
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

const connectionCreated = ref(false);

const steps = ref([
    {
        id: "1",
        name: t.value("collimato.workspace_setup.steps.name"),
        status: "current",
    },
    {
        id: "2",
        name: t.value("collimato.workspace_setup.steps.connection"),
        status: "upcoming",
    },
    {
        id: "3",
        name: t.value("collimato.workspace_setup.steps.data_model"),
        status: "upcoming",
    },
    {
        id: "4",
        name: t.value("collimato.workspace_setup.steps.finish"),
        status: "upcoming",
    },
]);

const rules = {
    name: { required },
    host: { required },
    port: { required },
    database: { required },
    username: { required },
    password: { required },
    displayname: { required },
};
const v$ = useVuelidate(rules, {
    name,
    host,
    port,
    database,
    username,
    password,
    displayname,
});

const currentStepIndex = computed(() => steps.value.findIndex((s) => s.status === "current"));

function nextStep() {
    const idx = currentStepIndex.value;

    steps.value[idx].status = "complete";
    steps.value[idx + 1].status = "current";
}

function previousStep() {
    const idx = currentStepIndex.value;

    steps.value[idx].status = "upcoming";
    steps.value[idx - 1].status = "current";
}

function add(object) {
    dataModel.value = object;
}

async function verifyName() {
    const isValid = await v$.value.name.$validate();

    if (!isValid) return;

    if (workspaceId.value) {
        nextStep();

        return;
    }

    collimatoService
        .createWorkspace({ name: name.value, description: description.value })
        .then((response) => {
            workspaceId.value = response.data.id;
            nextStep();
        });
}

async function addConnection() {
    if (connectionCreated.value) {
        nextStep();

        return;
    }

    const isValid = await v$.value.$validate();

    if (!isValid) return;

    error.value = null;
    testResult.value = null;
    connecting.value = true;

    collimatoService
        .createConnection(workspaceId.value, connectionPayload())
        .then(() => {
            connectionCreated.value = true;
            collimatoService.tables(workspaceId.value).then((response) => {
                tables.value = response.data;
            });
            nextStep();
        })
        .catch((err) => {
            error.value = extractErrorMessage(err);
        })
        .finally(() => {
            connecting.value = false;
        });
}

function reportSkipped(result) {
    const parts = [];

    if (result?.skipped?.length) {
        parts.push(
            t.value("collimato.workspace_setup.skipped_tables", {
                tables: result.skipped.join(", "),
            }),
        );
    }

    const columns = Object.entries(result?.skipped_columns ?? {}).map(
        ([table, cols]) => `${table}: ${cols.join(", ")}`,
    );

    if (columns.length) {
        parts.push(
            t.value("collimato.workspace_setup.skipped_columns", {
                columns: columns.join("; "),
            }),
        );
    }

    if (parts.length) {
        alertStore.showSuccess(parts.join(" "));
    }
}

function generateDataModel() {
    const result = {};

    for (const element of dataModel.value) {
        if (result[element.database] === undefined) {
            result[element.database] = {};
        }

        if (result[element.database][element.table] === undefined) {
            result[element.database][element.table] = [];
        }

        result[element.database][element.table].push({
            name: element.column,
            type: element.type,
        });
    }

    collimatoService.generate(workspaceId.value, result).then((response) => {
        reportSkipped(response.data);
        nextStep();
        collimatoService.finsihWorkspaceCreation(workspaceId.value).then(() => {
            router.push({
                name: "collimato-workspace",
                params: { workspaceId: workspaceId.value },
            });
        });
    });
}

onMounted(async () => {
    if (!usePermissionsStore().permissions.includes("create_collimato_workspace")) {
        router.push({ name: "workspaces" });

        return;
    }

    const resumeId = route.query.workspaceId;

    if (!resumeId) return;

    // Restore draft workspace, load name/description and figure out which step to resume
    const [wsRes, connRes] = await Promise.all([
        collimatoService.getWorkspaceById(resumeId),
        collimatoService.getWorkspaceConnections(resumeId),
    ]).catch(() => [null, null]);

    if (!wsRes) return;

    workspaceId.value = resumeId;
    name.value = wsRes.data.name;
    description.value = wsRes.data.description;
    steps.value[0].status = "complete";

    if (connRes?.data?.length > 0) {
        // Connection already done, jump straight to data model step
        connectionCreated.value = true;
        steps.value[1].status = "complete";
        steps.value[2].status = "current";
        collimatoService.tables(resumeId).then((res) => {
            tables.value = res.data;
        });
    } else {
        // No connection yet, resume at step 2
        steps.value[1].status = "current";
    }
});
</script>
