package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"revivers/internal/domain"
)

// PullRequestGet получает PR по ID
func (p *Pool) PullRequestGet(ctx context.Context, prID string) (*domain.PullRequest, error) {
	// Получаем основную информацию о PR
	var pr domain.PullRequest
	var authorInternalID int
	var createdAt sql.NullTime
	var mergedAt sql.NullTime
	var status string

	var prInternalID int
	err := p.DB.QueryRowContext(ctx,
		`SELECT id, pr_id, pr_name, author_id, status, created_at, merged_at
		 FROM pull_requests
		 WHERE pr_id = $1`,
		prID,
	).Scan(&prInternalID, &pr.PullRequestId, &pr.PullRequestName, &authorInternalID, &status, &createdAt, &mergedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &domain.NotFoundError{
				Resource: "pull_request",
				ID:       prID,
			}
		}
		return nil, fmt.Errorf("failed to get PR: %w", err)
	}

	pr.Status = domain.PullRequestStatus(status)
	if createdAt.Valid {
		pr.CreatedAt = &createdAt.Time
	}
	if mergedAt.Valid {
		pr.MergedAt = &mergedAt.Time
	}

	// Получаем user_id автора
	var authorUserID string
	err = p.DB.QueryRowContext(ctx,
		`SELECT user_id FROM users WHERE id = $1`,
		authorInternalID,
	).Scan(&authorUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get author user_id: %w", err)
	}
	pr.AuthorId = authorUserID

	// Получаем ревьюверов
	rows, err := p.DB.QueryContext(ctx,
		`SELECT u.user_id
		 FROM pull_request_reviewers prr
		 JOIN users u ON prr.user_id = u.id
		 WHERE prr.pr_id = (SELECT id FROM pull_requests WHERE pr_id = $1)`,
		prID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query reviewers: %w", err)
	}
	defer rows.Close()

	pr.AssignedReviewers = make([]string, 0)
	for rows.Next() {
		var reviewerID string
		if err := rows.Scan(&reviewerID); err != nil {
			return nil, fmt.Errorf("failed to scan reviewer: %w", err)
		}
		pr.AssignedReviewers = append(pr.AssignedReviewers, reviewerID)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate reviewers: %w", err)
	}

	return &pr, nil
}

// PullRequestsGetByReviewer получает все PR, где пользователь назначен ревьювером
func (p *Pool) PullRequestsGetByReviewer(ctx context.Context, userID string) ([]domain.PullRequestShort, error) {
	// Получаем internal ID пользователя
	var userInternalID int
	err := p.DB.QueryRowContext(ctx,
		`SELECT id FROM users WHERE user_id = $1`,
		userID,
	).Scan(&userInternalID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &domain.NotFoundError{
				Resource: "user",
				ID:       userID,
			}
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Получаем все PR, где пользователь ревьювер
	rows, err := p.DB.QueryContext(ctx,
		`SELECT pr.pr_id, pr.pr_name, pr.status, u.user_id
		 FROM pull_requests pr
		 JOIN pull_request_reviewers prr ON pr.id = prr.pr_id
		 JOIN users u ON pr.author_id = u.id
		 WHERE prr.user_id = $1
		 ORDER BY pr.created_at DESC`,
		userInternalID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query PRs: %w", err)
	}
	defer rows.Close()

	var prs []domain.PullRequestShort
	for rows.Next() {
		var pr domain.PullRequestShort
		var status string
		if err := rows.Scan(&pr.PullRequestId, &pr.PullRequestName, &status, &pr.AuthorId); err != nil {
			return nil, fmt.Errorf("failed to scan PR: %w", err)
		}
		pr.Status = domain.PullRequestShortStatus(status)
		prs = append(prs, pr)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate PRs: %w", err)
	}

	return prs, nil
}

