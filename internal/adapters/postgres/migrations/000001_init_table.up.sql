
CREATE TABLE teams (
    id          SERIAL PRIMARY KEY,
    team_name   TEXT UNIQUE NOT NULL,
    created_at  TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE users (
    id          SERIAL PRIMARY KEY,
    user_id     TEXT UNIQUE NOT NULL,     -- внешний ID из API
    username    TEXT NOT NULL,
    team_id     INT REFERENCES teams(id) ON DELETE SET NULL,
    is_active   BOOLEAN NOT NULL DEFAULT TRUE
);

CREATE TABLE pull_requests (
    id              SERIAL PRIMARY KEY,
    pr_id           TEXT UNIQUE NOT NULL,       -- pull_request_id
    pr_name         TEXT NOT NULL,              -- pull_request_name
    author_id       INT REFERENCES users(id),
    status          TEXT NOT NULL CHECK (status IN ('OPEN', 'MERGED')),
    created_at      TIMESTAMP NOT NULL DEFAULT NOW(),
    merged_at       TIMESTAMP
);

CREATE TABLE pull_request_reviewers (
    pr_id       INT REFERENCES pull_requests(id) ON DELETE CASCADE,
    user_id     INT REFERENCES users(id),
    PRIMARY KEY (pr_id, user_id)
);

-- Ограничение: максимум 2 ревьювера на PR
CREATE OR REPLACE FUNCTION check_reviewer_limit() RETURNS trigger AS $$
BEGIN
    IF (
        SELECT COUNT(*) FROM pull_request_reviewers
        WHERE pr_id = NEW.pr_id
    ) >= 2 THEN
        RAISE EXCEPTION 'Reviewer limit exceeded: max 2 reviewers per PR';
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_reviewer_limit
BEFORE INSERT ON pull_request_reviewers
FOR EACH ROW
EXECUTE FUNCTION check_reviewer_limit();
