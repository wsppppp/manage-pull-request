package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/wsppppp/manage-pull-request/internal/app"
	"github.com/wsppppp/manage-pull-request/internal/domain"
	"github.com/wsppppp/manage-pull-request/internal/repository"
)

const uniqueViolationCode = "23505"

type PgRepository struct {
	db *pgxpool.Pool
}

func New(db *pgxpool.Pool) repository.Repository {
	return &PgRepository{db: db}
}

func (r *PgRepository) CreateTeam(ctx context.Context, team domain.Team) (domain.Team, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return domain.Team{}, err
	}
	defer tx.Rollback(ctx)

	// Создаем запись о команде.
	_, err = tx.Exec(ctx, `INSERT INTO teams (team_name) VALUES ($1)`, team.Name)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == uniqueViolationCode {
			return domain.Team{}, &app.ErrTeamExists{TeamName: team.Name}
		}
		return domain.Team{}, err
	}

	// Добавляем или обновляем участников команды.
	for _, member := range team.Members {
		_, err = tx.Exec(ctx, `INSERT INTO users (user_id, username, is_active, team_name) VALUES ($1, $2, $3, $4)
                           ON CONFLICT (user_id) DO UPDATE SET username = $2, is_active = $3, team_name = $4`,
			member.ID, member.Username, member.IsActive, team.Name)
		if err != nil {
			return domain.Team{}, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.Team{}, err
	}
	return team, nil
}

func (r *PgRepository) GetTeamByName(ctx context.Context, teamName string) (domain.Team, error) {
	team := domain.Team{Name: teamName}

	// Находим всех участников указанной команды.
	rows, err := r.db.Query(ctx, `SELECT user_id, username, is_active FROM users WHERE team_name = $1`, teamName)
	if err != nil {
		return domain.Team{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var user domain.User
		if err := rows.Scan(&user.ID, &user.Username, &user.IsActive); err != nil {
			return domain.Team{}, err
		}
		user.TeamName = teamName
		team.Members = append(team.Members, user)
	}

	// Если не нашли участников, проверяем, существует ли хотя бы сама команда.
	if len(team.Members) == 0 {
		var exists bool
		err := r.db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM teams WHERE team_name = $1)`, teamName).Scan(&exists)
		if err != nil {
			return domain.Team{}, err
		}
		if !exists {
			return domain.Team{}, app.ErrNotFound
		}
	}

	return team, nil
}

func (r *PgRepository) SetUserActivity(ctx context.Context, userID string, isActive bool) (*domain.User, error) {
	user := &domain.User{}

	// Обновляем статус пользователя и возвращаем обновленную запись.
	err := r.db.QueryRow(ctx,
		`UPDATE users SET is_active = $1 WHERE user_id = $2 RETURNING user_id, username, is_active, team_name`,
		isActive, userID,
	).Scan(&user.ID, &user.Username, &user.IsActive, &user.TeamName)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, app.ErrUserNotFound
		}
		return nil, err
	}
	return user, nil
}

func (r *PgRepository) GetUserByID(ctx context.Context, userID string) (*domain.User, error) {
	user := &domain.User{}

	// Находим пользователя по его ID.
	err := r.db.QueryRow(ctx,
		`SELECT user_id, username, is_active, team_name FROM users WHERE user_id = $1`,
		userID,
	).Scan(&user.ID, &user.Username, &user.IsActive, &user.TeamName)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, app.ErrUserNotFound
		}
		return nil, err
	}
	return user, nil
}

func (r *PgRepository) CreatePullRequest(ctx context.Context, pr domain.PullRequest) (*domain.PullRequest, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	// Создаем основную запись о Pull Request.
	_, err = tx.Exec(ctx,
		`INSERT INTO pull_requests (pull_request_id, pull_request_name, author_id, status, created_at)
         VALUES ($1, $2, $3, $4, $5)`,
		pr.ID, pr.Name, pr.AuthorID, pr.Status, pr.CreatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == uniqueViolationCode {
			return nil, app.ErrPRExists
		}
		return nil, err
	}

	// Привязываем ревьюеров к созданному PR.
	for _, reviewerID := range pr.AssignedReviewers {
		_, err := tx.Exec(ctx, `INSERT INTO pr_reviewers (pr_id, reviewer_id) VALUES ($1, $2)`, pr.ID, reviewerID)
		if err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return &pr, nil
}

func (r *PgRepository) GetPullRequestByID(ctx context.Context, prID string) (*domain.PullRequest, error) {
	pr := &domain.PullRequest{}

	// Получаем основную информацию о PR.
	err := r.db.QueryRow(ctx,
		`SELECT pull_request_id, pull_request_name, author_id, status, created_at, merged_at
		 FROM pull_requests WHERE pull_request_id = $1`,
		prID).Scan(&pr.ID, &pr.Name, &pr.AuthorID, &pr.Status, &pr.CreatedAt, &pr.MergedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, app.ErrPRNotFound
		}
		return nil, err
	}

	// Получаем список ID ревьюеров для этого PR.
	rows, err := r.db.Query(ctx, `SELECT reviewer_id FROM pr_reviewers WHERE pr_id = $1`, prID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	pr.AssignedReviewers = []string{}
	for rows.Next() {
		var reviewerID string
		if err := rows.Scan(&reviewerID); err != nil {
			return nil, err
		}
		pr.AssignedReviewers = append(pr.AssignedReviewers, reviewerID)
	}
	return pr, nil
}

func (r *PgRepository) MergePullRequest(ctx context.Context, prID string) (*domain.PullRequest, error) {
	// Обновляем статус PR и время слияния.
	_, err := r.db.Exec(ctx,
		`UPDATE pull_requests SET status = $1, merged_at = NOW() WHERE pull_request_id = $2`,
		domain.StatusMerged, prID)
	if err != nil {
		return nil, err
	}
	return r.GetPullRequestByID(ctx, prID)
}

func (r *PgRepository) UpdatePullRequestReviewers(ctx context.Context, prID string, reviewers []string) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Удаляем всех текущих ревьюеров.
	if _, err := tx.Exec(ctx, `DELETE FROM pr_reviewers WHERE pr_id = $1`, prID); err != nil {
		return err
	}

	// Вставляем новый список ревьюеров.
	for _, reviewerID := range reviewers {
		_, err := tx.Exec(ctx, `INSERT INTO pr_reviewers (pr_id, reviewer_id) VALUES ($1, $2)`, prID, reviewerID)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (r *PgRepository) GetOpenPullRequestsByReviewer(ctx context.Context, userID string) ([]*domain.PullRequest, error) {
	// Находим ID всех открытых PR, назначенных этому пользователю.
	rows, err := r.db.Query(ctx,
		`SELECT pr.pull_request_id
		 FROM pull_requests pr
		 JOIN pr_reviewers r ON pr.pull_request_id = r.pr_id
		 WHERE r.reviewer_id = $1 AND pr.status = 'OPEN'`,
		userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var prIDs []string
	for rows.Next() {
		var prID string
		if err := rows.Scan(&prID); err != nil {
			return nil, err
		}
		prIDs = append(prIDs, prID)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}

	// Для каждого найденного ID получаем полную информацию о PR.
	pullRequests := make([]*domain.PullRequest, 0, len(prIDs))
	for _, id := range prIDs {
		pr, err := r.GetPullRequestByID(ctx, id)
		if err != nil {
			continue
		}
		pullRequests = append(pullRequests, pr)
	}

	return pullRequests, nil
}
