-- 1. Change user_id from INTEGER to VARCHAR to support UUIDs
ALTER TABLE user_task_summaries ALTER COLUMN user_id TYPE VARCHAR(255);

-- 2. Change tenant_id to BIGINT to match the Task Service
ALTER TABLE user_task_summaries ALTER COLUMN tenant_id TYPE BIGINT;