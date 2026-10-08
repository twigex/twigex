// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

const sameId = (a, b) => String(a) === String(b);

const copyOf = (value) =>
    Array.isArray(value) ? [...value] : value && typeof value === "object" ? { ...value } : value;

// setCardField returns the board's columns with field set to value on the
// task's card, in both the column's cards and its saved order. Columns the
// task is not in are returned as they were.
export function setCardField(columns, taskId, field, value) {
    return columns.map((column) => {
        let changed = false;
        const set = (list) =>
            Array.isArray(list)
                ? list.map((card) => {
                      if (!sameId(card.id, taskId)) return card;
                      changed = true;

                      return { ...card, [field]: copyOf(value) };
                  })
                : list;

        const tasks = set(column.tasks);
        const order = set(column.order);

        return changed ? { ...column, tasks, order } : column;
    });
}

// moveCardToColumn sets the field the board is grouped by and moves the card
// to the column for its new value, or to the column of cards with no value
// when the board has none for it. A card whose value names its own column
// stays where it is.
export function moveCardToColumn(columns, taskId, field, value) {
    const from = columns.findIndex((column) =>
        (column.tasks || []).some((card) => sameId(card.id, taskId)),
    );

    if (from === -1) return columns;

    const columnId = value?.id ?? value ?? null;
    let to = columns.findIndex((column) => columnId != null && sameId(column.id, columnId));

    if (to === -1) to = columns.findIndex((column) => column.id == null || column.isUnassigned);
    if (to === from) return setCardField(columns, taskId, field, value);

    const next = columns.slice();
    const source = columns[from];
    const card = source.tasks.find((c) => sameId(c.id, taskId));

    next[from] = {
        ...source,
        tasks: source.tasks.filter((c) => !sameId(c.id, taskId)),
        ...(Array.isArray(source.order)
            ? { order: source.order.filter((o) => !sameId(o.id, taskId)) }
            : {}),
    };
    if (to === -1) return next;

    const target = columns[to];
    const order = Array.isArray(target.order) ? target.order : null;

    next[to] = {
        ...target,
        tasks: [...(target.tasks || []), { ...card, [field]: copyOf(value) }],
        ...(order
            ? {
                  order: order.some((o) => sameId(o.id, taskId))
                      ? order
                      : [...order, { id: card.id }],
              }
            : {}),
    };

    return next;
}
