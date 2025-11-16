package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"revivers/internal/domain"

	"github.com/lib/pq"
)

// PullRequestCreate создает PR и назначает ревьюверов
func (p *Pool) PullRequestCreate(ctx context.Context, prID, prName, authorUserID string, reviewerUserIDs []string) (*domain.PullRequest, error) {
	tx, err := p.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Получаем internal ID автора и его команду
	var authorInternalID int
	var teamID sql.NullInt64
	err = tx.QueryRowContext(ctx,
		`SELECT id, team_id FROM users WHERE user_id = $1`,
		authorUserID,
	).Scan(&authorInternalID, &teamID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &domain.NotFoundError{
				Resource: "user",
				ID:       authorUserID,
			}
		}
		return nil, fmt.Errorf("failed to get author: %w", err)
	}

	// Проверяем, существует ли уже PR с таким ID
	var existingPRID int
	err = tx.QueryRowContext(ctx,
		`SELECT id FROM pull_requests WHERE pr_id = $1`,
		prID,
	).Scan(&existingPRID)
	if err == nil {
		// PR уже существует
		return nil, &domain.PullRequestExistsError{
			PullRequestID: prID,
		}
	} else if !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("failed to check PR existence: %w", err)
	}

	// Создаем PR
	var prInternalID int
	var createdAt time.Time
	err = tx.QueryRowContext(ctx,
		`INSERT INTO pull_requests (pr_id, pr_name, author_id, status)
		 VALUES ($1, $2, $3, 'OPEN')
		 RETURNING id, created_at`,
		prID, prName, authorInternalID,
	).Scan(&prInternalID, &createdAt)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" { // unique_violation
			return nil, &domain.PullRequestExistsError{
				PullRequestID: prID,
			}
		}
		return nil, fmt.Errorf("failed to create PR: %w", err)
	}

	// Назначаем ревьюверов
	for _, reviewerUserID := range reviewerUserIDs {
		var reviewerInternalID int
		err = tx.QueryRowContext(ctx,
			`SELECT id FROM users WHERE user_id = $1`,
			reviewerUserID,
		).Scan(&reviewerInternalID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, &domain.NotFoundError{
					Resource: "user",
					ID:       reviewerUserID,
				}
			}
			return nil, fmt.Errorf("failed to get reviewer %s: %w", reviewerUserID, err)
		}

		_, err = tx.ExecContext(ctx,
			`INSERT INTO pull_request_reviewers (pr_id, user_id)
			 VALUES ($1, $2)`,
			prInternalID, reviewerInternalID,
		)
		if err != nil {
			var pqErr *pq.Error
			if errors.As(err, &pqErr) {
				if pqErr.Code == "23505" { // unique_violation
					// Ревьювер уже назначен, пропускаем
					continue
				}
				if pqErr.Code == "P0001" { // trigger exception (reviewer limit)
					return nil, fmt.Errorf("reviewer limit exceeded: %w", err)
				}
			}
			return nil, fmt.Errorf("failed to assign reviewer %s: %w", reviewerUserID, err)
		}
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Получаем созданный PR с ревьюверами
	return p.PullRequestGet(ctx, prID)
}

