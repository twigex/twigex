<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div class="space-y-4">
        <BaseCheckbox v-model="enabled" :label="t('collimato.new_connection.ssl')" />

        <div v-if="enabled" class="space-y-4 pl-7">
            <BaseSelect
                v-model="verification"
                :label="t('collimato.new_connection.ssl_verification')"
                :options="verificationOptions"
            />

            <div v-if="verification !== SSL_MODE_REQUIRE">
                <label :for="caId" class="block text-sm font-medium leading-6 text-gray-900">
                    {{ t("collimato.new_connection.ssl_ca") }}
                </label>
                <textarea
                    :id="caId"
                    :value="caCert"
                    rows="4"
                    spellcheck="false"
                    placeholder="-----BEGIN CERTIFICATE-----"
                    class="mt-2 block w-full rounded-md border-0 py-1.5 font-mono text-xs text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600"
                    @input="$emit('update:caCert', $event.target.value)"
                />
                <p class="mt-1.5 text-sm text-gray-500">
                    {{ t("collimato.new_connection.ssl_ca_help") }}
                </p>
            </div>
        </div>
    </div>
</template>

<script setup>
import { computed, useId } from "vue";
import { t } from "@/i18n/index.js";
import BaseCheckbox from "@/components/BaseCheckbox.vue";
import BaseSelect from "@/components/BaseSelect.vue";
import {
    SSL_MODE_DISABLE,
    SSL_MODE_REQUIRE,
    SSL_MODE_VERIFY_CA,
    SSL_MODE_VERIFY_FULL,
} from "@/utils/collimato/connectionSecurity";

const props = defineProps({
    sslMode: {
        type: String,
        required: true,
    },
    caCert: {
        type: String,
        default: "",
    },
});

const emit = defineEmits(["update:sslMode", "update:caCert"]);

const caId = useId();

const verificationOptions = computed(() => [
    {
        value: SSL_MODE_REQUIRE,
        label: t.value("collimato.new_connection.ssl_verify_none"),
    },
    {
        value: SSL_MODE_VERIFY_CA,
        label: t.value("collimato.new_connection.ssl_verify_ca"),
    },
    {
        value: SSL_MODE_VERIFY_FULL,
        label: t.value("collimato.new_connection.ssl_verify_full"),
    },
]);

const enabled = computed({
    get: () => props.sslMode !== SSL_MODE_DISABLE,
    set: (on) => emit("update:sslMode", on ? SSL_MODE_REQUIRE : SSL_MODE_DISABLE),
});

const verification = computed({
    get: () => props.sslMode,
    set: (mode) => emit("update:sslMode", mode),
});
</script>
