// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/go-redis/redis"
	"github.com/twigex/twigex/model"
	"github.com/twigex/twigex/store/redisstore"
	"github.com/twigex/twigex/store/sqlstore"
	"github.com/twigex/twigex/tlog"
)

type Store struct {
	DB             *sql.DB
	File           FileStore
	Auth           AuthStore
	Activity       ActivityStore
	Collimato      CollimatoStore
	Preferences    PreferencesStore
	Share          ShareStore
	User           UserStore
	UserPhoto      UserPhotoStore
	Roles          RoleStore
	Notifications  NotificationStore
	PasswordReset  PasswordResetStore
	License        LicenseStore
	Storage        StorageStore
	Channels       ChannelStore
	Posts          PostStore
	Migrations     MigrationStore
	Workspace      WorkspaceStore
	UploadSession  UploadSessionStore
	SystemSettings SystemSettingsStore
	Jobs           JobsStore
	OIDC           OIDCProviderStore
	Groups         GroupStore
}

const (
	DBPingAttempts    = 18
	DBPingTimeoutSecs = 10
	ExitPing          = 102
)

func SetupConnection(settings model.SqlSettings, multiStatement bool) (*sql.DB, error) {
	if multiStatement {
		dsn := *settings.DataSource + "&multiStatements=true"
		settings.DataSource = &dsn
	}

	db, err := sqlstore.NewClient(settings)
	if err != nil {
		return nil, err
	}

	for i := 0; i < DBPingAttempts; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), DBPingTimeoutSecs*time.Second)
		defer cancel()
		err = db.PingContext(ctx)
		if err == nil {
			break
		} else {
			if i == DBPingAttempts-1 {
				tlog.Errorw("failed to reach database after max attempts",
					"attempts", DBPingAttempts,
				)
				time.Sleep(time.Second)
				os.Exit(ExitPing)
			} else {
				tlog.Warnw("failed to reach database, retrying",
					"retry_in_seconds", DBPingTimeoutSecs,
					"attempt", i+1,
				)
				time.Sleep(DBPingTimeoutSecs * time.Second)
			}
		}
	}

	return db, nil
}

func NewNewStore(c *model.ServerConfig) (*Store, error) {
	dbType := c.SqlSettings.DriverName

	auth, err := redisstore.NewRedisClient(&redis.Options{
		Addr:     *c.RedisSettings.Address,
		Password: *c.RedisSettings.Password, // no password set
		DB:       0,                         // use default DB
	})
	if err != nil {
		return nil, err
	}

	if *dbType == "mysql" {
		db, err := SetupConnection(c.SqlSettings, false)
		if err != nil {
			return nil, err
		}

		files, err := sqlstore.NewFileRepository(db)
		if err != nil {
			return nil, err
		}

		users, err := sqlstore.NewUserRepository(db)
		if err != nil {
			return nil, err
		}

		userPhotos := sqlstore.NewUserPhotoRepository(db)
		role, err := sqlstore.NewRoleRepository(db)
		if err != nil {
			return nil, err
		}

		collimato, err := sqlstore.NewCollimatoRepository(db)
		if err != nil {
			return nil, err
		}

		share, err := sqlstore.NewShareRepository(db)
		if err != nil {
			return nil, err
		}

		activity, err := sqlstore.NewActivityRepository(db)
		if err != nil {
			return nil, err
		}

		preferences, err := sqlstore.NewPreferencesRepository(db)
		if err != nil {
			return nil, err
		}

		notifications, err := sqlstore.NewNotificationRepository(db)
		if err != nil {
			return nil, err
		}

		pwReset, err := sqlstore.NewPasswordResetRepository(db)
		if err != nil {
			return nil, err
		}

		license, err := sqlstore.NewLicenseRepository(db)
		if err != nil {
			return nil, err
		}

		storage, err := sqlstore.NewStorageRepository(db)
		if err != nil {
			return nil, err
		}

		channels, err := sqlstore.NewChannelsRepository(db)
		if err != nil {
			return nil, err
		}

		posts, err := sqlstore.NewPostRepository(db)
		if err != nil {
			return nil, err
		}

		migrations, err := sqlstore.NewMigrationRepository(db)
		if err != nil {
			return nil, err
		}

		workspace, err := sqlstore.NewWorkspaceRepository(db)
		if err != nil {
			return nil, err
		}

		uploadSession := sqlstore.NewUploadSessionRepository(db)
		systemSettings := sqlstore.NewSystemSettingsRepository(db)

		jobs := sqlstore.NewJobsRepository(db)
		oidc := sqlstore.NewOIDCProviderRepository(db)
		groups, err := sqlstore.NewGroupRepository(db)
		if err != nil {
			return nil, err
		}

		return &Store{
			DB:             db,
			File:           files,
			Auth:           auth,
			Activity:       activity,
			Collimato:      collimato,
			Preferences:    preferences,
			Share:          share,
			User:           users,
			UserPhoto:      userPhotos,
			Roles:          role,
			Notifications:  notifications,
			PasswordReset:  pwReset,
			License:        license,
			Storage:        storage,
			Channels:       channels,
			Posts:          posts,
			Migrations:     migrations,
			Workspace:      workspace,
			UploadSession:  uploadSession,
			SystemSettings: systemSettings,
			Jobs:           jobs,
			OIDC:           oidc,
			Groups:         groups,
		}, nil
	}

	return &Store{}, nil
}

