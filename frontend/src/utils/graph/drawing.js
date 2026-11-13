// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { EdgeCurvedArrowProgram } from "@sigma/edge-curve";

// How Sigma draws labels on the graph's maps.

export const LABEL_FONT = "Inter, ui-sans-serif, system-ui, sans-serif";
export const LABEL_COLOR = "#1f2937";

// Labels go under the dots, and a group box shows its count inside.
export function drawLabel(context, data, settings) {
    if (data.isGroup && !data.faded) {
        context.font = `700 ${settings.labelSize - 1}px ${LABEL_FONT}`;
        context.fillStyle = "#ffffff";
        context.textAlign = "center";
        context.textBaseline = "middle";
        context.fillText(data.countText, data.x, data.y);
    }

    if (!data.label) {
        return;
    }

    let weight = 500;

    if (data.highlighted) {
        weight = 600;
    }

    context.font = `${weight} ${settings.labelSize}px ${LABEL_FONT}`;
    context.fillStyle = LABEL_COLOR;
    context.textAlign = "center";
    context.textBaseline = "top";
    context.fillText(data.label, data.x, data.y + data.size + 4);
}

// A dot in focus gets a soft ring and its label on a white background, so it
// stays readable over lines.
export function drawHover(context, data, settings) {
    context.beginPath();
    context.arc(data.x, data.y, data.size + 6, 0, Math.PI * 2);
    context.fillStyle = "rgba(79, 70, 229, 0.12)";
    context.fill();

    if (data.label) {
        context.font = `600 ${settings.labelSize}px ${LABEL_FONT}`;

        const width = context.measureText(data.label).width;
        const top = data.y + data.size + 3;

        context.fillStyle = "#ffffff";
        context.fillRect(data.x - width / 2 - 4, top, width + 8, settings.labelSize + 4);
    }

    drawLabel(context, data, settings);
}

// The Sigma settings both maps share: curved arrows, and labels in the app's
// font drawn by drawLabel and drawHover.
export function baseSigmaSettings() {
    return {
        renderEdgeLabels: true,
        defaultEdgeType: "curved",
        edgeProgramClasses: {
            curved: EdgeCurvedArrowProgram,
        },
        labelRenderedSizeThreshold: 0,
        labelFont: LABEL_FONT,
        labelColor: { color: LABEL_COLOR },
        edgeLabelFont: LABEL_FONT,
        edgeLabelSize: 10,
        edgeLabelColor: { color: "#4b5563" },
        defaultDrawNodeLabel: drawLabel,
        defaultDrawNodeHover: drawHover,
    };
}
