package repository

import (
	"context"

	"github.com/wsppppp/manage-pull-request/internal/domain"
)

// Repository определяет интерфейс для взаимодействия с хранилищем данных
type Repository interface {
	// команды
	CreateTeam(ctx context.Context, team domain.Team) (domain.Team, error)
	GetTeamByName(ctx context.Context, teamName string) (domain.Team, error)

	// юзеры
	SetUserActivity(ctx context.Context, userID string, isActive bool) (*domain.User, error)
	GetUserByID(ctx context.Context, userID string) (*domain.User, error)

	// pr
	CreatePullRequest(ctx context.Context, pr domain.PullRequest) (*domain.PullRequest, error)
	GetPullRequestByID(ctx context.Context, prID string) (*domain.PullRequest, error)
	MergePullRequest(ctx context.Context, prID string) (*domain.PullRequest, error)
	UpdatePullRequestReviewers(ctx context.Context, prID string, reviewers []string) error
	GetOpenPullRequestsByReviewer(ctx context.Context, userID string) ([]*domain.PullRequest, error)
}