type FileStore interface {
	GetByIDs(user model.User, IDS []string) ([]model.File, error)
	Get(id string) (*model.File, error)
	ResolveEffectiveShare(ctx context.Context, userID string, fileID string) ([]model.File, error)
	GetByParent(parent model.File, userID string) ([]model.File, error)
	GetChildren(ctx context.Context, folderID string) ([]model.File, error)
	GetBreadcrumbs(ctx context.Context, userID string, fileID string) ([]model.File, error)
	MoveAllowedWithinShare(ctx context.Context, userID string, fileID string, targetID string) (bool, error)
	GetGrantingRoot(ctx context.Context, fileID string, shareType int, shareWith string) (string, error)
	InheritedAccessLevel(ctx context.Context, fileID string, shareType int, shareWith string) (model.AccessLevel, error)
	GetSubFilesByFolder(folderID string, storage string) ([]model.File, error)
	CanMoveFolder(id string, targetID string) (bool, error)
	GetUserRoot(ctx context.Context, userID string, storage string) (*model.File, error)
	RenameStorageDrives(ctx context.Context, storageID string, name string) error
	DeleteDrive(id string) error
	GetStorageUsed(ownerID string) (int64, error)
	GetTotalStorage() (int64, error)
	GetFavourites(ctx context.Context, userID string) ([]model.File, error)
	GetFavouriteIDs(userID string, fileIDs []string) ([]string, error)
	GetDeleted(userID string) ([]model.File, error)
	Create(file model.File) (*string, error)
	CreateFolder(file model.File) (*string, error)
	UpdateParentID(rootIDS []string, subfileIDS []string, targetParentID string, storage string) error
	Delete(id string, storage string) (bool, error)
	Rename(file model.File, name string, newName string) error
	MoveToStorage(lockOwnerID string, rootIDS, subIDS []string, targetStorageID, targetParentID string) error
	Update(file model.File) error // Use this for updating single file
	AddFavourite(userID string, fileID string) error
	RemoveFavourite(userID string, fileID string) error
	GetFavourite(userID string, fileID string) (string, error)
	Trash(userID string, files, subFiles []model.File) error
	Restore(userID string, id []string) error
	CreateShare(file model.File, shareType int, user string, targetUser string, expiration int64, level model.AccessLevel) (*model.SharedFile, error)
	DeleteShares(fileID string) error
	GetShared(userID string) ([]model.File, error)
	UpdateSize(size int64, fileID string) error
	RemoveUserFromShare(fileID string, userID string) error
	RemoveGroupFromShare(ctx context.Context, fileID, groupID string) error
	GetRecents(ctx context.Context, userID string) ([]model.File, error)
	UpdateUserPermission(p model.FileSharePatch) error
	CreateMetadata(name string, metaType string, values any) (*model.FileMetadata, error)
	GetAllMetadata() ([]model.FileMetadata, error)
	CreateMetadataEntry(id string, value string, metadata model.FileMetadata) (*model.FileMetadataEntry, error) // metaid un metadata ir viens un tas pats!
	GetMetadataByID(id string) (*model.FileMetadata, error)
	DeleteMetadataEntry(fileID string, id string) error
	DeleteMetadata(id string) error
	UpdateMetadata(id string, name string, metaType string, values interface{}) error
	GetMetadataIDForFile(id string) (*string, error)
	UpdateMetadataEntry(id string, value interface{}) error
	GetMetadataEntriesForFiles(ids []string) ([]model.FileMetadataEntry, error)
	Lock(lockOwner string, fileIDs []string, ttl time.Duration) error
	Unlock(lockOwner string) error
	RenewLock(lockOwner string, ttl time.Duration) error
}

type AuthStore interface {
	Set(key string, expireTime time.Duration, value interface{}) error
	Get(key string, dest interface{}) error
	GetSessions(userID string) ([]model.SessionDetails, error)
	UpdateSession(session model.SessionDetails, sessionLength int) error
	UpdateLastActivity(session model.SessionDetails) error
	GetUserID(r *http.Request) (*string, error)
	Delete(key string) error
	Incr(key string, ttl time.Duration) (int64, error)
}

type ActivityStore interface {
	Create(userID, app, activityType, parentID, itemID, parameters string) error
	GetForFile(file model.File) ([]model.Activity, error)
}