// PullRequestMerge помечает PR как MERGED (идемпотентная операция)
func (p *Pool) PullRequestMerge(ctx context.Context, prID string) error {
	tx, err := p.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Проверяем существование PR и его текущий статус
	var status string
	var mergedAt sql.NullTime
	err = tx.QueryRowContext(ctx,
		`SELECT status, merged_at FROM pull_requests WHERE pr_id = $1`,
		prID,
	).Scan(&status, &mergedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &domain.NotFoundError{
				Resource: "pull_request",
				ID:       prID,
			}
		}
		return fmt.Errorf("failed to get PR: %w", err)
	}

	// Если уже MERGED, операция идемпотентна - ничего не делаем
	if status == "MERGED" {
		return nil
	}

	// Обновляем статус на MERGED
	_, err = tx.ExecContext(ctx,
		`UPDATE pull_requests
		 SET status = 'MERGED', merged_at = NOW()
		 WHERE pr_id = $1`,
		prID,
	)
	if err != nil {
		return fmt.Errorf("failed to merge PR: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// PullRequestReassignReviewer переназначает ревьювера на другого из его команды
func (p *Pool) PullRequestReassignReviewer(ctx context.Context, prID, oldUserID, newUserID string) error {
	tx, err := p.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Получаем PR
	var prInternalID int
	var authorInternalID int
	err = tx.QueryRowContext(ctx,
		`SELECT id, author_id FROM pull_requests WHERE pr_id = $1`,
		prID,
	).Scan(&prInternalID, &authorInternalID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &domain.NotFoundError{
				Resource: "pull_request",
				ID:       prID,
			}
		}
		return fmt.Errorf("failed to get PR: %w", err)
	}

	// Проверяем, что PR не MERGED
	var status string
	err = tx.QueryRowContext(ctx,
		`SELECT status FROM pull_requests WHERE pr_id = $1`,
		prID,
	).Scan(&status)
	if err != nil {
		return fmt.Errorf("failed to get PR status: %w", err)
	}
	if status == "MERGED" {
		return &domain.PullRequestMergedError{
			PullRequestID: prID,
		}
	}

	// Получаем internal ID старого ревьювера
	var oldReviewerInternalID int
	var oldReviewerTeamID sql.NullInt64
	err = tx.QueryRowContext(ctx,
		`SELECT id, team_id FROM users WHERE user_id = $1`,
		oldUserID,
	).Scan(&oldReviewerInternalID, &oldReviewerTeamID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &domain.NotFoundError{
				Resource: "user",
				ID:       oldUserID,
			}
		}
		return fmt.Errorf("failed to get old reviewer: %w", err)
	}

	// Проверяем, что старый ревьювер назначен на этот PR
	var count int
	err = tx.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM pull_request_reviewers
		 WHERE pr_id = $1 AND user_id = $2`,
		prInternalID, oldReviewerInternalID,
	).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check reviewer assignment: %w", err)
	}
	if count == 0 {
		return &domain.NotAssignedError{
			PullRequestID: prID,
			UserID:        oldUserID,
		}
	}

	// Получаем internal ID нового ревьювера и проверяем, что он в той же команде
	var newReviewerInternalID int
	var newReviewerTeamID sql.NullInt64
	err = tx.QueryRowContext(ctx,
		`SELECT id, team_id FROM users WHERE user_id = $1 AND is_active = true`,
		newUserID,
	).Scan(&newReviewerInternalID, &newReviewerTeamID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &domain.NotFoundError{
				Resource: "user",
				ID:       newUserID,
			}
		}
		return fmt.Errorf("failed to get new reviewer: %w", err)
	}

	// Проверяем, что новый ревьювер в той же команде, что и старый
	if !oldReviewerTeamID.Valid || !newReviewerTeamID.Valid || oldReviewerTeamID.Int64 != newReviewerTeamID.Int64 {
		return &domain.NoCandidateError{
			TeamName: "different team",
		}
	}

	// Проверяем, что новый ревьювер не является автором PR
	if newReviewerInternalID == authorInternalID {
		return &domain.NoCandidateError{
			TeamName: "author cannot be reviewer",
		}
	}

	// Удаляем старого ревьювера
	_, err = tx.ExecContext(ctx,
		`DELETE FROM pull_request_reviewers
		 WHERE pr_id = $1 AND user_id = $2`,
		prInternalID, oldReviewerInternalID,
	)
	if err != nil {
		return fmt.Errorf("failed to remove old reviewer: %w", err)
	}

	// Добавляем нового ревьювера
	_, err = tx.ExecContext(ctx,
		`INSERT INTO pull_request_reviewers (pr_id, user_id)
		 VALUES ($1, $2)`,
		prInternalID, newReviewerInternalID,
	)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			if pqErr.Code == "23505" { // unique_violation
				return fmt.Errorf("reviewer already assigned: %w", err)
			}
			if pqErr.Code == "P0001" { // trigger exception (reviewer limit)
				return fmt.Errorf("reviewer limit exceeded: %w", err)
			}
		}
		return fmt.Errorf("failed to assign new reviewer: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

