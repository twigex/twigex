CREATE TABLE IF NOT EXISTS `graph_entity_types` (
  `id` varchar(36) NOT NULL,
  `key` varchar(64) NOT NULL,
  `name` varchar(255) NOT NULL,
  `source` varchar(32) NOT NULL,
  `source_id` varchar(36) DEFAULT NULL,
  `key_property` varchar(255) DEFAULT NULL,
  `title_property` varchar(255) DEFAULT NULL,
  `description` text DEFAULT NULL,
  `plural_name` varchar(255) DEFAULT NULL,
  `icon` varchar(32) DEFAULT NULL,
  `color` varchar(16) DEFAULT NULL,
  `status` varchar(16) NOT NULL DEFAULT 'active',
  `built_in` tinyint(1) NOT NULL DEFAULT 0,
  `created_by` varchar(36) DEFAULT NULL,
  `updated_by` varchar(36) DEFAULT NULL,
  `created_at` bigint NOT NULL,
  `updated_at` bigint NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uniq_graph_entity_types_key` (`key`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `graph_relation_types` (
  `id` varchar(36) NOT NULL,
  `key` varchar(64) NOT NULL,
  `name` varchar(255) NOT NULL,
  `reverse_name` varchar(255) NOT NULL,
  `from_type_id` varchar(36) NOT NULL,
  `to_type_id` varchar(36) NOT NULL,
  `kind` varchar(16) NOT NULL,
  `kind_config` json DEFAULT NULL,
  `cardinality` varchar(16) DEFAULT NULL,
  `description` text DEFAULT NULL,
  `built_in` tinyint(1) NOT NULL DEFAULT 0,
  `created_by` varchar(36) DEFAULT NULL,
  `updated_by` varchar(36) DEFAULT NULL,
  `created_at` bigint NOT NULL,
  `updated_at` bigint NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uniq_graph_relation_types_key` (`key`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `graph_relations` (
  `id` varchar(36) NOT NULL,
  `relation_type_id` varchar(36) NOT NULL,
  `from_type_id` varchar(36) NOT NULL,
  `from_id` varchar(255) NOT NULL,
  `to_type_id` varchar(36) NOT NULL,
  `to_id` varchar(255) NOT NULL,
  `origin` varchar(32) NOT NULL,
  `created_by` varchar(36) DEFAULT NULL,
  `created_at` bigint NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uniq_graph_relations_relation` (`relation_type_id`, `from_type_id`, `from_id`, `to_type_id`, `to_id`),
  KEY `idx_graph_relations_from` (`from_type_id`, `from_id`),
  KEY `idx_graph_relations_to` (`to_type_id`, `to_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `graph_datasets` (
  `id` varchar(36) NOT NULL,
  `workspace_id` varchar(36) NOT NULL,
  `name` varchar(255) NOT NULL,
  `kind` varchar(10) NOT NULL,
  `created_at` bigint NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uniq_graph_datasets_workspace_name` (`workspace_id`, `name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `graph_dataset_fields` (
  `id` varchar(36) NOT NULL,
  `workspace_id` varchar(36) NOT NULL,
  `dataset_id` varchar(36) NOT NULL,
  `name` varchar(255) NOT NULL,
  `created_at` bigint NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uniq_graph_dataset_fields_workspace_name` (`workspace_id`, `name`),
  KEY `idx_graph_dataset_fields_dataset` (`dataset_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
