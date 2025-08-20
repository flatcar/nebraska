-- +migrate Up

-- Create junction table for channel-specific floors
CREATE TABLE IF NOT EXISTS channel_package_floors (
    channel_id uuid NOT NULL REFERENCES channel(id) ON DELETE CASCADE,
    package_id uuid NOT NULL REFERENCES package(id) ON DELETE CASCADE,
    floor_reason text,
    created_ts timestamptz DEFAULT current_timestamp,
    PRIMARY KEY (channel_id, package_id)
);

-- Create indexes for efficient querying
CREATE INDEX IF NOT EXISTS idx_channel_package_floors_channel ON channel_package_floors(channel_id);
CREATE INDEX IF NOT EXISTS idx_channel_package_floors_package ON channel_package_floors(package_id);

-- +migrate Down

-- Drop the junction table
DROP TABLE IF EXISTS channel_package_floors;