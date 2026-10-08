// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { nextTick, ref } from "vue";
import { useRoute } from "vue-router";
import { t } from "@/i18n/index.js";
import { useCollimatoStore } from "@/store/collimato";
import { useAlertStore } from "@/store/alerts";
import collimatoService from "@/services/collimatoService.js";
import { extractErrorMessage } from "@/utils/errors";
import { GridStack } from "gridstack";
import "gridstack/dist/gridstack.min.css";

export function useDashboardGrid({
    dashboard,
    charts,
    widgets,
    loaded,
    loadMapLayer,
    loadChartData,
}) {
    const route = useRoute();
    const collimatoStore = useCollimatoStore();
    const alertStore = useAlertStore();

    const edit = ref(false);
    const saving = ref(false);
    const layoutBeforeEdit = ref(null);

    let grid = null;

    function loadGrid() {
        grid = GridStack.init({
            cellHeight: "auto",
            resizable: { handles: "all" },
            animate: true,
            float: true,
            disableDrag: !edit.value,
            disableResize: !edit.value,
            acceptWidgets: () => true,
        });

        const myEvent = new Event("widget_resize");

        grid.on("drag", () => window.dispatchEvent(myEvent));
        grid.on("resizestop", () => window.dispatchEvent(myEvent));

        loaded.value = true;
    }

    function resizeWidget() {
        window.dispatchEvent(new Event("widget_resize"));
    }

    function addWidget(item) {
        item.position = { x: 0, y: 0, w: 3, h: 2 };
        charts.value.push(item);

        nextTick(() => {
            const lastWidgetIndex = charts.value.length - 1;

            if (widgets.value[lastWidgetIndex]) {
                const newWidgetEl = widgets.value[lastWidgetIndex].$el;

                grid.addWidget(
                    newWidgetEl,
                    item.position.x,
                    item.position.y,
                    item.position.w,
                    item.position.h,
                );

                if (item.chart_type === "map") {
                    item.configuration.layers.forEach((layer) => {
                        loadMapLayer(layer, lastWidgetIndex, item);
                    });
                } else {
                    loadChartData(item, lastWidgetIndex);
                }
            }
        });
    }

    function removeWidget(item) {
        const chartIndex = charts.value.findIndex((c) => c.id === item.id);

        if (chartIndex !== -1) {
            grid.removeWidget(widgets.value[chartIndex].$el);
            charts.value.splice(chartIndex, 1);
        }
    }

    function setEditing(on) {
        edit.value = on;
        grid.enableMove(on);
        grid.enableResize(on);

        setTimeout(() => resizeWidget(), 300);
    }

    function startEdit() {
        if (!collimatoStore.hasPermissionToEditDashboards) {
            alertStore.showError(t.value("collimato.dashboard.error.no_permission_edit"));

            return;
        }

        layoutBeforeEdit.value = charts.value.map((chart) => ({
            chart: chart,
            position: { ...chart.position },
        }));

        setEditing(true);
    }

    async function saveLayout() {
        saving.value = true;

        try {
            await collimatoService.updateDashboardCharts(
                route.params.workspaceId,
                dashboard.value.id,
                {
                    charts: charts.value.map((chart) => ({
                        id: chart.id,
                        position: { ...chart.position },
                    })),
                },
            );

            layoutBeforeEdit.value = null;
            setEditing(false);
        } catch (error) {
            alertStore.showError(extractErrorMessage(error));
        } finally {
            saving.value = false;
        }
    }

    async function cancelEdit() {
        const snapshot = layoutBeforeEdit.value ?? [];

        const kept = snapshot.map((entry) => entry.chart);

        widgets.value.forEach((widget, index) => {
            if (!kept.includes(charts.value[index]) && widget.$el.gridstackNode) {
                grid.removeWidget(widget.$el, false);
            }
        });

        snapshot.forEach((entry) => Object.assign(entry.chart.position, entry.position));
        charts.value = kept;

        await nextTick();

        grid.batchUpdate();

        widgets.value.forEach((widget, index) => {
            const element = widget.$el;
            const position = charts.value[index].position;

            if (!element.gridstackNode) {
                grid.makeWidget(element);
            }

            grid.update(element, {
                x: position.x,
                y: position.y,
                w: position.w,
                h: position.h,
            });
        });

        grid.commit();

        layoutBeforeEdit.value = null;
        setEditing(false);
    }

    return {
        edit,
        saving,
        loadGrid,
        resizeWidget,
        addWidget,
        removeWidget,
        startEdit,
        saveLayout,
        cancelEdit,
    };
}
