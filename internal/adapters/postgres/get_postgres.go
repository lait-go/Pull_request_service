package postgres

import (
	"context"
	"fmt"
	"revivers/internal/domain"
)

func (p *Pool) TeamGet(ctx context.Context, teamName string) (domain.Team, error) {
	rows, err := p.DB.QueryContext(ctx,
		`SELECT u.user_id, u.username, u.is_active, t.team_name
     FROM users u
     JOIN teams t ON u.team_id = t.id
     WHERE t.team_name = $1`,
		teamName,
	)
	if err != nil {
		return domain.Team{}, fmt.Errorf("failed to query team: %w", err)
	}
	defer rows.Close()

	var team domain.Team
	team.TeamName = teamName
	team.Members = make([]domain.TeamMember, 0)

	var hasRows bool
	for rows.Next() {
		hasRows = true
		var member domain.TeamMember
		err := rows.Scan(&member.UserId, &member.Username, &member.IsActive, &team.TeamName)
		if err != nil {
			return domain.Team{}, fmt.Errorf("failed to scan team row: %w", err)
		}
		team.Members = append(team.Members, member)
	}

	if err = rows.Err(); err != nil {
		return domain.Team{}, fmt.Errorf("failed to iterate team rows: %w", err)
	}

	// Если команда не найдена, возвращаем доменную ошибку
	if !hasRows {
		return domain.Team{}, &domain.NotFoundError{
			Resource: "team",
			ID:       teamName,
		}
	}

	return team, nil
}

func (p *Pool) TeamExists(ctx context.Context, teamName string) (bool, error) {
	var exists bool

	err := p.DB.QueryRowContext(ctx,
		`SELECT EXISTS(SELECT 1 FROM teams WHERE team_name = $1)`, teamName,
	).Scan(&exists)

	if err != nil {
		return false, fmt.Errorf("failed to check team existence: %w", err)
	}

	return exists, nil
}
