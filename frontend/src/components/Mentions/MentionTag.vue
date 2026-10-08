<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <span
        v-if="styled"
        ref="anchor"
        :class="classes"
        :role="interactive ? 'button' : undefined"
        :tabindex="interactive ? 0 : undefined"
        @click="onClick"
        @keydown.enter.prevent="onClick"
        @keydown.space.prevent="onClick"
        >@{{ label }}</span
    >
    <template v-else>@{{ label }}</template>

    <Teleport to="body" v-if="open">
        <MentionTooltip
            :open="open"
            :user="user"
            :style="floatingStyles"
            :ref="setFloating"
            @close="open = false"
        />
    </Teleport>
</template>

<script setup>
import { computed, ref } from "vue";
import { useFloating, autoUpdate, autoPlacement, offset, shift } from "@floating-ui/vue";
import { useUserStore } from "@/store/user";
import { useUser, useUserByUsername } from "@/composables/useUser";
import MentionTooltip from "@/components/Mentions/MentionTooltip.vue";
import { UNKNOWN_USER } from "@/utils/mentions";

const props = defineProps({
    username: { type: String, default: "" },
    userId: { type: String, default: "" },
});

const userStore = useUserStore();

const isBroadcast = computed(() => props.username === "all" || props.username === "here");

// Lazily loads only this mention (never @all/@here) and resolves reactively:
// an unresolved @mention upgrades to a styled tag once its user is fetched,
// without re-parsing the message.
const byId = useUser(() => props.userId || null);
const byUsername = useUserByUsername(() =>
    isBroadcast.value || props.userId ? null : props.username,
);

const user = computed(() => (props.userId ? byId.value : byUsername.value));

// A tombstone rather than the id, which would put a raw identifier
// mid-sentence.
const label = computed(() => {
    if (props.userId) {
        return user.value?.username || UNKNOWN_USER;
    }

    return props.username;
});

// Unknown handles stay plain text, matching a raw @word the reader can ignore.
// An id is always styled: its tombstone is not something the reader typed.
const styled = computed(() => isBroadcast.value || !!user.value || !!props.userId);

const interactive = computed(() => !isBroadcast.value && !!user.value);

const classes = computed(() => {
    let color = "bg-gray-200";

    if (isBroadcast.value) {
        color = "bg-indigo-200";
    } else if (!user.value) {
        // A tombstone opens nothing, so it does not invite a click.
        return "px-1 rounded mention bg-gray-100 text-gray-500 italic";
    } else if (userStore.user && user.value.id === userStore.user.id) {
        color = "bg-yellow-300";
    }

    return `px-1 rounded mention cursor-pointer hover:underline ${color}`;
});

const anchor = ref(null);
const floating = ref(null);
const open = ref(false);

const { floatingStyles } = useFloating(anchor, floating, {
    middleware: [
        offset(6),
        autoPlacement({ crossAxis: true, alignment: "start" }),
        shift({ padding: 8 }),
    ],
    whileElementsMounted: autoUpdate,
});

// The tooltip's root is v-if'd on open, so its ref resolves to a component
// instance whose $el we hand to floating-ui as the positioned element.
function setFloating(el) {
    if (!el) {
        floating.value = null;

        return;
    }

    if (el instanceof HTMLElement) {
        floating.value = el;

        return;
    }

    const dom = el.$el ?? el.$?.subTree?.el ?? null;

    floating.value = dom instanceof HTMLElement ? dom : null;
}

function onClick() {
    // @all/@here address everyone; a tombstone has nobody left.
    if (isBroadcast.value || !user.value) return;
    open.value = true;
}
</script>
