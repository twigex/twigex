// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { computed } from "vue";
import { t } from "@/i18n/index.js";
import { typeStyles } from "@/utils/graph/palette.js";
import { typeIcon } from "@/utils/graph/icons.js";
import { useGraphTypes } from "@/composables/graph/useGraphTypes.js";

const KIND_STYLES = {
    stored: "bg-indigo-50 text-indigo-700 ring-indigo-700/10",
    table: "bg-gray-50 text-gray-600 ring-gray-500/10",
    property: "bg-emerald-50 text-emerald-700 ring-emerald-600/20",
};

const STATUS_STYLES = {
    experimental: "bg-yellow-50 text-yellow-800 ring-yellow-600/20",
    active: "bg-green-50 text-green-700 ring-green-600/20",
    deprecated: "bg-gray-50 text-gray-600 ring-gray-500/10",
};

// Where Twigex reads each built-in relation from, in words.
const BUILT_IN_SOURCES = {
    chart_uses_dataset_field: "graph.ontology.source.chart_queries",
    dashboard_contains_chart: "graph.ontology.source.dashboard_charts",
    dashboard_uses_dataset_field: "graph.ontology.source.dashboard_filters",
    dataset_field_belongs_to_dataset: "graph.ontology.source.data_model",
};

// useOntology turns the shared type list into what the ontology pages show:
// each type's look, and each relation type's kind and source in words.
export function useOntology() {
    const graphTypes = useGraphTypes();

    const types = computed(() => graphTypes.types.value);

    const styles = computed(() => typeStyles(types.value.entity_types));

    const typeByID = computed(() => {
        return new Map(types.value.entity_types.map((entityType) => [entityType.id, entityType]));
    });

    // Your own types first, then the built-in ones.
    const sortedEntityTypes = computed(() => {
        const own = types.value.entity_types.filter((entityType) => !entityType.built_in);
        const builtIn = types.value.entity_types.filter((entityType) => entityType.built_in);

        return [...own, ...builtIn];
    });

    const sortedRelationTypes = computed(() => {
        const own = types.value.relation_types.filter((relationType) => !relationType.built_in);
        const builtIn = types.value.relation_types.filter((relationType) => relationType.built_in);

        return [...own, ...builtIn];
    });

    function typeName(id) {
        const entityType = typeByID.value.get(id);

        if (!entityType) {
            return "?";
        }

        return entityType.name;
    }

    function pluralName(entityType) {
        if (entityType.plural_name) {
            return entityType.plural_name;
        }

        return entityType.name;
    }

    function iconFor(entityType) {
        return typeIcon(entityType.icon);
    }

    function statusLabel(status) {
        return t.value(`graph.ontology.status.${status || "active"}`);
    }

    function statusClass(status) {
        return STATUS_STYLES[status || "active"] ?? STATUS_STYLES.active;
    }

    function kindLabel(relationType) {
        return t.value(`graph.ontology.kind.${relationType.kind}`);
    }

    function kindClass(relationType) {
        return KIND_STYLES[relationType.kind] ?? KIND_STYLES.table;
    }

    function cardinalityLabel(cardinality) {
        if (!cardinality) {
            return "—";
        }

        return t.value(`graph.ontology.cardinality.${cardinality}`);
    }

    function readFrom(relationType) {
        if (relationType.built_in) {
            return t.value(BUILT_IN_SOURCES[relationType.key] ?? "graph.ontology.source.twigex");
        }

        const propertyName = relationType.kind_config?.property_name;

        if (propertyName) {
            return propertyName;
        }

        return t.value("graph.ontology.source.a_field");
    }

    function readsFromField(relationType) {
        return !relationType.built_in && Boolean(relationType.kind_config?.property_name);
    }

    function sourceLabel(entityType) {
        if (entityType.built_in) {
            return t.value("graph.ontology.built_in");
        }

        if (entityType.source_name) {
            return t.value("graph.ontology.from_dataset", { dataset: entityType.source_name });
        }

        return t.value("graph.ontology.from_a_dataset");
    }

    function relationTypeCount(entityType) {
        return types.value.relation_types.filter((relationType) => {
            return (
                relationType.from_type_id === entityType.id ||
                relationType.to_type_id === entityType.id
            );
        }).length;
    }

    return {
        graphTypes,
        types,
        styles,
        typeByID,
        sortedEntityTypes,
        sortedRelationTypes,
        typeName,
        pluralName,
        iconFor,
        statusLabel,
        statusClass,
        kindLabel,
        kindClass,
        cardinalityLabel,
        readFrom,
        readsFromField,
        sourceLabel,
        relationTypeCount,
    };
}
