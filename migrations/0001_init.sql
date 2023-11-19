CREATE TABLE IF NOT EXISTS {{.TablesPrefix}}request (
  "id" TEXT NOT NULL PRIMARY KEY,
  "type" TEXT NOT NULL,
  "signature" TEXT NOT NULL UNIQUE,
  "client_id" TEXT NOT NULL,
  "requested_at" TIMESTAMP NOT NULL,
  "requested_scope" TEXT[],
  "granted_scope" TEXT[],
  "requested_audience" TEXT[],
  "granted_audience" TEXT[],
  "form" TEXT,
  "session" JSONB,
  "active" BOOLEAN NOT NULL DEFAULT TRUE
);