type CollimatoStore interface {
	CreateConnection(connection model.NewConnection) (*model.Connection, error)
	GetConnections(workspace string) ([]model.Connection, error)
	GetConnectionByID(id string) (*model.Connection, error)
	UpdateConnection(connection model.Connection) error
	DeleteConnection(id string) error
	AddWorkspaceUsers(workspaceID string, users []string, roles []string) error
	RemoveWorkspaceUser(workspaceID string, userID string) error
	GetWorkspaceUsers(workspaceID string) ([]model.CollimatoWorkspaceUser, error)
	GetWorkspaceUserByUserID(userID, workspaceID string) (*model.CollimatoWorkspaceUser, error)
	CreateChart(chart model.NewChart) (*model.Chart, error)
	GetCharts(workspaceID string) ([]model.Chart, error)
	GetChartByID(id string) (*model.Chart, error)
	UpdateChart(id string, chart model.Chart) (*model.Chart, error)
	DeleteChart(id string) error
	GetDashboards(workspaceID string) ([]model.Dashboard, error)
	GetDashboardByID(id string) (*model.Dashboard, error)
	CreateDashboard(dashboard model.Dashboard) (*model.Dashboard, error)
	UpdateDashboard(dashboard model.Dashboard) (*model.Dashboard, error)
	DeleteDashboard(id string) error
	AddDashboardFilter(dashboardID string, filter model.DashboardFilter) (*model.DashboardFilter, error)
	UpdateDashboardFilter(filter model.DashboardFilter) error
	DeleteDashboardFilter(dashboardID string, filterID string) error
	GetDashboardFilterByID(dashboardID string, filterID string) (*model.DashboardFilter, error)
	AddDashboardCharts(userID string, dashboardID string, charts []model.DashboardChartPatch) error
	CreateWorkspace(userID, name, description, status, secret string) (*model.CollimatoWorkspace, error)
	DeleteWorkspace(workspaceID string) error
	GetWorkspacesForUser(userID string) ([]model.CollimatoWorkspace, error)
	GetWorkspaces() ([]model.CollimatoWorkspace, error)
	CountWorkspaces(ctx context.Context) (int, error)
	GetWorkspaceByID(id string) (*model.CollimatoWorkspace, error)
	UpdateWorkspace(workspace model.CollimatoWorkspace) (*model.CollimatoWorkspace, error)
	GetWorkspaceRoles(workspaceID string) ([]model.CollimatoRole, error)
	GetWorkspaceRoleByID(workspaceID, roleID string) (*model.CollimatoRole, error)
	CreateWorkspaceRole(role model.CollimatoRole) (*model.CollimatoRole, error)
	UpdateWorkspaceRole(workspaceID string, role model.CollimatoRole) error
	DeleteWorkspaceRole(workspaceID, roleID string) error
	UpdateUserRoles(workspaceID string, userID string, roles []string) error
	GetRolesByName(roleNames []string, workspaceID string) ([]model.CollimatoRole, error)
	AddWorkspaceGroups(ctx context.Context, workspaceID string, groupIDs []string, roles []string, addedBy string) error
	GetWorkspaceGroups(ctx context.Context, workspaceID string) ([]model.CollimatoWorkspaceGroup, error)
	GetWorkspaceGroupMembers(ctx context.Context, workspaceID string) ([]model.CollimatoWorkspaceUser, error)
	GetWorkspaceMemberIDs(ctx context.Context, workspaceIDs []string) (map[string][]string, error)
	RemoveWorkspaceGroup(ctx context.Context, workspaceID, groupID string) error
	UpdateWorkspaceGroupRoles(ctx context.Context, workspaceID, groupID string, roles []string) error
	GetEffectiveRolesForUser(ctx context.Context, workspaceID, userID string) ([]string, error)
	GetWorkspaceFiles(workspaceID, fileType string) ([]model.CollimatoWorkspaceFile, error)
	GetWorkspaceFile(workspaceID, name, fileType string) (*model.CollimatoWorkspaceFile, error)
	UpdateWorkspaceFile(file model.CollimatoWorkspaceFile) (*model.CollimatoWorkspaceFile, error)
	DeleteWorkspaceFile(workspaceID, name, fileType string) error
}

type PreferencesStore interface {
	GetForUser(userID, category string) ([]model.Preference, error)
	GetNotifications(userID, category string) ([]model.Preference, error)
	Update(preferences []model.Preference) error
}

type SystemSettingsStore interface {
	Get(name string) (string, error)
	GetByPrefix(prefix string) (map[string]string, error)
	UpdateBatch(settings map[string]string) error
}

type ShareStore interface {
	Create(ctx context.Context, userID string, shareToken string, share model.Request) (*model.SharedLinks, error)
	Delete(ctx context.Context, owner string, token string) error
	Update(ctx context.Context, owner string, source string, share model.LinkUpdate) error
	GetByToken(ctx context.Context, token string) (*model.ExternalShare, error)
	GetActiveForFile(ctx context.Context, fileID string) ([]model.ExternalShare, error)
	DeactivateForFile(ctx context.Context, fileID string) error
	IncrementAccessed(ctx context.Context, token string) error
	IncrementDownloaded(ctx context.Context, token string) error
}

type UserPhotoStore interface {
	Get(ctx context.Context, userID string) (*model.UserPhoto, error)
	Replace(ctx context.Context, photo model.UserPhoto) (*model.UserPhoto, error)
	Delete(ctx context.Context, userID, photoID string) (bool, error)
	CountByStorage(ctx context.Context, storageID string) (int, error)
}

type UserStore interface {
	Create(user model.NewUser) (*model.User, error)
	Update(user model.UserPatch) error
	Deactivate(id string, deactivatedAt int64) (int64, error)
	UpdateTimezone(id string, timezone []byte) (bool, error)
	UpdateName(id string, name string, lastname string) (bool, error)
	UpdateEmail(id string, email string) (bool, error)
	UpdateUsername(id string, username string) (bool, error)
	UpdateProfile(id string, email string, name string, lastname string) (bool, error)
	UpdatePassword(id string, password string) (bool, error)
	GetAll() ([]model.User, error)
	GetAllPaged(ctx context.Context, query string, sort model.Sort, limit, offset int, includeDeactivated bool) ([]model.User, error)
	Count(ctx context.Context, query string, includeDeactivated bool) (int, error)
	Search(ctx context.Context, query string, limit int) ([]model.User, error)
	GetByEmail(email string) (*model.User, error)
	GetByAuthData(providerID, ID string) (*model.User, error)
	Get(id string) (*model.User, error)
	CountActive(ctx context.Context) (int, error)
	GetByUsername(username string) (*model.User, error)
	GetByIDs(ids []string) ([]model.User, error)
	GetByUsernames(ctx context.Context, usernames []string) ([]model.User, error)
	GetByAuthService(ctx context.Context, authService string) ([]model.User, error)
	GetActiveByIDs(ctx context.Context, ids []string) ([]model.User, error)
	GetSharedUsers(fileID string) ([]model.User, error) // use GetSharedUsersForFile
	GetSharedUsersForFile(ctx context.Context, file model.File) ([]model.SharedUsers, error)
	UpdateMfaSecret(user model.User, secret string) error
	UpdateMfaActive(user model.User, status bool) error
	UpdateAuthData(id string, authData string) (bool, error)
	UpdateStatus(userID string, status string, userDefined bool) (*model.UserStatus, error)
	GetStatus(userID string) (*model.UserStatus, error)
	GetStatuses() ([]model.UserStatus, error)
}

