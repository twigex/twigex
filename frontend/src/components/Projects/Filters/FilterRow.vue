<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div class="flex w-full items-start gap-3">
        <!-- Field selector -->
        <Listbox v-model="filter.field" class="w-1/3">
            <div class="relative">
                <ListboxButton
                    ref="fieldButton"
                    class="fr-select w-full"
                    @mousedown="onFieldButtonDown"
                >
                    <span class="block truncate">{{ fieldLabel }}</span>
                    <ChevronUpDownIcon class="h-5 w-5 text-gray-400 flex-shrink-0" />
                </ListboxButton>
                <Teleport to="body">
                    <ListboxOptions
                        ref="fieldFloating"
                        class="fr-dropdown"
                        :style="fieldDropdownStyle"
                    >
                        <ListboxOption
                            v-for="header in filteredHeaders"
                            :key="header.name"
                            :value="header.name"
                            v-slot="{ active, selected }"
                        >
                            <li
                                :class="[
                                    'fr-option',
                                    active ? 'bg-indigo-600 text-white' : 'text-gray-900',
                                ]"
                            >
                                <span :class="['block truncate', selected && 'font-semibold']">
                                    {{ header.display_name || header.name }}
                                </span>
                                <CheckIcon
                                    v-if="selected"
                                    :class="[
                                        'h-5 w-5 flex-shrink-0',
                                        active ? 'text-white' : 'text-indigo-600',
                                    ]"
                                />
                            </li>
                        </ListboxOption>
                    </ListboxOptions>
                </Teleport>
            </div>
        </Listbox>

        <!-- Operator selector -->
        <Listbox v-model="filter.operator" class="w-1/4">
            <div class="relative">
                <ListboxButton ref="operatorButton" class="fr-select w-full">
                    <span class="block truncate">{{ operatorLabel }}</span>
                    <ChevronUpDownIcon class="h-5 w-5 text-gray-400 flex-shrink-0" />
                </ListboxButton>
                <Teleport to="body">
                    <ListboxOptions
                        ref="operatorFloating"
                        class="fr-dropdown"
                        :style="operatorDropdownStyle"
                    >
                        <ListboxOption
                            v-for="op in operators"
                            :key="op.value"
                            :value="op.value"
                            v-slot="{ active, selected }"
                        >
                            <li
                                :class="[
                                    'fr-option',
                                    active ? 'bg-indigo-600 text-white' : 'text-gray-900',
                                ]"
                            >
                                <span :class="['block truncate', selected && 'font-semibold']">{{
                                    op.label
                                }}</span>
                                <CheckIcon
                                    v-if="selected"
                                    :class="[
                                        'h-5 w-5 flex-shrink-0',
                                        active ? 'text-white' : 'text-indigo-600',
                                    ]"
                                />
                            </li>
                        </ListboxOption>
                    </ListboxOptions>
                </Teleport>
            </div>
        </Listbox>

        <!-- Value -->
        <div class="w-1/3">
            <template v-if="hideValue"> </template>

            <template v-else-if="isUserField">
                <Combobox v-model="filter.value" as="div">
                    <div ref="valueButton" class="relative">
                        <ComboboxInput
                            class="fr-select w-full pr-8"
                            :display-value="userName"
                            :placeholder="t('projects.filter_row.select_user')"
                            @change="searchPeople($event.target.value)"
                        />
                        <ComboboxButton class="absolute inset-y-0 right-0 flex items-center pr-2">
                            <ChevronUpDownIcon class="h-5 w-5 text-gray-400" aria-hidden="true" />
                        </ComboboxButton>
                    </div>
                    <Teleport to="body">
                        <ComboboxOptions
                            ref="valueFloating"
                            class="fr-dropdown"
                            :style="valueDropdownStyle"
                        >
                            <li v-if="shownUsers.length === 0" class="fr-option text-gray-500">
                                {{ t("projects.filter_row.no_people_found") }}
                            </li>
                            <ComboboxOption
                                v-for="user in shownUsers"
                                :key="user.id"
                                :value="user.id"
                                v-slot="{ active, selected }"
                                as="template"
                            >
                                <li
                                    :class="[
                                        'fr-option',
                                        active ? 'bg-indigo-600 text-white' : 'text-gray-900',
                                    ]"
                                >
                                    <span :class="['block truncate', selected && 'font-semibold']">
                                        {{ user.name }} {{ user.lastname }}
                                    </span>
                                    <CheckIcon
                                        v-if="selected"
                                        :class="[
                                            'h-5 w-5 flex-shrink-0',
                                            active ? 'text-white' : 'text-indigo-600',
                                        ]"
                                    />
                                </li>
                            </ComboboxOption>
                        </ComboboxOptions>
                    </Teleport>
                </Combobox>
            </template>

            <!-- DTR status uses a standard Listbox (small list, no virtual scroll needed) -->
            <template v-else-if="isLinkedField && route.name === 'detailed-task-report'">
                <Listbox v-model="filter.value">
                    <div class="relative">
                        <ListboxButton ref="valueButton" class="fr-select w-full">
                            <span class="block truncate">{{ linkedLabel }}</span>
                            <ChevronUpDownIcon class="h-5 w-5 text-gray-400 flex-shrink-0" />
                        </ListboxButton>
                        <Teleport to="body">
                            <ListboxOptions
                                ref="valueFloating"
                                class="fr-dropdown"
                                :style="valueDropdownStyle"
                            >
                                <ListboxOption
                                    v-for="option in workspaceStore.getStatusOptions"
                                    :key="option.id"
                                    :value="option.id"
                                    v-slot="{ active, selected }"
                                >
                                    <li
                                        :class="[
                                            'fr-option',
                                            active ? 'bg-indigo-600 text-white' : 'text-gray-900',
                                        ]"
                                    >
                                        <span
                                            :class="['block truncate', selected && 'font-semibold']"
                                            >{{ option.name }}</span
                                        >
                                        <CheckIcon
                                            v-if="selected"
                                            :class="[
                                                'h-5 w-5 flex-shrink-0',
                                                active ? 'text-white' : 'text-indigo-600',
                                            ]"
                                        />
                                    </li>
                                </ListboxOption>
                            </ListboxOptions>
                        </Teleport>
                    </div>
                </Listbox>
            </template>

            <!-- All other linked fields: custom dropdown so overflow-y doesn't conflict with virtual scroll -->
            <template v-else-if="isLinkedField">
                <div class="relative w-full">
                    <button
                        type="button"
                        class="fr-select w-full"
                        ref="valueButton"
                        @click.stop="toggleLinkedDropdown"
                    >
                        <span class="block truncate">{{ linkedLabel }}</span>
                        <ChevronUpDownIcon class="h-5 w-5 text-gray-400 flex-shrink-0" />
                    </button>
                    <Teleport to="body">
                        <div
                            v-if="linkedDropdownOpen"
                            ref="valueFloating"
                            class="fr-linked-dropdown"
                            :style="valueDropdownStyle"
                            @click.stop
                        >
                            <div class="px-1 pt-1 pb-1">
                                <input
                                    ref="linkedSearchInput"
                                    v-model="linkedSearch"
                                    type="text"
                                    :placeholder="t('projects.filter_row.search_placeholder')"
                                    class="block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6"
                                />
                            </div>
                            <div
                                v-if="linkedOptionsLoading && !filteredLinkedOptions.length"
                                class="fr-option text-gray-400 cursor-default"
                            >
                                <span class="block truncate">{{
                                    t("projects.filter_row.loading_options")
                                }}</span>
                            </div>
                            <div
                                v-else
                                ref="vsContainer"
                                class="overflow-y-auto border-t border-gray-100"
                                style="max-height: 200px"
                                @scroll="onVsScroll"
                            >
                                <div
                                    :style="{
                                        height: filteredLinkedOptions.length * VS_ITEM_H + 'px',
                                        position: 'relative',
                                    }"
                                >
                                    <div
                                        v-for="item in visibleLinkedOptions"
                                        :key="item.id"
                                        :style="{
                                            position: 'absolute',
                                            top: item._top + 'px',
                                            left: 0,
                                            right: 0,
                                            height: VS_ITEM_H + 'px',
                                        }"
                                        :class="[
                                            'fr-option cursor-pointer',
                                            item.id === filter.value
                                                ? 'bg-indigo-600 text-white'
                                                : 'text-gray-900 hover:bg-indigo-600 hover:text-white',
                                        ]"
                                        @click="selectLinkedOption(item)"
                                    >
                                        <span
                                            :class="[
                                                'block truncate',
                                                item.id === filter.value && 'font-semibold',
                                            ]"
                                            >{{ item.name }}</span
                                        >
                                        <CheckIcon
                                            v-if="item.id === filter.value"
                                            class="h-5 w-5 flex-shrink-0 text-white"
                                        />
                                    </div>
                                </div>
                            </div>
                        </div>
                    </Teleport>
                </div>
            </template>

            <template v-else-if="isDateField">
                <Listbox v-model="filter.value">
                    <div class="relative">
                        <ListboxButton ref="valueButton" class="fr-select w-full">
                            <span class="block truncate">{{ dateLabel }}</span>
                            <ChevronUpDownIcon class="h-5 w-5 text-gray-400 flex-shrink-0" />
                        </ListboxButton>
                        <Teleport to="body">
                            <ListboxOptions
                                ref="valueFloating"
                                class="fr-dropdown"
                                :style="valueDropdownStyle"
                            >
                                <ListboxOption
                                    v-for="option in dateOptions"
                                    :key="option.value"
                                    :value="option.value"
                                    v-slot="{ active, selected }"
                                >
                                    <li
                                        :class="[
                                            'fr-option',
                                            active ? 'bg-indigo-600 text-white' : 'text-gray-900',
                                        ]"
                                    >
                                        <span
                                            :class="['block truncate', selected && 'font-semibold']"
                                            >{{ option.name }}</span
                                        >
                                        <CheckIcon
                                            v-if="selected"
                                            :class="[
                                                'h-5 w-5 flex-shrink-0',
                                                active ? 'text-white' : 'text-indigo-600',
                                            ]"
                                        />
                                    </li>
                                </ListboxOption>
                            </ListboxOptions>
                        </Teleport>
                    </div>
                </Listbox>

                <input
                    v-if="needsDatePicker"
                    v-model="filter.date"
                    type="date"
                    class="mt-2 block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6"
                />

                <div v-if="needsRangePicker" class="flex gap-2 mt-2">
                    <input
                        v-model="filter.date"
                        type="date"
                        class="block w-1/2 rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6"
                        :placeholder="t('projects.filter_row.start_date')"
                    />
                    <input
                        v-model="filter.date2"
                        type="date"
                        class="block w-1/2 rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6"
                        :placeholder="t('projects.filter_row.end_date')"
                    />
                </div>
            </template>

            <template v-else>
                <input
                    v-model="filter.value"
                    type="text"
                    class="block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6"
                    :placeholder="t('projects.filter_row.enter_value')"
                />
            </template>
        </div>

        <BaseButton
            type="button"
            variant="secondary"
            class="flex-none self-start"
            color="!shadow-none text-gray-400 hover:bg-red-50 hover:text-red-600"
            :prepend-icon="TrashIcon"
            :aria-label="t('common.button.remove')"
            @click="$emit('remove')"
        />
    </div>
