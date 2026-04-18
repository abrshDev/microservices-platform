-- This table will serve as our "Memory" for idempotency
CREATE TABLE IF NOT EXISTS processed_events (
    event_id UUID PRIMARY KEY,         -- This will be the TaskID from Kafka
    status VARCHAR(20) NOT NULL,       -- To track 'PENDING' vs 'COMPLETED'
    processed_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);