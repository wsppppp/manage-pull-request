package http

import (
	"github.com/wsppppp/manage-pull-request/internal/domain"
	"time"
)

// TeamMemberDTO - модель участника команды для API.
type TeamMemberDTO struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	IsActive bool   `json:"is_active"`
}

// TeamDTO - модель команды для API.
type TeamDTO struct {
	Name    string          `json:"team_name"`
	Members []TeamMemberDTO `json:"members"`
}

// toDomainTeam конвертирует DTO в доменную модель Team.
func toDomainTeam(dto TeamDTO) domain.Team {
	members := make([]domain.User, len(dto.Members))
	for i, m := range dto.Members {
		members[i] = domain.User{
			ID:       m.UserID,
			Username: m.Username,
			IsActive: m.IsActive,
			TeamName: dto.Name,
		}
	}
	return domain.Team{
		Name:    dto.Name,
		Members: members,
	}
}

// fromDomainTeam конвертирует доменную модель Team в DTO.
func fromDomainTeam(team domain.Team) TeamDTO {
	members := make([]TeamMemberDTO, len(team.Members))
	for i, m := range team.Members {
		members[i] = TeamMemberDTO{
			UserID:   m.ID,
			Username: m.Username,
			IsActive: m.IsActive,
		}
	}
	return TeamDTO{
		Name:    team.Name,
		Members: members,
	}
}

// SetUserActivityRequest - модель запроса для установки активности пользователя.
type SetUserActivityRequest struct {
	UserID   string `json:"user_id"`
	IsActive bool   `json:"is_active"`
}

// UserDTO - модель пользователя для API ответа.
type UserDTO struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	TeamName string `json:"team_name"`
	IsActive bool   `json:"is_active"`
}

// fromDomainUser конвертирует доменную модель User (указатель) в DTO.
func fromDomainUser(user *domain.User) UserDTO { // Принимает *domain.User
	if user == nil {
		return UserDTO{}
	}
	return UserDTO{
		UserID:   user.ID,
		Username: user.Username,
		TeamName: user.TeamName,
		IsActive: user.IsActive,
	}
}

// CreatePullRequestRequest - модель запроса для создания PR.
type CreatePullRequestRequest struct {
	ID       string `json:"pull_request_id"`
	Name     string `json:"pull_request_name"`
	AuthorID string `json:"author_id"`
}

// MergePullRequestRequest - модель запроса для слияния PR.
type MergePullRequestRequest struct {
	PullRequestID string `json:"pull_request_id"`
}

// PullRequestDTO - модель PR для API ответа.
type PullRequestDTO struct {
	ID                string     `json:"pull_request_id"`
	Name              string     `json:"pull_request_name"`
	AuthorID          string     `json:"author_id"`
	Status            string     `json:"status"`
	AssignedReviewers []string   `json:"assigned_reviewers"`
	CreatedAt         time.Time  `json:"createdAt"`
	MergedAt          *time.Time `json:"mergedAt,omitempty"`
}

// fromDomainPR конвертирует доменную модель PullRequest (указатель) в DTO.
func fromDomainPR(pr *domain.PullRequest) PullRequestDTO {
	if pr == nil {
		return PullRequestDTO{}
	}
	return PullRequestDTO{
		ID:                pr.ID,
		Name:              pr.Name,
		AuthorID:          pr.AuthorID,
		Status:            string(pr.Status),
		AssignedReviewers: pr.AssignedReviewers,
		CreatedAt:         pr.CreatedAt,
		MergedAt:          pr.MergedAt,
	}
}

// ReassignReviewerRequest - модель запроса для переназначения.
type ReassignReviewerRequest struct {
	PullRequestID string `json:"pull_request_id"`
	OldReviewerID string `json:"old_reviewer_id"`
}

// PullRequestShortDTO - укороченная версия для /users/getReview.
type PullRequestShortDTO struct {
	ID       string `json:"pull_request_id"`
	Name     string `json:"pull_request_name"`
	AuthorID string `json:"author_id"`
	Status   string `json:"status"`
}

func fromDomainPRtoShort(pr *domain.PullRequest) PullRequestShortDTO {
	return PullRequestShortDTO{
		ID:       pr.ID,
		Name:     pr.Name,
		AuthorID: pr.AuthorID,
		Status:   string(pr.Status),
	}
}
