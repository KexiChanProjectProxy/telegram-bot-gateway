-- Rollback migration 002_api_key_permissions

DROP TABLE IF EXISTS api_key_feedback_permissions;
DROP TABLE IF EXISTS api_key_bot_permissions;
