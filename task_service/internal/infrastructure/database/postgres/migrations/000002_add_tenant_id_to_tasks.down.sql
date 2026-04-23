DROP INDEX IF EXISTS idx_tasks_tenant_id;
ALTER TABLE tasks DROP COLUMN IF EXISTS tenant_id;