</template>

<script setup>
import { t } from "@/i18n/index.js";
import { computed, watch, ref, reactive, nextTick, onMounted, onUnmounted } from "vue";
import { useAnchoredPopup } from "@/composables/useAnchoredPopup";
import { storeToRefs } from "pinia";
import { useWorkspaceStore } from "@/store/workspaces";
import { useUserStore } from "@/store/user";
import { useRoute } from "vue-router";
import workspaceService from "@/services/workspaceService";
import { useAlertStore } from "@/store/alerts";
import { extractErrorMessage } from "@/utils/errors";
import {
    Combobox,
    ComboboxButton,
    ComboboxInput,
    ComboboxOption,
    ComboboxOptions,
    Listbox,
    ListboxButton,
    ListboxOptions,
    ListboxOption,
} from "@headlessui/vue";
import { ChevronUpDownIcon, CheckIcon, TrashIcon } from "@heroicons/vue/20/solid";
import BaseButton from "@/components/BaseButton.vue";

const workspaceStore = useWorkspaceStore();
const userStore = useUserStore();
const { getWorkspaceTablesForLinked, getStatusOptions } = storeToRefs(workspaceStore);
const route = useRoute();

// Each Listbox has its own popup, anchored to its own button, so opening one
// never moves another and it opens in place however it was opened.
const fieldButton = ref(null);
const operatorButton = ref(null);
const valueButton = ref(null);
const { floating: fieldFloating, floatingStyles: fieldDropdownStyle } = useAnchoredPopup({
    matchWidth: true,
    anchor: fieldButton,
});
const { floating: operatorFloating, floatingStyles: operatorDropdownStyle } = useAnchoredPopup({
    matchWidth: true,
    anchor: operatorButton,
});
const { floating: valueFloating, floatingStyles: valueDropdownStyle } = useAnchoredPopup({
    matchWidth: true,
    anchor: valueButton,
});

