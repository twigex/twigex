// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package app

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"hash/fnv"
	"io"
	"math"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/twigex/twigex/internal/linkpreview"
	"github.com/twigex/twigex/model"
	"github.com/twigex/twigex/safehttp"
	"github.com/twigex/twigex/tlog"
)

const (
	linkPreviewUserAgent     = "linkpreview/1.0"
	linkPreviewTimeout       = 10 * time.Second
	linkPreviewMaxBody       = 2 << 20 // 2 MiB
	linkPreviewMaxRedirects  = 5
	linkPreviewMaxImageBytes = 5 << 20 // 5 MiB
)

var linkPreviewClient = safehttp.NewClient(safehttp.Config{
	Timeout:      linkPreviewTimeout,
	MaxRedirects: linkPreviewMaxRedirects,
})

func (a *App) AttachLinkPreview(post *model.Post) {
	if post == nil {
		return
	}

	if !*a.ConfigStore.Config.ChannelSettings.LinkPreviews {
		return
	}

	urls := extractURLs(post.Message)
	if len(urls) == 0 {
		return
	}

	linkHash := createLinkHash(urls[0])

	exists, err := a.Store.Posts.GetLinkMetadataByHash(linkHash)
	if err != nil {
		tlog.Errorw("Failed to retrieve link metadata",
			"link_hash", linkHash,
			"error", err,
		)
	}

	if exists != nil {
		if time.Since(time.UnixMilli(exists.UpdatedAt)) > 24*time.Hour {
			a.createAndAttachLinkPreview(post, urls, linkHash, true)
			return
		}

		if err = a.UpdatePostLinkMetadata(post.ChannelID, post.ID, *exists); err != nil {
			tlog.Errorw("Failed to update post link metadata",
				"post_id", post.ID,
				"channel_id", post.ChannelID,
				"error", err,
			)
		}

		return
	}

	a.createAndAttachLinkPreview(post, urls, linkHash, false)
}

func (a *App) createAndAttachLinkPreview(post *model.Post, urls []string, linkHash int64, update bool) {
	data, err := fetchAndParseLinkPreview(urls[0])
	if err != nil {
		tlog.Errorw("Failed to generate link preview",
			"url", urls[0],
			"error", err,
		)
		return
	}

	var linkData model.LinkData
	if err = json.Unmarshal(data, &linkData); err != nil {
		tlog.Errorw("Failed to unmarshal link preview data",
			"url", urls[0],
			"error", err,
		)
		return
	}

	// Link to the posted URL, not the redirected one our fetch resolved to.
	linkData.Url = urls[0]
	linkData.Video = detectVideo(urls[0])

	// YouTube serves servers a consent page with no usable metadata, so enrich
	// from the video id and oEmbed rather than the scrape.
	if linkData.Video != nil {
		a.enrichVideoMetadata(&linkData)
	}

	// A detected video is worth showing on its id/oEmbed data even if the scrape was blocked.
	if linkData.Video == nil && (linkData.Title == "" || linkData.Description == "") {
		return
	}

	// Store the image ourselves so clients don't load it from a third-party host
	// (which leaks viewer IPs and lets the origin swap it). The served URL is not
	// persisted; it is derived from the hash at read time so the route can change
	// without rewriting stored rows.
	var storageID string
	var imageSize int64
	if linkData.Image.URL != "" {
		storageID, imageSize, err = a.storeLinkPreviewImage(linkData.Image.URL, linkHash)
		if err != nil {
			tlog.Warnw("Failed to store link preview image, dropping image",
				"url", urls[0],
				"error", err,
			)
			linkData.Image = model.LinkImage{}
		} else {
			linkData.Image.URL = ""
		}
	}

	t := time.Now().UnixMilli()
	link := model.LinkMetadata{
		Hash:      linkHash,
		URL:       urls[0],
		Type:      linkData.Type,
		Data:      linkData,
		StorageID: storageID,
		Size:      imageSize,
		CreatedAt: t,
		UpdatedAt: t,
	}

	if update {
		err = a.Store.Posts.UpdateLinkMetadata(link)
	} else {
		err = a.Store.Posts.CreateLinkMetadata(link)
	}

	if err != nil {
		tlog.Errorw("Failed to save link metadata",
			"url", urls[0],
			"error", err,
		)
		return
	}

	if err = a.UpdatePostLinkMetadata(post.ChannelID, post.ID, link); err != nil {
		tlog.Errorw("Failed to update post with link metadata",
			"post_id", post.ID,
			"channel_id", post.ChannelID,
			"error", err,
		)
	}
}

