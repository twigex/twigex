// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { describe, expect, it } from "vitest";
import { toActivityRow } from "../taskComment";

const people = {
    anna: { name: "Anna", lastname: "Ozola" },
    bob: { name: "Bob", lastname: "Berg" },
};
const deps = { getUser: (id) => people[id], translate: (key) => `<${key.split(".").pop()}>` };
const entry = (fields) => ({ id: "e1", user_id: "anna", created_at: 1700000000, ...fields });

describe("toActivityRow", () => {
    it("shows a comment as its author's text", () => {
        const row = toActivityRow(entry({ activity_type: "Comment", content: "Looks good" }), deps);

        expect(row).toMatchObject({
            type: "commented",
            userId: "anna",
            person: { name: "Anna Ozola" },
            comment: "Looks good",
        });
    });

    it("says what was done, by whom", () => {
        expect(toActivityRow(entry({ activity_type: "task_created" }), deps)).toMatchObject({
            userId: "anna",
            person: { name: "Anna Ozola" },
            comment: "<created_the_task>",
        });
    });

    it("names who a task was assigned to, and who assigned it", () => {
        const row = toActivityRow(
            entry({ activity_type: "task_assigned", affected_user: "bob" }),
            deps,
        );

        expect(row).toMatchObject({
            userId: "anna",
            person: { name: "Anna Ozola" },
            comment: "<assigned_the_task_to> Bob Berg.",
        });
    });

    it("shows an invitation as the invited person's", () => {
        const row = toActivityRow(
            entry({ activity_type: "project_invite", affected_user: "bob" }),
            deps,
        );

        expect(row).toMatchObject({
            userId: "bob",
            person: { name: "Bob Berg" },
            comment: "<was_invited_to_the_project>",
        });
    });

    it("keeps the text of an activity it does not know, and names an unknown user", () => {
        const row = toActivityRow(
            entry({ activity_type: "something_new", user_id: "gone", content: "did a thing" }),
            deps,
        );

        expect(row).toMatchObject({ person: { name: "Unknown User" }, comment: "did a thing" });
    });

    it("dates an entry from its seconds", () => {
        expect(toActivityRow(entry({ activity_type: "task_deleted" }), deps).dateTime).toBe(
            "2023-11-14T22:13:20.000Z",
        );
    });
});