const onFieldButtonDown = () => {
    fieldUserInteracting.value = true;
};

const props = defineProps({
    filter: { type: Object, required: true },
    users: { type: Array, default: () => [] },
    // Finds people on the server by what is typed; without it the people
    // given are filtered here.
    searchUsers: { type: Function, default: null },
});

defineEmits(["remove"]);

const peopleQuery = ref("");
const searchedUsers = ref(null);
let peopleSearch = null;

function searchPeople(query) {
    peopleQuery.value = query;
    clearTimeout(peopleSearch);
    const text = query.trim();

    if (!text) {
        searchedUsers.value = null;

        return;
    }

    if (!props.searchUsers) {
        const needle = text.toLowerCase();

        searchedUsers.value = props.users.filter((u) =>
            `${u.name} ${u.lastname}`.toLowerCase().includes(needle),
        );

        return;
    }

    peopleSearch = setTimeout(async () => {
        try {
            const found = await props.searchUsers(text);

            if (peopleQuery.value === query) searchedUsers.value = found;
        } catch {
            if (peopleQuery.value === query) searchedUsers.value = [];
        }
    }, 250);
}

const shownUsers = computed(() => searchedUsers.value ?? props.users);

function userName(id) {
    if (!id) return "";
    const u =
        (searchedUsers.value || []).find((u) => u.id === id) ||
        props.users.find((u) => u.id === id) ||
        userStore.getUserById(id);

    return u ? `${u.name} ${u.lastname}` : id;
}

