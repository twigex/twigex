// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { createApp } from "vue";
import { createPinia } from "pinia"; // Initialize Pinia here

import App from "./App.vue";
import router from "./router";
import { setupInterceptors } from "./services/api";

import "./index.css";
import "./global.css";

// Create Pinia instance
const pinia = createPinia();

// Create the Vue app and set up the stores
const app = createApp(App);

// Apply Pinia before router and other plugins
app.use(pinia);
app.use(router);

// Attach axios interceptors now that Pinia and the router both exist
setupInterceptors(router);

// Mount the app after everything is set up
app.mount("#app");
