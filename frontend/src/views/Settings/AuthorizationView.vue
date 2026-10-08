<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div class="flex flex-col h-full overflow-hidden">
        <div v-if="loaded" class="flex flex-row justify-center flex-1 overflow-y-auto p-2">
            <div class="space-y-6 w-full md:w-3/4">
                <div class="pb-4 border-b border-gray-900/10">
                    <h2 class="text-base font-semibold leading-7 text-gray-900">
                        {{ t("settings.authorization.login_methods") }}
                    </h2>
                    <p class="mt-1 text-sm leading-6 text-gray-600">
                        {{ t("settings.authorization.login_methods_description") }}
                    </p>

                    <div class="mt-6 space-y-5">
                        <!-- Lockout warning -->
                        <div
                            v-if="
                                !settings.localAuthEnabled &&
                                !settings.ldapEnabled &&
                                !settings.oidcEnabled
                            "
                            class="rounded-md bg-yellow-50 border border-yellow-200 p-4"
                        >
                            <p class="text-sm text-yellow-800">
                                {{ t("settings.authorization.no_login_methods_warning") }}
                            </p>
                        </div>

                        <!-- Email & Password -->
                        <div class="relative flex gap-x-3">
                            <div class="flex h-6 items-center">
                                <input
                                    v-model="settings.localAuthEnabled"
                                    id="local-auth"
                                    type="checkbox"
                                    class="h-4 w-4 rounded border-gray-300 text-indigo-600 focus:ring-indigo-600"
                                />
                            </div>
                            <div class="text-sm leading-6">
                                <label for="local-auth" class="font-medium text-gray-900">{{
                                    t("settings.authorization.email_and_password")
                                }}</label>
                                <p class="text-gray-500">
                                    {{ t("settings.authorization.email_and_password_description") }}
                                </p>
                            </div>
                        </div>

                        <!-- LDAP -->
                        <div
                            class="relative flex gap-x-3"
                            :class="{ 'opacity-50 pointer-events-none': !hasLdapLicense }"
                        >
                            <div class="flex h-6 items-center">
                                <input
                                    v-model="settings.ldapEnabled"
                                    id="ldap-auth"
                                    type="checkbox"
                                    :disabled="settings.ldapEnvLocked"
                                    class="h-4 w-4 rounded border-gray-300 text-indigo-600 focus:ring-indigo-600 disabled:opacity-50 disabled:cursor-not-allowed"
                                />
                            </div>
                            <div class="text-sm leading-6">
                                <label
                                    for="ldap-auth"
                                    class="font-medium text-gray-900 inline-flex items-center gap-x-2"
                                    >{{ t("settings.authorization.ldap") }}
                                    <span
                                        v-if="!hasLdapLicense"
                                        class="inline-flex items-center gap-x-1 rounded-md bg-amber-50 px-1.5 py-0.5 text-xs font-medium text-amber-800 ring-1 ring-inset ring-amber-600/20"
                                    >
                                        <LockClosedIcon class="h-3 w-3" />
                                        {{ t("settings.license.enterprise_plan") }}
                                    </span>
                                </label>
                                <p class="text-gray-500">
                                    {{ t("settings.authorization.ldap_description") }}
                                </p>
                                <p
                                    v-if="settings.ldapEnvLocked"
                                    class="mt-0.5 text-xs text-yellow-600"
                                >
                                    {{ t("settings.env_managed_notice") }}
                                </p>
                            </div>
                        </div>

                        <!-- LDAP config -->
                        <AuthorizationLdapConfig
                            v-if="settings.ldapEnabled"
                            v-model:ldap="settings.ldap"
                            v-model:security-option="ldapSecurityOption"
                            :env-locked="settings.ldapEnvLocked"
                            :default-port="defaultLdapPort"
                            :security-options="ldapSecurityOptions"
                            :uses-tls="ldapUsesTls"
                            :testing="testingLDAP"
                            :test-result="ldapTestResult"
                            :syncing="syncingLDAP"
                            :sync-message="ldapSyncMessage"
                            @test="testLDAP"
                            @sync="runLDAPSync"
                        />

                        <!-- OpenID Connect -->
                        <div
                            class="relative flex gap-x-3"
                            :class="{ 'opacity-50 pointer-events-none': !hasOAuthLicense }"
                        >
                            <div class="flex h-6 items-center">
                                <input
                                    v-model="settings.oidcEnabled"
                                    id="oidc-auth"
                                    type="checkbox"
                                    class="h-4 w-4 rounded border-gray-300 text-indigo-600 focus:ring-indigo-600"
                                />
                            </div>
                            <div class="text-sm leading-6">
                                <label
                                    for="oidc-auth"
                                    class="font-medium text-gray-900 inline-flex items-center gap-x-2"
                                    >{{ t("settings.authorization.oidc") }}
                                    <span
                                        v-if="!hasOAuthLicense"
                                        class="inline-flex items-center gap-x-1 rounded-md bg-amber-50 px-1.5 py-0.5 text-xs font-medium text-amber-800 ring-1 ring-inset ring-amber-600/20"
                                    >
                                        <LockClosedIcon class="h-3 w-3" />
                                        {{ t("settings.license.business_plan") }}
                                    </span>
                                </label>
                                <p class="text-gray-500">
                                    {{ t("settings.authorization.oidc_description") }}
                                </p>
                            </div>
                        </div>

                        <!-- OIDC Providers list -->
                        <AuthorizationOidcProviderList
                            v-if="settings.oidcEnabled"
                            :providers="oidcProviders"
                            @toggle="toggleProvider"
                            @edit="editProvider"
                            @delete="confirmDeleteProvider"
                            @add="openCreateProviderModal"
                        />
                    </div>
                </div>

                <div class="pb-4 border-b border-gray-900/10">
                    <h2 class="text-base font-semibold leading-7 text-gray-900">
                        {{ t("settings.authorization.login_page") }}
                    </h2>
                    <p class="mt-1 text-sm leading-6 text-gray-600">
                        {{ t("settings.authorization.login_page_description") }}
                    </p>

                    <div class="mt-6 space-y-4">
                        <div
                            class="relative flex gap-x-3"
                            :class="!settings.oidcEnabled ? 'opacity-50' : ''"
                        >
                            <div class="flex h-6 items-center">
                                <input
                                    v-model="settings.oidcFirst"
                                    id="oidc-first"
                                    type="checkbox"
                                    :disabled="!settings.oidcEnabled"
                                    class="h-4 w-4 rounded border-gray-300 text-indigo-600 focus:ring-indigo-600"
                                />
                            </div>
                            <div class="text-sm leading-6">
                                <label for="oidc-first" class="font-medium text-gray-900">{{
                                    t("settings.authorization.show_sso_first")
                                }}</label>
                                <p class="text-gray-500">
                                    {{ t("settings.authorization.show_sso_first_description") }}
                                </p>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </div>

        <!-- Save bar -->
        <div class="shrink-0 flex items-center justify-end gap-x-3 border-t bg-gray-50 px-6 py-4">
            <button
                :disabled="!hasChanges || saving"
                @click.prevent="saveSettings"
                class="rounded-md bg-indigo-600 px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-indigo-500 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600 disabled:opacity-50 disabled:cursor-not-allowed"
            >
                {{ t("common.button.save") }}
            </button>
        </div>
    </div>

    <AuthorizationOidcProviderDialog
        v-model:form="providerForm"
        :open="showProviderModal"
        :editing="isEditingProvider"
        :saving="savingProvider"
        :valid="!!isProviderFormValid"
        @close="closeProviderModal"
        @save="saveProvider"
        @copy="copyRedirectURL"
    />

    <AuthorizationOidcDeleteDialog
        :target="deleteTarget"
        @cancel="deleteTarget = null"
        @confirm="deleteProvider"
    />