watch(
    () => props.filter.value,
    () => {
        peopleQuery.value = "";
        searchedUsers.value = null;
    },
);

const tableHeaders = computed({
    get() {
        return workspaceStore.getTableHeaders;
    },
    set(value) {
        workspaceStore.setTableHeaders(value);
    },
});

const filteredHeaders = computed(() =>
    (tableHeaders.value || []).filter(
        (h) => h.name !== "workspace_name" && h.name !== "table_name" && h.name !== "link_to_table",
    ),
);

const operators = computed(() => [
    { value: "is", label: t.value("projects.filter_row.is") },
    { value: "is_not", label: t.value("projects.filter_row.is_not") },
    { value: "is_set", label: t.value("projects.filter_row.is_set") },
    { value: "is_not_set", label: t.value("projects.filter_row.is_not_set") },
]);

const fieldLabel = computed(() => {
    if (!props.filter.field) return t.value("projects.filter_row.where");
    const h = filteredHeaders.value.find((h) => h.name === props.filter.field);

    return h ? h.display_name || h.name : props.filter.field;
});

const operatorLabel = computed(() => {
    const op = operators.value.find((o) => o.value === props.filter.operator);

    return op ? op.label : props.filter.operator || "—";
});

const linkedLabel = computed(() => {
    if (!props.filter.value) return t.value("projects.filter_row.select_option");
    if (linkedTableId.value && resolvedNames[props.filter.value])
        return resolvedNames[props.filter.value];
    // Show loading text while options are being fetched so the user sees a clear state.
    if (linkedOptionsLoading.value) return t.value("projects.filter_row.loading_options");
    const opts =
        route.name === "detailed-task-report"
            ? workspaceStore.getStatusOptions
            : linkedOptions.value;
    const fv = props.filter.value;
    const o = (opts || []).find((o) => {
        if (Array.isArray(o.id)) {
            return Array.isArray(fv) ? o.id[0] === fv[0] : o.id.includes(fv);
        }

        return o.id === fv;
    });

    if (o) return o.name;
    if (linkedTableId.value && resolvedNames[fv]) return resolvedNames[fv];

    return Array.isArray(fv) ? fv[0] : fv;
});

