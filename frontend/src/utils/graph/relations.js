// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

const FLIPPED = {
    one_to_many: "many_to_one",
    many_to_one: "one_to_many",
};

// The relation types that touch one type, each read from that type's side:
// from its "from" end with the name, from its "to" end with the reverse name
// and the cardinality turned around. A relation type from the type to itself
// is listed once from each end.
export function relationsOf(entityType, relationTypes, typeByID) {
    const relations = [];

    for (const relationType of relationTypes) {
        if (relationType.from_type_id === entityType.id) {
            relations.push({
                id: `${relationType.key}:out`,
                relationType: relationType,
                name: relationType.name,
                other: typeByID.get(relationType.to_type_id),
                cardinality: relationType.cardinality ?? "",
            });
        }

        if (relationType.to_type_id === entityType.id) {
            const cardinality = relationType.cardinality ?? "";

            relations.push({
                id: `${relationType.key}:in`,
                relationType: relationType,
                name: relationType.reverse_name,
                other: typeByID.get(relationType.from_type_id),
                cardinality: FLIPPED[cardinality] ?? cardinality,
            });
        }
    }

    return relations;
}
