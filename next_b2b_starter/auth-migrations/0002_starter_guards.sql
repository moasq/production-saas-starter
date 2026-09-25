-- Starter invariants supplement the schema generated from Better Auth 1.7.6.
ALTER TABLE member ADD CONSTRAINT member_single_role CHECK (role IN ('admin','manager','member'));
ALTER TABLE invitation ADD CONSTRAINT invitation_single_role CHECK (role IN ('admin','manager','member'));
CREATE UNIQUE INDEX member_user_organization_key ON member ("userId", "organizationId");
CREATE UNIQUE INDEX user_email_normalized_key ON "user" (lower(email));

CREATE TABLE auth_rate_limit (
  key text PRIMARY KEY,
  count integer NOT NULL,
  expires_at timestamptz NOT NULL
);
CREATE INDEX auth_rate_limit_expiry_idx ON auth_rate_limit (expires_at);
CREATE TABLE signup_intent (
  id text PRIMARY KEY,
  email text NOT NULL,
  name text NOT NULL,
  organization_name text NOT NULL,
  expires_at timestamptz NOT NULL,
  organization_id text REFERENCES organization(id)
);
CREATE INDEX signup_intent_expiry_idx ON signup_intent (expires_at);

CREATE FUNCTION preserve_last_admin() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
  IF OLD.role = 'admin' AND (TG_OP = 'DELETE' OR NEW.role <> 'admin') THEN
    -- Serialize all admin demotions/removals for one workspace, including concurrent requests.
    PERFORM 1 FROM organization WHERE id = OLD."organizationId" FOR UPDATE;
    IF FOUND AND NOT EXISTS (
      SELECT 1 FROM member WHERE "organizationId" = OLD."organizationId" AND role = 'admin' AND id <> OLD.id
    ) THEN
      RAISE EXCEPTION 'A workspace must retain at least one administrator' USING ERRCODE = '23514';
    END IF;
  END IF;
  IF TG_OP = 'DELETE' THEN RETURN OLD; END IF;
  RETURN NEW;
END;
$$;
CREATE TRIGGER member_last_admin BEFORE DELETE OR UPDATE OF role ON member
FOR EACH ROW EXECUTE FUNCTION preserve_last_admin();
