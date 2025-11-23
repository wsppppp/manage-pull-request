-- Создаем таблицу команд
CREATE TABLE IF NOT EXISTS teams (
                                     team_name VARCHAR(255) PRIMARY KEY
    );

-- Создаем таблицу пользователей
CREATE TABLE IF NOT EXISTS users (
                                     user_id VARCHAR(255) PRIMARY KEY,
    username VARCHAR(255) NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    team_name VARCHAR(255) REFERENCES teams(team_name) ON DELETE SET NULL
    );

-- Создаем тип для статуса PR
CREATE TYPE pr_status AS ENUM ('OPEN', 'MERGED');

-- Создаем таблицу pull requests
CREATE TABLE IF NOT EXISTS pull_requests (
                                             pull_request_id VARCHAR(255) PRIMARY KEY,
    pull_request_name VARCHAR(255) NOT NULL,
    author_id VARCHAR(255) REFERENCES users(user_id) NOT NULL,
    status pr_status NOT NULL DEFAULT 'OPEN',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    merged_at TIMESTAMPTZ
    );

-- Создаем связующую таблицу для ревьюеров PR
-- Это более гибкое решение, чем хранить массив ID в таблице pull_requests
CREATE TABLE IF NOT EXISTS pr_reviewers (
    pr_id VARCHAR(255) REFERENCES pull_requests(pull_request_id) ON DELETE CASCADE,
    reviewer_id VARCHAR(255) REFERENCES users(user_id) ON DELETE CASCADE,
    PRIMARY KEY (pr_id, reviewer_id)
    );

-- Индексы для ускорения выборок
CREATE INDEX IF NOT EXISTS idx_users_team_name ON users(team_name);
CREATE INDEX IF NOT EXISTS idx_pr_author_id ON pull_requests(author_id);