-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS teams (
                                     name VARCHAR(255) PRIMARY KEY,
                                     created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS users (
                                     user_id VARCHAR(50) PRIMARY KEY,
                                     username VARCHAR(255) NOT NULL,
                                     team_id VARCHAR(255) NOT NULL REFERENCES teams(name) ON DELETE CASCADE,
                                     is_active BOOLEAN DEFAULT FALSE,
                                     created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
                                     updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS pull_requests (
                                             id VARCHAR(100) PRIMARY KEY,
                                             name VARCHAR(255) NOT NULL,
                                             author_id VARCHAR(50) NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
                                             status VARCHAR(20) DEFAULT 'OPEN' CHECK (status IN ('OPEN', 'MERGED')),
                                             need_more_reviewers BOOLEAN DEFAULT FALSE,
                                             created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
                                             merged_at TIMESTAMP WITH TIME ZONE NULL
);

CREATE TABLE IF NOT EXISTS pr_reviewers (
                                            pr_id VARCHAR(100) REFERENCES pull_requests(id) ON DELETE CASCADE,
                                            user_id VARCHAR(50) REFERENCES users(user_id) ON DELETE CASCADE,
                                            assigned_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
                                            PRIMARY KEY (pr_id, user_id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS pr_reviewers;
DROP TABLE IF EXISTS pull_requests;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS teams;
-- +goose StatementEnd
