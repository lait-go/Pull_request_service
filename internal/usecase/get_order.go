package usecase

import (
	"net/http"
	"revivers/internal/domain"
)

// Получить команду с участниками
// (GET /team/get)
func (_ Unimplemented) GetTeamGet(w http.ResponseWriter, r *http.Request, params domain.GetTeamGetParams) {
	w.WriteHeader(http.StatusNotImplemented)
}

// Получить PR'ы, где пользователь назначен ревьювером
// (GET /users/getReview)
func (_ Unimplemented) GetUsersGetReview(w http.ResponseWriter, r *http.Request, params domain.GetUsersGetReviewParams) {
	w.WriteHeader(http.StatusNotImplemented)
}

// Создать PR и автоматически назначить до 2 ревьюверов из команды автора
// (POST /pullRequest/create)
func (_ Unimplemented) PostPullRequestCreate(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}

// Пометить PR как MERGED (идемпотентная операция)
// (POST /pullRequest/merge)
func (_ Unimplemented) PostPullRequestMerge(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}

// Переназначить конкретного ревьювера на другого из его команды
// (POST /pullRequest/reassign)
func (_ Unimplemented) PostPullRequestReassign(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}

// Создать команду с участниками (создаёт/обновляет пользователей)
// (POST /team/add)
func (_ Unimplemented) PostTeamAdd(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}

// Установить флаг активности пользователя
// (POST /users/setIsActive)
func (_ Unimplemented) PostUsersSetIsActive(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}
