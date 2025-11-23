package app

import (
	"context"
	"errors"
	"math/rand"
	"time"

	"github.com/wsppppp/manage-pull-request/internal/domain"
	"github.com/wsppppp/manage-pull-request/internal/repository"
)

// Service инкапсулирует бизнес-логику приложения.
type Service struct {
	repo repository.Repository
}

func New(repo repository.Repository) *Service {
	return &Service{
		repo: repo,
	}
}

// CreateTeam создает команду. Возвращает ошибку ErrTeamExists, если команда уже существует.
func (s *Service) CreateTeam(ctx context.Context, team domain.Team) (domain.Team, error) {
	return s.repo.CreateTeam(ctx, team)
}

// GetTeam получает команду по имени. Возвращает ErrNotFound, если команда не найдена.
func (s *Service) GetTeam(ctx context.Context, teamName string) (domain.Team, error) {
	return s.repo.GetTeamByName(ctx, teamName)
}

// SetUserActivity обновляет статус активности пользователя.
func (s *Service) SetUserActivity(ctx context.Context, userID string, isActive bool) (*domain.User, error) {
	return s.repo.SetUserActivity(ctx, userID, isActive)
}

func (s *Service) CreatePullRequest(ctx context.Context, prID, prName, authorID string) (*domain.PullRequest, error) {
	author, err := s.repo.GetUserByID(ctx, authorID)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return nil, ErrAuthorNotFound
		}
		return nil, err
	}

	if author.TeamName == "" {
		return nil, ErrAuthorNotFound
	}

	team, err := s.repo.GetTeamByName(ctx, author.TeamName)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, ErrAuthorNotFound
		}
		return nil, err
	}

	candidates := make([]domain.User, 0)
	for _, member := range team.Members {
		if member.IsActive && member.ID != authorID {
			candidates = append(candidates, member)
		}
	}

	rand.Shuffle(len(candidates), func(i, j int) {
		candidates[i], candidates[j] = candidates[j], candidates[i]
	})

	reviewers := make([]string, 0, 2)
	for i := 0; i < len(candidates) && i < 2; i++ {
		reviewers = append(reviewers, candidates[i].ID)
	}
	pr := domain.PullRequest{
		ID:                prID,
		Name:              prName,
		AuthorID:          authorID,
		Status:            domain.StatusOpen,
		AssignedReviewers: reviewers,
		CreatedAt:         time.Now(),
	}
	return s.repo.CreatePullRequest(ctx, pr)
}

// MergePullRequest мерджит pr.
func (s *Service) MergePullRequest(ctx context.Context, prID string) (*domain.PullRequest, error) {
	pr, err := s.repo.GetPullRequestByID(ctx, prID)
	if err != nil {
		return nil, err // Пробрасываем ошибку (например, ErrPRNotFound)
	}

	if pr.Status == domain.StatusMerged {
		return nil, ErrPRMerged
	}

	return s.repo.MergePullRequest(ctx, prID)
}

// ReassignReviewer заменяет ревьюера на нового
func (s *Service) ReassignReviewer(ctx context.Context, prID, oldReviewerID string) (*domain.PullRequest, string, error) {
	pr, err := s.repo.GetPullRequestByID(ctx, prID)
	if err != nil {
		return nil, "", err
	}
	if pr.Status == domain.StatusMerged {
		return nil, "", ErrPRMerged
	}

	isAssigned := false
	for _, reviewer := range pr.AssignedReviewers {
		if reviewer == oldReviewerID {
			isAssigned = true
			break
		}
	}
	if !isAssigned {
		return nil, "", ErrReviewerNotAssigned
	}

	author, err := s.repo.GetUserByID(ctx, pr.AuthorID)
	if err != nil {
		return nil, "", ErrAuthorNotFound
	}
	team, err := s.repo.GetTeamByName(ctx, author.TeamName)
	if err != nil {
		return nil, "", ErrAuthorNotFound
	}

	currentReviewers := make(map[string]struct{})
	for _, r := range pr.AssignedReviewers {
		currentReviewers[r] = struct{}{}
	}
	candidates := make([]string, 0)
	for _, member := range team.Members {
		_, isReviewer := currentReviewers[member.ID]
		if member.IsActive && member.ID != pr.AuthorID && !isReviewer {
			candidates = append(candidates, member.ID)
		}
	}

	if len(candidates) == 0 {
		return nil, "", ErrNoCandidates
	}

	rand.Shuffle(len(candidates), func(i, j int) { candidates[i], candidates[j] = candidates[j], candidates[i] })
	newReviewerID := candidates[0]

	newReviewersList := make([]string, 0, len(pr.AssignedReviewers))
	for _, reviewer := range pr.AssignedReviewers {
		if reviewer == oldReviewerID {
			newReviewersList = append(newReviewersList, newReviewerID) // Замена
		} else {
			newReviewersList = append(newReviewersList, reviewer)
		}
	}
	pr.AssignedReviewers = newReviewersList

	if err := s.repo.UpdatePullRequestReviewers(ctx, prID, pr.AssignedReviewers); err != nil {
		return nil, "", err
	}

	return pr, newReviewerID, nil
}

// GetPullRequestsByReviewer возвращает все открытые pr для ревьюера.
func (s *Service) GetPullRequestsByReviewer(ctx context.Context, userID string) ([]*domain.PullRequest, error) {
	return s.repo.GetOpenPullRequestsByReviewer(ctx, userID)
}
