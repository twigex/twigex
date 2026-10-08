// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package model

import "fmt"

const (
	NOTIFICATION_FILE_CREATE                 = "file_create"
	NOTIFICATION_FILE_SHARE                  = "file_share"
	NOTIFICATION_FILE_UPLOAD                 = "file_upload"
	NOTIFICATION_FILE_RENAME                 = "file_rename"
	NOTIFICATION_FILE_DELETE                 = "file_delete"
	NOTIFICATION_FILE_RESTORE                = "file_restore"
	NOTIFICATION_FILE_COMMENT                = "file_comment"
	NOTIFICATION_FILE_DOWNLOAD               = "file_download"
	NOTIFICATION_FILE_PUBLIC_DOWNLOAD        = "file_public_download"
	NOTIFICATION_PROJECT_TASK_CREATED        = "task_created"
	NOTIFICATION_PROJECT_DELETED             = "project_deleted"
	NOTIFICATION_PROJECT_INVITE              = "project_invite"
	NOTIFICATION_PROJECT_TASK_DELETED        = "task_deleted"
	NOTIFICATION_PROJECT_TASK_ASSIGNED       = "task_assigned"
	NOTIFICATION_PROJECT_TASK_STATUS_CHANGED = "task_status_changed"
	NOTIFICATION_TASK_COMMENT                = "task_comment"

	NOTIFICATION_CHANNEL_UPDATE        = "channel_update"
	NOTIFICATION_CHANNEL_POST          = "post"
	NOTIFICATION_CHANNEL_POST_REACTION = "post_reaction"
	NOTIFICATION_CHANNEL_POST_UPDATE   = "post_update"
	NOTIFICATION_CHANNEL_POST_DELETE   = "post_delete"
	NOTIFICATION_CHANNEL_MENTION       = "post_mention"
	NOTIFICATION_CHANNEL_TYPING        = "typing"
	NOTIFICATION_CHANNEL_LINK_PREVIEW  = "link_preview_update"
	NOTIFICATION_CHANNEL_PRESENCE      = "presence"

	NOTIFICATION_PROJECT_CHANGE = "project_change"

	NOTIFICATION_CHANNEL_MEETING_SCHEDULED = "meeting_scheduled"
	NOTIFICATION_CHANNEL_MEETING_UPDATED   = "meeting_updated"
	NOTIFICATION_CHANNEL_MEETING_CANCELLED = "meeting_cancelled"
	NOTIFICATION_CHANNEL_MEETING_STARTED   = "meeting_started"
	NOTIFICATION_CHANNEL_MEETING_ENDED     = "meeting_ended"
	NOTIFICATION_CHANNEL_MEETING_PRESENCE  = "meeting_presence"
	NOTIFICATION_CHANNEL_MEETINGS_REFRESH  = "meetings_refresh"

	NOTIFICATION_CHANNEL_INCOMING_CALL  = "incoming_call"
	NOTIFICATION_CHANNEL_CALL_DECLINED  = "call_declined"
	NOTIFICATION_CHANNEL_CALL_CONNECTED = "call_connected"
	NOTIFICATION_CHANNEL_CALL_CANCELLED = "call_cancelled"
	NOTIFICATION_CHANNEL_CALL_HANDLED   = "call_handled"
)

var DefaultNotifications = map[string]string{
	"notify_email_file_or_folder_created":                 "true",
	"notify_app_file_or_folder_created":                   "true",
	"notify_email_file_or_folder_renamed":                 "true",
	"notify_app_file_or_folder_renamed":                   "true",
	"notify_email_file_or_folder_deleted":                 "true",
	"notify_app_file_or_folder_deleted":                   "true",
	"notify_email_file_or_folder_restored":                "true",
	"notify_app_file_or_folder_restored":                  "true",
	"notify_email_file_or_folder_add_or_remove_favorites": "true",
	"notify_app_file_or_folder_add_or_remove_favorites":   "true",
	"notify_email_file_or_folder_shared":                  "true",
	"notify_app_file_or_folder_shared":                    "true",
	"notify_email_file_or_folder_downloaded":              "true",
	"notify_app_file_or_folder_downloaded":                "true",
	"notify_email_file_or_folder_public_downloaded":       "true",
	"notify_app_file_or_folder_public_downloaded":         "true",
	"notify_email_project_invited":                        "true",
	"notify_app_project_invited":                          "true",
	"notify_email_project_deleted":                        "true",
	"notify_app_project_deleted":                          "true",
	"notify_email_task_deleted":                           "false",
	"notify_app_task_deleted":                             "false",
	"notify_email_task_created":                           "false",
	"notify_app_task_created":                             "false",
	"notify_email_task_assigned":                          "true",
	"notify_app_task_assigned":                            "true",
	"notify_email_task_comment":                           "true",
	"notify_app_task_comment":                             "true",
	"notify_email_assigned_task_status_changed":           "true",
	"notify_app_assigned_task_status_changed":             "true",
	"notify_email_meeting_scheduled":                      "true",
	"notify_app_meeting_scheduled":                        "true",
	"notify_email_meeting_updated":                        "true",
	"notify_app_meeting_updated":                          "true",
	"notify_email_meeting_cancelled":                      "true",
	"notify_app_meeting_cancelled":                        "true",
	"notify_email_channel_mention":                        "true",
	"notify_app_channel_mention":                          "true",
	"send_email_notifications":                            "immediately",
}

type NotificationMessage struct {
	ID               string                 `json:"id"`
	Sender           string                 `json:"sender"`
	Subject          string                 `json:"-"` // for email notification subject
	Receiver         string                 `json:"receiver"`
	App              string                 `json:"app"`
	NotificationType string                 `json:"type"`
	ItemID           string                 `json:"item"`
	Details          map[string]interface{} `json:"details"`
	ReadAt           int64                  `json:"read_at"`
	SentAt           int64                  `json:"-"`
	CreatedAt        int64                  `json:"created_at"`
	UpdatedAt        int64                  `json:"updated_at"`
	DeletedAt        int64                  `json:"deleted_at"`
}

func (n *NotificationMessage) ToMap() map[string]any {
	safeDetails := make(map[string]interface{})
	for k, v := range n.Details {
		switch val := v.(type) {
		case string, int, int64, float64, bool, nil:
			safeDetails[k] = val
		default:
			safeDetails[k] = fmt.Sprintf("%v", val)
		}
	}

	return map[string]any{
		"id":         n.ID,
		"sender":     n.Sender,
		"receiver":   n.Receiver,
		"app":        n.App,
		"type":       n.NotificationType,
		"item":       n.ItemID,
		"details":    safeDetails,
		"read_at":    n.ReadAt,
		"sent_at":    n.SentAt,
		"created_at": n.CreatedAt,
		"updated_at": n.UpdatedAt,
		"deleted_at": n.DeletedAt,
	}
}
