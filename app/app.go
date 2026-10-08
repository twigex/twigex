// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package app

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/twigex/twigex/app/jobs"
	"github.com/twigex/twigex/config"
	"github.com/twigex/twigex/crypto"
	"github.com/twigex/twigex/interfaces"
	"github.com/twigex/twigex/internal/licensing"
	"github.com/twigex/twigex/model"

	lksdk "github.com/livekit/server-sdk-go"
	"github.com/twigex/twigex/internal/filestore"
	"github.com/twigex/twigex/internal/i18n"
	"github.com/twigex/twigex/internal/klipy"
	"github.com/twigex/twigex/store"
	"github.com/twigex/twigex/tlog"
)

type App struct {
	Server             Server
	Logger             *tlog.Logger
	ConfigStore        config.ConfigStore
	Store              store.Store // Define all methods here
	FileStore          filestore.FileBackend
	FileStorageObjects map[string]filestore.FileBackend
	GIF                *klipy.Client
	LDAPAuth           interfaces.LDAP
	OIDCAuth           interfaces.OIDC
	CollimatoRoles     interfaces.CollimatoRoles
	WorkspaceRoles     interfaces.WorkspaceRoles
	ctx                context.Context
	cancel             context.CancelFunc
	LiveKit            *liveKitRooms
	directCalls        directCallRegistry
	QueryEngine        QueryEngine
}

func (a *App) dbCtx(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, time.Duration(*a.ConfigStore.Config.SqlSettings.QueryTimeout)*time.Second)
}

// clientLeft reports whether the caller cancelled the request, as a view
// does when the user switches away. The failure that follows is expected and
// not worth logging, unlike a query that ran past its timeout.
func clientLeft(ctx context.Context) bool {
	return errors.Is(ctx.Err(), context.Canceled)
}

func NewApp() *App {
	l := tlog.NewLogger()
	tlog.InitGlobalLogger(l)

	c, err := config.NewConfig()
	if err != nil {
		tlog.Errorw("Failed to initialize config", "error", err)
		os.Exit(1)
	}

	st, err := store.NewNewStore(c.Config)
	if err != nil {
		tlog.Errorw("Failed to initialize store", "error", err)
		os.Exit(1)
	}

	s, err := NewServer()
	if err != nil {
		tlog.Errorw("Failed to start server", "error", err)
		os.Exit(1)
	}

	err = i18n.Init("i18n")
	if err != nil {
		tlog.Errorw("Failed to initialize i18n", "error", err)
		os.Exit(1)
	}

	a := &App{
		Server:      s,
		Logger:      l,
		ConfigStore: *c,
		Store:       *st,
	}

	a.QueryEngine = &cubeEngine{
		client:    cubeHTTPClient,
		apiURL:    func() string { return *a.ConfigStore.Config.CollimatoSettings.CubeAPIURL },
		mintToken: a.GenerateToken,
	}

	if err := a.LoadSystemSettings(); err != nil {
		tlog.Errorw("Failed to load system settings", "error", err)
		os.Exit(1)
	}

	if *a.ConfigStore.Config.ChannelSettings.KlipyGIF.Enabled {
		a.GIF = klipy.New(*a.ConfigStore.Config.ChannelSettings.KlipyGIF.APIKey)
	}

	normalizedHost := a.NormalizeLiveKitHost(
		*a.ConfigStore.Config.ChannelSettings.Host,
		model.ProtocolHTTPS,
	)
	a.LiveKit = &liveKitRooms{
		RoomServiceClient: lksdk.NewRoomServiceClient(
			normalizedHost,
			*a.ConfigStore.Config.ChannelSettings.Key,
			*a.ConfigStore.Config.ChannelSettings.Secret,
		),
	}

	if ldapInterface != nil {
		a.LDAPAuth = ldapInterface(a)
	}

	if oidcInterface != nil {
		a.OIDCAuth = oidcInterface(a)
	}

	if collimatoRolesInterface != nil {
		a.CollimatoRoles = collimatoRolesInterface(a)
	}

	if workspaceRolesInterface != nil {
		a.WorkspaceRoles = workspaceRolesInterface(a)
	}

	a.Server.NotificationHub.app = a
	return a
}

func (a *App) InitJobs() error {
	a.ctx, a.cancel = context.WithCancel(context.Background())

	handlers := []jobs.Job{
		jobs.NewMoveFilesJob(a.Store.File, a),
		jobs.NewCleanupJob(a.Store.Jobs),
		jobs.NewDeleteChannelJob(a),
	}

	if a.LDAPAuth != nil {
		handlers = append(handlers, a.LDAPAuth.SyncJob())
	}

	jobs.NewRegistry(a.Store.Jobs, handlers...).Start(a.ctx)
	a.startUserInactivityMonitor()

	return nil
}

