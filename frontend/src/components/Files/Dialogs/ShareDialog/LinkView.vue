<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div class="mt-5 flex flex-col gap-y-6">
        <div>
            <h3 class="text-sm font-semibold leading-6 text-gray-900">
                {{ t("files.share_dialog.link.access_level") }}
            </h3>
            <RadioGroup
                v-model="accessLevel"
                class="mt-2 grid auto-rows-fr grid-cols-1 gap-3 px-1 sm:grid-cols-2"
            >
                <RadioGroupOption
                    v-for="option in accessOptions"
                    :key="option.value"
                    :value="option.value"
                    as="template"
                    v-slot="{ checked }"
                >
                    <div
                        class="relative flex h-full cursor-pointer rounded-lg border bg-white p-4 transition-all focus:outline-none"
                        :class="
                            checked
                                ? 'border-indigo-600 ring-2 ring-indigo-600'
                                : 'border-gray-300 hover:border-gray-400'
                        "
                    >
                        <div class="flex items-start gap-x-3">
                            <component
                                :is="option.icon"
                                class="mt-0.5 size-5 shrink-0"
                                :class="checked ? 'text-indigo-600' : 'text-gray-400'"
                                aria-hidden="true"
                            />
                            <div>
                                <p
                                    class="text-sm font-semibold"
                                    :class="checked ? 'text-indigo-900' : 'text-gray-900'"
                                >
                                    {{ option.label }}
                                </p>
                                <p class="mt-0.5 text-xs text-gray-500">
                                    {{ option.description }}
                                </p>
                            </div>
                        </div>
                        <CheckCircleIcon
                            v-if="checked"
                            class="absolute right-3 top-3 size-4 text-indigo-600"
                            aria-hidden="true"
                        />
                    </div>
                </RadioGroupOption>
            </RadioGroup>
        </div>

        <div class="divide-y divide-gray-200 rounded-lg ring-1 ring-gray-200 mb-1">
            <SwitchGroup
                v-if="allowViewSelected"
                as="div"
                class="flex items-center gap-3 px-4 py-3.5"
            >
                <span
                    class="flex h-9 w-9 shrink-0 items-center justify-center rounded-lg bg-indigo-50"
                >
                    <PencilSquareIcon class="h-5 w-5 text-indigo-600" aria-hidden="true" />
                </span>
                <span class="flex min-w-0 flex-1 flex-col">
                    <SwitchLabel as="span" class="text-sm font-medium text-gray-900" passive>
                        {{ t("files.share_dialog.link.allow_editing") }}
                    </SwitchLabel>
                    <SwitchDescription as="span" class="text-sm text-gray-500">
                        {{ t("files.share_dialog.link.allow_editing_description") }}
                    </SwitchDescription>
                </span>
                <Switch
                    v-model="allowEdit"
                    :class="[
                        allowEdit ? 'bg-indigo-600' : 'bg-gray-200',
                        'relative inline-flex h-6 w-11 shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors duration-200 ease-in-out focus:outline-none focus:ring-2 focus:ring-indigo-600 focus:ring-offset-2',
                    ]"
                >
                    <span
                        aria-hidden="true"
                        :class="[
                            allowEdit ? 'translate-x-5' : 'translate-x-0',
                            'pointer-events-none inline-block size-5 transform rounded-full bg-white shadow ring-0 transition duration-200 ease-in-out',
                        ]"
                    />
                </Switch>
            </SwitchGroup>

            <SwitchGroup as="div" class="flex items-center gap-3 px-4 py-3.5">
                <span
                    class="flex h-9 w-9 shrink-0 items-center justify-center rounded-lg bg-indigo-50"
                >
                    <LockClosedIcon class="h-5 w-5 text-indigo-600" aria-hidden="true" />
                </span>
                <span class="flex min-w-0 flex-1 flex-col">
                    <SwitchLabel as="span" class="text-sm font-medium text-gray-900" passive>
                        {{ t("files.share_dialog.link.password_protected") }}
                    </SwitchLabel>
                    <SwitchDescription as="span" class="text-sm text-gray-500">
                        {{ t("files.share_dialog.link.password_protected_description") }}
                    </SwitchDescription>
                </span>
                <Switch
                    v-model="passwordProtected"
                    :class="[
                        passwordProtected ? 'bg-indigo-600' : 'bg-gray-200',
                        'relative inline-flex h-6 w-11 shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors duration-200 ease-in-out focus:outline-none focus:ring-2 focus:ring-indigo-600 focus:ring-offset-2',
                    ]"
                >
                    <span
                        aria-hidden="true"
                        :class="[
                            passwordProtected ? 'translate-x-5' : 'translate-x-0',
                            'pointer-events-none inline-block size-5 transform rounded-full bg-white shadow ring-0 transition duration-200 ease-in-out',
                        ]"
                    />
                </Switch>
            </SwitchGroup>

            <TransitionRoot
                :show="passwordProtected"
                enter="transition-opacity duration-300"
                enter-from="opacity-0"
                enter-to="opacity-100"
                leave="transition-opacity duration-0"
                leave-from="opacity-0"
                leave-to="opacity-0"
            >
                <div class="px-4 py-3.5">
                    <input
                        v-model="password"
                        type="password"
                        autocomplete="new-password"
                        :placeholder="t('files.share_dialog.link.password_placeholder')"
                        class="block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6"
                    />
                </div>
            </TransitionRoot>

            <div class="flex items-center gap-3 px-4 py-3.5">
                <span
                    class="flex h-9 w-9 shrink-0 items-center justify-center rounded-lg bg-indigo-50"
                >
                    <CalendarIcon class="h-5 w-5 text-indigo-600" aria-hidden="true" />
                </span>
                <span class="flex min-w-0 flex-col">
                    <span class="text-sm font-medium text-gray-900">
                        {{ t("files.share_dialog.link.expiry") }}
                    </span>
                    <span class="text-sm text-gray-500">
                        {{ t("files.share_dialog.link.expiry_description") }}
                    </span>
                </span>
                <div class="ml-auto flex items-center gap-2">
                    <Popover as="div" class="relative">
                        <PopoverButton
                            ref="reference"
                            class="flex w-40 items-center rounded-md bg-white py-2 pl-3 pr-10 text-left text-sm text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 focus:outline-none focus:ring-2 focus:ring-indigo-600"
                        >
                            <span
                                class="block flex-1 truncate"
                                :class="expiryDate ? 'text-gray-900' : 'text-gray-400'"
                            >
                                {{ expiryDate || t("files.share_dialog.no_expiration") }}
                            </span>
                            <span
                                class="pointer-events-none absolute inset-y-0 right-0 flex items-center pr-3"
                            >
                                <CalendarDaysIcon
                                    class="h-5 w-5 text-gray-400"
                                    aria-hidden="true"
                                />
                            </span>
                        </PopoverButton>
                        <Teleport to="body">
                            <PopoverPanel
                                ref="floating"
                                :style="floatingStyles"
                                class="z-[100] rounded-xl bg-white p-3 shadow-lg ring-1 ring-black/5"
                                v-slot="{ close }"
                            >
                                <DatePicker
                                    :modelValue="expiryDate"
                                    @update:modelValue="
                                        (val) => {
                                            expiryDate = val;
                                            close();
                                        }
                                    "
                                />
                            </PopoverPanel>
                        </Teleport>
                    </Popover>
                    <button
                        v-if="expiryDate"
                        type="button"
                        class="flex h-9 w-9 shrink-0 items-center justify-center rounded-md text-gray-400 ring-1 ring-inset ring-gray-300 hover:bg-gray-50 hover:text-gray-600"
                        :title="t('common.button.clear')"
                        @click="expiryDate = ''"
                    >
                        <XMarkIcon class="h-5 w-5" aria-hidden="true" />
                    </button>
                </div>
            </div>

            <TransitionRoot
                :show="allowDownload"
                enter="transition-opacity duration-300"
                enter-from="opacity-0"
                enter-to="opacity-100"
                leave="transition-opacity duration-0"
                leave-from="opacity-0"
                leave-to="opacity-0"
            >
                <div class="flex items-center gap-3 px-4 py-3.5">
                    <span
                        class="flex h-9 w-9 shrink-0 items-center justify-center rounded-lg bg-indigo-50"
                    >
                        <ArrowDownTrayIcon class="h-5 w-5 text-indigo-600" aria-hidden="true" />
                    </span>
                    <span class="flex min-w-0 flex-1 flex-col">
                        <span class="text-sm font-medium text-gray-900">
                            {{ t("files.share_dialog.link.download_limit") }}
                        </span>
                        <span class="text-sm text-gray-500">
                            {{ t("files.share_dialog.link.download_limit_description") }}
                        </span>
                    </span>
                    <input
                        v-model.number="maxDownloads"
                        type="number"
                        min="0"
                        class="block w-20 rounded-md border-0 py-1.5 text-center text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6"
                    />
                </div>
            </TransitionRoot>

            <div class="px-4 py-3.5">
                <label for="link-message" class="text-sm font-medium text-gray-900">
                    {{ t("files.share_dialog.link.message") }}
                </label>
                <textarea
                    id="link-message"
                    v-model="message"
                    rows="3"
                    :placeholder="t('files.share_dialog.link.message_placeholder')"
                    class="mt-2 block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6"
                />
            </div>
        </div>
    </div>
