package models

import "time"

type PRStatus = string

const (
	PRStatusOpen   PRStatus = "OPEN"
	PRStatusMerged PRStatus = "MERGED"
)

type ErrorCode string

const (
	ErrorTeamExists  ErrorCode = "TEAM_EXISTS"
	ErrorPRExists    ErrorCode = "PR_EXISTS"
	ErrorPRMerged    ErrorCode = "PR_MERGED"
	ErrorNotAssigned ErrorCode = "NOT_ASSIGNED"
	ErrorNoCandidate ErrorCode = "NO_CANDIDATE"
	ErrorNotFound    ErrorCode = "NOT_FOUND"
	ErrorInternal    ErrorCode = "INTERNAL"
	ErrorBadRequest  ErrorCode = "BAD_REQUEST"
)

type ErrorBody struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
}

type ErrorResponse struct {
	Error ErrorBody `json:"error"`
}
type TeamMember struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	IsActive bool   `json:"is_active"`
}

type Team struct {
	TeamName string       `json:"team_name"`
	Members  []TeamMember `json:"members"`
}

type User struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	TeamName string `json:"team_name"`
	IsActive bool   `json:"is_active"`
}

type PullRequest struct {
	PullRequestID     string     `json:"pull_request_id"`
	PullRequestName   string     `json:"pull_request_name"`
	AuthorID          string     `json:"author_id"`
	Status            string     `json:"status"` // OPEN, MERGED
	AssignedReviewers []string   `json:"assigned_reviewers"`
	CreatedAt         *time.Time `json:"createdAt,omitempty"`
	MergedAt          *time.Time `json:"mergedAt,omitempty"`
}

type PullRequestShort struct {
	PullRequestID   string `json:"pull_request_id"`
	PullRequestName string `json:"pull_request_name"`
	AuthorID        string `json:"author_id"`
	Status          string `json:"status"` // OPEN, MERGED
}

type Stats struct {
	TotalTeams       int `json:"total_teams"`
	TotalUsers       int `json:"total_users"`
	ActiveUsers      int `json:"active_users"`
	TotalPRs         int `json:"total_prs"`
	OpenPRs          int `json:"open_prs"`
	MergedPRs        int `json:"merged_prs"`
	TotalAssignments int `json:"total_assignments"`
}

type UsersStats struct {
	TotalUsers  int `json:"total_users"`
	ActiveUsers int `json:"active_users"`
}

type PRStats struct {
	TotalPRs  int `json:"total_prs"`
	OpenPRs   int `json:"open_prs"`
	MergedPRs int `json:"merged_prs"`
}
