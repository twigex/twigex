<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <a
        :href="link.url"
        target="_blank"
        rel="noopener noreferrer"
        class="group block w-full md:w-[430px] mb-1"
    >
        <div
            class="flex w-full overflow-hidden rounded-md border border-slate-700/6 bg-white transition"
        >
            <div class="w-1 bg-slate-500/70 group-hover:bg-slate-400/90"></div>

            <div class="flex flex-col min-w-0 gap-3 p-3">
                <div class="min-w-0 flex-1">
                    <p class="mb-1 text-xs text-slate-400">
                        {{ link.site_name }}
                    </p>

                    <p
                        class="line-clamp-2 text-sm font-semibold text-sky-400 group-hover:text-sky-300"
                    >
                        {{ link.title }}
                    </p>

                    <p class="mt-1 line-clamp-3 text-sm text-gray-500">
                        {{ link.description }}
                    </p>
                </div>

                <div
                    v-if="link?.image?.url && !imageFailed"
                    class="relative w-full flex-none overflow-hidden rounded-md bg-gray-200"
                    :style="imageWrapperStyle"
                >
                    <iframe
                        v-if="videoPlaying && link.video"
                        :src="youtubeEmbedUrl(link.video.id)"
                        class="absolute inset-0 h-full w-full"
                        frameborder="0"
                        sandbox="allow-scripts allow-same-origin allow-presentation allow-popups allow-popups-to-escape-sandbox"
                        allow="autoplay; encrypted-media; picture-in-picture; fullscreen"
                        allowfullscreen
                    ></iframe>

                    <template v-else>
                        <img
                            :src="link.image.url"
                            class="absolute inset-0 h-full w-full object-cover"
                            loading="lazy"
                            decoding="async"
                            @error="onImgError"
                        />

                        <button
                            v-if="link.video"
                            type="button"
                            class="group/play absolute inset-0 flex items-center justify-center"
                            :aria-label="t('channels.message.play_video')"
                            @click.prevent.stop="emit('update:videoPlaying', true)"
                        >
                            <span
                                class="flex h-[48px] w-[68px] items-center justify-center rounded-xl bg-[#212121]/80 transition group-hover/play:bg-[#ff0000]"
                            >
                                <svg
                                    viewBox="0 0 24 24"
                                    class="h-7 w-7 fill-white"
                                    aria-hidden="true"
                                >
                                    <path d="M8 5v14l11-7z" />
                                </svg>
                            </span>
                        </button>
                    </template>
                </div>
            </div>
        </div>
    </a>
</template>

<script setup>
import { computed, ref } from "vue";
import { t } from "@/i18n/index.js";

const props = defineProps({
    link: {
        type: Object,
        required: true,
    },
    videoPlaying: {
        type: Boolean,
        default: false,
    },
});

const emit = defineEmits(["update:videoPlaying"]);

const imageFailed = ref(false);

const imageWrapperStyle = computed(() => {
    const img = props.link?.image;

    const width = img?.width || 1200;
    const height = img?.height || 628;

    return {
        aspectRatio: `${width} / ${height}`,
    };
});

// Build the embed src ourselves and re-validate the id, so nothing untrusted
// reaches the iframe even if the stored value were tampered with.
const youtubeEmbedUrl = (id) => {
    if (!/^[A-Za-z0-9_-]{11}$/.test(id)) {
        return "";
    }

    return `https://www.youtube-nocookie.com/embed/${id}?autoplay=1&rel=0`;
};

function onImgError() {
    imageFailed.value = true;
}
</script>