type RoleStore interface {
	GetAll() ([]model.Role, error)
	GetByID(id string) (*model.Role, error)
	GetByName(name string) (*model.Role, error)
	GetByNames(roleNames []string) ([]model.Role, error)
	CreateOrUpdate(role model.Role) (*model.Role, error)
	Delete(id string) error
}

type NotificationStore interface {
	Create(sender model.User, receiver string, app string, notificationType string, item string, props map[string]interface{}) (*model.NotificationMessage, error)
	CreateBulk(sender model.User, receivers []string, app string, notificationType string, item string, props map[string]interface{}) ([]model.NotificationMessage, error)
	Delete(id string, user string) error
	GetAllForReceiver(receiverID string) ([]model.NotificationMessage, error)
	GetByID(id string) (*model.NotificationMessage, error)
	MarkAsSent(messages []model.NotificationMessage) error
	MarkAsRead(id []string, receiver string) error
}

type PasswordResetStore interface {
	Create(user model.User, token string) error
	Delete(token string) error
	Get(token string) (*model.PasswordReset, error)
}

type LicenseStore interface {
	Create(licenseBytes string) error
	DeleteActive() error
	GetAll() ([]model.ActiveLicense, error)
}

type StorageStore interface {
	Create(storage *model.Storage) (*model.Storage, error)
	GetAll() ([]model.Storage, error)
	GetByID(id string) (*model.Storage, error)
	GetPrimary() (*model.Storage, error)
	UpdatePrimary(id string) error
	Delete(id string) error
	UpdateCredentials(id string, storage *model.Storage) error
	Update(id string, storage *model.Storage) error
}

type ChannelStore interface {
	Create(ch *model.Channel) (*model.Channel, error)
	Update(channel *model.Channel) error
	GetAll() ([]model.Channel, error)
	GetAllByType(channelType string) ([]model.Channel, error)
	GetDirectMessage(userID, userID2 string) (*model.Channel, error)
	Get(channelID string) (*model.Channel, error)
	UpdateArchived(ctx context.Context, channelID string, archivedAt int64) error
	DeletePermanently(ctx context.Context, channelID string) error
	CreateOrPromoteMembers(ctx context.Context, members []model.ChannelMember) error
	RemoveUser(channelID, user string) error
	UpdateUserRole(channelID, user, role string) error
	IsMember(channelID, userID string) (bool, error)
	GetMembers(channelID string) ([]model.ChannelMember, error)
	SearchMembers(ctx context.Context, channelID, query string, limit int) ([]model.User, error)
	GetMembersByUsernames(ctx context.Context, channelID string, usernames []string) ([]model.User, error)
	GetMembersByIDs(ctx context.Context, channelID string, ids []string) ([]model.User, error)
	IncrementMentionCount(ctx context.Context, channelID string, userIDs []string) error
	IncrementAllMentionCounts(ctx context.Context, channelID, excludeUserID string) error
	GetDirectMembersForUser(channelID, userID string) ([]model.ChannelMember, error)
	GetAllForUser(userID string) ([]model.Channel, error)
	CreateMeeting(ctx context.Context, meeting model.ChannelMeeting) error
	GetMeeting(ctx context.Context, id string) (*model.ChannelMeeting, error)
	GetActiveMeetings(ctx context.Context, channelID string) ([]model.ChannelMeeting, error)
	GetScheduledMeetings(ctx context.Context, channelID string) ([]model.ChannelMeeting, error)
	StartMeeting(ctx context.Context, id string, startedAt int64) error
	EndMeeting(ctx context.Context, id string, endedAt int64) error
	UpdateMeetingHost(ctx context.Context, id, hostID string) error
	UpdateMeetingInviteOnly(ctx context.Context, id string, inviteOnly bool) error
	UpdateMeeting(ctx context.Context, meeting model.ChannelMeeting) (int, error)
	IncrementMeetingSequence(ctx context.Context, id string) (int, error)
	GetMeetingsForUser(ctx context.Context, userID string) ([]model.ChannelMeeting, error)
	AddMeetingInvitees(ctx context.Context, meetingID string, userIDs []string) error

	CreateVideoGuestLink(ctx context.Context, links []model.ChannelGuestLink) error
	GetVideoGuestLink(ctx context.Context, id string) (*model.ChannelGuestLink, error)
	TouchVideoGuestLink(ctx context.Context, id string, usedAt int64) error
	RevokeVideoGuestLink(ctx context.Context, id string, revokedAt int64) error
	RevokeMeetingGuestLinks(ctx context.Context, meetingID string, revokedAt int64) error
	GetActiveMeetingGuestLinks(ctx context.Context, meetingID string, now int64) ([]model.ChannelGuestLink, error)

	AddGroups(ctx context.Context, channelID string, groupIDs []string, addedBy string) error
	RemoveGroup(ctx context.Context, channelID, groupID string) error
	GetGroups(ctx context.Context, channelID string) ([]model.ChannelGroup, error)
	MaterializeGroups(ctx context.Context, channelID string, groupIDs []string) error
	DematerializeGroup(ctx context.Context, channelID, groupID string) error
	MaterializeUserInGroupChannels(ctx context.Context, userIDs []string, groupID string) error
	DematerializeUserFromGroupChannels(ctx context.Context, userID, groupID string) error
	RemoveGroupFromAllChannels(ctx context.Context, groupID string) error

	AddMeetingGroups(ctx context.Context, meetingID string, groupIDs []string, addedBy string) error
	DeleteMeetingGroups(ctx context.Context, meetingID string) error
	GetMeetingGroups(ctx context.Context, meetingID string) ([]model.ChannelMeetingGroup, error)
	UserInMeetingGroup(ctx context.Context, meetingID, userID string) (bool, error)
	GetMeetingGroupMemberIDs(ctx context.Context, meetingID string) ([]string, error)
}

