CREATE TABLE IF NOT EXISTS events (
      id UUID PRIMARY KEY,
      user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
      event_date TIMESTAMP NOT NULL,
      event_name TEXT NOT NULL,
      description TEXT,
      status TEXT NOT NULL DEFAULT 'active',
      reminder_time TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_events_user_date ON events(user_id, event_date);
CREATE INDEX IF NOT EXISTS idx_events_status ON events(status);
CREATE INDEX IF NOT EXISTS idx_events_reminder ON events(reminder_time) WHERE reminder_time IS NOT NULL;