const dateLabel = computed(() => {
    if (!props.filter.value) return t.value("projects.filter_row.choose_value");
    const o = dateOptions.find((o) => o.value === props.filter.value);

    return o ? o.name : props.filter.value;
});

const linkedOptions = ref([]);
const linkedOptionsLoading = ref(false);
// True only when the user physically clicked the field dropdown; prevents
// programmatic filter updates (fast filters) from triggering value clear.
const fieldUserInteracting = ref(false);

// Custom linked dropdown state (not using HeadlessUI Listbox; avoids overflow conflict)
const linkedDropdownOpen = ref(false);
const linkedSearch = ref("");
const vsContainer = ref(null);
const linkedSearchInput = ref(null);
const vsScrollTop = ref(0);
const VS_ITEM_H = 36;

function toggleLinkedDropdown() {
    if (linkedDropdownOpen.value) {
        linkedDropdownOpen.value = false;
    } else {
        const searched = linkedSearch.value !== "";

        linkedSearch.value = "";
        vsScrollTop.value = 0;
        linkedDropdownOpen.value = true;
        if (!searched && !linkedOptions.value.length) loadLinkedPage(true);
        nextTick(() => linkedSearchInput.value?.focus());
    }
}

function handleOutsideClick() {
    linkedDropdownOpen.value = false;
}

onMounted(() => document.addEventListener("click", handleOutsideClick));
onUnmounted(() => document.removeEventListener("click", handleOutsideClick));

const filteredLinkedOptions = computed(() => {
    if (linkedTableId.value) return linkedOptions.value;
    const q = linkedSearch.value.trim().toLowerCase();

    return q
        ? linkedOptions.value.filter((o) => o.name?.toLowerCase().includes(q))
        : linkedOptions.value;
});

const visibleLinkedOptions = computed(() => {
    const items = filteredLinkedOptions.value;

    if (!items.length) return [];
    const containerH = 192;
    const start = Math.max(0, Math.floor(vsScrollTop.value / VS_ITEM_H) - 3);
    const end = Math.min(items.length, Math.ceil((vsScrollTop.value + containerH) / VS_ITEM_H) + 3);

    return items.slice(start, end).map((o, i) => ({ ...o, _top: (start + i) * VS_ITEM_H }));
});

function onVsScroll(e) {
    vsScrollTop.value = e.target.scrollTop;
    if (e.target.scrollTop + e.target.clientHeight >= e.target.scrollHeight - VS_ITEM_H * 5) {
        loadLinkedPage(false);
    }
}

function selectLinkedOption(option) {
    resolvedNames[option.id] = option.name;
    props.filter.value = option.id;
}

// A linked table can hold a great many rows, so its records are searched
// and paged on the server; status options are few and come from the store.
const LINKED_PAGE = 50;
const linkedTableId = ref("");
const linkedHasMore = ref(true);
const resolvedNames = reactive({});
let linkedRequest = 0;

async function loadLinkedPage(reset) {
    const tableId = linkedTableId.value;

    if (!tableId) return;
    if (!reset && (linkedOptionsLoading.value || !linkedHasMore.value)) return;

    const seq = ++linkedRequest;

    if (reset) {
        linkedOptions.value = [];
        linkedHasMore.value = true;
    }

    linkedOptionsLoading.value = true;
    try {
        const res = await workspaceService.getLinkedRecordsLite(route.params.id, tableId, {
            q: linkedSearch.value.trim(),
            limit: LINKED_PAGE,
            offset: linkedOptions.value.length,
        });

        if (seq !== linkedRequest) return;
        const page = (res.data || []).map((item) => ({ id: item.id, name: item.name }));

        linkedOptions.value = [...linkedOptions.value, ...page];
        linkedHasMore.value = page.length === LINKED_PAGE;
    } catch (err) {
        if (seq === linkedRequest) useAlertStore().showError(extractErrorMessage(err));
    } finally {
        if (seq === linkedRequest) linkedOptionsLoading.value = false;
    }
}

