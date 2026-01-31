CREATE TABLE IF NOT EXISTS clicks
(
    id             UUID DEFAULT generateUUIDv4(),
    link_id        UUID NOT NULL,
    short_code     String NOT NULL,
    clicked_at     DateTime64(3) NOT NULL,
    ip_address     String NOT NULL,
    user_agent     String DEFAULT '',
    referer        String DEFAULT '',
    country_code   LowCardinality(String) DEFAULT '',
    region         String DEFAULT '',
    city           String DEFAULT '',
    browser        LowCardinality(String) DEFAULT '',
    browser_version String DEFAULT '',
    os             LowCardinality(String) DEFAULT '',
    os_version     String DEFAULT '',
    device_type    LowCardinality(String) DEFAULT '',
    is_bot         UInt8 DEFAULT 0,
    utm_source     String DEFAULT '',
    utm_medium     String DEFAULT '',
    utm_campaign   String DEFAULT ''
)
ENGINE = MergeTree()
PARTITION BY toYYYYMM(clicked_at)
ORDER BY (link_id, clicked_at)
TTL toDateTime(clicked_at) + INTERVAL 2 YEAR
SETTINGS index_granularity = 8192;