</template>

<script setup>
import { t } from "@/i18n/index.js";
import { ref, reactive, computed, onMounted } from "vue";
import { useAlertStore } from "@/store/alerts";
import { extractErrorMessage } from "@/utils/errors";
import settingsService from "@/services/settingsService";
import { useSettingsStore } from "@/store/settings";
import { LockClosedIcon } from "@heroicons/vue/20/solid";
import AuthorizationLdapConfig from "@/components/Settings/Authorization/AuthorizationLdapConfig.vue";
import AuthorizationOidcProviderList from "@/components/Settings/Authorization/AuthorizationOidcProviderList.vue";
import AuthorizationOidcProviderDialog from "@/components/Settings/Authorization/AuthorizationOidcProviderDialog.vue";
import AuthorizationOidcDeleteDialog from "@/components/Settings/Authorization/AuthorizationOidcDeleteDialog.vue";

const settingsStore = useSettingsStore();
const hasOAuthLicense = computed(() => settingsStore.getLicenseFeature("oauth_providers"));
const hasLdapLicense = computed(() => settingsStore.getLicenseFeature("ldap"));

const loaded = ref(false);
const saving = ref(false);
const testingLDAP = ref(false);
const ldapTestResult = ref(null);
const syncingLDAP = ref(false);
const ldapSyncResult = ref(null);
const ldapSyncStatus = ref(null);
const oidcProviders = ref([]);

const showProviderModal = ref(false);
const isEditingProvider = ref(false);
const savingProvider = ref(false);
const editingProviderId = ref(null);
const deleteTarget = ref(null);
const preparedProviderID = ref(null);