async function resolveLinkedName(tableId, id) {
    if (!tableId || !id || typeof id !== "string" || resolvedNames[id]) return;
    try {
        const res = await workspaceService.getLinkedRecordsLite(route.params.id, tableId, {
            ids: id,
        });
        const row = (res.data || [])[0];

        if (row) resolvedNames[id] = row.name;
    } catch {
        // The raw id stays as the label.
    }
}

let linkedSearchTimer = null;

watch(linkedSearch, () => {
    if (!linkedTableId.value || !linkedDropdownOpen.value) return;
    clearTimeout(linkedSearchTimer);
    linkedSearchTimer = setTimeout(() => {
        vsScrollTop.value = 0;
        loadLinkedPage(true);
    }, 250);
});
onUnmounted(() => clearTimeout(linkedSearchTimer));
onUnmounted(() => clearTimeout(peopleSearch));

const dateOptionsRequiringDatePicker = ["Exact date", "Before date", "After date"];

const dateOptionsRequiringRange = ["Date range"];

const needsDatePicker = computed(() => {
    if (!isDateField.value) return false;

    return dateOptionsRequiringDatePicker.includes(props.filter.value);
});

const needsRangePicker = computed(() => {
    if (!isDateField.value) return false;

    return dateOptionsRequiringRange.includes(props.filter.value);
});

const isUserField = computed(() => {
    const header = tableHeaders.value.find((h) => h.name === props.filter.field);

    if (!header) return false;

    return ["default_assignee", "assignee"].includes(header.header_usage);
});

const isLinkedField = computed(() => {
    const header = tableHeaders.value.find((h) => h.name === props.filter.field);

    if (!header) return false;

    if (route.name === "detailed-task-report" && header.name === "status") {
        return true;
    }

    if (header.header_usage === "status") return true;

    return !!header.linked_id || !!header.parent_table_id;
});

watch(
    // Watch all three sources so FilterRow re-resolves when any of them change.
    () => [props.filter.field, getWorkspaceTablesForLinked.value, getStatusOptions.value],
    async ([newField], [oldField] = []) => {
        // Clear value only when the user physically picked a different field.
        // Only reset fieldUserInteracting when the field actually changed; if the
        // watch fires due to statusOptions/tables loading, leave the flag alone so
        // a concurrent field-dropdown interaction still triggers the clear.
        if (oldField !== newField) {
            if (fieldUserInteracting.value && oldField) {
                props.filter.value = "";
                props.filter.operator = "is";
            }

            fieldUserInteracting.value = false;
            linkedSearch.value = "";
            vsScrollTop.value = 0;
        }

        if (route.name === "detailed-task-report" && newField === "status") {
            linkedTableId.value = "";

            return;
        }

        const header = tableHeaders.value?.find((h) => h.name === newField);

        // Status fields: read directly from the store (preloaded by the view's loadData).
        // The watch source includes getStatusOptions so this re-fires when it populates.
        // DTR is handled by the early return above.
        if (header?.header_usage === "status") {
            linkedTableId.value = "";
            const stored = getStatusOptions.value || [];

            linkedOptions.value = stored.map((o) => ({
                id: Array.isArray(o.id) ? o.id[0] : o.id,
                name: o.name,
            }));
            linkedOptionsLoading.value = !stored.length;

            return;
        }

        if (header && (header.parent_table_id || header.linked_id)) {
            const tableId =
                header.single_select && header.linked_id
                    ? header.linked_id
                    : header.parent_table_id || header.linked_id;

            if (oldField !== newField || linkedTableId.value !== tableId) {
                linkedRequest++;
                linkedOptions.value = [];
                linkedHasMore.value = true;
                linkedOptionsLoading.value = false;
            }

            linkedTableId.value = tableId;
            resolveLinkedName(tableId, props.filter.value);
            if (linkedDropdownOpen.value && !linkedOptions.value.length) loadLinkedPage(true);
        } else {
            linkedTableId.value = "";
            linkedOptions.value = [];
        }
    },
    { immediate: true },
);

