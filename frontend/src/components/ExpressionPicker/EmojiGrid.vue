<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div class="pb-2">
        <div class="flex flex-1 justify-center lg:justify-end w-full">
            <div class="grid w-full">
                <input
                    v-model="searchQuery"
                    type="text"
                    name="search"
                    aria-label="Search"
                    class="col-start-1 row-start-1 block w-full rounded-md bg-white py-1.5 pr-3 pl-10 text-base text-gray-600 sm:text-sm/6"
                    :placeholder="t('channels.emoji_picker.search_placeholder')"
                />
                <MagnifyingGlassIcon
                    class="pointer-events-none col-start-1 row-start-1 ml-3 size-5 self-center text-gray-600"
                    aria-hidden="true"
                />
            </div>
        </div>
    </div>

    <!-- Category Navigation -->
    <div class="flex-shrink-0 h-10">
        <nav class="isolate flex divide-x divide-gray-100 rounded-lg">
            <a
                v-if="recents.length > 0"
                :key="'Recents'"
                :class="[
                    'group relative min-w-0 flex-1 overflow-hidden px-2 py-2 text-center text-sm font-medium hover:bg-gray-100 cursor-pointer rounded',
                    currentVisibleCategory === 'Recents' ? 'bg-indigo-100' : 'bg-white',
                ]"
                @click="scrollToCategory('Recents')"
            >
                <span>🕘</span>
            </a>

            <a
                v-for="(icon, slug) in categoryEmoji"
                :key="slug"
                :class="[
                    'group relative min-w-0 flex-1 overflow-hidden px-2 py-2 text-center text-sm font-medium hover:bg-gray-100 cursor-pointer rounded',
                    currentVisibleCategory === icon.name ? 'bg-indigo-100' : 'bg-white',
                ]"
                @click="scrollToCategory(icon.name)"
            >
                <span>{{ icon.icon }}</span>
            </a>
        </nav>
    </div>

    <!-- Emoji Container -->
    <div class="flex flex-col h-full min-h-0">
        <div class="flex-1 overflow-y-auto relative" ref="scrollContainer" @scroll="handleScroll">
            <div v-if="searchQuery" class="space-y-4 flex flex-col h-full">
                <div
                    v-if="searchedEmojis.length > 0"
                    :key="'Recents'"
                    :data-category="'Recents'"
                    class="category-section"
                >
                    <div class="text-sm text-gray-500 py-2 sticky top-0 bg-white z-10">
                        {{ t("channels.emoji_picker.category.search") }}
                    </div>
                    <div class="grid grid-cols-8 gap-1">
                        <div
                            v-for="emoji in searchedEmojis"
                            :key="emoji.short_name"
                            class="flex flex-col items-center justify-center cursor-pointer p-1 hover:bg-gray-100 rounded"
                            @click="selectEmoji(emoji)"
                            :title="`:${emoji.short_name}:`"
                        >
                            <div :style="emojiStyle(emoji)" />
                        </div>
                    </div>
                </div>
                <div class="flex flex-col h-full" v-else>
                    <div
                        class="text-center flex flex-col h-full w-full justify-center items-center"
                    >
                        <MagnifyingGlassIcon
                            class="mx-auto h-12 w-12 text-gray-400"
                            aria-hidden="true"
                        />

                        <h3 class="mt-2 text-sm font-semibold text-gray-600">
                            {{ t("channels.emoji_picker.no_results") }}
                        </h3>
                        <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
                            {{ t("channels.emoji_picker.try_searching") }}
                        </p>
                    </div>
                </div>
            </div>

            <div v-if="!searchQuery" class="space-y-4">
                <div
                    v-if="recents.length > 0"
                    :key="'Recents'"
                    :data-category="'Recents'"
                    class="category-section"
                >
                    <div class="text-sm text-gray-500 py-2 sticky top-0 bg-white z-10">
                        {{ t("channels.emoji_picker.categories.recents") }}
                    </div>
                    <div class="grid grid-cols-8 gap-1">
                        <div
                            v-for="emoji in recentEmojis"
                            :key="emoji.short_name"
                            class="flex flex-col items-center justify-center cursor-pointer p-1 hover:bg-gray-100 rounded"
                            @click="selectEmoji(emoji)"
                            :title="`:${emoji.short_name}:`"
                        >
                            <div :style="emojiStyle(emoji)" />
                        </div>
                    </div>
                </div>

                <div
                    v-for="category in visibleCategories"
                    :key="category.name"
                    :data-category="category.name"
                    class="category-section"
                >
                    <div class="text-sm text-gray-500 py-2 sticky top-0 bg-white z-10">
                        {{ category.displayName }}
                    </div>
                    <div class="grid grid-cols-8 gap-1">
                        <div
                            v-for="emoji in emojisInCategory(category.name)"
                            :key="emoji.short_name"
                            class="flex flex-col items-center justify-center cursor-pointer p-1 hover:bg-gray-100 rounded"
                            @click="selectEmoji(emoji)"
                            :title="`:${emoji.short_name}:`"
                        >
                            <div :style="emojiStyle(emoji)" />
                        </div>
                    </div>
                </div>
            </div>
        </div>
    </div>
</template>

<script setup>
import { t } from "@/i18n/index.js";

import { ref, onMounted, nextTick, computed } from "vue";
import { emojiByShortName, emojiStyle, emojisInCategory, searchableEmojis } from "@/utils/emoji";
import { MagnifyingGlassIcon } from "@heroicons/vue/24/outline";