func (a *App) Start(port string) error {
	err := a.Server.Start(port)
	if err != nil {
		return err
	}

	return nil
}

func (a *App) StartTLS(port string, cert string, key string) error {
	err := a.Server.StartTLS(cert, key)
	if err != nil {
		return err
	}

	return nil
}

func (a *App) LoadLicense() error {
	now := time.Now().Unix()

	stored, err := a.Store.License.GetAll()
	if err != nil {
		return err
	}

	envStr := strings.TrimSpace(os.Getenv("TWIGEX_LICENSE"))

	var envLicense *model.License
	if envStr != "" {
		envLicense, err = licensing.VerifyLicense(envStr)
		if err != nil {
			tlog.Warnw("Failed to verify license from ENV, ignoring",
				"error", err,
			)
			envLicense = nil
		}
	}

	candidates := make([]*model.License, 0, len(stored)+1)
	envStored := false

	for _, row := range stored {
		if row.Bytes == envStr {
			envStored = true
		}

		parsed, verifyErr := licensing.VerifyLicense(row.Bytes)
		if verifyErr != nil {
			tlog.Warnw("Failed to verify a stored license, ignoring",
				"row_id", row.ID,
				"error", verifyErr,
			)
			continue
		}

		candidates = append(candidates, parsed)
	}

	// Each call inserts, so a restart with an unchanged licence adds another row.
	if envLicense != nil && !envStored {
		if err := a.Store.License.Create(envStr); err != nil {
			return err
		}

		candidates = append(candidates, envLicense)
	}

	license := model.SelectLicense(candidates, now)

	if envLicense != nil && envLicense.StartsAt <= now && envLicense.ExpiresAt >= now {
		license = envLicense
	}

	head := model.OrderLicenses(license, candidates, now)
	if head == nil {
		tlog.Infow("No license in force")
		a.Server.License = nil
		return nil
	}

	for c := head; c != nil; c = c.Next {
		c.SetDefaults()
	}

	a.Server.License = head

	if license == nil {
		tlog.Infow("No license in force, renewal queued",
			"starts_at", head.StartsAt,
		)
		return nil
	}

	tlog.Infow("License applied",
		"plan", license.SkuName,
		"expires_at", license.ExpiresAt,
	)

	return nil
}

func (a *App) LoadStorages() (map[string]filestore.FileBackend, error) {
	storages, err := a.Store.Storage.GetAll()
	if err != nil {
		return nil, err
	}

	encryptKey := *a.ConfigStore.Config.ServerSettings.AtRestEncryptKey
	so := make(map[string]filestore.FileBackend)

	for _, s := range storages {
		// Decrypt S3 secret key before use
		secretKey := s.SecretKey
		if s.Type == model.S3Storage && secretKey != "" {
			decrypted, err := crypto.Decrypt(encryptKey, secretKey)
			if err != nil {
				return nil, fmt.Errorf("failed to decrypt secret key for storage %s: %w", s.ID, err)
			}

			secretKey = decrypted
		}

		backend, err := filestore.NewFileBackend(filestore.FileBackendSettings{
			DriverName:        s.DriverName(),
			Directory:         s.Directory,
			S3AccessKeyId:     s.AccessKey,
			S3SecretAccessKey: secretKey,
			S3Bucket:          s.Bucket,
			S3Endpoint:        s.Endpoint,
			S3SSL:             s.SSL,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to initialize storage %s (%s): %w", s.Label, s.ID, err)
		}

		so[s.ID] = backend
	}

	return so, nil
}

func (a *App) startUserInactivityMonitor() {
	go func() {
		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()

		for {
			<-ticker.C
			for _, userID := range a.Server.NotificationHub.connectedUsers() {
				status, err := a.Store.User.GetStatus(userID)
				if err != nil {
					tlog.Errorw("failed to get user status", "user_id", userID, "error", err)
					continue
				}

				if status == nil {
					tlog.Errorw("user status is nil for user", "user_id", userID)
					continue
				}

				if status.Status == model.StatusAway {
					// User is already away, skip inactivity check
					continue
				}

				if status.UserDefined {
					continue // User defined status, skip inactivity check
				}

				lastActivity := time.Unix(status.LastActivity, 0)
				if time.Since(lastActivity) > 10*time.Minute {
					a.SetUserStatus(userID, model.StatusAway, false)
				}
			}

		}
	}()
}
