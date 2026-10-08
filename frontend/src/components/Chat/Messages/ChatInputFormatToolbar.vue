<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div class="flex items-center gap-x-0.5 px-2 pt-1.5 pb-1 border-b border-gray-100">
        <button
            @mousedown.prevent="emits('format', 'bold')"
            type="button"
            class="flex h-7 w-7 items-center justify-center rounded text-gray-500 hover:bg-gray-100 hover:text-gray-700 transition-colors"
            title="Bold"
        >
            <span class="text-sm font-bold leading-none">B</span>
        </button>
        <button
            @mousedown.prevent="emits('format', 'italic')"
            type="button"
            class="flex h-7 w-7 items-center justify-center rounded text-gray-500 hover:bg-gray-100 hover:text-gray-700 transition-colors"
            title="Italic"
        >
            <span class="text-sm italic font-serif leading-none">I</span>
        </button>
        <button
            @mousedown.prevent="emits('format', 'strikethrough')"
            type="button"
            class="flex h-7 w-7 items-center justify-center rounded text-gray-500 hover:bg-gray-100 hover:text-gray-700 transition-colors"
            title="Strikethrough"
        >
            <span class="text-sm line-through leading-none">S</span>
        </button>

        <div class="w-px h-4 bg-gray-200 mx-0.5" />

        <button
            @mousedown.prevent="emits('format', 'code')"
            type="button"
            class="flex h-7 w-7 items-center justify-center rounded text-gray-500 hover:bg-gray-100 hover:text-gray-700 transition-colors"
            title="Inline code"
        >
            <CodeBracketIcon class="size-4" />
        </button>
        <button
            @mousedown.prevent="emits('format', 'codeblock')"
            type="button"
            class="flex h-7 w-7 items-center justify-center rounded text-gray-500 hover:bg-gray-100 hover:text-gray-700 transition-colors"
            title="Code block"
        >
            <CodeBracketSquareIcon class="size-4" />
        </button>

        <div class="w-px h-4 bg-gray-200 mx-0.5" />

        <button
            @mousedown.prevent="emits('format', 'quote')"
            type="button"
            class="flex h-7 w-7 items-center justify-center rounded text-gray-500 hover:bg-gray-100 hover:text-gray-700 transition-colors"
            title="Blockquote"
        >
            <ChatBubbleLeftIcon class="size-4" />
        </button>
        <button
            @mousedown.prevent="emits('format', 'ul')"
            type="button"
            class="flex h-7 w-7 items-center justify-center rounded text-gray-500 hover:bg-gray-100 hover:text-gray-700 transition-colors"
            title="Bullet list"
        >
            <ListBulletIcon class="size-4" />
        </button>
        <button
            @mousedown.prevent="emits('format', 'ol')"
            type="button"
            class="flex h-7 w-7 items-center justify-center rounded text-gray-500 hover:bg-gray-100 hover:text-gray-700 transition-colors"
            title="Numbered list"
        >
            <QueueListIcon class="size-4" />
        </button>
    </div>
</template>

<script setup>
import {
    CodeBracketIcon,
    CodeBracketSquareIcon,
    ListBulletIcon,
    QueueListIcon,
    ChatBubbleLeftIcon,
} from "@heroicons/vue/24/outline";

const emits = defineEmits(["format"]);
</script>
