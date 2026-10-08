<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div class="pb-2">
        <div class="flex flex-1 items-center justify-center lg:justify-end w-full gap-x-1">
            <ArrowLeftIcon
                v-if="trendingSelected"
                @click="clearData()"
                class="size-6 text-gray-600 cursor-pointer hover:text-gray-800"
                aria-hidden="true"
            />
            <div class="relative w-full">
                <!-- search input -->
                <input
                    v-model="searchQuery"
                    type="text"
                    name="search"
                    aria-label="Search"
                    class="block w-full rounded-md bg-white py-1.5 pr-10 pl-10 text-base text-gray-600 sm:text-sm/6"
                    :placeholder="t('channels.gif_picker.search_placeholder')"
                />

                <!-- left icon -->
                <MagnifyingGlassIcon
                    class="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 size-5 text-gray-600"
                    aria-hidden="true"
                />

                <!-- clear icon (X) -->
                <XMarkIcon
                    v-if="searchQuery"
                    @click="clearData()"
                    class="absolute right-3 top-1/2 -translate-y-1/2 size-5 text-gray-600 cursor-pointer hover:text-gray-800"
                    aria-hidden="true"
                />
            </div>
        </div>
    </div>

    <div @scroll="onScroll" class="h-full overflow-y-auto">
        <div
            v-if="searchQuery.length > 0 || trendingSelected"
            class="flex flex-row h-full gap-1 px-2"
        >
            <div class="flex flex-col gap-y-1 h-full w-1/2">
                <div
                    @click="selectGif(gif)"
                    v-for="gif in leftColumn"
                    :key="gif.id"
                    class="cursor-pointer"
                >
                    <video
                        :src="gif.media_formats.nanomp4.url"
                        :poster="gif.media_formats.gifpreview.url || ''"
                        autoplay
                        loop
                        muted
                        playsinline
                        preload="metadata"
                        class="w-full object-contain rounded-md border-2 border-transparent hover:border-indigo-500"
                    ></video>
                </div>
            </div>

            <div class="flex flex-col gap-y-1 h-60 w-1/2">
                <div
                    @click="selectGif(gif)"
                    v-for="gif in rightColumn"
                    :key="gif.id"
                    class="cursor-pointer"
                >
                    <video
                        :src="gif.media_formats.nanomp4.url"
                        :poster="gif.media_formats.gifpreview.url || ''"
                        autoplay
                        loop
                        muted
                        playsinline
                        preload="metadata"
                        class="w-full object-contain rounded-md border-2 border-transparent hover:border-indigo-500"
                    ></video>
                </div>
            </div>
        </div>
        <div v-else class="columns-2 gap-1 px-2">
            <div @click="selectTrending()" class="mb-1 break-inside-avoid cursor-pointer">
                <div
                    class="relative text-center text-sm text-gray-500 mb-1 h-20 rounded-md overflow-hidden hover:border-2 hover:border-indigo-500"
                >
                    <video
                        :src="trendingUrl"
                        autoplay
                        loop
                        muted
                        playsinline
                        preload="metadata"
                        class="rounded-md w-full object-contain"
                    ></video>

                    <div
                        class="absolute inset-0 bg-black/40 text-white flex items-center justify-center text-xs font-semibold"
                    >
                        {{ t("channels.gif_picker.trending") }}
                    </div>
                </div>
            </div>
            <div
                @click="selectCategory(cat)"
                v-for="cat in categories"
                :key="cat.name"
                class="mb-1 break-inside-avoid cursor-pointer"
            >
                <div
                    class="relative text-center text-sm text-gray-500 mb-1 h-20 rounded-md overflow-hidden hover:border-2 hover:border-indigo-500"
                >
                    <img
                        :src="cat.image"
                        alt="GIF Category"
                        class="w-full h-full object-cover rounded-md"
                    />

                    <!-- overlay -->
                    <div
                        class="absolute inset-0 bg-black/40 text-white flex items-center justify-center text-xs font-semibold"
                    >
                        {{ cat.name }}
                    </div>
                </div>
            </div>
        </div>
    </div>
</template>

<script setup>
import { t } from "@/i18n/index.js";

import { onMounted, ref, watch, computed } from "vue";
import chatService from "@/services/chatService";
import { MagnifyingGlassIcon, XMarkIcon, ArrowLeftIcon } from "@heroicons/vue/24/outline";

const emits = defineEmits(["select-gif", "close"]);

const leftColumn = computed(() => gifs.value.filter((_, i) => i % 2 === 0));

const rightColumn = computed(() => gifs.value.filter((_, i) => i % 2 !== 0));

const searchQuery = ref("");
const gifs = ref([]);
const categories = ref([]);
const next = ref(null);
const trendingUrl = ref("");
const trendingSelected = ref(false);
const locked = ref(false);
const timeout = ref(null);

function selectGif(gif) {
    emits("select-gif", gif.id);
    emits("close");

    clearData();
}

function selectCategory(category) {
    searchQuery.value = category.searchterm;
    // You can emit an event or handle the category selection as needed
}

function selectTrending() {
    trendingSelected.value = true;
    searchQuery.value = "";
    gifs.value = [];

    const trendingResponse = loadTrendingGifs(20);

    trendingResponse.then((response) => {
        gifs.value = response.data.results;
        next.value = response.data.next;
    });
}

async function loadTrendingGifs(limit = 1) {
    const trendingResponse = await chatService.getKlipyTrending(limit, next.value);

    return trendingResponse;
}

async function searchGifs(query, limit = 20) {
    const response = await chatService.searchKlipyGif(query, limit, next.value);

    return response;
}

function onScroll(event) {
    if (searchQuery.value.length === 0 && !trendingSelected.value) {
        return;
    }

    if (event.target.scrollHeight - (event.target.scrollTop + event.target.clientHeight) <= 400) {
        if (locked.value) {
            return;
        }

        if (trendingSelected.value) {
            loadTrendingGifs(20, next.value).then((res) => {
                gifs.value.push(...res.data.results);
                next.value = res.data.next;
                locked.value = false;
            });
        } else {
            searchGifs(searchQuery.value, 20).then((res) => {
                gifs.value.push(...res.data.results);
                next.value = res.data.next;
                locked.value = false;
            });
        }

        locked.value = true;
    }
}

function clearData() {
    searchQuery.value = "";
    trendingSelected.value = false;
    gifs.value = [];
    next.value = null;
}

onMounted(async () => {
    try {
        const trendingResponse = await loadTrendingGifs();

        if (trendingResponse.data.results) {
            trendingUrl.value = trendingResponse.data.results[0].media_formats.nanomp4.url;
        }

        const categoriesResponse = await chatService.getKlipyCategories();

        categories.value = categoriesResponse.data.tags;
    } catch (error) {
        console.error("Error fetching GIFs:", error);
    }
});

watch(searchQuery, async (newQuery) => {
    clearTimeout(timeout.value);

    if (newQuery.length === 0) {
        gifs.value = [];

        return;
    }

    timeout.value = setTimeout(async () => {
        try {
            const response = await searchGifs(newQuery, 20);

            gifs.value = response.data.results;
            next.value = response.data.next;
        } catch (error) {
            console.error("Error searching GIFs:", error);
        }
    }, 300);
});
</script>
