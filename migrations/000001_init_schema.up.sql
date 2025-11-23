-- таблица команд
CREATE TABLE IF NOT EXISTS teams (
    team_name VARCHAR(255) PRIMARY KEY
    );

-- таблица пользователей
CREATE TABLE IF NOT EXISTS users (
    user_id VARCHAR(255) PRIMARY KEY,
    username VARCHAR(255) NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    team_name VARCHAR(255) REFERENCES teams(team_name) ON DELETE SET NULL
    );

-- тип для статуса pr
CREATE TYPE pr_status AS ENUM ('OPEN', 'MERGED');

-- таблица для pr
CREATE TABLE IF NOT EXISTS pull_requests (
    pull_request_id VARCHAR(255) PRIMARY KEY,
    pull_request_name VARCHAR(255) NOT NULL,
    author_id VARCHAR(255) REFERENCES users(user_id) NOT NULL,
    status pr_status NOT NULL DEFAULT 'OPEN',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    merged_at TIMESTAMPTZ
    );

-- связующая таблица для ревьюеров PR
CREATE TABLE IF NOT EXISTS pr_reviewers (
    pr_id VARCHAR(255) REFERENCES pull_requests(pull_request_id) ON DELETE CASCADE,
    reviewer_id VARCHAR(255) REFERENCES users(user_id) ON DELETE CASCADE,
    PRIMARY KEY (pr_id, reviewer_id)
    );

-- индексы
CREATE INDEX IF NOT EXISTS idx_users_team_name ON users(team_name);
CREATE INDEX IF NOT EXISTS idx_pr_author_id ON pull_requests(author_id);