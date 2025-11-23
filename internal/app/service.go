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

// New создает новый экземпляр Service.
func New(repo repository.Repository) *Service {
	return &Service{
		repo: repo,
	}
}

// CreateTeam создает команду. Возвращает ошибку ErrTeamExists, если команда уже существует.
func (s *Service) CreateTeam(ctx context.Context, team domain.Team) (domain.Team, error) {
	// Проверка на существование вынесена на уровень БД (PRIMARY KEY),
	// но можно добавить и здесь, если требуется.
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

// MergePullRequest выполняет слияние PR.
func (s *Service) MergePullRequest(ctx context.Context, prID string) (*domain.PullRequest, error) {
	// Сначала получаем PR, чтобы проверить его статус.
	// Метод GetPullRequestByID нужно будет добавить в интерфейс и реализацию репозитория.
	pr, err := s.repo.GetPullRequestByID(ctx, prID)
	if err != nil {
		return nil, err // Пробрасываем ошибку (например, ErrPRNotFound)
	}

	// Проверяем, не был ли PR уже смержен.
	if pr.Status == domain.StatusMerged {
		return nil, ErrPRMerged
	}

	// Если все в порядке, вызываем метод репозитория для слияния.
	return s.repo.MergePullRequest(ctx, prID)
}

func (s *Service) ReassignReviewer(ctx context.Context, prID string) (*domain.PullRequest, error) {
	// 1. Получаем PR и проверяем базовые случаи
	pr, err := s.repo.GetPullRequestByID(ctx, prID)
	if err != nil {
		return nil, err // Обрабатывает ErrPRNotFound
	}
	if pr.Status == domain.StatusMerged {
		return nil, ErrPRMerged
	}
	if len(pr.AssignedReviewers) == 0 {
		return nil, ErrNoReviewersToReassign
	}

	// 2. Находим команду автора
	author, err := s.repo.GetUserByID(ctx, pr.AuthorID)
	if err != nil {
		return nil, ErrAuthorNotFound // Если автор удален после создания PR
	}
	team, err := s.repo.GetTeamByName(ctx, author.TeamName)
	if err != nil {
		return nil, ErrAuthorNotFound // Если команда удалена
	}

	// 3. Собираем карту текущих ревьюеров для быстрой проверки
	currentReviewers := make(map[string]struct{}, len(pr.AssignedReviewers))
	for _, r := range pr.AssignedReviewers {
		currentReviewers[r] = struct{}{}
	}

	// 4. Ищем кандидатов для замены
	candidates := make([]string, 0)
	for _, member := range team.Members {
		// Кандидат: активен, не автор, и еще не в ревьюерах
		_, isReviewer := currentReviewers[member.ID]
		if member.IsActive && member.ID != pr.AuthorID && !isReviewer {
			candidates = append(candidates, member.ID)
		}
	}

	if len(candidates) == 0 {
		return nil, ErrNoAvailableReviewers
	}

	// 5. Выбираем, кого меняем и на кого меняем
	rand.Shuffle(len(candidates), func(i, j int) { candidates[i], candidates[j] = candidates[j], candidates[i] })
	newReviewerID := candidates[0] // Берем первого случайного кандидата

	rand.Shuffle(len(pr.AssignedReviewers), func(i, j int) {
		pr.AssignedReviewers[i], pr.AssignedReviewers[j] = pr.AssignedReviewers[j], pr.AssignedReviewers[i]
	})
	pr.AssignedReviewers[0] = newReviewerID // Заменяем первого случайного старого ревьюера

	// 6. Обновляем список ревьюеров в базе
	if err := s.repo.UpdatePullRequestReviewers(ctx, prID, pr.AssignedReviewers); err != nil {
		return nil, err
	}

	// Возвращаем обновленную модель PR
	return pr, nil
}

// GetPullRequestsByReviewer возвращает все открытые PR, назначенные пользователю.
func (s *Service) GetPullRequestsByReviewer(ctx context.Context, userID string) ([]*domain.PullRequest, error) {
	// Здесь можно добавить проверку, существует ли пользователь, но для данной задачи это избыточно.
	return s.repo.GetOpenPullRequestsByReviewer(ctx, userID)
}
