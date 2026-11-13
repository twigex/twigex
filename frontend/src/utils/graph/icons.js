// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import {
    BanknotesIcon,
    BriefcaseIcon,
    BuildingOffice2Icon,
    ChartBarIcon,
    CircleStackIcon,
    ClipboardDocumentListIcon,
    CubeIcon,
    DocumentTextIcon,
    GlobeAltIcon,
    HashtagIcon,
    MapPinIcon,
    ShoppingCartIcon,
    Squares2X2Icon,
    TagIcon,
    TruckIcon,
    UserIcon,
    UsersIcon,
    WrenchScrewdriverIcon,
} from "@heroicons/vue/24/outline";

// The icons a type can have, by name. The server accepts any simple name, so
// a name this list doesn't know draws the plain cube.
export const TYPE_ICONS = {
    "building-office": BuildingOffice2Icon,
    "shopping-cart": ShoppingCartIcon,
    truck: TruckIcon,
    user: UserIcon,
    users: UsersIcon,
    "document-text": DocumentTextIcon,
    "clipboard-document-list": ClipboardDocumentListIcon,
    cube: CubeIcon,
    banknotes: BanknotesIcon,
    "map-pin": MapPinIcon,
    tag: TagIcon,
    briefcase: BriefcaseIcon,
    "globe-alt": GlobeAltIcon,
    "wrench-screwdriver": WrenchScrewdriverIcon,
    "chart-bar": ChartBarIcon,
    "squares-2x2": Squares2X2Icon,
    "circle-stack": CircleStackIcon,
    hashtag: HashtagIcon,
};

export function typeIcon(name) {
    return TYPE_ICONS[name] ?? CubeIcon;
}
