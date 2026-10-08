// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { describe, it, expect } from "vitest";
import { defineComponent, h, nextTick, ref } from "vue";
import { mount } from "@vue/test-utils";
import { useCallSidebarScroll } from "@/composables/chat/useCallSidebarScroll";

describe("useCallSidebarScroll", () => {
    it("re-checks the scroll arrows when the participant count changes", async () => {
        const participants = ref([{ identity: "a" }]);
        let scroll;

        mount(
            defineComponent({
                setup() {
                    scroll = useCallSidebarScroll(ref(false), participants);

                    return () => h("div");
                },
            }),
        );

        await nextTick();

        scroll.sidebarScroll.value = {
            scrollTop: 0,
            clientHeight: 100,
            scrollHeight: 300,
        };

        expect(scroll.canScrollDown.value).toBe(false);

        participants.value.push({ identity: "b" });
        await nextTick();
        await nextTick();

        expect(scroll.canScrollDown.value).toBe(true);
        expect(scroll.canScrollUp.value).toBe(false);
    });
});
