// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { it, expect, vi } from "vitest";
import { flushPromises } from "@vue/test-utils";
import { useEntityList } from "@/composables/graph/useEntityList.js";

// Each read waits for its own answer, so the test decides the order.
const reads = [];

vi.mock("@/services/graphService", () => ({
    default: {
        searchEntities: (type, search, signal) => {
            return new Promise((resolve, reject) => {
                signal.addEventListener("abort", () => reject(new Error("canceled")));

                reads.push({ search: search, answer: resolve });
            });
        },
    },
}));

function page(title) {
    return { data: { items: [{ type: "company", id: title, title: title }], next: "" } };
}

it("keeps the newest search when an older answer arrives late", async () => {
    const list = useEntityList();

    list.start("company", "tal");
    list.start("company", "tallinn");

    reads[1].answer(page("Tallinn Tech"));
    reads[0].answer(page("Talsi Foods"));

    await flushPromises();

    expect(list.items.value.map((item) => item.title)).toEqual(["Tallinn Tech"]);
    expect(list.loading.value).toBe(false);
    expect(list.errorMessage.value).toBe("");
});
