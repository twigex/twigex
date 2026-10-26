// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { describe, it, expect, beforeEach } from "vitest";
import { setActivePinia, createPinia } from "pinia";
import { useSettingsStore } from "@/store/settings";

beforeEach(() => {
    setActivePinia(createPinia());
});

describe("isFeatureEnabled", () => {
    it("is on when the server sends the flag as on", () => {
        const settingsStore = useSettingsStore();

        settingsStore.setConfig({ FeatureFlags: { KnowledgeGraph: true } });

        expect(settingsStore.isFeatureEnabled("KnowledgeGraph")).toBe(true);
    });

    it("is off when the server sends the flag as off", () => {
        const settingsStore = useSettingsStore();

        settingsStore.setConfig({ FeatureFlags: { KnowledgeGraph: false } });

        expect(settingsStore.isFeatureEnabled("KnowledgeGraph")).toBe(false);
    });

    it("is off before the config has loaded", () => {
        const settingsStore = useSettingsStore();

        expect(settingsStore.isFeatureEnabled("KnowledgeGraph")).toBe(false);
    });

    it("is off for anonymous users, who get no flags", () => {
        const settingsStore = useSettingsStore();

        settingsStore.setConfig({ DefaultLocale: "en" });

        expect(settingsStore.isFeatureEnabled("KnowledgeGraph")).toBe(false);
    });
});
