// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

// Date ranges come with thin spaces around the dash. Compare with plain
// spaces so the expected labels stay readable.
export function plainSpaces(labels) {
    return labels.map((label) => label.replaceAll("\u2009", " "));
}