const defaultSettings = () => ({
    localAuthEnabled: true,
    ldapEnabled: false,
    oidcEnabled: false,
    oidcFirst: false,
    ldapEnvLocked: false,
    ldap: {
        serverURL: "",
        baseDN: "",
        bindDN: "",
        bindPassword: "",
        userFilter: "(uid={0})",
        usernameAttr: "uid",
        idAttr: "",
        firstNameAttr: "givenName",
        lastNameAttr: "sn",
        emailAttr: "mail",
        allowedGroups: "",
        insecureSkipVerify: false,
        connectionSecurity: "none",
        caCertificate: "",
        timeoutSeconds: 10,
        syncEnabled: false,
        syncIntervalMinutes: 60,
        syncMaxDeactivatePercent: 20,
        host: "",
        port: 0,
    },
});

const defaultLdapPort = computed(() =>
    settings.ldap.connectionSecurity === "tls" ? "636" : "389",
);

const ldapSecurityOptions = [
    { value: "none", labelKey: "settings.authorization.ldap_security_none" },
    { value: "tls", labelKey: "settings.authorization.ldap_security_tls" },
    {
        value: "starttls",
        labelKey: "settings.authorization.ldap_security_starttls",
    },
];

// A stored empty value predates the explicit choice and means no encryption, so
// it resolves to the first option rather than leaving the control blank.
const ldapSecurityOption = computed({
    get: () =>
        ldapSecurityOptions.find((o) => o.value === settings.ldap.connectionSecurity) ??
        ldapSecurityOptions[0],
    set: (o) => {
        settings.ldap.connectionSecurity = o.value;
    },
});

const ldapUsesTls = computed(() => ["tls", "starttls"].includes(settings.ldap.connectionSecurity));

const defaultProviderForm = () => ({
    name: "",
    enabled: true,
    discovery_url: "",
    client_id: "",
    client_secret: "",
    redirect_url: "",
    scopes: "openid,profile,email",
    button_text: "",
    button_color: "#4f46e5",
});

const settings = reactive(defaultSettings());
const savedSettings = ref(JSON.stringify(defaultSettings()));
const providerForm = reactive(defaultProviderForm());

const hasChanges = computed(() => JSON.stringify(settings) !== savedSettings.value);

const isProviderFormValid = computed(
    () =>
        providerForm.name.trim() &&
        providerForm.discovery_url.trim() &&
        providerForm.client_id.trim() &&
        (isEditingProvider.value || providerForm.client_secret.trim()),
);

onMounted(async () => {
    await Promise.all([fetchSettings(), fetchOIDCProviders()]);
    loaded.value = true;
});

async function fetchSettings() {
    try {
        const response = await settingsService.getAuthSettings();
        const { ldapSyncStatus: syncStatus, ...editable } = response.data;

        Object.assign(settings, { ...defaultSettings(), ...editable });
        ldapSyncStatus.value = syncStatus ?? null;
        savedSettings.value = JSON.stringify(settings);
    } catch {}
}

async function saveSettings() {
    saving.value = true;
    try {
        await settingsService.updateAuthSettings({ ...settings });
        savedSettings.value = JSON.stringify(settings);
        useAlertStore().showSuccess(t.value("common.success.settings_updated"));
    } catch (e) {
        useAlertStore().showError(extractErrorMessage(e));
    } finally {
        saving.value = false;
    }
}

async function testLDAP() {
    testingLDAP.value = true;
    ldapTestResult.value = null;
    try {
        await settingsService.testLDAPConnection({ ...settings.ldap });
        ldapTestResult.value = {
            ok: true,
            message: t.value("settings.authorization.ldap_connected_successfully"),
        };
    } catch (e) {
        ldapTestResult.value = {
            ok: false,
            message:
                e.response?.data?.message ||
                t.value("settings.authorization.ldap_connection_failed"),
        };
    } finally {
        testingLDAP.value = false;
    }
}

async function runLDAPSync() {
    syncingLDAP.value = true;
    ldapSyncResult.value = null;
    try {
        await settingsService.runLDAPSync();
        ldapSyncResult.value = {
            ok: true,
            message: t.value("settings.authorization.ldap_sync_started"),
        };
        await waitForLDAPSync();
    } catch (e) {
        ldapSyncResult.value = {
            ok: false,
            message: extractErrorMessage(e),
        };
    } finally {
        syncingLDAP.value = false;
    }
}

