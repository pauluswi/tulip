-- +migrate Up
-- SQL in section 'Up' is executed when this migration is applied

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS paytokens (
    "id" UUID NOT NULL PRIMARY KEY,
    "token" VARCHAR NOT NULL,
    "token_date" DATE NOT NULL,
    "customer_id" VARCHAR NOT NULL,
    "valid_until" TIMESTAMP WITH TIME ZONE NOT NULL,
    "metadata" JSONB NOT NULL DEFAULT '{}',
    "created_at" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    "updated_at" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now()
);

-- Add unique index to ensure token is only unique for current date
CREATE UNIQUE INDEX IF NOT EXISTS idx_unq_tokens_token_token_date ON paytokens (token, token_date);
