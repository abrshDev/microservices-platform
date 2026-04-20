CREATE TABLE IF NOT EXISTS user_task_summaries (
    user_id UUID PRIMARY KEY,
    username VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL,
    total_tasks INTEGER DEFAULT 0,
    last_task_title VARCHAR(255),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Index for fast lookups during reporting
CREATE INDEX idx_summaries_email ON user_task_summaries(email);