package app

import (
	"errors"
	"fmt"
)

// Определяем кастомные ошибки для бизнес-логики для удобной обработки в HTTP-слое.
var (
	ErrUserNotFound          = errors.New("user not found")
	ErrAuthorNotFound        = errors.New("pr author not found")
	ErrPRExists              = errors.New("pull request already exists")
	ErrPRNotFound            = errors.New("pull request not found")
	ErrPRMerged              = errors.New("pull request is already merged")
	ErrNotFound              = errors.New("not found")
	ErrNoReviewersToReassign = errors.New("no reviewers assigned to reassign")
	ErrNoAvailableReviewers  = errors.New("no available reviewers found to reassign")
)

// ErrTeamExists - это тип ошибки, который содержит имя команды.
type ErrTeamExists struct {
	TeamName string
}

// Error делает ErrTeamExists совместимым с интерфейсом error.
func (e *ErrTeamExists) Error() string {
	return fmt.Sprintf("team %q already exists", e.TeamName)
}
