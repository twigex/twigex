# Frontend Code Guidelines

## Stack

- **Vue 3** with `<script setup>` syntax
- **Pinia** for state management
- **Vue Router 4** for routing
- **Tailwind CSS / Tailwind UI** for styling
- **Headless UI** for accessible unstyled components
- **Axios** for HTTP
- **Vite** as build tool
- **JavaScript** (no TypeScript enforcement; `.js` files)

---

## Directory Structure

```
src/
├── views/          # Page-level route components
├── components/     # Reusable UI components (subdirectories by feature)
├── store/          # Pinia stores
├── services/       # API service layer + axios instance
├── composables/    # Reusable composition functions
├── constants/      # Static constants and enums
├── utils/          # Pure helper functions
├── i18n/           # Translation files and setup
├── router/         # Vue Router config
└── assets/         # Static assets
```

**Rules:**
- Views live in `views/`, components in `components/`.
- Group related components in a subdirectory: `components/Projects/`, `components/Chat/`.
- Folders follow the app's sections (Files, Chat, Projects, Collimato, Settings). Inside a section, a subfolder holds a group of related components, such as `components/Chat/Calls/`. A folder with one file in it is flattened into its parent.
- Never put business logic directly in views — delegate to composables, stores, or services.

### Composables and utils by section

- Code with no Vue in it goes in `utils/`. Code that uses Vue reactivity, lifecycle hooks, the router or a store goes in `composables/` as `use[Feature]`.
- Use a section folder (`composables/projects/`, `utils/files/`) when only that section uses it, and the top level when several sections do. When a second section needs it, move it up instead of importing across sections.
- Extract code into one when it is reused, or when a component's script grows large and a piece of it can stand on its own. Being used once is not a reason to keep it inline.

---

## Naming Conventions

| Thing | Convention | Example |
|---|---|---|
| Component files | `PascalCase.vue` | `TaskComments.vue` |
| View files | `PascalCase.vue` | `LoginView.vue` |
| Store files | `camelCase.js` | `workspaces.js` |
| Service files | `camelCase.js` | `chatService.js` |
| Composable files | `use[Feature].js` | `useNotifications.js` |
| Constant files | `camelCase.js` | `permissions.js` |
| Util files | `camelCase.js` | `date.js` |
| JS variables & functions | `camelCase` | `currentPage`, `handleClick` |
| Primitive constants | `UPPER_SNAKE_CASE` | `DRAFT_KEY_PREFIX` |
| Object constants | `camelCase` | `channelRoles` |
| Props | `camelCase` | `isLoading`, `itemsPerPage` |
| Emitted events | `camelCase` string | `emit("rowClick")` |
| CSS custom classes | `kebab-case` | `.grid-column` |