type PostStore interface {
	GetAllForChannel(ctx context.Context, channelID, postID, param string) (*model.PostResponse, error)
	Get(ctx context.Context, postID string) (*model.Post, error)
	Create(post model.Post) (*model.Post, error)
	Update(id string, message string, remove []string, attachments []model.PostFileAttachment) (bool, error)
	Delete(postID string) error
	GetBatchForChannel(ctx context.Context, channelID string, limit int) ([]string, error)
	GetAttachmentsForMany(ctx context.Context, postIDs []string) ([]model.PostFileAttachment, error)
	GetByIDs(ctx context.Context, postIDs []string) ([]model.Post, error)
	DeleteBatch(ctx context.Context, postIDs []string) error
	MarkAllRead(userID, channelID string) error
	CreateReaction(postID, channelID, userID, reaction string) (*model.PostReaction, error)
	DeleteReaction(postID, channelID, userID, reaction string) error
	CreateAttachment(files []model.PostFileAttachment) ([]model.PostFileAttachment, error)
	GetAttachments(postID string) ([]model.PostFileAttachment, error)
	GetAttachmentByID(id string) (*model.PostFileAttachment, error)
	CreateLinkMetadata(metadata model.LinkMetadata) error
	GetLinkMetadataByHash(hash int64) (*model.LinkMetadata, error)
	UpdateLinkMetadata(metadata model.LinkMetadata) error
}

type MigrationStore interface {
	Create(name string) error
	IsMigrated(migrationName string) (bool, error)
}

