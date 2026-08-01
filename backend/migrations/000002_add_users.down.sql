DROP INDEX IF EXISTS idx_history_user_created;
DROP INDEX IF EXISTS idx_outfits_user_id;

ALTER TABLE history DROP COLUMN IF EXISTS user_id;
ALTER TABLE outfits DROP COLUMN IF EXISTS user_id;

DROP TABLE IF EXISTS email_otps;
DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS users;
