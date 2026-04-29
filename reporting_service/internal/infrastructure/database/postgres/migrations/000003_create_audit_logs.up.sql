CREATE TABLE IF NOT EXISTS audit_logs(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,
    tenant_id BIGINT NOT NULL,
    action_type VARCHAR(50) NOT NULL,
    previous_total INTEGER NOT NULL,
    new_total INTEGER NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_audit_user_tenant ON audit_logs(user_id, tenant_id);