</template>

<script setup>
import { ref, computed, watch } from "vue";
import {
    RadioGroup,
    RadioGroupOption,
    Switch,
    SwitchGroup,
    SwitchLabel,
    SwitchDescription,
    TransitionRoot,
    Popover,
    PopoverButton,
    PopoverPanel,
} from "@headlessui/vue";
import { useFloating, flip, shift, offset, autoUpdate } from "@floating-ui/vue";
import {
    EyeIcon,
    ArrowDownTrayIcon,
    ArrowUpTrayIcon,
    ArrowsRightLeftIcon,
    LockClosedIcon,
    CalendarIcon,
    CalendarDaysIcon,
    XMarkIcon,
    PencilSquareIcon,
} from "@heroicons/vue/24/outline";
import { CheckCircleIcon } from "@heroicons/vue/24/solid";
import DatePicker from "@/components/DatePicker/DatePicker.vue";
import { t } from "@/i18n";

const props = defineProps({
    isFolder: {
        type: Boolean,
        default: false,
    },
    modelValue: {
        type: Object,
        default: () => ({}),
    },
});

const emit = defineEmits(["update:modelValue"]);

const reference = ref(null);
const floating = ref(null);
const { floatingStyles } = useFloating(reference, floating, {
    strategy: "fixed",
    middleware: [offset(4), flip(), shift({ padding: 8 })],
    whileElementsMounted: autoUpdate,
});