type WorkspaceStore interface {
	RenameTaskAndKanbanSectionTx(workspaceID, tableID, taskID, field, newName, oldName, linkedTableName, linkedTableID string) error
	CreateTask(workspaceID, tableID, name, userID, parentTaskID, taskID string, timestamp int64, fields map[string]any) (*map[string]interface{}, error)
	SyncTaskKanban(workspaceID, tableID, taskID string, selects map[string]string, task map[string]interface{}) error
	CreateTx(workspaceID, userID, name, description, prefix string, adminRole, userRole model.ProjectWorkspaceRole, member model.WorkspaceMember) error
	UpdateTaskTx(workspaceID, tableID, taskID, field, value string, typedValue interface{}, hasUpdatedAt, isKanbanField, syncColors bool) (*map[string]interface{}, error)
	GenerateUniquePrefix() (string, error)
	GetAll(ctx context.Context) ([]model.Workspace, error)
	GetAllForUser(userID string) ([]model.Workspace, error)
	GetRow(workspaceID string) (*model.Workspace, error)
	GetFolders(workspaceID string) ([]model.WorkspaceFolder, error)
	GetTableViewMetadata(workspaceID, userID string, tableIDs []string) (map[string]*model.WorkspaceTableViewMetadata, error)
	GetTableList(workspaceID, userID string) ([]model.WorkspaceTable, error)
	GetAllTablesBasic(workspaceID string) ([]model.WorkspaceTable, error)
	CreateTable(id string, workspaceID string, name string, safeName string, linked bool, parent_table_id string, userID string, single_select bool, linkBothDirections bool, secondTableID string, folderID string, taskOrderJSON string, defaultStatuses []model.KanbanStatusOption, statusTableID string, statusTablePhysicalName string, mainViewID string) (*model.WorkspaceTable, error)
	GetTableParentID(tableID string) string
	GetLinkedTableIDByName(tableID, name string) string
	GetTableColumnTypes(tableName string) ([]*sql.ColumnType, error)
	GetTasksByDateRange(ctx context.Context, tableName string, from, to int64, filter model.SQLFilter, limit int) ([]map[string]interface{}, int, error)
	GetTasksByCalendarRange(ctx context.Context, tableName string, from, to int64, filter model.SQLFilter) ([]map[string]interface{}, error)
	GetLinkedRecordsLite(ctx context.Context, tableName, search string, ids []string, limit, offset int, access model.SQLFilter) ([]map[string]interface{}, error)
	GetTaskAssignees(ctx context.Context, tableName string, taskIDs []string) (map[string]string, error)
	GetTableRows(tableName string) ([]map[string]interface{}, error)
	GetTableRowsFiltered(ctx context.Context, tableName string, filter model.SQLFilter) ([]map[string]interface{}, error)
	GetRootTasksPagedWithSubtasks(ctx context.Context, tableName string, access, filter model.SQLFilter, sortSQL string, limit, offset int, count bool) ([]map[string]interface{}, int, int, error)
	CountBranches(ctx context.Context, tableName string, access, filter model.SQLFilter) (int, error)
	GetTableRowByID(tableName, itemID string) (map[string]interface{}, error)
	GetBranchPosition(ctx context.Context, tableName, taskID string, access, filter model.SQLFilter, sortSQL string) (int, error)
	GetSubtaskIDs(ctx context.Context, tableName, taskID string, skipStatusIDs []string, access model.SQLFilter) ([]string, error)
	GetTaskParentID(tableName, taskID string) (string, error)
	BuildFilterSQLInline(filters model.FilterPayload, headers []model.WorkspaceHeaders, loc *time.Location, tableNames map[string]string) model.SQLFilter
	GetTableNamesByIDs(ids []string) (map[string]string, error)
	GetLinkedIDs(linkedID string, itemTableID string) ([]model.LinkedItem, error)
	GetLinkedIDsBatch(ctx context.Context, linkedID string, taskIDs []string) (map[string][]model.LinkedItem, error)
	GetSingleSelectValues(linkedID string, itemID string) (*model.TaskOrderField, error)
	GetStatusTypeMap(tableID string) (map[string]string, error)
	GetStatusOptions(ctx context.Context, tableID string) ([]map[string]string, error)
	CreateTableField(params model.CreateFieldParams) (*model.WorkspaceHeaders, error)
	ChangeTaskLinks(ctx context.Context, workspaceID, tableID, taskID, field string, add, remove []string) (*map[string]interface{}, error)
	GetMembers(workspaceID string) ([]model.WorkspaceMember, error)
	GetMembersForWorkspaces(ctx context.Context, workspaceIDs []string) (map[string][]model.WorkspaceMember, error)
	GetGroupCounts(ctx context.Context, workspaceIDs []string) (map[string]int, error)
	GetUserRoleNamesByWorkspace(ctx context.Context, userID string, workspaceIDs []string) (map[string][]string, error)
	GetRolesForWorkspaces(ctx context.Context, workspaceIDs []string) (map[string][]model.ProjectWorkspaceRole, error)
	GetTablePermissionsForRoleIDs(ctx context.Context, roleIDs []string) ([]model.TablePermission, error)
	AddMember(members []model.WorkspaceMember) ([]model.WorkspaceMember, error)
	CreateView(workspaceID string, tableID string, order string, name string, viewType string, userID string, parentTableID string, viewID string) (*model.WorkspaceView, error)
	ViewExistsByName(tableID, name string) (bool, error)
	GetTableName(tableID string) (string, error)
	GetTableWorkspaceID(ctx context.Context, tableID string) (string, error)
	GetMemberUserIDs(ctx context.Context, workspaceID string, userIDs []string) ([]string, error)
	GetViewVisibility(ctx context.Context, viewID string) (bool, string, error)
	GetView(ctx context.Context, viewID string) (*model.WorkspaceView, error)
	UpdateTaskOrderGridField(workspaceID string, tableID string, field model.WorkspaceHeaders) error
	GetLinkedTableID(tableID, fieldName string) (string, error)
	IsTableSingleSelect(tableID string) (bool, error)
	IsOptionTableOf(ctx context.Context, tableID, optionTableID string) (bool, error)
	GetKanbanStatusOptions(tableName string) ([]model.KanbanStatusOption, error)
	GetTaskFieldValue(tableName, taskID, fieldName string) (string, error)
	GetColumnNames(ctx context.Context, tableName string) ([]string, error)
	TableHasColumn(tableName, column string) (bool, error)
	UpdateView(itemID string, workspaceID string, tableID string, order string, name string, viewType string) (*model.WorkspaceView, error)
	UpdateViewName(itemID string, workspaceID string, tableID string, name string) error
	ChangeViewOrder(ctx context.Context, viewID, workspaceID, tableID string, change func(order string) (string, error)) (string, error)
	GetKanbanRestIDs(ctx context.Context, tableName string, filter model.SQLFilter, upTo string, limit int) ([]string, error)
	UpdateViewPublic(viewID, workspaceID, tableID string, isPublic bool) error
	GetViewSharedUserIDs(viewID string) ([]string, error)
	UpdateViewShares(viewID, workspaceID string, userIDs []string) error
	GetSharedViewIDsForUser(userID, workspaceID string) ([]string, error)
	CreateFieldValue(workspaceID string, tableID string, linkedTableID string, tableName string, name string, field string, optionID string, color string) (*model.TaskOrderField, error)
	GetLinkedTaskData(parentTableID string, parentTaskID string) (*model.TaskOrderField, error)
	UpdateFieldDisplayName(workspaceID, tableID, fieldName, displayName string) error
	UpdateFieldFormula(ctx context.Context, workspaceID, tableID, fieldName string, formula *model.FormulaSpec) error
	CreateMemberToTask(workspaceID string, tableID string, taskID string, userID string, fieldName string) error
	Delete(workspaceID string) (bool, error)
	DeleteTable(workspaceID string, tableID string, deletedAt int64) ([]map[string]interface{}, error)
	GetTableMeta(tableID string) (tableName string, isSingleSelect bool, err error)
	GetSingleSelectCascadeInfo(tableID string) (parentTableName, fieldName string, err error)
	GetJunctionTablesForLinkedTable(tableID string) ([]string, error)
	GetJunctionTablesForMainTable(tableID string) ([]string, error)
	DeleteTaskTx(tableName string, isSingleSelect bool, taskIDs []string, cascades []model.DeleteCascade) (bool, error)
	DeleteTableView(workspaceID string, tableID string, viewID string) (bool, error)
	DeleteTableField(workspaceID string, tableID string, fieldID string, parentFieldID string) error
	GetFieldParentInfo(fieldID string) (parentFieldID string, parentTableID string, err error)
	DeleteTableSingleField(workspaceID string, tableID string, fieldID string, linkedTableID string, linkedTableName string) (bool, error)
	GetTableMetasForUser(userID string) ([]model.WorkspaceTable, error)
	GetTasksAcrossTables(ctx context.Context, sources []model.TaskSource, limit int) ([]map[string]interface{}, error)
	GetTableRowsPage(ctx context.Context, tableName string, filter model.SQLFilter, firstIDs []string, limit int) ([]map[string]interface{}, error)
	TaskOrParentMatches(ctx context.Context, tableName, taskID string, filter model.SQLFilter) (bool, error)
	GetTableHeaderMeta(tableID string) (*model.TableHeaderMeta, error)
	GetSingleSelectOptions(ctx context.Context, linkedID string, ids []string) (map[string]*model.TaskOrderField, error)
	StreamTableRows(ctx context.Context, tableName string, filter model.SQLFilter, orderSQL string, fn func(map[string]interface{}) error) error
	GetTaskPageKeys(ctx context.Context, sources []model.TaskSource, sort []model.SortParam, limit, offset int) ([]model.TaskKey, error)
	GetRoleNamesForUsers(ctx context.Context, workspaceID string, userIDs []string) (map[string][]string, error)
	GetAttachmentNames(ctx context.Context, tableID, fieldID string, taskIDs []string) (map[string][]string, error)
	GetAttachmentsForTasks(ctx context.Context, tableID, fieldID string, taskIDs []string) (map[string][]map[string]interface{}, error)
	CountTasksAcrossTables(ctx context.Context, sources []model.TaskSource, groups []model.SQLFilter) ([]int, error)
	GetComments(taskID string) ([]model.TaskComment, error)
	GetTaskActivities(taskID string) ([]model.Activity, error)
	CreateTaskComment(workspaceID string, tableID string, taskID string, comment string, userID string, commentID string) (*model.TaskComment, error)
	GetTaskAssigneeID(tableName, taskID string) (string, error)
	GetMemberByUserID(workspaceID, userID string) (*model.WorkspaceMember, error)
	GetTaskNameByID(workspaceID string, tableID string, taskID string) (*string, error)
	GetNameByID(workspaceID string) (*string, error)
	CreateFolder(ctx context.Context, workspaceID, name, parentFolderID, folderID string) (*model.WorkspaceFolder, error)
	CreateFilter(workspaceID, tableID string, filters model.FilterPayload, name string, isPrivate bool, userID, id string) (string, error)
	Update(workspaceID, name string, description *string) (*string, error)
	UpdateTable(workspaceID string, tableID string, name string) (*string, error)
	DeleteMember(workspaceID string, memberID string) (bool, error)
	DeleteFolder(ctx context.Context, workspaceID, folderID string, withTables bool) ([]string, []map[string]interface{}, error)
	GetFolderTableIDs(ctx context.Context, workspaceID, folderID string) ([]string, error)
	GetLinkingTables(ctx context.Context, workspaceID string, tableIDs []string) ([]model.LinkingTable, error)
	UpdateFolder(workspaceID string, folderID string, name string) (bool, error)
	MoveTable(ctx context.Context, workspaceID, tableID, folderID string) error
	MoveFolder(ctx context.Context, workspaceID, folderID, parentID string) error
	GetTableMetas(workspaceIDs []string) ([]model.WorkspaceTable, error)
	GetTablesByIDs(tableIDs []string) ([]model.WorkspaceTable, error)
	GetUserTimezone(userID string) (*model.UserTimezone, error)
	UpdateMemberRole(workspaceID string, memberID string, role string, user model.User) (bool, error)
	GetViewItemOrder(workspaceID, tableID, viewID string) ([]byte, error)
	GetGridViewItemOrder(workspaceID, tableID, viewID string) ([]byte, error)
	UpdateViewItemOrder(workspaceID, tableID, viewID string, itemOrder []byte) error
	GetCustomFields(workspaceID, tableID string) ([]model.TaskOrderField, error)
	UpdateGridSort(workspace_id, table_id, view_id, sortData, sortID string) (bool, error)
	GetUserByUserID(userID, workspaceId string) (*model.WorkspaceMember, error)
	GetRolesByName(roleNames []string, workspaceID string) ([]model.ProjectWorkspaceRole, error)
	CreateRole(role model.ProjectWorkspaceRole) (*model.ProjectWorkspaceRole, error)
	GetRoleByID(workspaceID, roleID string) (*model.ProjectWorkspaceRole, error)
	UpdateRole(workspaceID string, role *model.ProjectWorkspaceRole) error
	GetRoles(workspaceID string) ([]model.ProjectWorkspaceRole, error)
	DeleteRole(ctx context.Context, workspaceID, roleID, roleName, fallback string) error
	GetAllActiveIDs() ([]string, error)
	GetIDsForGroup(groupID string) ([]string, error)
	UserHasOtherGroupAccess(workspaceID, groupID, userID string) (bool, error)
	UserHasAnyGroupAccess(workspaceID, userID string) (bool, error)
	ClearAssigneeByUserID(tableName, fieldName, userID string) error
	GetPersonFieldNamesForTable(tableID string) ([]string, error)
	GetFieldType(ctx context.Context, tableID, fieldName string) (string, error)
	GetAllIDsForUser(userID string) ([]string, error)
	GetStatusNameByID(tableName, id string) (string, error)
	GetMainViewByTableID(workspaceID, tableID string) (string, error)
	GetMainViewIDs(workspaceByTable map[string]string) (map[string]string, error)
	UpdateColumnWidth(workspaceID, tableID, viewID, columnName string, width int) error
	IsMember(workspaceID, userID string) (bool, error)
	UpdateAttachment(attachment *model.WorkspaceAttachment) error
	GetAttachment(fileID, workspace_id string) (*model.WorkspaceAttachment, error)
	DeleteAttachment(fileID, workspaceID string) error
	GetUserAttachment(fileID string) (*model.WorkspaceAttachment, error)
	GetAttachmentByID(fileID string) (*model.WorkspaceAttachment, error)
	CreateSavedFilter(workspaceID, tableID, viewID, filterID, userID, id string) (string, error)
	GetTableFilters(workspaceID, tableID, userID string) ([]model.SavedFilterWithPayload, error)
	DeactivateFiltersForView(workspaceID, tableID, viewID string) error
	UpdateFilter(filterID, workspaceID, tableID string, filters model.FilterPayload, name string, isPrivate bool, userID string) error
	UpdateFilterActiveStatus(workspaceID, tableID, viewID, savedFilterID string, isActive bool) (bool, error)
	UpdateSavedFilter(savedFilterID, workspaceID, tableID, viewID string, isActive bool) (bool, error)
	GetSavedFilterOwner(savedFilterID, workspaceID string) (bool, string, error)
	DeleteSavedFilter(savedFilterID, workspaceID string) error
	GetTablePermissionsForRole(roleID, workspaceID string) ([]model.TablePermission, error)
	UpdateTablePermissionsForRole(roleID, workspaceID string, perms []model.TablePermission) error
	GetTablePermissionsForRoles(roleIDs []string, workspaceID string) ([]model.TablePermission, error)
	AddGroups(ctx context.Context, workspaceID string, groupIDs []string, roles []string, addedBy string) error
	GetGroups(ctx context.Context, workspaceID string) ([]model.WorkspaceGroup, error)
	RemoveGroup(ctx context.Context, workspaceID, groupID string) error
	UpdateGroupRoles(ctx context.Context, workspaceID, groupID string, roles []string) error
	SearchMembers(ctx context.Context, workspaceIDs []string, query string, limit, offset int) ([]model.User, error)
	GetGroupRolesForUser(userID, workspaceID string) ([]string, error)
}

