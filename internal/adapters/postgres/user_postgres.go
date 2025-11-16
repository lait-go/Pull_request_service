package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"revivers/internal/domain"
)

// UserSetActive устанавливает флаг активности пользователя
func (p *Pool) UserSetActive(ctx context.Context, userID string, isActive bool) error {
	res, err := p.DB.ExecContext(ctx,
		`UPDATE users SET is_active = $1 WHERE user_id = $2`,
		isActive, userID,
	)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if affected == 0 {
		return &domain.NotFoundError{
			Resource: "user",
			ID:       userID,
		}
	}

	return nil
}

// UserGetByID получает пользователя по user_id и возвращает его internal ID и teamID
func (p *Pool) UserGetByID(ctx context.Context, userID string) (int, int, error) {
	var internalID int
	var teamID sql.NullInt64

	err := p.DB.QueryRowContext(ctx,
		`SELECT id, team_id
		 FROM users
		 WHERE user_id = $1`,
		userID,
	).Scan(&internalID, &teamID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, 0, &domain.NotFoundError{
				Resource: "user",
				ID:       userID,
			}
		}
		return 0, 0, fmt.Errorf("failed to get user: %w", err)
	}

	teamIDInt := 0
	if teamID.Valid {
		teamIDInt = int(teamID.Int64)
	}

	return internalID, teamIDInt, nil
}

// ActiveUsersGetByTeamID получает активных пользователей команды, исключая указанного пользователя
func (p *Pool) ActiveUsersGetByTeamID(ctx context.Context, teamID int, excludeUserID int) ([]struct{ UserID string; InternalID int }, error) {
	rows, err := p.DB.QueryContext(ctx,
		`SELECT user_id, id
		 FROM users
		 WHERE team_id = $1 AND is_active = true AND id != $2
		 ORDER BY user_id`,
		teamID, excludeUserID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query active users: %w", err)
	}
	defer rows.Close()

	var users []struct{ UserID string; InternalID int }
	for rows.Next() {
		var user struct {
			UserID     string
			InternalID int
		}
		if err := rows.Scan(&user.UserID, &user.InternalID); err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate users: %w", err)
	}

	return users, nil
}

// UserGet получает пользователя по user_id с информацией о команде
func (p *Pool) UserGet(ctx context.Context, userID string) (*domain.User, error) {
	var user domain.User
	var teamName sql.NullString

	err := p.DB.QueryRowContext(ctx,
		`SELECT u.user_id, u.username, u.is_active, t.team_name
		 FROM users u
		 LEFT JOIN teams t ON u.team_id = t.id
		 WHERE u.user_id = $1`,
		userID,
	).Scan(&user.UserId, &user.Username, &user.IsActive, &teamName)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &domain.NotFoundError{
				Resource: "user",
				ID:       userID,
			}
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if teamName.Valid {
		user.TeamName = teamName.String
	} else {
		user.TeamName = ""
	}

	return &user, nil
}