Component names follow the [Vue style guide](https://vuejs.org/style-guide/):

- **Multi-word names**, so they never clash with an HTML element: `ChatInput`, not `Input`.
- **`Base` prefix** for the shared, app-wide building blocks: `BaseButton`, `BaseSelect`, `BaseTooltip`.
- **Parent name as a prefix** for a component that only exists as part of another, so the two sort together: `FileProgressCard` belongs to `FileProgressCards`, `ChatMessageReactions` to `ChatMessage`.

---

## Component Structure

Every component uses `<script setup>`. File block order: `<template>` → `<script setup>` → `<style>`.

Order within the script block:

1. Imports
2. `defineProps`
3. `defineEmits`
4. Store / composable instances
5. Refs and reactive state
6. Computed properties
7. Functions / event handlers
8. Lifecycle hooks

```vue
<template>
    ...
</template>

<script setup>
import { ref, computed, onMounted } from "vue";
import { t } from "@/i18n";
import { useWorkspacesStore } from "@/store/workspaces";
import workspaceService from "@/services/workspaceService";

const props = defineProps({
    workspaceId: {
        type: String,
        required: true,
    },
    isLoading: {
        type: Boolean,
        default: false,
    },
});

const emit = defineEmits(["update", "close"]);

const store = useWorkspacesStore();

const items = ref([]);
const isOpen = ref(false);

const filteredItems = computed(() => items.value.filter((i) => i.active));

async function loadItems() {
    const { data } = await workspaceService.getItems(props.workspaceId);
    items.value = data;
}

onMounted(loadItems);
</script>
```

---

## Props

Always define `type`. Always set `default` for optional props. Use `validator` for constrained string values.

```javascript
const props = defineProps({
    variant: {
        type: String,
        default: "primary",
        validator: (v) => ["primary", "secondary", "danger"].includes(v),
    },
    count: {
        type: Number,
        default: 0,
    },
    items: {
        type: Array,
        default: () => [],
    },
});
```

---

## Events

Declare all emitted events with `defineEmits`. Use `camelCase` event names.

```javascript
const emit = defineEmits(["update", "close", "rowClick"]);

function handleRowClick(row) {
    emit("rowClick", row);
}
```

Listen in templates with kebab-case: `@row-click="handler"`.

---

## Templates

Keep templates readable. Extract deeply nested or repeated blocks into sub-components.

```vue
<template>
    <!-- Conditional rendering: v-if for expensive mounts, v-show for frequent toggling -->
    <div v-if="isVisible">...</div>
    <div v-show="isToggled">...</div>

    <!-- Lists always have a unique :key -->
    <div v-for="item in items" :key="item.id">{{ item.name }}</div>

    <!-- Class binding -->
    <div
        :class="[
            'base-class font-semibold',
            isActive ? 'bg-indigo-600 text-white' : 'bg-gray-100 text-gray-700',
        ]"
    />

    <!-- Events -->
    <button @click="handleClick" @keydown.enter="submit">Submit</button>

    <!-- Named and scoped slots -->
    <slot name="header" :item="currentItem">Default header</slot>

    <!-- Portals for modals/dropdowns -->
    <Teleport to="body">
        <ModalDialog v-if="showModal" @close="showModal = false" />
    </Teleport>
</template>
```

---

## Tailwind CSS

Use utility classes directly in templates. No custom CSS unless Tailwind cannot express it.

```vue
<div class="flex items-center justify-between gap-x-4 px-3 py-2">
    <span class="truncate text-sm font-medium text-gray-900">{{ label }}</span>
    <button
        class="rounded-md bg-indigo-600 px-2.5 py-1.5 text-sm text-white shadow-sm hover:bg-indigo-500 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600"
    >
        Save
    </button>
</div>
```

Custom scoped styles only for things Tailwind cannot do (e.g., dynamic `grid-template-columns` values driven by JS):

```vue
<style scoped>
.dynamic-grid {
    grid-template-columns: v-bind(gridColumns);
}
</style>
```

---

## Pinia Stores

One store per domain. File name is the domain noun (`workspaces.js`), export is `use[Domain]Store`.

```javascript
import { defineStore } from "pinia";
import workspaceService from "@/services/workspaceService";

export const useWorkspacesStore = defineStore("workspaces", {
    state: () => ({
        workspaces: [],
        currentWorkspaceId: null,
    }),

    getters: {
        currentWorkspace: (state) =>
            state.workspaces.find((w) => w.id === state.currentWorkspaceId) ?? null,

        getWorkspaceById: (state) => (id) =>
            state.workspaces.find((w) => w.id === id) ?? null,
    },

    actions: {
        setCurrentWorkspace(id) {
            this.currentWorkspaceId = id;
        },

        async fetchWorkspaces() {
            const { data } = await workspaceService.getAll();
            this.workspaces = data;
        },
    },
});
```

**Rules:**
- Getters that need a parameter return a factory function.
- Actions call the service layer — never call axios directly from a store.

### Writing to store state

Pinia has no mutations, and writing a scalar field from a component is
supported and fine. An action that only assigns one field adds a layer and
says nothing:

```javascript
alertStore.show = false;
mediaStore.index = newIndex;
```

Use an action when the write is more than that:

- it touches a collection (push, splice, filter, map-and-reassign),
- it is a read-modify-write of state the store owns,
- it has to hold an invariant, such as ordering, deduplication or a cap,
- more than one component writes the same field.

```javascript
// Wrong: every caller repeats the reshaping, and nothing is named in devtools
channelsStore.channels = channelsStore.channels.filter((c) => c.id !== id);

// Right
channelsStore.removeChannel(id);
```

Read state through `storeToRefs` when a reactive handle is needed. Do not wrap
a store field in a hand-rolled writable `computed`: a getter that returns the
array makes `push` bypass the setter, so some writes go through the store and
some do not, and a reader cannot tell which.

---

## Services

One service file per API resource. Export a plain object (not a class) with RESTful method names.

```javascript
import axios from "./api";

const workspaceService = {
    getAll() {
        return axios.get("/workspaces");
    },

    getById(id) {
        return axios.get(`/workspaces/${id}`);
    },

    create(data) {
        return axios.post("/workspaces", data);
    },

    update(id, data) {
        return axios.put(`/workspaces/${id}`, data);
    },

    delete(id) {
        return axios.delete(`/workspaces/${id}`);
    },
};

export default workspaceService;
```

**Rules:**
- Services return the raw axios promise. The caller destructures `{ data }`.
- Do not catch errors inside services — handle them in the store or composable.
- Do not import stores inside services.

---

## Composables

Encapsulate reusable stateful logic that doesn't belong in a store. Name the file and function `use[Feature]`.

```javascript
// composables/useFileOperations.js
import { ref } from "vue";
import { t } from "@/i18n";
import fileService from "@/services/fileService";
import { useAlertStore } from "@/store/alerts";

export function useFileOperations() {
    const isUploading = ref(false);
    const alertStore = useAlertStore();

    async function uploadFile(folderId, file) {
        isUploading.value = true;
        try {
            const { data } = await fileService.upload(folderId, file);
            return data;
        } catch {
            alertStore.showError(t.value("files.error.upload_failed"));
        } finally {
            isUploading.value = false;
        }
    }

    return { isUploading, uploadFile };
}
```

**Rules:**
- Always return a plain object — never return reactive wrappers of the whole state.
- Composables may call services and stores.
- Composables may use Vue lifecycle hooks and watchers.
- If logic is only used in one component, keep it in that component.

### Never read inside an effect what that effect writes

A `watchEffect`, a `computed`, a watcher's source function and a render all
track every reactive value they read. Writing one of those values from inside
the same effect makes the effect trigger itself, and Vue stops it with
`Maximum recursive updates exceeded`.

A lazy loader is where this hides: the natural way to write one is to check the
cache, and then fill it.

```javascript
// Wrong — the effect tracks usersMap[id], then fills it
watchEffect(() => {
    if (!store.usersMap[id]) store.fetchUser(id);
});

// Right — the effect depends on the id it was given, not on the cache
watchEffect(() => {
    const wanted = toValue(id);
    queueMicrotask(() => requestUser(store, wanted));
});
```

Tracking only happens while the effect runs synchronously, so moving the work
into a microtask takes it out of the tracked scope. Reading stays reactive
through the returned `computed`, which is where it belongs.

The same applies to a store action called from an effect: it must be
idempotent. An action that rebuilds and rewrites its entries on every call
looks like a change even when nothing changed, and re-triggers whatever called
it.

When this error appears, `onRenderTriggered` names the dependency outright:

```javascript
onRenderTriggered((e) => console.log(e.type, e.key, e.target));
```

---

## Constants

Group related constants in one file. Use named exports, never default exports.

```javascript
// constants/channels.js
const channelRoles = {
    Admin: "admin",
    Member: "member",
};

const channelTypes = {
    Public: "O",
    Private: "P",
    Direct: "D",
};

const LAST_VIEWED_CHANNEL = "last_viewed_channel";

export { channelRoles, channelTypes, LAST_VIEWED_CHANNEL };
```

**Rules:**
- Object constants use `camelCase` keys and names.
- Standalone primitive constants use `UPPER_SNAKE_CASE`.
- Never hardcode magic strings in components — import from constants.

---

## Utils

Pure functions with no side effects and no Vue/store dependencies.

```javascript
// utils/date.js
export function formatDate(dateStr) {
    return new Date(dateStr).toLocaleDateString();
}

// utils/utils.js
export function convertSize(bytes) {
    const units = ["Bytes", "KB", "MB", "GB", "TB"];
    if (bytes === 0) return `0 ${units[0]}`;
    const i = Math.floor(Math.log(bytes) / Math.log(1000));
    return `${(bytes / Math.pow(1000, i)).toFixed(1)} ${units[i]}`;
}
```

**Rules:**
- Named exports only.
- No imports from `@/store`, `@/services`, or Vue reactive APIs.
- If a function needs reactive state or store access it belongs in a composable, not utils.

---

## i18n

`t` is a computed ref — call `t.value('key')` in `<script setup>` and `t('key')` directly in templates.

```javascript
import { t } from "@/i18n";

// In script setup — use t.value()
const title = computed(() => t.value("workspace.settings.title"));
alertStore.showError(t.value("tasks.error.create_failed"));

// With variable interpolation
t.value("files.items_count", { count: total });
```

```vue
<template>
    <!-- In template — call t() directly -->
    <span>{{ t("workspace.settings.title") }}</span>
    <span>{{ t("files.items_count", { count: total }) }}</span>
</template>
```

For HTML content (sanitized with DOMPurify):

```javascript
import { tHtml } from "@/i18n";
```

```vue
<p v-html="tHtml('onboarding.welcome_html', { name: user.name })" />
```

**Rules:**
- Every user-facing string must use a translation key — no raw strings in templates.
- Key structure: `[feature].[section].[label]` (e.g., `projects.task.due_date`).
- Add the key to `en.json`, then run `make sync-i18n` to create it in the other
  locales (`lv`, `kk`, `pl`), and translate it in each of them when a user will
  read it. English alone is fine for internal or admin-only diagnostics.
- Use `tHtml` only when the translation contains HTML markup.

---

## Error Handling

Handle errors at the composable or store action level, not inside services. Use `useAlertStore` to surface feedback.

```javascript
import { useAlertStore } from "@/store/alerts";
import { t } from "@/i18n";

const alertStore = useAlertStore();

async function saveTask(data) {
    try {
        const { data: task } = await taskService.create(data);
        tasks.value.push(task);
        alertStore.showSuccess(t.value("tasks.success.created"));
    } catch {
        alertStore.showError(t.value("tasks.error.create_failed"));
    }
}
```

Do not silently swallow errors. Always surface feedback to the user via `alertStore.showError` or `alertStore.showSuccess`.

---

## Headless UI

Use Headless UI components for interactive patterns (dropdowns, modals, transitions).

```vue
<template>
    <Menu as="div" class="relative">
        <MenuButton class="...">Options</MenuButton>
        <MenuItems class="absolute right-0 mt-2 w-48 rounded-md bg-white shadow-lg ring-1 ring-black/5">
            <MenuItem v-slot="{ active }">
                <button :class="[active ? 'bg-gray-100' : '', 'block w-full px-4 py-2 text-sm text-gray-700']">
                    Edit
                </button>
            </MenuItem>
        </MenuItems>
    </Menu>
</template>

<script setup>
import { Menu, MenuButton, MenuItems, MenuItem } from "@headlessui/vue";
</script>
```

---

## Icons

Use Hero Icons from `@heroicons/vue`.

```javascript
import { PencilIcon, TrashIcon } from "@heroicons/vue/24/outline";
import { CheckCircleIcon } from "@heroicons/vue/24/solid";
```

```vue
<PencilIcon class="h-5 w-5 text-gray-400" aria-hidden="true" />
```

Always set `aria-hidden="true"` on decorative icons.

---

## Formatting

Code is formatted with **Prettier**. Config lives in `frontend/.prettierrc`:

- 4-space indentation, no tabs
- Lines target 100 characters
- Everything else uses Prettier defaults

Run before committing:

```bash
npm run format
```

CI runs `make lint-frontend`, which fails if any file is not formatted or ESLint reports an error.

ESLint (`frontend/eslint.config.js`) checks for mistakes and spacing. `npm run lint` shows what it finds, and `npm run lint:fix` adds the blank lines it asks for between steps; it changes nothing else.

Never manually reformat code to fight Prettier — if a line looks odd, trust the formatter. Don't reformat code unrelated to your change.

---

## Comments

Default to writing no comments. Well-named variables, functions, and components are self-documenting.

**Only add a comment when the *why* is non-obvious:** a hidden constraint, a subtle invariant, a workaround for a specific external bug, or behaviour that would genuinely surprise a reader.

```vue
<template>
    <!-- Teleport required: this dropdown escapes an overflow:hidden ancestor -->
    <Teleport to="body">
        <DropdownMenu />
    </Teleport>
</template>
```

```javascript
// ace-builds workers use web workers; disable to avoid Vite bundling issues
editor.setOptions({ useWorker: false });
```

**Format rules:**

- Prefer one short line; use a short multi-line block only when the *why* genuinely needs more explanation.
- Plain prose only: no ASCII art, no box-drawing characters, no separator lines.
- Never use comments as section dividers or headings inside a file.

**Never do this:**

```vue
<!-- ── Roles section ── -->
<!-- ================= -->
<!-- // Filters -->
```

```javascript
// ── Lifecycle hooks ──
// =================
// Filters
// --------------------
```

**The test:** if removing the comment wouldn't confuse a future reader, don't write it. If it describes *what* the code does rather than *why*, delete it.

---

## Do / Don't

| Do                                               | Don't                                       |
| ------------------------------------------------ | ------------------------------------------- |
| Use `<script setup>`                             | Use Options API or `defineComponent`        |
| Call services from stores/composables/components | Call axios directly from components         |
| Use an action to write a collection or invariant | Reshape store collections from a component  |
| Use named exports in constants/utils             | Use default exports in constants/utils      |
| Keep components focused and small                | Build 1000+ line components                 |
| Use Tailwind utilities                           | Write component-scoped CSS unless necessary |
| Translate every user-facing string               | Hardcode UI strings in templates            |
| Handle errors and show user feedback             | Silently catch and ignore errors            |
| Group related components in subdirectories       | Dump all components at the top level        |
