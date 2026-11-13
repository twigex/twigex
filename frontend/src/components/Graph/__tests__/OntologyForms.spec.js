// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { it, expect, vi, beforeEach } from "vitest";
import { mount, flushPromises } from "@vue/test-utils";
import { createPinia, setActivePinia } from "pinia";
import TypeFormDialog from "@/components/Graph/TypeFormDialog.vue";
import RelationTypeFormDialog from "@/components/Graph/RelationTypeFormDialog.vue";

const sent = [];

vi.mock("@/services/graphService", () => ({
    default: {
        getDatasetChoices: async () => ({
            data: [
                {
                    id: "ds-companies",
                    name: "companies",
                    workspace_name: "Sales",
                    fields: [
                        { id: "f-key", name: "companies.key", title: "Company ID" },
                        { id: "f-name", name: "companies.name", title: "Company" },
                    ],
                },
                {
                    id: "ds-orders",
                    name: "orders",
                    workspace_name: "Sales",
                    fields: [{ id: "f-company-id", name: "orders.company_id" }],
                },
            ],
        }),
        updateEntityType: async (id, input) => {
            sent.push({ call: "updateEntityType", id: id, input: { ...input } });

            return { data: { id: id, ...input } };
        },
        updateRelationType: async (id, input) => {
            sent.push({ call: "updateRelationType", id: id, input: { ...input } });

            return { data: { id: id, ...input } };
        },
    },
}));

globalThis.ResizeObserver = class {
    observe() {}
    unobserve() {}
    disconnect() {}
};

const company = {
    id: "t-company",
    key: "company",
    name: "Company",
    plural_name: "Companies",
    icon: "building-office",
    color: "cyan",
    status: "active",
    description: "From the ERP.",
    source: "dataset",
    source_id: "ds-companies",
    key_property: "f-key",
    title_property: "f-name",
};

const order = {
    id: "t-order",
    key: "order",
    name: "Order",
    source: "dataset",
    source_id: "ds-orders",
};

beforeEach(() => {
    sent.length = 0;
    setActivePinia(createPinia());
    document.body.innerHTML = "";
});

function submit() {
    const form = document.body.querySelector("form");

    form.dispatchEvent(new Event("submit", { bubbles: true, cancelable: true }));
}

it("opens a type filled in and saves every field", async () => {
    const wrapper = mount(TypeFormDialog, {
        props: { open: true, entityType: company },
        attachTo: document.body,
    });

    await flushPromises();

    expect(document.body.querySelector("#type-name").value).toBe("Company");
    expect(document.body.querySelector("#type-key").disabled).toBe(true);
    expect(document.body.textContent).toContain("Sales · companies");

    submit();
    await flushPromises();

    expect(sent[0].call).toBe("updateEntityType");
    expect(sent[0].input).toMatchObject({
        name: "Company",
        plural_name: "Companies",
        icon: "building-office",
        color: "cyan",
        source_id: "ds-companies",
        key_property: "f-key",
        title_property: "f-name",
    });
    expect(wrapper.emitted("saved")).toHaveLength(1);

    wrapper.unmount();
});

it("opens a relation type filled in, with its preview, and saves it", async () => {
    const places = {
        id: "l-places",
        key: "company_places_order",
        name: "places",
        reverse_name: "placed by",
        from_type_id: "t-company",
        to_type_id: "t-order",
        cardinality: "one_to_many",
        kind_config: { property: "f-company-id" },
    };

    const wrapper = mount(RelationTypeFormDialog, {
        props: { open: true, relationType: places, entityTypes: [company, order] },
        attachTo: document.body,
    });

    await flushPromises();

    const text = document.body.textContent.replace(/\s+/g, " ");

    expect(text).toContain("Company places Order");
    expect(text).toContain("read from orders.company_id");

    submit();
    await flushPromises();

    expect(sent[0].call).toBe("updateRelationType");
    expect(sent[0].input).toMatchObject({
        name: "places",
        reverse_name: "placed by",
        property: "f-company-id",
        from_property: "",
        cardinality: "one_to_many",
    });

    wrapper.unmount();
});

it("gives a new type a key even when its name has no Latin letters", async () => {
    const wrapper = mount(TypeFormDialog, {
        props: { open: true, entityType: null },
        attachTo: document.body,
    });

    await flushPromises();

    const name = document.body.querySelector("#type-name");
    const key = document.body.querySelector("#type-key");

    name.value = "Компания";
    name.dispatchEvent(new Event("input"));
    await flushPromises();

    const firstKey = key.value;

    name.value = "Компаниялар";
    name.dispatchEvent(new Event("input"));
    await flushPromises();

    expect(firstKey).toMatch(/^type_[a-z0-9]+$/);
    expect(key.value).toBe(firstKey);

    wrapper.unmount();
});
