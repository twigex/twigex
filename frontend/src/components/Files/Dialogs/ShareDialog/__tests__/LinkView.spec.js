// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { describe, it, expect, beforeEach } from "vitest";
import { mount } from "@vue/test-utils";
import { nextTick } from "vue";
import { setActivePinia, createPinia } from "pinia";
import moment from "moment-timezone";
import LinkView from "@/components/Files/Dialogs/ShareDialog/LinkView.vue";
import { useUserStore } from "@/store/user";

function livingIn(timezone) {
    useUserStore().user = {
        ...useUserStore().user,
        timezone: { useAutomaticTimezone: "false", manualTimezone: timezone },
    };
}

function mountLink(expiration) {
    return mount(LinkView, {
        props: { modelValue: { expiration }, isFolder: false },
        global: { stubs: { teleport: true } },
    });
}

function lastExpiration(wrapper) {
    return wrapper.emitted("update:modelValue").at(-1)[0].expiration;
}

const endOf28InRiga = moment.tz("2026-10-28 23:59:59", "Europe/Riga").unix();

describe("LinkView expiry", () => {
    beforeEach(() => setActivePinia(createPinia()));

    it("keeps a saved expiry nobody changed, wherever the editor is", () => {
        livingIn("Pacific/Auckland");

        const wrapper = mountLink(endOf28InRiga);

        expect(lastExpiration(wrapper)).toBe(endOf28InRiga);
    });

    it("ends a newly picked day at its last second where the editor is", async () => {
        livingIn("Pacific/Auckland");

        const wrapper = mountLink(endOf28InRiga);

        wrapper.vm.expiryDate = "2026-10-30";
        await nextTick();

        const expected = moment.tz("2026-10-30 23:59:59", "Pacific/Auckland").unix();

        expect(lastExpiration(wrapper)).toBe(expected);
    });

    it("has no expiry when none is set", () => {
        livingIn("Europe/Riga");

        expect(lastExpiration(mountLink(0))).toBe(0);
    });
});
