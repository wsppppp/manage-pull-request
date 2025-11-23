package http

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/wsppppp/manage-pull-request/internal/app"
)

// Handler обрабатывает HTTP-запросы.
type Handler struct {
	service *app.Service
}

// NewHandler создает новый экземпляр Handler.
func NewHandler(service *app.Service) *Handler {
	return &Handler{service: service}
}

// --- Обработчики (остаются без изменений) ---

func (h *Handler) createTeam(w http.ResponseWriter, r *http.Request) {
	var teamDTO TeamDTO
	if err := json.NewDecoder(r.Body).Decode(&teamDTO); err != nil {
		writeError(w, "INVALID_REQUEST", "invalid request body", http.StatusBadRequest, err)
		return
	}

	team, err := h.service.CreateTeam(r.Context(), toDomainTeam(teamDTO))
	if err != nil {
		var teamExistsErr *app.ErrTeamExists
		if errors.As(err, &teamExistsErr) {
			writeError(w, "TEAM_EXISTS", teamExistsErr.Error(), http.StatusConflict, teamExistsErr)
			return
		}
		// --- КОНЕЦ ИСПРАВЛЕНИЯ ---

		writeError(w, "INTERNAL_ERROR", "internal server error", http.StatusInternalServerError, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
	// В случае успеха, переменная `team` содержит корректные данные от сервиса.
	json.NewEncoder(w).Encode(map[string]any{"team": fromDomainTeam(team)})
}
func (h *Handler) getTeam(w http.ResponseWriter, r *http.Request) {
	teamName := r.URL.Query().Get("team_name")
	if teamName == "" {
		writeError(w, "INVALID_REQUEST", "team_name is required", http.StatusBadRequest, nil)
		return
	}

	team, err := h.service.GetTeam(r.Context(), teamName)
	if err != nil {
		if errors.Is(err, app.ErrNotFound) {
			writeError(w, "NOT_FOUND", "team not found", http.StatusNotFound, err)
			return
		}
		writeError(w, "INTERNAL_ERROR", "internal server error", http.StatusInternalServerError, err)
		return
	}

	json.NewEncoder(w).Encode(fromDomainTeam(team))
}

func (h *Handler) setUserActivity(w http.ResponseWriter, r *http.Request) {
	var req SetUserActivityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "INVALID_REQUEST", "invalid request body", http.StatusBadRequest, err)
		return
	}

	user, err := h.service.SetUserActivity(r.Context(), req.UserID, req.IsActive)
	if err != nil {
		if errors.Is(err, app.ErrUserNotFound) {
			writeError(w, "NOT_FOUND", "user not found", http.StatusNotFound, err)
			return
		}
		writeError(w, "INTERNAL_ERROR", "internal server error", http.StatusInternalServerError, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{"user": fromDomainUser(user)})
}

func (h *Handler) createPullRequest(w http.ResponseWriter, r *http.Request) {
	var req CreatePullRequestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "INVALID_REQUEST", "invalid request body", http.StatusBadRequest, err)
		return
	}

	pr, err := h.service.CreatePullRequest(r.Context(), req.ID, req.Name, req.AuthorID)
	if err != nil {
		switch {
		case errors.Is(err, app.ErrAuthorNotFound):
			writeError(w, "NOT_FOUND", "author or author's team not found", http.StatusNotFound, err)
		case errors.Is(err, app.ErrPRExists):
			writeError(w, "PR_EXISTS", "PR id already exists", http.StatusConflict, err)
		default:
			writeError(w, "INTERNAL_ERROR", "internal server error", http.StatusInternalServerError, err)
		}
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]any{"pr": fromDomainPR(pr)})
}

// --- Helpers & Middleware ---

func writeError(w http.ResponseWriter, code, message string, httpStatus int, err error) {
	if err != nil {
		log.Printf("ERROR: %v", err)
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(httpStatus)
	json.NewEncoder(w).Encode(map[string]any{
		"error": map[string]string{
			"code":    code,
			"message": message,
		},
	})
}

// setContentTypeJSON теперь является методом Handler.
func (h *Handler) setContentTypeJSON(next http.Handler) http.Handler { // <--- ИЗМЕНЕНИЕ ЗДЕСЬ
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		next.ServeHTTP(w, r)
	})
}

func (h *Handler) mergePullRequest(w http.ResponseWriter, r *http.Request) {
	var req MergePullRequestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "INVALID_REQUEST", "invalid request body", http.StatusBadRequest, err)
		return
	}

	pr, err := h.service.MergePullRequest(r.Context(), req.PullRequestID)
	if err != nil {
		switch {
		case errors.Is(err, app.ErrPRNotFound):
			writeError(w, "NOT_FOUND", "pull request not found", http.StatusNotFound, err)
		case errors.Is(err, app.ErrPRMerged):
			writeError(w, "PR_ALREADY_MERGED", "pull request is already merged", http.StatusConflict, err) // 409 Conflict
		default:
			writeError(w, "INTERNAL_ERROR", "internal server error", http.StatusInternalServerError, err)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{"pr": fromDomainPR(pr)})
}

func (h *Handler) reassignReviewer(w http.ResponseWriter, r *http.Request) {
	var req ReassignReviewerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "INVALID_REQUEST", "invalid request body", http.StatusBadRequest, err)
		return
	}

	pr, err := h.service.ReassignReviewer(r.Context(), req.PullRequestID)
	if err != nil {
		switch {
		case errors.Is(err, app.ErrPRNotFound):
			writeError(w, "NOT_FOUND", "pull request not found", http.StatusNotFound, err)
		case errors.Is(err, app.ErrPRMerged):
			writeError(w, "PR_ALREADY_MERGED", "cannot reassign reviewer for a merged pr", http.StatusConflict, err)
		case errors.Is(err, app.ErrNoReviewersToReassign):
			writeError(w, "NO_REVIEWERS", "pr has no reviewers to reassign", http.StatusBadRequest, err)
		case errors.Is(err, app.ErrNoAvailableReviewers):
			writeError(w, "NO_CANDIDATES", "no available candidates to assign", http.StatusNotFound, err)
		default:
			writeError(w, "INTERNAL_ERROR", "internal server error", http.StatusInternalServerError, err)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{"pr": fromDomainPR(pr)})
}

func (h *Handler) getReviews(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		writeError(w, "INVALID_REQUEST", "user_id is required", http.StatusBadRequest, nil)
		return
	}

	prs, err := h.service.GetPullRequestsByReviewer(r.Context(), userID)
	if err != nil {
		// В данном случае, даже если пользователь не найден, мы вернем пустой список, а не ошибку.
		// Обрабатываем только внутренние ошибки сервера.
		writeError(w, "INTERNAL_ERROR", "internal server error", http.StatusInternalServerError, err)
		return
	}

	// Конвертируем слайс доменных моделей в слайс DTO
	prDTOs := make([]PullRequestDTO, 0, len(prs))
	for _, pr := range prs {
		prDTOs = append(prDTOs, fromDomainPR(pr))
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{"pull_requests": prDTOs})
}
