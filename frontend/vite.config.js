// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { defineConfig } from "vite";
import vue from "@vitejs/plugin-vue";
import fs from "fs";
import path from "path";

const pdfjsWasmDir = path.resolve(__dirname, "node_modules/pdfjs-dist/wasm");

const wasmMimeTypes = {
    ".js": "text/javascript",
    ".wasm": "application/wasm",
};

// pdf.js fetches these only when a PDF actually uses JPEG2000 or JBIG2 (scanned
// pages); without them those images render blank with just a console warning.
function pdfjsWasm() {
    let outDir;

    return {
        name: "pdfjs-wasm",
        configResolved(config) {
            outDir = path.resolve(config.root, config.build.outDir);
        },
        configureServer(server) {
            const mount = server.config.base + "pdfjs/wasm";

            server.middlewares.use(mount, (req, res, next) => {
                let file;

                try {
                    const base = fs.realpathSync(pdfjsWasmDir);

                    file = fs.realpathSync(path.resolve(base, "." + req.url.split("?")[0]));
                    if (!file.startsWith(base + path.sep) || !fs.statSync(file).isFile()) {
                        next();

                        return;
                    }
                } catch {
                    next();

                    return;
                }

                res.setHeader(
                    "content-type",
                    wasmMimeTypes[path.extname(file)] ?? "application/octet-stream",
                );
                fs.createReadStream(file)
                    .on("error", () => res.destroy())
                    .pipe(res);
            });
        },
        writeBundle() {
            fs.cpSync(pdfjsWasmDir, path.join(outDir, "pdfjs/wasm"), {
                recursive: true,
            });
        },
    };
}

export default defineConfig(({ mode }) => ({
    plugins: [vue(), pdfjsWasm()],
    test: {
        environment: "jsdom",
        globals: true,
    },
    resolve: {
        alias: {
            "@": path.resolve(__dirname, "src"),
            "frappe-gantt/dist/frappe-gantt.css": path.resolve(
                __dirname,
                "node_modules/frappe-gantt/dist/frappe-gantt.css",
            ),
        },
    },
    define: {
        __VUE_PROD_DEVTOOLS__: mode === "development",
    },
    worker: {
        format: "es",
    },
    server: {
        proxy: {
            "/api": {
                target: "http://localhost:3000",
                changeOrigin: true,
                ws: true,
            },
        },
    },
    build: {
        rollupOptions: {
            input: path.resolve(__dirname, "index.html"),
        },
    },
}));
