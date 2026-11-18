// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package queryengine

import "testing"

func TestMakeTitle(t *testing.T) {
	cases := []struct {
		name string
		want string
	}{
		{"count", "Count"},
		{"online_deliveries", "Online Deliveries"},
		{"total_value_usd", "Total Value Usd"},
		{"api_url", "Api Url"},
		{"x2_value", "X2 Value"},
		{"id", "ID"},
		{"customer_id", "Customer ID"},
		{"stores_id", "Stores ID"},
		{"user_ids", "User IDs"},
		{"html_and_json", "Html and Json"},
		{"deliveredAt", "Delivered at"},
		{"and_more", "And More"},
		{"HTMLParser", "Html Parser"},
		{"userIDs", "User IDs"},
		{"sourceURLs", "Source Urls"},
		{"APIsList", "Apis List"},
		{"userIDsURLs", "User IDs Urls"},
		{"userIDsABs", "User IDs Abs"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			got := MakeTitle(testCase.name)

			if got != testCase.want {
				t.Errorf("want %q, got %q", testCase.want, got)
			}
		})
	}
}
