package app

import (
	"errors"
	"fmt"
)

// Определяем кастомные ошибки для бизнес-логики для удобной обработки в HTTP-слое.
var (
	ErrUserNotFound        = errors.New("user not found")
	ErrAuthorNotFound      = errors.New("pr author not found")
	ErrPRExists            = errors.New("pull request already exists")
	ErrPRNotFound          = errors.New("pull request not found")
	ErrPRMerged            = errors.New("pull request is already merged")
	ErrNotFound            = errors.New("not found")
	ErrReviewerNotAssigned = fmt.Errorf("reviewer is not assigned to this pull request")
	ErrNoCandidates        = fmt.Errorf("no active replacement candidate in team")
)

type ErrTeamExists struct {
	TeamName string
}

func (e *ErrTeamExists) Error() string {
	return fmt.Sprintf("team %q already exists", e.TeamName)
}
