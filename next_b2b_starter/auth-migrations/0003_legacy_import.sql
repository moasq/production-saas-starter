-- Record completed imports so rerunning the cutover tool cannot reinstate a removed member.
CREATE TABLE legacy_import (
  source_kind text NOT NULL CHECK (source_kind IN ('organization','account')),
  source_id bigint NOT NULL,
  target_id text NOT NULL,
  imported_at timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY (source_kind, source_id)
);
