package http

import (
	"net/http"
	"revivers/internal/domain"
)

// ServerInterface represents all server handlers.
type ServerInterface interface {
	// Создать PR и автоматически назначить до 2 ревьюверов из команды автора
	// (POST /pullRequest/create)
	PostPullRequestCreate(w http.ResponseWriter, r *http.Request)
	// Пометить PR как MERGED (идемпотентная операция)
	// (POST /pullRequest/merge)
	PostPullRequestMerge(w http.ResponseWriter, r *http.Request)
	// Переназначить конкретного ревьювера на другого из его команды
	// (POST /pullRequest/reassign)
	PostPullRequestReassign(w http.ResponseWriter, r *http.Request)
	// Создать команду с участниками (создаёт/обновляет пользователей)
	// (POST /team/add)
	PostTeamAdd(w http.ResponseWriter, r *http.Request)
	// Получить команду с участниками
	// (GET /team/get)
	GetTeamGet(w http.ResponseWriter, r *http.Request, params domain.GetTeamGetParams)
	// Получить PR'ы, где пользователь назначен ревьювером
	// (GET /users/getReview)
	GetUsersGetReview(w http.ResponseWriter, r *http.Request, params domain.GetUsersGetReviewParams)
	// Установить флаг активности пользователя
	// (POST /users/setIsActive)
	PostUsersSetIsActive(w http.ResponseWriter, r *http.Request)
}

