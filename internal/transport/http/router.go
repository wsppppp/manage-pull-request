package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// NewRouter настраивает и возвращает роутер.
func (h *Handler) NewRouter() http.Handler {
	r := chi.NewRouter()

	// Подключаем middleware
	r.Use(middleware.RequestID) // Добавляет ID каждому запросу
	r.Use(middleware.RealIP)    // Определяет реальный IP клиента
	r.Use(middleware.Logger)    // Логирует запросы
	r.Use(middleware.Recoverer) // Восстанавливается после паник
	r.Use(h.setContentTypeJSON) // Устанавливает Content-Type: application/json

	// Группа роутов для команд
	r.Route("/team", func(r chi.Router) {
		r.Post("/add", h.createTeam)
		r.Get("/get", h.getTeam)
	})

	// Группа роутов для пользователей
	r.Route("/users", func(r chi.Router) {
		r.Post("/setIsActive", h.setUserActivity)
		r.Get("/getReview", h.getReviews)
	})

	// Группа роутов для Pull Request'ов
	r.Route("/pullRequest", func(r chi.Router) {
		r.Post("/create", h.createPullRequest)
		r.Post("/reassign", h.reassignReviewer)
		r.Post("/merge", h.mergePullRequest)
	})

	// Health-check эндпоинт
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"status": "Service is running!"}`))
	})

	return r
}