const ldapSyncMessage = computed(() => {
    if (ldapSyncResult.value) {
        return ldapSyncResult.value;
    }

    const status = ldapSyncStatus.value;

    if (!status) {
        return null;
    }

    const key = (name) => `settings.authorization.${name}`;

    switch (status.status) {
        case "never":
            return { ok: true, message: t.value(key("ldap_sync_never")) };
        case "pending":
            return { ok: true, message: t.value(key("ldap_sync_queued")) };
        case "running":
            return { ok: true, message: t.value(key("ldap_syncing")) };
        case "refused":
            return {
                ok: false,
                message: t.value(key("ldap_sync_refused"), {
                    reason: status.reason,
                }),
            };
        case "failed":
        case "cancelled":
            return {
                ok: false,
                message: t.value(key("ldap_sync_failed"), {
                    reason: status.reason,
                }),
            };
        case "completed":
            if (status.stale) {
                return { ok: false, message: t.value(key("ldap_sync_stale")) };
            }

            return {
                ok: true,
                message: t.value(key("ldap_sync_finished"), {
                    deactivated: status.deactivated,
                    reactivated: status.reactivated,
                    updated: status.updated,
                }),
            };
        default:
            return null;
    }
});

// The run is queued rather than performed in the request, so the outcome is
// polled for a while; leaving the page early is safe, the status is stored.
async function fetchLDAPSyncStatus() {
    try {
        const response = await settingsService.getAuthSettings();

        ldapSyncStatus.value = response.data.ldapSyncStatus ?? null;
    } catch {}
}

async function waitForLDAPSync() {
    for (let attempt = 0; attempt < 20; attempt++) {
        await new Promise((resolve) => setTimeout(resolve, 3000));
        await fetchLDAPSyncStatus();

        const status = ldapSyncStatus.value?.status;

        if (status && status !== "pending" && status !== "running") {
            ldapSyncResult.value = null;

            return;
        }
    }
}

async function fetchOIDCProviders() {
    try {
        const response = await settingsService.getOIDCProviders();

        oidcProviders.value = response.data;
    } catch {}
}

async function openCreateProviderModal() {
    isEditingProvider.value = false;
    editingProviderId.value = null;
    preparedProviderID.value = null;
    Object.assign(providerForm, defaultProviderForm());
    showProviderModal.value = true;

    try {
        const response = await settingsService.prepareOIDCProvider();

        preparedProviderID.value = response.data.id;
        providerForm.redirect_url = response.data.redirect_url;
    } catch {
        useAlertStore().showError(t.value("settings.authorization.oidc_prepare_failed"));
        showProviderModal.value = false;
    }
}

function editProvider(provider) {
    isEditingProvider.value = true;
    editingProviderId.value = provider.id;
    preparedProviderID.value = null;
    Object.assign(providerForm, {
        name: provider.name,
        enabled: provider.enabled,
        discovery_url: provider.discovery_url,
        client_id: provider.client_id,
        client_secret: "",
        redirect_url: provider.redirect_url || "",
        scopes: provider.scopes,
        button_text: provider.button_text || "",
        button_color: provider.button_color || "#4f46e5",
    });
    showProviderModal.value = true;
}

function closeProviderModal() {
    showProviderModal.value = false;
}

async function saveProvider() {
    if (!isProviderFormValid.value) return;
    savingProvider.value = true;
    try {
        const payload = { ...providerForm };

        delete payload.redirect_url;

        if (isEditingProvider.value) {
            if (!payload.client_secret) delete payload.client_secret;
            await settingsService.updateOIDCProvider(editingProviderId.value, payload);
        } else {
            payload.id = preparedProviderID.value;
            await settingsService.createOIDCProvider(payload);
        }

        await fetchOIDCProviders();
        closeProviderModal();
        useAlertStore().showSuccess(t.value("settings.authorization.oidc_provider_saved"));
    } catch (e) {
        useAlertStore().showError(extractErrorMessage(e));
    } finally {
        savingProvider.value = false;
    }
}

async function toggleProvider(provider) {
    try {
        await settingsService.updateOIDCProvider(provider.id, {
            enabled: !provider.enabled,
        });
        provider.enabled = !provider.enabled;
    } catch (e) {
        useAlertStore().showError(extractErrorMessage(e));
    }
}

function confirmDeleteProvider(provider) {
    deleteTarget.value = provider;
}

async function deleteProvider() {
    if (!deleteTarget.value) return;
    try {
        await settingsService.deleteOIDCProvider(deleteTarget.value.id);
        oidcProviders.value = oidcProviders.value.filter((p) => p.id !== deleteTarget.value.id);
        deleteTarget.value = null;
        useAlertStore().showSuccess(t.value("settings.authorization.oidc_provider_deleted"));
    } catch (e) {
        useAlertStore().showError(extractErrorMessage(e));
    }
}

async function copyRedirectURL(url) {
    try {
        await navigator.clipboard.writeText(url);
        useAlertStore().showSuccess(t.value("settings.authorization.oidc_copied_to_clipboard"));
    } catch {
        useAlertStore().showError(t.value("settings.authorization.oidc_copy_failed"));
    }
}
</script>
