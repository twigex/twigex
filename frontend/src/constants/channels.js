// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

const channelRoles = {
    Admin: "admin",
    Member: "member",
};

const channelTypes = {
    Public: "O",
    Private: "P",
    Direct: "D",
};

const postFetchDirection = {
    Previous: "previous",
    Next: "next",
    History: "history",
};

const LAST_VIEWED_CHANNEL = "last_viewed_channel";

export { channelRoles, channelTypes, postFetchDirection, LAST_VIEWED_CHANNEL };
