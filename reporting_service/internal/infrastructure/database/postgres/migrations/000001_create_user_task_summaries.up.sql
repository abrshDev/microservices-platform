CREATE TABLE IF NOT EXISTS user_task_summaries (
    user_id INTEGER NOT NULL,
    tenant_id INTEGER NOT NULL,
    total_tasks INTEGER DEFAULT 0,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, tenant_id)
);

CREATE INDEX idx_summaries_tenant ON user_task_summaries(tenant_id);