// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { ref, computed } from "vue";
import { t } from "@/i18n/index.js";
import graphService from "@/services/graphService";
import { extractErrorMessage } from "@/utils/errors";
import { usePermissions } from "@/composables/usePermissions";
import { useAlertStore } from "@/store/alerts";
import { useGraphTypes } from "@/composables/graph/useGraphTypes.js";

// useTypeEditing runs the ontology's forms: which one is open, deleting after
// a confirmation, and reading the types again after any change. Only people
// who manage the graph get the buttons; the server checks again on save.
export function useTypeEditing() {
    const graphTypes = useGraphTypes();
    const alerts = useAlertStore();
    const { can } = usePermissions();

    const canEdit = computed(() => can("manage_knowledge_graph"));

    const typeForm = ref({ open: false, entityType: null });
    const relationTypeForm = ref({ open: false, relationType: null, fromTypeId: "" });
    const deleting = ref({ open: false, kind: "", item: null });

    function newType() {
        typeForm.value = { open: true, entityType: null };
    }

    function editType(entityType) {
        typeForm.value = { open: true, entityType: entityType };
    }

    function newRelationType(fromTypeId) {
        relationTypeForm.value = { open: true, relationType: null, fromTypeId: fromTypeId ?? "" };
    }

    function editRelationType(relationType) {
        relationTypeForm.value = { open: true, relationType: relationType, fromTypeId: "" };
    }

    function closeForms() {
        typeForm.value = { open: false, entityType: null };
        relationTypeForm.value = { open: false, relationType: null, fromTypeId: "" };
    }

    async function saved() {
        closeForms();

        await graphTypes.reload();
    }

    function askDelete(kind, item) {
        deleting.value = { open: true, kind: kind, item: item };
    }

    function cancelDelete() {
        deleting.value = { open: false, kind: "", item: null };
    }

    const deleteMessage = computed(() => {
        const item = deleting.value.item;

        if (!item) {
            return "";
        }

        if (deleting.value.kind === "type") {
            return t.value("graph.form.delete.type_message", { name: item.name });
        }

        return t.value("graph.form.delete.relation_message", { name: item.name });
    });

    // A type a relation type still uses can't be deleted; the server says so,
    // and the message is shown as it is.
    async function confirmDelete() {
        const { kind, item } = deleting.value;

        cancelDelete();

        try {
            if (kind === "type") {
                await graphService.deleteEntityType(item.id);
            } else {
                await graphService.deleteRelationType(item.id);
            }

            await graphTypes.reload();

            return true;
        } catch (e) {
            alerts.showError(extractErrorMessage(e));

            return false;
        }
    }

    return {
        canEdit,
        typeForm,
        relationTypeForm,
        deleting,
        deleteMessage,
        newType,
        editType,
        newRelationType,
        editRelationType,
        closeForms,
        saved,
        askDelete,
        cancelDelete,
        confirmDelete,
    };
}
