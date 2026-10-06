-- 0003_user_time_start_date: the day the /time page starts from when no range
-- is given. NULL keeps the previous behaviour (January 1 of the current year).
-- Stored as "2006-01-02" text, the same shape the page's date inputs submit.

ALTER TABLE users ADD COLUMN time_start_date TEXT;
