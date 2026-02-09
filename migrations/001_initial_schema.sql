-- Migration: 001_initial_schema
-- Description: Create all tables for the Telegram Bot Gateway

-- Users table
CREATE TABLE IF NOT EXISTS users (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    username VARCHAR(100) NOT NULL UNIQUE,
    email VARCHAR(255) UNIQUE,
    password VARCHAR(255) NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_users_username (username),
    INDEX idx_users_email (email),
    INDEX idx_users_active (is_active)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Roles table
CREATE TABLE IF NOT EXISTS roles (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(50) NOT NULL UNIQUE,
    description VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_roles_name (name)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Permissions table
CREATE TABLE IF NOT EXISTS permissions (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    description VARCHAR(255),
    resource VARCHAR(50) NOT NULL,
    action VARCHAR(50) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_permissions_name (name),
    INDEX idx_permissions_resource_action (resource, action)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- User-Role mapping (many-to-many)
CREATE TABLE IF NOT EXISTS user_roles (
    user_id BIGINT UNSIGNED NOT NULL,
    role_id BIGINT UNSIGNED NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, role_id),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE,
    INDEX idx_user_roles_user (user_id),
    INDEX idx_user_roles_role (role_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Role-Permission mapping (many-to-many)
CREATE TABLE IF NOT EXISTS role_permissions (
    role_id BIGINT UNSIGNED NOT NULL,
    permission_id BIGINT UNSIGNED NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (role_id, permission_id),
    FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE,
    FOREIGN KEY (permission_id) REFERENCES permissions(id) ON DELETE CASCADE,
    INDEX idx_role_permissions_role (role_id),
    INDEX idx_role_permissions_permission (permission_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Bots table
CREATE TABLE IF NOT EXISTS bots (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    username VARCHAR(100) NOT NULL UNIQUE,
    token VARCHAR(255) NOT NULL UNIQUE,
    display_name VARCHAR(255),
    description TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    webhook_url VARCHAR(512),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_bots_username (username),
    INDEX idx_bots_active (is_active)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Chats table
CREATE TABLE IF NOT EXISTS chats (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    bot_id BIGINT UNSIGNED NOT NULL,
    telegram_id BIGINT NOT NULL,
    type VARCHAR(50) NOT NULL,
    title VARCHAR(255),
    username VARCHAR(100),
    first_name VARCHAR(255),
    last_name VARCHAR(255),
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY idx_bot_telegram_chat (bot_id, telegram_id),
    FOREIGN KEY (bot_id) REFERENCES bots(id) ON DELETE CASCADE,
    INDEX idx_chats_bot (bot_id),
    INDEX idx_chats_telegram_id (telegram_id),
    INDEX idx_chats_type (type),
    INDEX idx_chats_active (is_active)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- API Keys table
CREATE TABLE IF NOT EXISTS api_keys (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    `key` VARCHAR(100) NOT NULL UNIQUE,
    hashed_key VARCHAR(255) NOT NULL UNIQUE,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    scopes TEXT,
    rate_limit INT DEFAULT 1000,
    is_active BOOLEAN DEFAULT TRUE,
    expires_at TIMESTAMP NULL,
    last_used_at TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_api_keys_key (`key`),
    INDEX idx_api_keys_active (is_active),
    INDEX idx_api_keys_expires (expires_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Chat Permissions table (chat-level ACL)
CREATE TABLE IF NOT EXISTS chat_permissions (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    chat_id BIGINT UNSIGNED NOT NULL,
    user_id BIGINT UNSIGNED NULL,
    api_key_id BIGINT UNSIGNED NULL,
    can_read BOOLEAN DEFAULT FALSE,
    can_send BOOLEAN DEFAULT FALSE,
    can_manage BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY idx_chat_user (chat_id, user_id),
    UNIQUE KEY idx_chat_apikey (chat_id, api_key_id),
    FOREIGN KEY (chat_id) REFERENCES chats(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (api_key_id) REFERENCES api_keys(id) ON DELETE CASCADE,
    INDEX idx_chat_permissions_chat (chat_id),
    INDEX idx_chat_permissions_user (user_id),
    INDEX idx_chat_permissions_apikey (api_key_id),
    CHECK (
        (user_id IS NOT NULL AND api_key_id IS NULL) OR
        (user_id IS NULL AND api_key_id IS NOT NULL)
    )
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Messages table
CREATE TABLE IF NOT EXISTS messages (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    chat_id BIGINT UNSIGNED NOT NULL,
    telegram_id BIGINT NOT NULL,
    from_user_id BIGINT NULL,
    from_username VARCHAR(100),
    from_first_name VARCHAR(255),
    from_last_name VARCHAR(255),
    direction VARCHAR(20) NOT NULL,
    message_type VARCHAR(50) NOT NULL,
    text TEXT,
    raw_data LONGTEXT,
    reply_to_message_id BIGINT NULL,
    sent_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (chat_id) REFERENCES chats(id) ON DELETE CASCADE,
    INDEX idx_messages_chat_sent (chat_id, sent_at DESC),
    INDEX idx_messages_telegram_id (telegram_id),
    INDEX idx_messages_direction (direction),
    INDEX idx_messages_reply_to (reply_to_message_id),
    INDEX idx_messages_created (created_at DESC)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Webhooks table
CREATE TABLE IF NOT EXISTS webhooks (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    url VARCHAR(512) NOT NULL,
    secret VARCHAR(255) NOT NULL,
    scope VARCHAR(20) NOT NULL,
    chat_id BIGINT UNSIGNED NULL,
    reply_to_message_id BIGINT NULL,
    events TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (chat_id) REFERENCES chats(id) ON DELETE CASCADE,
    INDEX idx_webhooks_chat (chat_id),
    INDEX idx_webhooks_scope (scope),
    INDEX idx_webhooks_active (is_active)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Webhook Deliveries table
CREATE TABLE IF NOT EXISTS webhook_deliveries (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    webhook_id BIGINT UNSIGNED NOT NULL,
    message_id BIGINT UNSIGNED NOT NULL,
    status VARCHAR(20) NOT NULL,
    attempt_count INT DEFAULT 0,
    last_error TEXT,
    next_retry_at TIMESTAMP NULL,
    delivered_at TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (webhook_id) REFERENCES webhooks(id) ON DELETE CASCADE,
    FOREIGN KEY (message_id) REFERENCES messages(id) ON DELETE CASCADE,
    INDEX idx_webhook_deliveries_webhook_status (webhook_id, status),
    INDEX idx_webhook_deliveries_next_retry (next_retry_at),
    INDEX idx_webhook_deliveries_status (status),
    INDEX idx_webhook_deliveries_created (created_at DESC)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Refresh Tokens table
CREATE TABLE IF NOT EXISTS refresh_tokens (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED NOT NULL,
    token VARCHAR(255) NOT NULL UNIQUE,
    expires_at TIMESTAMP NOT NULL,
    revoked_at TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    INDEX idx_refresh_tokens_user (user_id),
    INDEX idx_refresh_tokens_token (token),
    INDEX idx_refresh_tokens_expires (expires_at),
    INDEX idx_refresh_tokens_revoked (revoked_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Insert default roles
INSERT INTO roles (name, description) VALUES
    ('admin', 'Full system access'),
    ('operator', 'Can manage bots and view messages'),
    ('viewer', 'Read-only access')
ON DUPLICATE KEY UPDATE description=VALUES(description);

-- Insert default permissions
INSERT INTO permissions (name, description, resource, action) VALUES
    ('users:read', 'View users', 'users', 'read'),
    ('users:create', 'Create users', 'users', 'create'),
    ('users:update', 'Update users', 'users', 'update'),
    ('users:delete', 'Delete users', 'users', 'delete'),
    ('bots:read', 'View bots', 'bots', 'read'),
    ('bots:create', 'Create bots', 'bots', 'create'),
    ('bots:update', 'Update bots', 'bots', 'update'),
    ('bots:delete', 'Delete bots', 'bots', 'delete'),
    ('chats:read', 'View chats', 'chats', 'read'),
    ('messages:read', 'View messages', 'messages', 'read'),
    ('messages:send', 'Send messages', 'messages', 'send'),
    ('webhooks:read', 'View webhooks', 'webhooks', 'read'),
    ('webhooks:create', 'Create webhooks', 'webhooks', 'create'),
    ('webhooks:update', 'Update webhooks', 'webhooks', 'update'),
    ('webhooks:delete', 'Delete webhooks', 'webhooks', 'delete'),
    ('apikeys:read', 'View API keys', 'apikeys', 'read'),
    ('apikeys:create', 'Create API keys', 'apikeys', 'create'),
    ('apikeys:delete', 'Delete API keys', 'apikeys', 'delete')
ON DUPLICATE KEY UPDATE description=VALUES(description);

-- Assign permissions to admin role
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
CROSS JOIN permissions p
WHERE r.name = 'admin'
ON DUPLICATE KEY UPDATE role_id=role_id;

-- Assign read permissions to viewer role
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
CROSS JOIN permissions p
WHERE r.name = 'viewer' AND p.action = 'read'
ON DUPLICATE KEY UPDATE role_id=role_id;

-- Assign operator permissions
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
CROSS JOIN permissions p
WHERE r.name = 'operator' AND (
    p.name IN ('bots:read', 'bots:create', 'bots:update', 'chats:read', 'messages:read', 'messages:send', 'webhooks:read', 'webhooks:create', 'webhooks:update')
)
ON DUPLICATE KEY UPDATE role_id=role_id;
