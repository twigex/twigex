// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { describe, it, expect, vi } from "vitest";
import { mount } from "@vue/test-utils";
import TimeDimensions from "@/components/Collimato/Charts/Sidebar/TimeDimensions.vue";

vi.mock("@/components/DatePicker/DatePicker.vue", () => ({
    default: { name: "DatePicker", template: "<div />" },
}));

const emptyTime = {
    dimension: null,
    dateRange: "",
    granularity: "",
    startTime: "",
    endTime: "",
};

function factory(time = emptyTime) {
    return mount(TimeDimensions, {
        props: { time, cubes: [] },
        global: {
            stubs: {
                Disclosure: { template: "<div><slot :open='true' /></div>" },
                DisclosureButton: { template: "<button><slot /></button>" },
                DisclosurePanel: { template: "<div><slot /></div>" },
                TransitionRoot: { template: "<div><slot /></div>" },
                TransitionChild: { template: "<div><slot /></div>" },
                Dialog: { template: "<div><slot /></div>" },
                DialogPanel: { template: "<div><slot /></div>" },
                DialogTitle: { template: "<div><slot /></div>" },
            },
        },
    });
}

describe("TimeDimensions", () => {
    it("emits one patch when cleared, not one per field", () => {
        const w = factory({
            dimension: { name: "orders.created_at", title: "Created" },
            dateRange: "custom",
            granularity: "day",
            startTime: "2026-01-01",
            endTime: "2026-01-31",
        });

        w.findAll("button")
            .find((b) => b.text() === "Clear")
            .trigger("click");

        const events = w.emitted("update:time");

        expect(events).toHaveLength(1);
        expect(events[0][0]).toEqual(emptyTime);
    });

    it("emits only the key that changed", async () => {
        const w = factory({ ...emptyTime, dateRange: "custom" });

        await w.find("#start_time").setValue("2026-02-01");

        const events = w.emitted("update:time");

        expect(events).toHaveLength(1);
        expect(events[0][0]).toEqual({ startTime: "2026-02-01" });
    });

    it("follows the prop rather than any local copy", async () => {
        const w = factory({ ...emptyTime, dateRange: "custom" });

        expect(w.find("#start_time").element.value).toBe("");

        await w.setProps({
            time: { ...emptyTime, dateRange: "custom", startTime: "2026-03-05" },
        });

        expect(w.find("#start_time").element.value).toBe("2026-03-05");

        await w.setProps({ time: emptyTime });

        expect(w.find("#start_time").exists()).toBe(false);
    });

    it("shows the custom range fields only for a custom range", () => {
        expect(factory().find("#start_time").exists()).toBe(false);
        expect(
            factory({ ...emptyTime, dateRange: "custom" })
                .find("#start_time")
                .exists(),
        ).toBe(true);
    });

    it("renders the values it is given", () => {
        const w = factory({
            dimension: { name: "orders.created_at", title: "Created at" },
            dateRange: "custom",
            granularity: "day",
            startTime: "2026-01-01",
            endTime: "2026-01-31",
        });

        expect(w.text()).toContain("Created at");
        expect(w.find("#start_time").element.value).toBe("2026-01-01");
        expect(w.find("#end_time").element.value).toBe("2026-01-31");
    });
});
