// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

// Orders entity types so that every relation goes from an earlier type to a
// later one, leaving out a relation from a type to itself. When there is no
// such order (two types relate to each other both ways, or a longer loop),
// hasLoop is set and the types caught in the loop come last.
export function columnOrder(entityTypes, relationTypes) {
    const incoming = new Map();
    const outgoing = new Map();

    for (const entityType of entityTypes) {
        incoming.set(entityType.id, new Set());
        outgoing.set(entityType.id, new Set());
    }

    for (const relationType of relationTypes) {
        const from = relationType.from_type_id;
        const to = relationType.to_type_id;

        if (from === to || !incoming.has(from) || !incoming.has(to)) {
            continue;
        }

        incoming.get(to).add(from);
        outgoing.get(from).add(to);
    }

    const order = [];
    const placed = new Set();

    let ready = entityTypes.filter((entityType) => incoming.get(entityType.id).size === 0);

    while (ready.length > 0) {
        const next = ready.shift();

        order.push(next.id);
        placed.add(next.id);

        for (const target of outgoing.get(next.id)) {
            incoming.get(target).delete(next.id);

            const nowReady = incoming.get(target).size === 0 && !placed.has(target);
            const alreadyWaiting = ready.some((entityType) => entityType.id === target);

            if (nowReady && !alreadyWaiting) {
                const entityType = entityTypes.find((candidate) => candidate.id === target);

                ready.push(entityType);
            }
        }
    }

    const leftOver = entityTypes.filter((entityType) => !placed.has(entityType.id));

    for (const entityType of leftOver) {
        order.push(entityType.id);
    }

    return { order: order, hasLoop: leftOver.length > 0 };
}
