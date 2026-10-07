-- 0004_user_vocabulary: the wording preset the UI uses for this user (clients
-- vs projects, see presets.* in the locale files). "default" keeps the
-- previous wording.

ALTER TABLE users ADD COLUMN vocabulary TEXT NOT NULL DEFAULT 'default';
