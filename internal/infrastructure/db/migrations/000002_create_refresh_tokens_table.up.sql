CREATE TABLE refresh_tokens (
    id              UUID                PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         BIGINT              NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash      TEXT                UNIQUE NOT NULL,

    -- Rotation chain
    parent_id       UUID                REFERENCES refresh_tokens(id),
    replaced_by     UUID                REFERENCES refresh_tokens(id),

    -- Revocation tracking
    revoked_at      TIMESTAMPTZ         NULL,

    -- Metadata
    user_agent      TEXT                NOT NULL,
    ip_address      TEXT                NOT NULL,
    created_at      TIMESTAMPTZ         NOT NULL DEFAULT NOW(),
    expires_at      TIMESTAMPTZ         NOT NULL,

    -- GeoIP snapshot
    country         TEXT                NULL,
    region          TEXT                NULL,
    city            TEXT                NULL,
    latitude        DOUBLE PRECISION    NULL,
    longitude       DOUBLE PRECISION    NULL,
    asn             INTEGER             NULL,
    asn_org         TEXT                NULL,
    is_vpn          BOOLEAN             DEFAULT FALSE,
    is_proxy        BOOLEAN             DEFAULT FALSE,
    is_tor          BOOLEAN             DEFAULT FALSE,
    is_datacenter   BOOLEAN             DEFAULT FALSE
);

CREATE INDEX ON refresh_tokens (user_id);
CREATE INDEX ON refresh_tokens (token_hash);
CREATE INDEX ON refresh_tokens (parent_id);
CREATE INDEX ON refresh_tokens (replaced_by);