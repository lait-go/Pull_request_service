package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"revivers/internal/domain"

	"github.com/lib/pq"
)

func (p *Pool) CreateTeam(ctx context.Context, data domain.Team) error {
	tx, err := p.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	var teamID int
	err = tx.QueryRowContext(ctx,
		`INSERT INTO teams (team_name)
		 VALUES ($1)
		 RETURNING id`,
		data.TeamName,
	).Scan(&teamID)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" { // unique_violation
			return &domain.TeamAlreadyExistsError{TeamName: data.TeamName}
		}
		return fmt.Errorf("failed to create team: %w", err)
	}

	for _, m := range data.Members {
		var userExists bool
		err = tx.QueryRowContext(ctx,
			`SELECT EXISTS(SELECT 1 FROM users WHERE user_id = $1)`,
			m.UserId,
		).Scan(&userExists)
		if err != nil {
			return fmt.Errorf("failed checking user: %w", err)
		}

		if userExists {
			res, err := tx.ExecContext(ctx,
				`UPDATE users
				 SET team_id = $1,
				     username = $2,
				     is_active = $3
				 WHERE user_id = $4`,
				teamID,
				m.Username,
				m.IsActive,
				m.UserId,
			)
			if err != nil {
				return fmt.Errorf("failed to update user %s: %w", m.UserId, err)
			}
			affected, _ := res.RowsAffected()
			if affected == 0 {
				// Пользователь не найден (маловероятно, но возможно при race condition)
				return &domain.NotFoundError{
					Resource: "user",
					ID:       m.UserId,
				}
			}
		} else {
			_, err = tx.ExecContext(ctx,
				`INSERT INTO users (user_id, username, team_id, is_active)
				 VALUES ($1, $2, $3, $4)`,
				m.UserId,
				m.Username,
				teamID,
				m.IsActive,
			)
			if err != nil {
				var pqErr *pq.Error
				if errors.As(err, &pqErr) && pqErr.Code == "23505" { // unique_violation
					return &domain.NotFoundError{
						Resource: "user",
						ID:       m.UserId,
					}
				}
				return fmt.Errorf("failed to insert user %s: %w", m.UserId, err)
			}
		}
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (p *Pool) UpdateTeam(ctx context.Context, data domain.Team) error {
	tx, err := p.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	var teamID int

	err = tx.QueryRowContext(ctx,
		`SELECT id FROM teams WHERE team_name = $1`,
		data.TeamName,
	).Scan(&teamID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &domain.NotFoundError{
				Resource: "team",
				ID:       data.TeamName,
			}
		}
		return fmt.Errorf("failed to get team: %w", err)
	}

	_, err = tx.ExecContext(ctx,
		`UPDATE users
		 SET team_id = NULL
		 WHERE team_id = $1`,
		teamID,
	)
	if err != nil {
		return fmt.Errorf("failed to clear team members: %w", err)
	}

	for _, m := range data.Members {
		var userExists bool
		err = tx.QueryRowContext(ctx,
			`SELECT EXISTS(SELECT 1 FROM users WHERE user_id = $1)`,
			m.UserId,
		).Scan(&userExists)
		if err != nil {
			return fmt.Errorf("failed checking user: %w", err)
		}

		if userExists {
			res, err := tx.ExecContext(ctx,
				`UPDATE users
				 SET team_id = $1,
				     username = $2,
				     is_active = $3
				 WHERE user_id = $4`,
				teamID,
				m.Username,
				m.IsActive,
				m.UserId,
			)
			if err != nil {
				return fmt.Errorf("failed to update user %s: %w", m.UserId, err)
			}
			affected, _ := res.RowsAffected()
			if affected == 0 {
				return &domain.NotFoundError{
					Resource: "user",
					ID:       m.UserId,
				}
			}
		} else {
			// Создаем пользователя, если его нет
			_, err = tx.ExecContext(ctx,
				`INSERT INTO users (user_id, username, team_id, is_active)
				 VALUES ($1, $2, $3, $4)`,
				m.UserId,
				m.Username,
				teamID,
				m.IsActive,
			)
			if err != nil {
				return fmt.Errorf("failed to insert user %s: %w", m.UserId, err)
			}
		}
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
