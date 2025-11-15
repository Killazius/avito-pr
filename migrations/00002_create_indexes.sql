-- +goose Up
-- +goose StatementBegin
CREATE INDEX IF NOT EXISTS idx_users_team_name ON users(team_name);
CREATE INDEX IF NOT EXISTS idx_users_is_active ON users(is_active);
CREATE INDEX IF NOT EXISTS idx_users_team_active ON users(team_name, is_active);
CREATE INDEX IF NOT EXISTS idx_pull_requests_author_id ON pull_requests(author_id);
CREATE INDEX IF NOT EXISTS idx_pull_requests_status ON pull_requests(status);
CREATE INDEX IF NOT EXISTS idx_pull_requests_created_at ON pull_requests(created_at);
CREATE INDEX IF NOT EXISTS idx_pr_reviewers_user_id ON pr_reviewers(user_id);
CREATE INDEX IF NOT EXISTS idx_pr_reviewers_assigned_at ON pr_reviewers(assigned_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_pr_reviewers_assigned_at;
DROP INDEX IF EXISTS idx_pr_reviewers_user_id;
DROP INDEX IF EXISTS idx_pull_requests_created_at;
DROP INDEX IF EXISTS idx_pull_requests_status;
DROP INDEX IF EXISTS idx_pull_requests_author_id;
DROP INDEX IF EXISTS idx_users_team_active;
DROP INDEX IF EXISTS idx_users_is_active;
DROP INDEX IF EXISTS idx_users_team_name;
-- +goose StatementEnd
