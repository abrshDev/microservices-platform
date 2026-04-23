ALTER TABLE tasks ADD COLUMN tenant_id INTEGER NOT NULL DEFAULT 0;
CREATE INDEX idx_tasks_tenant_id ON tasks(tenant_id);