const isDateField = computed(() => ["start_date", "due_date"].includes(props.filter.field));

const hideValue = computed(() => ["is_set", "is_not_set"].includes(props.filter.operator));

const dateOptions = [
    { name: t.value("projects.filter_row.today"), value: "Today" },
    { name: t.value("projects.filter_row.yesterday"), value: "Yesterday" },
    { name: t.value("projects.filter_row.tomorrow"), value: "Tomorrow" },
    { name: t.value("projects.filter_row.next_7_days"), value: "Next 7 days" },
    { name: t.value("projects.filter_row.last_7_days"), value: "Last 7 days" },
    { name: t.value("projects.filter_row.last_week"), value: "Last week" },
    { name: t.value("projects.filter_row.this_week"), value: "This week" },
    { name: t.value("projects.filter_row.next_week"), value: "Next week" },
    { name: t.value("projects.filter_row.last_month"), value: "Last month" },
    { name: t.value("projects.filter_row.this_month"), value: "This month" },
    { name: t.value("projects.filter_row.next_month"), value: "Next month" },
    {
        name: t.value("projects.filter_row.today_and_earlier"),
        value: "Today & Earlier",
    },
    {
        name: t.value("projects.filter_row.last_quarter"),
        value: "Last quarter",
    },
    {
        name: t.value("projects.filter_row.next_quarter"),
        value: "Next quarter",
    },
    { name: t.value("projects.filter_row.overdue"), value: "Overdue" },
    {
        name: t.value("projects.filter_row.later_than_today"),
        value: "Later than Today",
    },
    { name: t.value("projects.filter_row.next_year"), value: "Next year" },
    { name: t.value("projects.filter_row.this_year"), value: "This year" },
    { name: t.value("projects.filter_row.last_year"), value: "Last year" },
    { name: t.value("projects.filter_row.exact_date"), value: "Exact date" },
    { name: t.value("projects.filter_row.before_date"), value: "Before date" },
    { name: t.value("projects.filter_row.after_date"), value: "After date" },
    { name: t.value("projects.filter_row.date_range"), value: "Date range" },
];

watch(
    () => props.filter.value,
    (newValue) => {
        if (isDateField.value) {
            if (!needsDatePicker.value && !needsRangePicker.value) {
                props.filter.date = null;
            }

            if (!dateOptionsRequiringDatePicker.includes(newValue)) {
                props.filter.date = null;
            }
        }
    },
    { immediate: true },
);
</script>

<style scoped>
/* Tailwind UI's Listbox, at its regular size. */
.fr-select {
    display: inline-flex;
    align-items: center;
    justify-content: space-between;
    gap: 8px;
    height: 36px;
    padding: 6px 8px 6px 12px;
    font-size: 14px;
    line-height: 24px;
    color: #111827;
    border-radius: 6px;
    background: #fff;
    cursor: default;
    box-shadow:
        inset 0 0 0 1px #d1d5db,
        0 1px 2px 0 rgb(0 0 0 / 0.05);
    text-align: left;
    white-space: nowrap;
    overflow: hidden;
}
.fr-select:focus {
    outline: none;
    box-shadow: inset 0 0 0 2px #4f46e5;
}
.fr-dropdown {
    z-index: 60;
    max-height: 240px;
    overflow-y: auto;
    border-radius: 6px;
    background: #fff;
    padding: 4px 0;
    font-size: 14px;
    box-shadow:
        0 10px 15px -3px rgb(0 0 0 / 0.1),
        0 4px 6px -4px rgb(0 0 0 / 0.1),
        0 0 0 1px rgb(0 0 0 / 0.05);
}
.fr-option {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 8px 12px;
    cursor: default;
    user-select: none;
    list-style: none;
}
.fr-linked-dropdown {
    z-index: 60;
    border-radius: 6px;
    background: #fff;
    padding: 0;
    font-size: 14px;
    box-shadow:
        0 10px 15px -3px rgb(0 0 0 / 0.1),
        0 4px 6px -4px rgb(0 0 0 / 0.1),
        0 0 0 1px rgb(0 0 0 / 0.05);
    min-width: 160px;
}
</style>