type UploadSessionStore interface {
	Create(session model.UploadSession) (*model.UploadSession, error)
	Get(id string) (*model.UploadSession, error)
	UpdateParts(
		id string,
		partsJSON string,
		uploadedSize int64,
		uploadedParts int,
		now int64,
	) error
	UpdateStatus(id string, status model.UploadSessionStatus, now int64) error

	Delete(id string) error
}

type JobsStore interface {
	Create(job model.Job) (*model.Job, error)
	ClaimNextPending() (*model.Job, error)
	UpdateStatus(id string, status string, result string) error
	IsCancelled(jobID string) (bool, error)
	UpdateProgress(id string, progress int) error
	GetRunningForUser(userID string) ([]model.Job, error)
	GetForUserByID(userID, id string) (*model.Job, error)
	Get(id string) (*model.Job, error)
	GetLatestByType(jobType string) (*model.Job, error)
	CountActiveByType(jobType string) (int, error)
	UpdatePayload(id string, payload json.RawMessage) error
	DeleteOld(finishedBefore int64) error
	AcknowledgeForUser(userID, jobID string) error
}

type OIDCProviderStore interface {
	Create(p *model.OIDCProvider) (*model.OIDCProvider, error)
	GetAll() ([]model.OIDCProvider, error)
	GetAllEnabled() ([]model.OIDCProvider, error)
	GetByID(id string) (*model.OIDCProvider, error)
	Update(p *model.OIDCProvider) (*model.OIDCProvider, error)
	UpdateSecret(id, encryptedSecret string) error
	Delete(id string) error
	Count() (int, error)
}