func fetchAndParseLinkPreview(rawURL string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), linkPreviewTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", linkPreviewUserAgent)

	resp, err := linkPreviewClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("linkpreview: unexpected status %d", resp.StatusCode)
	}

	body := io.LimitReader(resp.Body, linkPreviewMaxBody)
	return linkpreview.Parse(body, resp.Request.URL.String(), linkpreview.Options{
		Title:       true,
		Description: true,
		Image:       true,
		Favicon:     true,
		SiteName:    true,
	})
}

func (a *App) storeLinkPreviewImage(imageURL string, linkHash int64) (string, int64, error) {
	primary, err := a.Store.Storage.GetPrimary()
	if err != nil {
		return "", 0, err
	}

	if primary == nil {
		return "", 0, errors.New("no primary storage configured")
	}

	backend, ok := a.FileStorageObjects[primary.ID]
	if !ok {
		return "", 0, fmt.Errorf("storage backend not available: %s", primary.ID)
	}

	ctx, cancel := context.WithTimeout(context.Background(), linkPreviewTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, imageURL, nil)
	if err != nil {
		return "", 0, err
	}

	req.Header.Set("User-Agent", linkPreviewUserAgent)

	resp, err := linkPreviewClient.Do(req)
	if err != nil {
		return "", 0, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", 0, fmt.Errorf("unexpected status %d", resp.StatusCode)
	}

	buf, err := io.ReadAll(io.LimitReader(resp.Body, linkPreviewMaxImageBytes))
	if err != nil {
		return "", 0, err
	}

	// Reject non-images so we never serve attacker-supplied HTML/scripts from our origin.
	if ct := http.DetectContentType(buf); !strings.HasPrefix(ct, "image/") {
		return "", 0, fmt.Errorf("not an image: %s", ct)
	}

	storagePath, err := a.buildLinkPreviewImagePath(primary.ID, linkHash)
	if err != nil {
		return "", 0, err
	}

	size := int64(len(buf))
	if err = backend.WriteFile(ctx, storagePath, bytes.NewReader(buf), size); err != nil {
		return "", 0, err
	}

	return primary.ID, size, nil
}

func (a *App) enrichVideoMetadata(data *model.LinkData) {
	if data.Video.Provider != "youtube" {
		return
	}

	// Thumbnail from the id has no consent gate; 16:9 dims let the facade crop
	// hqdefault's letterbox via object-cover.
	data.Image = model.LinkImage{
		URL:    "https://i.ytimg.com/vi/" + data.Video.ID + "/hqdefault.jpg",
		Width:  1280,
		Height: 720,
	}
	data.SiteName = "YouTube"

	oe, err := fetchYouTubeOEmbed(data.Url)
	if err != nil {
		tlog.Warnw("YouTube oEmbed failed", "url", data.Url, "error", err)
		return
	}

	if oe.Title != "" {
		data.Title = oe.Title
	}

	if oe.AuthorName != "" {
		data.Description = oe.AuthorName
	}
}

type youtubeOEmbed struct {
	Title      string `json:"title"`
	AuthorName string `json:"author_name"`
}

func fetchYouTubeOEmbed(watchURL string) (*youtubeOEmbed, error) {
	endpoint := "https://www.youtube.com/oembed?format=json&url=" + url.QueryEscape(watchURL)

	ctx, cancel := context.WithTimeout(context.Background(), linkPreviewTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", linkPreviewUserAgent)

	resp, err := linkPreviewClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d", resp.StatusCode)
	}

	var oe youtubeOEmbed
	if err := json.NewDecoder(io.LimitReader(resp.Body, 64<<10)).Decode(&oe); err != nil {
		return nil, err
	}

	return &oe, nil
}

func (a *App) buildLinkPreviewImagePath(storageID string, linkHash int64) (string, error) {
	return a.BuildFilePath(storageID, "linkpreviews/"+strconv.FormatInt(linkHash, 10), model.AppChat)
}

