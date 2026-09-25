ALTER TABLE signup_intent ADD COLUMN claim_token text;
ALTER TABLE signup_intent ADD COLUMN claimed_at timestamptz;