type GroupStore interface {
	Create(ctx context.Context, ownerID, name, description string, roles []string) (*model.Group, error)
	Get(ctx context.Context, id string) (*model.Group, error)
	GetByIDs(ctx context.Context, ids []string) ([]model.Group, error)
	GetAllPaged(ctx context.Context, query string, sort model.Sort, limit, offset int) ([]model.Group, error)
	CountAll(ctx context.Context, query string) (int, error)
	SearchForUser(ctx context.Context, userID, query string, limit int) ([]model.Group, error)
	Update(ctx context.Context, id, name, description string, roles []string) error
	SoftDelete(ctx context.Context, id string) error
	AddMembers(ctx context.Context, groupID string, userIDs []string, role string) error
	RemoveMember(ctx context.Context, groupID, userID string) error
	RemoveAllMembershipsForUser(ctx context.Context, userID string) error
	GetMembers(ctx context.Context, groupID string) ([]model.GroupMember, error)
	GetMembersPaged(ctx context.Context, groupID, query string, sort model.Sort, limit, offset int) ([]model.GroupMember, error)
	CountMembers(ctx context.Context, groupID, query string) (int, error)
	GetSharesForFile(ctx context.Context, fileID string) ([]model.SharedGroup, error)
	GetIDsForUser(ctx context.Context, userID string) ([]string, error)
	GetRolesForUser(ctx context.Context, userID string) ([]string, error)
}