func (a *App) GetLinkMetadata(hash int64) (*model.LinkMetadata, *model.AppError) {
	meta, err := a.Store.Posts.GetLinkMetadataByHash(hash)
	if err != nil {
		tlog.Errorw("Failed to get link metadata", "hash", hash, "error", err)
		return nil, model.NewAppError("post.retrieval_failed", http.StatusInternalServerError)
	}

	return meta, nil
}

func (a *App) ServeLinkPreviewImage(hash int64, w http.ResponseWriter, r *http.Request) *model.AppError {
	meta, appErr := a.GetLinkMetadata(hash)
	if appErr != nil {
		return appErr
	}

	if meta == nil || meta.StorageID == "" {
		return model.NewAppError("post.retrieval_failed", http.StatusNotFound)
	}

	backend, ok := a.FileStorageObjects[meta.StorageID]
	if !ok {
		tlog.Errorw("Storage backend not available for link preview image",
			"hash", hash, "storage_id", meta.StorageID)
		return model.NewAppError("storage.not_available", http.StatusInternalServerError)
	}

	filePath, err := a.buildLinkPreviewImagePath(meta.StorageID, hash)
	if err != nil {
		tlog.Errorw("Failed to build link preview image path", "hash", hash, "error", err)
		return model.NewAppError("post.retrieval_failed", http.StatusInternalServerError)
	}

	if err = backend.ServeFile(filePath, w, r); err != nil {
		tlog.Errorw("Failed to serve link preview image", "hash", hash, "error", err)
		return model.NewAppError("post.retrieval_failed", http.StatusInternalServerError)
	}

	return nil
}

// Derive the served URL from the hash so the route stays out of the database.
func hydrateLinkPreviewImage(meta *model.LinkMetadata) {
	if meta != nil && meta.StorageID != "" {
		meta.Data.Image.URL = linkPreviewImagePath(meta.Hash)
	}
}

func linkPreviewImagePath(linkHash int64) string {
	return "/api/link-preview/image/" + strconv.FormatInt(linkHash, 10)
}

func extractURLs(s string) []string {
	// Don't match markdown images as urls
	re := regexp.MustCompile(`!\[([^\]]*)\]\(([^)]+)\)`)

	clean := re.ReplaceAllString(s, "")
	urlRegex := regexp.MustCompile(`\b(?:https?://|www\.)[^\s]+`)

	urlmatches := urlRegex.FindAllString(clean, -1)
	urls := make([]string, 0, len(urlmatches))

	for _, m := range urlmatches {
		m = strings.TrimRight(m, ".,!?;:")

		// Normalize scheme-less hosts so write and read hash the same string.
		if strings.HasPrefix(m, "www.") {
			m = "https://" + m
		}

		if _, err := url.ParseRequestURI(m); err == nil {
			urls = append(urls, m)
		}
	}

	return urls
}

func createLinkHash(url string) int64 {
	h := fnv.New64a()
	h.Write([]byte(url))
	return int64(h.Sum64() & math.MaxInt64)
}

var youtubeIDRegex = regexp.MustCompile(`^[A-Za-z0-9_-]{11}$`)

// detectVideo returns an embeddable video descriptor for known providers, or
// nil. The ID is strictly validated so nothing untrusted reaches the client's
// iframe src.
func detectVideo(rawURL string) *model.LinkVideo {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil
	}

	host := strings.TrimPrefix(strings.ToLower(u.Hostname()), "www.")
	host = strings.TrimPrefix(host, "m.")

	var id string
	switch host {
	case "youtube.com", "youtube-nocookie.com":
		switch {
		case u.Path == "/watch":
			id = u.Query().Get("v")
		case strings.HasPrefix(u.Path, "/embed/"):
			id = strings.TrimPrefix(u.Path, "/embed/")
		case strings.HasPrefix(u.Path, "/shorts/"):
			id = strings.TrimPrefix(u.Path, "/shorts/")
		}
	case "youtu.be":
		id = strings.TrimPrefix(u.Path, "/")
	default:
		return nil
	}

	if i := strings.IndexByte(id, '/'); i >= 0 {
		id = id[:i]
	}

	if !youtubeIDRegex.MatchString(id) {
		return nil
	}

	return &model.LinkVideo{Provider: "youtube", ID: id}
}
