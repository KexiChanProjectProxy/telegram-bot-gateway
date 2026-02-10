-- Add webhook_secret column to bots table
ALTER TABLE bots ADD COLUMN webhook_secret VARCHAR(64) AFTER webhook_url;
CREATE UNIQUE INDEX idx_bots_webhook_secret ON bots(webhook_secret);