const accessLevels = {
    view: { allowView: true, allowDownload: false, allowUpload: false },
    download: { allowView: true, allowDownload: true, allowUpload: false },
    collaborate: { allowView: true, allowDownload: true, allowUpload: true },
    drop: { allowView: false, allowDownload: false, allowUpload: true },
};

function levelFromFlags(v) {
    if (v.allowUpload && !v.allowView && !v.allowDownload) return "drop";
    if (v.allowUpload) return "collaborate";
    if (v.allowDownload) return "download";

    return "view";
}

const accessLevel = ref(levelFromFlags(props.modelValue));
const passwordProtected = ref(props.modelValue.passwordProtected ?? false);
const password = ref(props.modelValue.password ?? "");
const expiryDate = ref(unixToDate(props.modelValue.expiration));
const maxDownloads = ref(props.modelValue.maxDownloads ?? 0);
const message = ref(props.modelValue.message ?? "");
const allowEdit = ref(props.modelValue.allowEdit ?? false);

const allowDownload = computed(() => accessLevels[accessLevel.value].allowDownload);

// Editing implies viewing, so the toggle is only offered on view-capable levels.
const allowViewSelected = computed(() => accessLevels[accessLevel.value].allowView);

const accessOptions = computed(() => {
    const options = [
        {
            value: "view",
            icon: EyeIcon,
            label: t.value("files.share_dialog.link.level_view"),
            description: t.value("files.share_dialog.link.level_view_description"),
        },
        {
            value: "download",
            icon: ArrowDownTrayIcon,
            label: t.value("files.share_dialog.link.level_download"),
            description: t.value("files.share_dialog.link.level_download_description"),
        },
    ];

    if (props.isFolder) {
        options.push(
            {
                value: "collaborate",
                icon: ArrowsRightLeftIcon,
                label: t.value("files.share_dialog.link.level_collaborate"),
                description: t.value("files.share_dialog.link.level_collaborate_description"),
            },
            {
                value: "drop",
                icon: ArrowUpTrayIcon,
                label: t.value("files.share_dialog.link.level_drop"),
                description: t.value("files.share_dialog.link.level_drop_description"),
            },
        );
    }

    return options;
});

function unixToDate(seconds) {
    if (!seconds) return "";

    return new Date(seconds * 1000).toISOString().slice(0, 10);
}

function dateToUnix(date) {
    if (!date) return 0;

    return Math.floor(new Date(date).getTime() / 1000);
}

watch(
    [accessLevel, passwordProtected, password, expiryDate, maxDownloads, message, allowEdit],
    () => {
        const flags = accessLevels[accessLevel.value];

        emit("update:modelValue", {
            allowView: flags.allowView,
            allowDownload: flags.allowDownload,
            allowUpload: flags.allowUpload,
            allowEdit: flags.allowView && allowEdit.value,
            passwordProtected: passwordProtected.value,
            password: password.value,
            expiration: dateToUnix(expiryDate.value),
            maxDownloads: maxDownloads.value || 0,
            message: message.value,
        });
    },
    { immediate: true },
);
</script>