const categoryEmoji = {
    Smileys: {
        name: "Smileys & Emotion",
        displayName: t.value("channels.emoji_picker.categories.smileys"),
        icon: "😀",
    },
    People: {
        name: "People & Body",
        displayName: t.value("channels.emoji_picker.categories.people"),
        icon: "👨",
    },
    Animals: {
        name: "Animals & Nature",
        displayName: t.value("channels.emoji_picker.categories.animals"),
        icon: "🐶",
    },
    Food: {
        name: "Food & Drink",
        displayName: t.value("channels.emoji_picker.categories.food"),
        icon: "🍔",
    },
    Travel: {
        name: "Travel & Places",
        displayName: t.value("channels.emoji_picker.categories.travel"),
        icon: "🚗",
    },
    Activities: {
        name: "Activities",
        displayName: t.value("channels.emoji_picker.categories.activities"),
        icon: "🎳",
    },
    Objects: {
        name: "Objects",
        displayName: t.value("channels.emoji_picker.categories.objects"),
        icon: "📦",
    },
    Symbols: {
        name: "Symbols",
        displayName: t.value("channels.emoji_picker.categories.symbols"),
        icon: "❤️",
    },
    Flags: {
        name: "Flags",
        displayName: t.value("channels.emoji_picker.categories.flags"),
        icon: "🏳️",
    },
};

const categories = Object.values(categoryEmoji);
const renderedCategories = ref(1);
const visibleCategories = computed(() => categories.slice(0, renderedCategories.value));

const recents = ref(JSON.parse(localStorage.getItem("recentEmojis") || "[]"));
const searchQuery = ref("");
const currentVisibleCategory = ref(
    localStorage.getItem("recentEmojis") && localStorage.getItem("recentEmojis").length > 0
        ? "Recents"
        : categoryEmoji["Smileys"].name,
);
const scrollContainer = ref(null);
const categoryPositions = ref({});
const emojiPickerRef = ref(null);
const emits = defineEmits(["select-emoji", "close"]);

const searchedEmojis = computed(() => {
    if (!searchQuery.value) {
        return;
    }

    const query = searchQuery.value.toLowerCase();

    return searchableEmojis.filter(
        (emoji) =>
            emoji.short_name.includes(query) ||
            (emoji.name && emoji.name.toLowerCase().includes(query)) ||
            (emoji.keywords && emoji.keywords.some((kw) => kw.toLowerCase().includes(query))),
    );
});

const measureCategoryPositions = () => {
    const container = scrollContainer.value;

    if (!container) {
        return;
    }

    const sections = container.querySelectorAll(".category-section");
    const containerRect = container.getBoundingClientRect();

    sections.forEach((section) => {
        const rect = section.getBoundingClientRect();

        categoryPositions.value[section.dataset.category] =
            rect.top - containerRect.top + container.scrollTop;
    });

    handleScroll();
};

// All nine categories at once is ~3800 nodes in one synchronous render, which
// is what delays the first paint.
onMounted(() => {
    const total = categories.length;

    const renderNext = () => {
        if (renderedCategories.value < total) {
            renderedCategories.value++;
            requestAnimationFrame(renderNext);

            return;
        }

        nextTick(measureCategoryPositions);
    };

    requestAnimationFrame(renderNext);

    //add click outside listener
    document.addEventListener("click", handleClickOutside);
});

const handleClickOutside = (event) => {
    if (emojiPickerRef.value && !emojiPickerRef.value.contains(event.target)) {
        emits("close");
    }
};

const handleScroll = () => {
    const scrollPosition = scrollContainer.value.scrollTop + 10; // Adding slight offset

    for (const [category, position] of Object.entries(categoryPositions.value)) {
        const nextPosition = getNextCategoryPosition(category);

        if (
            scrollPosition >= position &&
            (nextPosition === undefined || scrollPosition < nextPosition)
        ) {
            if (currentVisibleCategory.value !== category) {
                currentVisibleCategory.value = category;
            }

            break;
        }
    }
};

const getNextCategoryPosition = (currentCategory) => {
    const names = categories.map((c) => c.name);
    const currentIndex = names.indexOf(currentCategory);

    if (currentIndex < names.length - 1) {
        return categoryPositions.value[names[currentIndex + 1]];
    }

    return undefined;
};

const scrollToCategory = (category) => {
    //If clicked on recents then scroll to top
    if (category === "Recents") {
        scrollContainer.value.scrollTo({
            top: 0,
            behavior: "smooth",
        });

        return;
    }

    const position = categoryPositions.value[category];

    if (position !== undefined) {
        scrollContainer.value.scrollTo({
            top: position,
            behavior: "smooth",
        });
    }
};

const selectEmoji = (emoji) => {
    // add to recents. push to start of array, remove if already exists
    const index = recents.value.indexOf(emoji.short_name);

    if (index !== -1) {
        recents.value.splice(index, 1);
    }

    recents.value.unshift(emoji.short_name);
    if (recents.value.length > 20) {
        recents.value.pop();
    }

    localStorage.setItem("recentEmojis", JSON.stringify(recents.value));

    emits("select-emoji", `:${emoji.short_name}:`);
};

const recentEmojis = computed(() => recents.value.map(emojiByShortName).filter(Boolean));
</script>

<style scoped>
.category-section {
    scroll-margin-top: 50px; /* Ensures space for sticky header */
}
</style>
