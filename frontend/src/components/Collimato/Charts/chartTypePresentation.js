// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import {
    HashtagIcon,
    PresentationChartLineIcon,
    ChartBarIcon,
    CalendarDaysIcon,
    ChartPieIcon,
    MapIcon,
} from "@heroicons/vue/24/outline";
import AxisChartCustomization from "@/components/Collimato/Charts/Configurations/AxisChartCustomization.vue";
import PieChartCustomization from "@/components/Collimato/Charts/Configurations/PieChartCustomization.vue";

const LABELS = "collimato.charts.new_chart.types.";

export const CHART_TYPE_PRESENTATION = {
    big_number: {
        icon: HashtagIcon,
        labelKey: `${LABELS}big_number`,
        shortLabelKey: `${LABELS}short.big_number`,
        customization: null,
        customizationProps: {},
        listStyle: {
            bg: "bg-indigo-500",
            badge: "bg-indigo-50 text-indigo-700 ring-indigo-700/10",
            activePill: "bg-indigo-600 text-white",
        },
    },
    line: {
        icon: PresentationChartLineIcon,
        labelKey: `${LABELS}line_chart`,
        shortLabelKey: `${LABELS}short.line`,
        customization: AxisChartCustomization,
        customizationProps: {},
        listStyle: {
            bg: "bg-blue-500",
            badge: "bg-blue-50 text-blue-700 ring-blue-700/10",
            activePill: "bg-blue-600 text-white",
        },
    },
    bar: {
        icon: ChartBarIcon,
        labelKey: `${LABELS}bar_chart`,
        shortLabelKey: `${LABELS}short.bar`,
        customization: AxisChartCustomization,
        customizationProps: { showOrientation: true, showBarValues: true },
        listStyle: {
            bg: "bg-emerald-500",
            badge: "bg-emerald-50 text-emerald-700 ring-emerald-700/10",
            activePill: "bg-emerald-600 text-white",
        },
    },
    time_bar: {
        icon: CalendarDaysIcon,
        labelKey: `${LABELS}time_bar_chart`,
        shortLabelKey: `${LABELS}short.time_bar`,
        customization: AxisChartCustomization,
        customizationProps: { showOrientation: true, showBarValues: true },
        listStyle: {
            bg: "bg-amber-500",
            badge: "bg-amber-50 text-amber-700 ring-amber-700/10",
            activePill: "bg-amber-500 text-white",
        },
    },
    pie: {
        icon: ChartPieIcon,
        labelKey: `${LABELS}pie_chart`,
        shortLabelKey: `${LABELS}short.pie`,
        customization: PieChartCustomization,
        customizationProps: {},
        listStyle: {
            bg: "bg-violet-500",
            badge: "bg-violet-50 text-violet-700 ring-violet-700/10",
            activePill: "bg-violet-600 text-white",
        },
    },
    map: {
        icon: MapIcon,
        labelKey: `${LABELS}map`,
        shortLabelKey: `${LABELS}short.map`,
        customization: null,
        customizationProps: {},
        listStyle: {
            bg: "bg-cyan-500",
            badge: "bg-cyan-50 text-cyan-700 ring-cyan-700/10",
            activePill: "bg-cyan-600 text-white",
        },
    },
};

export const FALLBACK_PRESENTATION = {
    icon: ChartBarIcon,
    labelKey: null,
    shortLabelKey: null,
    customization: null,
    customizationProps: {},
    listStyle: {
        bg: "bg-gray-500",
        badge: "bg-gray-50 text-gray-700 ring-gray-700/10",
        activePill: "bg-gray-600 text-white",
    },
};

export function presentationFor(chartType) {
    return CHART_TYPE_PRESENTATION[chartType] ?? FALLBACK_PRESENTATION;
}
