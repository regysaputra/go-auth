CREATE TABLE email_verification_codes (
    id BIGSERIAL PRIMARY KEY,
    email TEXT UNIQUE NOT NULL,
    code_hash TEXT NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL
);