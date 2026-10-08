// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package app

import (
	"github.com/avct/uasurfer"
)

var osNames = map[uasurfer.OSName]string{
	uasurfer.OSUnknown:      "Unknown",
	uasurfer.OSWindowsPhone: "Windows Phone",
	uasurfer.OSWindows:      "Windows",
	uasurfer.OSMacOSX:       "Mac OS",
	uasurfer.OSiOS:          "iOS",
	uasurfer.OSAndroid:      "Android",
	uasurfer.OSBlackberry:   "BlackBerry",
	uasurfer.OSChromeOS:     "Chrome OS",
	uasurfer.OSKindle:       "Kindle",
	uasurfer.OSWebOS:        "webOS",
	uasurfer.OSLinux:        "Linux",
}

var browserNames = map[uasurfer.BrowserName]string{
	uasurfer.BrowserUnknown:    "Unknown",
	uasurfer.BrowserChrome:     "Chrome",
	uasurfer.BrowserIE:         "Internet Explorer",
	uasurfer.BrowserSafari:     "Safari",
	uasurfer.BrowserFirefox:    "Firefox",
	uasurfer.BrowserAndroid:    "Android",
	uasurfer.BrowserOpera:      "Opera",
	uasurfer.BrowserBlackberry: "BlackBerry",
}

func getOSName(ua *uasurfer.UserAgent) string {
	return osNames[ua.OS.Name]
}

func getBrowserName(ua *uasurfer.UserAgent) string {
	return browserNames[ua.Browser.Name]
}
