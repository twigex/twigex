CREATE TABLE `channels` (
  `id` varchar(36) NOT NULL,
  `type` varchar(1) NOT NULL,
  `display_name` varchar(255) NOT NULL,
  `name` varchar(255) NOT NULL,
  `header` text DEFAULT NULL,
  `description` varchar(255) DEFAULT NULL,
  `last_post` bigint DEFAULT NULL,
  `msg_count` bigint NOT NULL,
  `created_by` varchar(36) NOT NULL,
  `created_at` bigint NOT NULL,
  `updated_at` bigint NOT NULL,
  `deleted_at` bigint NOT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_channels_name` (`name`),
  KEY `idx_channels_type_deleted_at` (`type`,`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `workspace_roles` (
  `id` varchar(36) NOT NULL,
  `workspace_id` varchar(36) NOT NULL,
  `name` varchar(255) DEFAULT NULL,
  `displayname` varchar(255) DEFAULT NULL,
  `description` text DEFAULT NULL,
  `permissions` text DEFAULT NULL,
  `created_at` bigint DEFAULT NULL,
  `updated_at` bigint DEFAULT NULL,
  `per_table_mode` tinyint(1) NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  KEY `idx_workspace_roles_workspace_name` (`workspace_id`,`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `workspace_tables` (
  `id` varchar(36) NOT NULL,
  `workspace_id` varchar(36) NOT NULL,
  `name` varchar(255) NOT NULL,
  `display_name` varchar(255) DEFAULT NULL,
  `linked` tinyint(1) NOT NULL,
  `single_select` tinyint(1) NOT NULL,
  `parent_table_id` varchar(36) DEFAULT NULL,
  `both_direction_link` tinyint(1) NOT NULL,
  `second_table_id` varchar(36) DEFAULT NULL,
  `folder_id` varchar(36) DEFAULT NULL,
  `deleted_at` bigint DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_workspace_tables_parent_table_id` (`parent_table_id`),
  KEY `idx_workspace_tables_workspace_name` (`workspace_id`,`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `channel_meetings` (
  `id` varchar(36) NOT NULL,
  `channel_id` varchar(36) NOT NULL,
  `host_id` varchar(36) NOT NULL,
  `created_by` varchar(36) NOT NULL,
  `title` varchar(255) NOT NULL,
  `guest_password` varchar(255) DEFAULT NULL,
  `duration_minutes` int NOT NULL DEFAULT 60,
  `sequence` int NOT NULL DEFAULT 0,
  `timezone` varchar(64) NOT NULL DEFAULT '',
  `invite_only` tinyint(1) NOT NULL DEFAULT 0,
  `scheduled_at` bigint DEFAULT NULL,
  `started_at` bigint DEFAULT NULL,
  `ended_at` bigint DEFAULT NULL,
  `created_at` bigint NOT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_cm_channel_id` (`channel_id`),
  KEY `idx_cm_ended_at` (`ended_at`),
  CONSTRAINT `fk_cm_channel_id` FOREIGN KEY (`channel_id`) REFERENCES `channels` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `activity` (
  `id` varchar(36) NOT NULL,
  `app` varchar(255) NOT NULL,
  `type` varchar(255) NOT NULL,
  `user_id` varchar(36) NOT NULL,
  `affected_user` varchar(36) NOT NULL,
  `item_id` varchar(36) NOT NULL,
  `parent_id` varchar(36) NOT NULL,
  `parameters` longtext DEFAULT NULL,
  `created_at` bigint NOT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_activity_parent_created` (`parent_id`,`created_at`),
  KEY `idx_activity_item_created` (`item_id`,`created_at`),
  KEY `idx_activity_type_created` (`type`,`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `app_migrations` (
  `id` varchar(36) NOT NULL,
  `name` varchar(255) DEFAULT NULL,
  `migrated_at` bigint DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uq_app_migrations_name` (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `channel_groups` (
  `id` varchar(36) NOT NULL,
  `channel_id` varchar(36) NOT NULL,
  `group_id` varchar(36) NOT NULL,
  `added_by` varchar(36) NOT NULL,
  `added_at` bigint NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `unique_channel_group` (`channel_id`,`group_id`),
  KEY `idx_channel_groups_group_id` (`group_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `channel_guest_links` (
  `id` varchar(36) NOT NULL,
  `channel_id` varchar(36) NOT NULL,
  `meeting_id` varchar(36) NOT NULL,
  `created_by` varchar(36) NOT NULL,
  `invited_email` varchar(255) DEFAULT NULL,
  `expires_at` bigint NOT NULL,
  `used_at` bigint DEFAULT NULL,
  `revoked_at` bigint DEFAULT NULL,
  `created_at` bigint NOT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_cgl_channel_id` (`channel_id`),
  KEY `idx_cgl_meeting_id` (`meeting_id`),
  CONSTRAINT `fk_cgl_channel_id` FOREIGN KEY (`channel_id`) REFERENCES `channels` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_cgl_meeting_id` FOREIGN KEY (`meeting_id`) REFERENCES `channel_meetings` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `channel_meeting_groups` (
  `id` varchar(36) NOT NULL,
  `meeting_id` varchar(36) NOT NULL,
  `group_id` varchar(36) NOT NULL,
  `added_by` varchar(36) NOT NULL,
  `added_at` bigint NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `unique_meeting_group` (`meeting_id`,`group_id`),
  KEY `idx_channel_meeting_groups_group_id` (`group_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `channel_meeting_invitees` (
  `meeting_id` varchar(36) NOT NULL,
  `user_id` varchar(36) NOT NULL,
  PRIMARY KEY (`meeting_id`,`user_id`),
  KEY `idx_cmi_user_id` (`user_id`),
  CONSTRAINT `fk_cmi_meeting_id` FOREIGN KEY (`meeting_id`) REFERENCES `channel_meetings` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `channel_members` (
  `id` varchar(36) NOT NULL,
  `channel_id` varchar(36) NOT NULL,
  `user_id` varchar(36) NOT NULL,
  `role` text DEFAULT NULL,
  `notify_props` json NOT NULL,
  `msg_count` bigint NOT NULL,
  `mention_count` bigint NOT NULL,
  `last_viewed_at` bigint NOT NULL,
  `updated_at` bigint NOT NULL,
  `date_joined` bigint NOT NULL,
  `is_direct` tinyint(1) NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  UNIQUE KEY `unique_channel_member` (`channel_id`,`user_id`),
  KEY `idx_channel_members_user_channel` (`user_id`,`channel_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `collimato_charts` (
  `id` varchar(36) NOT NULL,
  `name` varchar(255) NOT NULL,
  `chart_type` varchar(255) NOT NULL,
  `configuration` json NOT NULL,
  `data` json NOT NULL,
  `owner_id` varchar(36) NOT NULL,
  `workspace_id` varchar(36) NOT NULL,
  `created_at` bigint NOT NULL,
  `updated_at` bigint DEFAULT NULL,
  `deleted_at` bigint DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_charts_workspace_id` (`workspace_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `collimato_dashboard_charts` (
  `id` varchar(36) NOT NULL,
  `dashboard_id` varchar(36) NOT NULL,
  `chart_id` varchar(36) NOT NULL,
  `created_by` varchar(36) NOT NULL,
  `position` json NOT NULL,
  `created_at` bigint NOT NULL,
  `updated_at` bigint DEFAULT NULL,
  `deleted_at` bigint DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uq_dashboards_charts_dashboard_id_chart_id` (`dashboard_id`,`chart_id`),
  KEY `idx_dashboards_charts_chart_id` (`chart_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `collimato_dashboard_filters` (
  `id` varchar(36) NOT NULL,
  `dashboard_id` varchar(36) NOT NULL,
  `name` varchar(255) DEFAULT NULL,
  `table_name` varchar(255) DEFAULT NULL,
  `column_name` varchar(255) DEFAULT NULL,
  `filter_values` text DEFAULT NULL,
  `apply_to` text DEFAULT NULL,
  `created_at` bigint DEFAULT NULL,
  `updated_at` bigint DEFAULT NULL,
  `deleted_at` bigint DEFAULT NULL,
  `operator` varchar(50) NOT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_dashboard_filters_dashboard_id` (`dashboard_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `collimato_dashboards` (
  `id` varchar(36) NOT NULL,
  `title` varchar(255) NOT NULL,
  `description` text DEFAULT NULL,
  `owner_id` varchar(36) NOT NULL,
  `workspace_id` varchar(36) NOT NULL,
  `created_at` bigint NOT NULL,
  `updated_at` bigint DEFAULT NULL,
  `deleted_at` bigint DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_dashboards_workspace_id` (`workspace_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `collimato_workspace_connections` (
  `id` varchar(36) NOT NULL,
  `workspace_id` varchar(36) NOT NULL,
  `display_name` varchar(255) NOT NULL,
  `type` varchar(255) NOT NULL,
  `host` varchar(255) DEFAULT NULL,
  `port` int DEFAULT NULL,
  `database_name` varchar(255) DEFAULT NULL,
  `username` varchar(255) DEFAULT NULL,
  `password` varchar(255) DEFAULT NULL,
  `is_default` tinyint(1) NOT NULL,
  `created_by` varchar(36) NOT NULL,
  `created_at` bigint NOT NULL,
  `updated_at` bigint DEFAULT NULL,
  `deleted_at` bigint DEFAULT NULL,
  `ssl_mode` varchar(16) NOT NULL DEFAULT 'disable',
  `config` json DEFAULT NULL,
  `encrypted_config` text DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_collimato_workspace_connections_workspace_id` (`workspace_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `collimato_workspace_files` (
  `id` varchar(36) NOT NULL,
  `workspace_id` varchar(36) NOT NULL,
  `name` varchar(255) NOT NULL,
  `file_type` varchar(50) NOT NULL,
  `content` longtext NOT NULL,
  `encrypted` tinyint(1) NOT NULL DEFAULT 0,
  `created_at` bigint NOT NULL,
  `updated_at` bigint NOT NULL,
  `builder_model` longtext DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `workspace_file` (`workspace_id`,`file_type`,`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `collimato_workspace_groups` (
  `id` varchar(36) NOT NULL,
  `workspace_id` varchar(36) NOT NULL,
  `group_id` varchar(36) NOT NULL,
  `roles` text NOT NULL,
  `added_by` varchar(36) NOT NULL,
  `added_at` bigint NOT NULL,
  `updated_at` bigint DEFAULT NULL,
  `deleted_at` bigint NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  UNIQUE KEY `unique_workspace_group` (`workspace_id`,`group_id`),
  KEY `idx_collimato_workspace_groups_group_id` (`group_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `collimato_workspace_roles` (
  `id` varchar(36) NOT NULL,
  `workspace_id` varchar(36) NOT NULL,
  `name` varchar(255) DEFAULT NULL,
  `displayname` varchar(255) DEFAULT NULL,
  `description` text DEFAULT NULL,
  `permissions` text DEFAULT NULL,
  `table_level_permissions` text DEFAULT NULL,
  `column_level_permissions` text DEFAULT NULL,
  `row_level_permissions` text DEFAULT NULL,
  `auto_update` tinyint(1) DEFAULT 0,
  `created_at` bigint DEFAULT NULL,
  `updated_at` bigint DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_collimato_workspace_roles_workspace_name` (`workspace_id`,`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `collimato_workspace_users` (
  `id` varchar(36) NOT NULL,
  `workspace_id` varchar(36) NOT NULL,
  `user_id` varchar(36) NOT NULL,
  `role` text NOT NULL,
  `created_at` bigint NOT NULL,
  `updated_at` bigint DEFAULT NULL,
  `deleted_at` bigint DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `unique_workspace_user` (`workspace_id`,`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `collimato_workspaces` (
  `id` varchar(36) NOT NULL,
  `name` varchar(255) NOT NULL,
  `description` text DEFAULT NULL,
  `status` varchar(30) NOT NULL,
  `server_id` varchar(36) NOT NULL,
  `created_by` varchar(36) NOT NULL,
  `created_at` bigint NOT NULL,
  `updated_at` bigint DEFAULT NULL,
  `deleted_at` bigint DEFAULT NULL,
  `secret` text NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `external_share` (
  `id` varchar(36) NOT NULL,
  `owner` varchar(36) NOT NULL,
  `file_id` varchar(36) NOT NULL,
  `password_protected` tinyint(1) DEFAULT NULL,
  `password` varchar(255) NOT NULL,
  `source` varchar(255) DEFAULT NULL,
  `parent` varchar(255) DEFAULT NULL,
  `share_token` varchar(255) NOT NULL,
  `share_time` bigint DEFAULT NULL,
  `message` longtext DEFAULT NULL,
  `expiration` bigint NOT NULL DEFAULT 0,
  `allow_download` tinyint(1) DEFAULT NULL,
  `allow_upload` tinyint(1) DEFAULT NULL,
  `downloaded` int DEFAULT NULL,
  `accessed` int DEFAULT NULL,
  `active` tinyint(1) NOT NULL,
  `created_at` bigint NOT NULL,
  `updated_at` bigint NOT NULL,
  `allow_view` tinyint(1) NOT NULL DEFAULT 1,
  `last_accessed_at` bigint NOT NULL DEFAULT 0,
  `max_downloads` int NOT NULL DEFAULT 0,
  `allow_edit` tinyint(1) NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  UNIQUE KEY `share_token` (`share_token`),
  KEY `idx_external_share_file_active` (`file_id`,`active`),
  KEY `idx_external_share_owner_source` (`owner`,`source`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `favorites` (
  `id` varchar(36) NOT NULL,
  `user_id` varchar(36) NOT NULL,
  `app` varchar(255) NOT NULL,
  `item_id` varchar(36) NOT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_favorite_item_id` (`item_id`),
  UNIQUE KEY `uq_favorites_user_app_item` (`user_id`,`app`,`item_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `file_metadata` (
  `id` varchar(36) NOT NULL,
  `type` text DEFAULT NULL,
  `title` text DEFAULT NULL,
  `fields` text DEFAULT NULL,
  `created_at` bigint DEFAULT NULL,
  `updated_at` bigint DEFAULT NULL,
  `created_by` varchar(36) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `file_metadata_entries` (
  `id` varchar(36) NOT NULL,
  `file_id` varchar(36) NOT NULL,
  `metadata_id` varchar(36) DEFAULT NULL,
  `value` text DEFAULT NULL,
  `created_at` bigint DEFAULT NULL,
  `updated_at` bigint DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_file_metadata_entries_file_id` (`file_id`),
  KEY `idx_file_metadata_entries_metadata_id` (`metadata_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `file_share` (
  `id` varchar(36) NOT NULL,
  `file_id` varchar(36) NOT NULL,
  `initiator` varchar(255) NOT NULL,
  `share_type` int NOT NULL,
  `share_with` varchar(36) NOT NULL,
  `expiration` bigint NOT NULL DEFAULT 0,
  `time_shared` bigint NOT NULL,
  `created_at` bigint NOT NULL,
  `updated_at` bigint NOT NULL,
  `access_level` int NOT NULL DEFAULT 10,
  PRIMARY KEY (`id`),
  UNIQUE KEY `unique_share` (`file_id`,`share_type`,`share_with`),
  KEY `idx_file_share_share_with_type` (`share_with`,`share_type`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `file_storage` (
  `id` varchar(36) NOT NULL,
  `label` varchar(255) NOT NULL,
  `description` varchar(255) DEFAULT NULL,
  `type` int(1) NOT NULL,
  `directory` varchar(500) DEFAULT NULL,
  `endpoint` varchar(500) DEFAULT NULL,
  `access_key` varchar(500) DEFAULT NULL,
  `secret_key` text DEFAULT NULL,
  `bucket` varchar(255) DEFAULT NULL,
  `use_ssl` tinyint(1) NOT NULL DEFAULT 1,
  `created_at` bigint DEFAULT NULL,
  `is_primary` tinyint(1) NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `file_trash` (
  `id` varchar(36) NOT NULL,
  `file_id` varchar(36) NOT NULL,
  `owner` varchar(36) NOT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_file_trash_file_id` (`file_id`),
  KEY `idx_file_trash_owner` (`owner`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `files` (
  `id` varchar(36) NOT NULL,
  `owner` varchar(36) NOT NULL,
  `parent` varchar(36) NOT NULL,
  `storage` varchar(255) NOT NULL,
  `name` varchar(255) NOT NULL,
  `displayname` varchar(255) NOT NULL,
  `size` bigint NOT NULL,
  `type` varchar(255) NOT NULL,
  `is_folder` tinyint(1) NOT NULL,
  `created_at` bigint NOT NULL,
  `modified_at` bigint NOT NULL,
  `deleted_at` bigint NOT NULL DEFAULT 0,
  `lock_owner` varchar(64) DEFAULT NULL,
  `lock_expires_at` bigint DEFAULT NULL,
  `version` bigint NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  KEY `idx_file_storage` (`storage`),
  KEY `idx_files_lock_owner` (`lock_owner`),
  KEY `idx_file_parent_deleted` (`parent`,`deleted_at`),
  KEY `idx_file_owner_type` (`owner`,`type`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `group_members` (
  `id` varchar(36) NOT NULL,
  `group_id` varchar(36) NOT NULL,
  `user_id` varchar(36) NOT NULL,
  `role` varchar(32) NOT NULL DEFAULT 'member',
  `joined_at` bigint NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `unique_group_member` (`group_id`,`user_id`),
  KEY `idx_group_members_user_id` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `jobs` (
  `id` varchar(36) NOT NULL,
  `type` varchar(50) NOT NULL,
  `status` varchar(20) NOT NULL,
  `user_id` varchar(36) DEFAULT NULL,
  `payload` json NOT NULL,
  `progress` int NOT NULL DEFAULT 0,
  `error` text DEFAULT NULL,
  `created_at` bigint NOT NULL,
  `started_at` bigint NOT NULL DEFAULT 0,
  `finished_at` bigint NOT NULL DEFAULT 0,
  `updated_at` bigint NOT NULL DEFAULT 0,
  `acknowledged_at` bigint DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_jobs_user` (`user_id`),
  KEY `idx_jobs_status_created` (`status`,`created_at`),
  KEY `idx_jobs_type_created` (`type`,`created_at`),
  KEY `idx_jobs_type_status` (`type`,`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `licenses` (
  `id` varchar(36) NOT NULL,
  `created` bigint NOT NULL,
  `active` tinyint(1) NOT NULL,
  `bytes` text NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `linkmetadata` (
  `hash` bigint NOT NULL,
  `url` text NOT NULL,
  `type` varchar(50) NOT NULL,
  `data` json NOT NULL,
  `created_at` bigint NOT NULL,
  `updated_at` bigint NOT NULL,
  `storage_id` varchar(36) DEFAULT NULL,
  `size` bigint NOT NULL DEFAULT 0,
  PRIMARY KEY (`hash`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `notifications` (
  `id` varchar(36) NOT NULL,
  `sender` varchar(36) NOT NULL,
  `receiver` varchar(36) NOT NULL,
  `app` varchar(255) NOT NULL,
  `notification_type` varchar(255) NOT NULL,
  `item_id` varchar(36) NOT NULL,
  `details` json DEFAULT NULL,
  `read_at` bigint NOT NULL DEFAULT 0,
  `sent_at` bigint NOT NULL,
  `created_at` bigint NOT NULL,
  `updated_at` bigint NOT NULL,
  `deleted_at` bigint DEFAULT 0,
  PRIMARY KEY (`id`),
  KEY `idx_notifications_receiver_created` (`receiver`,`created_at`),
  KEY `idx_notifications_receiver_unread` (`receiver`,`read_at`,`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `oidc_providers` (
  `id` varchar(36) NOT NULL,
  `name` varchar(255) NOT NULL,
  `enabled` tinyint(1) NOT NULL DEFAULT 1,
  `discovery_url` varchar(500) NOT NULL,
  `client_id` varchar(500) NOT NULL,
  `client_secret` text NOT NULL,
  `scopes` varchar(500) NOT NULL DEFAULT 'openid,profile,email',
  `button_text` varchar(255) DEFAULT NULL,
  `button_color` varchar(50) DEFAULT NULL,
  `created_at` bigint NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `password_resets` (
  `id` varchar(36) NOT NULL,
  `user_id` varchar(36) NOT NULL,
  `token` varchar(255) NOT NULL,
  `created_at` bigint NOT NULL,
  `expires_at` bigint NOT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_password_resets_token` (`token`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `post_attachments` (
  `id` varchar(36) NOT NULL,
  `channel_id` varchar(36) NOT NULL,
  `post_id` varchar(36) NOT NULL,
  `user_id` varchar(36) NOT NULL,
  `name` varchar(255) NOT NULL,
  `size` bigint NOT NULL,
  `mime_type` varchar(255) NOT NULL,
  `width` int NOT NULL,
  `height` int NOT NULL,
  `created_at` bigint NOT NULL,
  `updated_at` bigint NOT NULL,
  `deleted_at` bigint DEFAULT 0,
  `url` text NOT NULL,
  `provider` varchar(50) NOT NULL DEFAULT '',
  `kind` varchar(50) NOT NULL DEFAULT 'file',
  `storage_id` varchar(36) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_post_attachments_post_id` (`post_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `post_reactions` (
  `id` varchar(36) NOT NULL,
  `channel_id` varchar(36) NOT NULL,
  `post_id` varchar(36) NOT NULL,
  `user_id` varchar(36) NOT NULL,
  `reaction` varchar(64) NOT NULL,
  `created_at` bigint NOT NULL,
  `updated_at` bigint NOT NULL,
  `deleted_at` bigint DEFAULT 0,
  PRIMARY KEY (`id`),
  UNIQUE KEY `unique_post_user_reaction` (`post_id`,`user_id`,`reaction`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `posts` (
  `id` varchar(36) NOT NULL,
  `user_id` varchar(36) NOT NULL,
  `channel_id` varchar(36) NOT NULL,
  `reply_id` varchar(36) DEFAULT NULL,
  `message` text NOT NULL,
  `type` varchar(36) DEFAULT NULL,
  `created_at` bigint NOT NULL,
  `updated_at` bigint NOT NULL,
  `deleted_at` bigint NOT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_posts_channel_created` (`channel_id`,`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `preferences` (
  `user_id` varchar(36) NOT NULL,
  `category` varchar(255) NOT NULL,
  `name` varchar(255) NOT NULL,
  `value` text NOT NULL,
  PRIMARY KEY (`user_id`,`category`,`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `roles` (
  `id` varchar(36) NOT NULL,
  `name` varchar(255) DEFAULT NULL,
  `displayname` varchar(128) DEFAULT NULL,
  `description` text DEFAULT NULL,
  `permissions` text DEFAULT NULL,
  `created_at` bigint DEFAULT NULL,
  `updated_at` bigint DEFAULT NULL,
  `built_in` tinyint(1) NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  KEY `idx_role_name` (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `status` (
  `user_id` varchar(36) NOT NULL,
  `status` varchar(255) NOT NULL,
  `last_activity` bigint NOT NULL,
  `user_defined` tinyint(1) NOT NULL DEFAULT 0,
  PRIMARY KEY (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `system_settings` (
  `name` varchar(255) NOT NULL,
  `value` text NOT NULL,
  PRIMARY KEY (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `upload_sessions` (
  `id` varchar(36) NOT NULL,
  `upload_id` varchar(255) NOT NULL,
  `user_id` varchar(36) NOT NULL,
  `context_type` varchar(32) NOT NULL,
  `context_id` varchar(36) DEFAULT NULL,
  `storage` varchar(32) NOT NULL,
  `file_name` varchar(255) NOT NULL,
  `mime_type` varchar(128) NOT NULL,
  `total_size` bigint NOT NULL,
  `uploaded_size` bigint NOT NULL DEFAULT 0,
  `uploaded_parts` int NOT NULL DEFAULT 0,
  `parts_json` text DEFAULT NULL,
  `backend_metadata` text DEFAULT NULL,
  `status` enum('pending','uploading','completed','failed','aborted') NOT NULL DEFAULT 'pending',
  `error_reason` varchar(255) DEFAULT NULL,
  `created_at` bigint NOT NULL,
  `updated_at` bigint NOT NULL,
  `expires_at` bigint NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `user_groups` (
  `id` varchar(36) NOT NULL,
  `name` varchar(255) NOT NULL,
  `description` text DEFAULT NULL,
  `owner_id` varchar(36) NOT NULL,
  `created_at` bigint NOT NULL,
  `updated_at` bigint DEFAULT NULL,
  `deleted_at` bigint NOT NULL DEFAULT 0,
  `roles` text DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `users` (
  `id` varchar(36) NOT NULL,
  `email` varchar(180) NOT NULL,
  `role` varchar(255) NOT NULL,
  `password` varchar(255) NOT NULL,
  `auth_service` varchar(32) DEFAULT NULL,
  `username` varchar(255) NOT NULL,
  `name` varchar(255) NOT NULL,
  `lastname` varchar(255) NOT NULL,
  `timezone` longtext DEFAULT NULL,
  `mfa_active` tinyint(1) DEFAULT NULL,
  `mfa_secret` varchar(128) DEFAULT NULL,
  `created_at` bigint NOT NULL,
  `updated_at` bigint NOT NULL,
  `deactivated_at` bigint NOT NULL,
  `auth_data` varchar(128) NOT NULL,
  `storage_limit` bigint NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  UNIQUE KEY `email` (`email`),
  UNIQUE KEY `username` (`username`),
  KEY `idx_users_directory` (`deactivated_at`,`name`,`lastname`),
  KEY `idx_users_auth` (`auth_service`,`auth_data`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `user_photos` (
  `user_id` varchar(36) NOT NULL,
  `photo_id` varchar(255) NOT NULL,
  `storage_id` varchar(36) DEFAULT NULL,
  PRIMARY KEY (`user_id`),
  KEY `fk_user_photos_storage` (`storage_id`),
  CONSTRAINT `fk_user_photos_storage` FOREIGN KEY (`storage_id`) REFERENCES `file_storage` (`id`),
  CONSTRAINT `fk_user_photos_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `workspace_attachments` (
  `id` varchar(36) NOT NULL,
  `workspace_id` varchar(36) NOT NULL,
  `table_id` varchar(36) NOT NULL,
  `field_id` varchar(36) NOT NULL,
  `task_id` varchar(36) NOT NULL,
  `user_id` varchar(36) NOT NULL,
  `name` varchar(255) NOT NULL,
  `size` bigint NOT NULL,
  `mime_type` varchar(255) NOT NULL,
  `width` int NOT NULL,
  `height` int NOT NULL,
  `created_at` bigint NOT NULL,
  `updated_at` bigint NOT NULL,
  `deleted_at` bigint DEFAULT 0,
  `storage_id` varchar(36) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_workspace_attachments_task_id` (`task_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `workspace_comments` (
  `id` varchar(36) NOT NULL,
  `user_id` varchar(36) NOT NULL,
  `workspace_id` varchar(36) NOT NULL,
  `table_id` varchar(36) NOT NULL,
  `item_id` varchar(36) NOT NULL,
  `content` text NOT NULL,
  `created_at` bigint NOT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_workspace_comments_item_created` (`item_id`,`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `workspace_fields` (
  `id` varchar(36) NOT NULL,
  `workspace_id` varchar(36) NOT NULL,
  `table_id` varchar(36) NOT NULL,
  `field_name` varchar(255) NOT NULL,
  `field_display_name` varchar(255) DEFAULT NULL,
  `field_type` varchar(255) NOT NULL,
  `parent_field_id` varchar(36) DEFAULT NULL,
  `created_at` bigint NOT NULL,
  `updated_at` bigint DEFAULT NULL,
  `deleted_at` bigint DEFAULT NULL,
  `formula` json DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_workspace_fields_table_field` (`table_id`,`field_name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `workspace_filters` (
  `id` varchar(36) NOT NULL,
  `workspace_id` varchar(36) NOT NULL,
  `table_id` varchar(36) NOT NULL,
  `filter_settings` json DEFAULT NULL,
  `created_at` bigint NOT NULL,
  `updated_at` bigint NOT NULL,
  `name` varchar(150) NOT NULL,
  `is_private` tinyint(1) NOT NULL DEFAULT 0,
  `created_by` varchar(36) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_filters_table_id` (`table_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `workspace_folders` (
  `id` varchar(36) NOT NULL,
  `workspace_id` varchar(36) NOT NULL,
  `parent_folder_id` varchar(36) DEFAULT NULL,
  `name` varchar(255) NOT NULL,
  `description` text DEFAULT NULL,
  `created_at` bigint NOT NULL,
  `updated_at` bigint DEFAULT NULL,
  `deleted_at` bigint DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_folders_workspace_id` (`workspace_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `workspace_groups` (
  `id` varchar(36) NOT NULL,
  `workspace_id` varchar(36) NOT NULL,
  `group_id` varchar(36) NOT NULL,
  `roles` varchar(255) NOT NULL DEFAULT 'user',
  `added_by` varchar(36) NOT NULL,
  `added_at` bigint NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `unique_workspace_group` (`workspace_id`,`group_id`),
  KEY `idx_workspace_groups_group_id` (`group_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `workspace_members` (
  `id` varchar(36) NOT NULL,
  `user_id` varchar(36) NOT NULL,
  `workspace_id` varchar(36) NOT NULL,
  `role` varchar(255) NOT NULL,
  `date_joined` bigint NOT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_workspace_members` (`user_id`),
  UNIQUE KEY `uq_workspace_members_workspace_user` (`workspace_id`,`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `workspace_relationships` (
  `id` varchar(36) NOT NULL,
  `workspace_id` varchar(36) NOT NULL,
  `table_id` varchar(36) NOT NULL,
  `linked_table_id` varchar(36) NOT NULL,
  `table_name` varchar(255) NOT NULL,
  `created_at` bigint NOT NULL,
  `updated_at` bigint DEFAULT NULL,
  `deleted_at` bigint DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_workspace_relationships_linked_table_id` (`linked_table_id`),
  KEY `idx_workspace_relationships_table_id_name` (`table_id`,`table_name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `workspace_saved_filters` (
  `id` varchar(36) NOT NULL,
  `workspace_id` varchar(36) NOT NULL,
  `table_id` varchar(36) NOT NULL,
  `view_id` varchar(36) DEFAULT NULL,
  `filter_id` varchar(36) NOT NULL,
  `user_id` varchar(36) NOT NULL,
  `is_active` tinyint(1) NOT NULL DEFAULT 0,
  `created_at` bigint NOT NULL,
  `updated_at` bigint NOT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_wsf_table_view` (`table_id`,`view_id`),
  KEY `idx_wsf_filter` (`filter_id`),
  KEY `idx_wsf_view_user_active` (`view_id`,`user_id`,`is_active`,`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `workspace_sort` (
  `id` varchar(36) NOT NULL,
  `workspace_id` varchar(36) NOT NULL,
  `table_id` varchar(36) NOT NULL,
  `view_id` varchar(36) NOT NULL,
  `sort_settings` json DEFAULT NULL,
  `created_at` bigint NOT NULL,
  `updated_at` bigint NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uq_sort_view` (`view_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `workspace_table_permissions` (
  `id` varchar(36) NOT NULL,
  `role_id` varchar(36) NOT NULL,
  `workspace_id` varchar(36) NOT NULL,
  `table_id` varchar(36) NOT NULL,
  `action` varchar(64) NOT NULL,
  `created_at` bigint NOT NULL,
  `updated_at` bigint NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uq_role_table_action` (`role_id`,`table_id`,`action`),
  KEY `idx_wtp_table_id` (`table_id`),
  CONSTRAINT `fk_wtp_role` FOREIGN KEY (`role_id`) REFERENCES `workspace_roles` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_wtp_table` FOREIGN KEY (`table_id`) REFERENCES `workspace_tables` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `workspace_view` (
  `id` varchar(36) NOT NULL,
  `workspace_id` varchar(36) NOT NULL,
  `table_id` varchar(36) NOT NULL,
  `view_type` varchar(255) NOT NULL,
  `parent_table_id` varchar(36) DEFAULT NULL,
  `name` varchar(255) NOT NULL,
  `item_order` longtext DEFAULT NULL,
  `created_by` varchar(36) NOT NULL,
  `main_view` tinyint(1) DEFAULT 0,
  `created_at` bigint NOT NULL,
  `updated_at` bigint DEFAULT NULL,
  `deleted_at` bigint DEFAULT NULL,
  `is_public` tinyint(1) NOT NULL DEFAULT 1,
  PRIMARY KEY (`id`),
  KEY `idx_workspace_view_parent_table_id` (`parent_table_id`),
  KEY `idx_workspace_view_table_type` (`table_id`,`view_type`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `workspace_view_shares` (
  `id` varchar(36) NOT NULL,
  `view_id` varchar(36) NOT NULL,
  `workspace_id` varchar(36) NOT NULL,
  `user_id` varchar(36) NOT NULL,
  `created_at` bigint NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uq_view_user` (`view_id`,`user_id`),
  KEY `idx_view_shares_workspace_user` (`workspace_id`,`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `workspaces` (
  `id` varchar(36) NOT NULL,
  `title` varchar(255) NOT NULL,
  `description` text DEFAULT NULL,
  `start_date` bigint DEFAULT NULL,
  `end_date` bigint DEFAULT NULL,
  `pre_fix` varchar(255) NOT NULL,
  `created_at` bigint NOT NULL,
  `updated_at` bigint DEFAULT NULL,
  `deleted_at` bigint DEFAULT NULL,
  `created_by` varchar(36) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_workspaces_pre_fix` (`pre_fix`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
