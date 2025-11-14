package http

import (
	"net/http"
	"revivers/internal/domain"

	"github.com/go-chi/chi"
)

// Handler creates http.Handler with routing matching OpenAPI spec.
func Handler(si ServerInterface) http.Handler {
	return HandlerWithOptions(si, ChiServerOptions{})
}

// HandlerFromMux creates http.Handler with routing matching OpenAPI spec based on the provided mux.
func HandlerFromMux(si ServerInterface, r chi.Router) http.Handler {
	return HandlerWithOptions(si, ChiServerOptions{
		BaseRouter: r,
	})
}

func HandlerFromMuxWithBaseURL(si ServerInterface, r chi.Router, baseURL string) http.Handler {
	return HandlerWithOptions(si, ChiServerOptions{
		BaseURL:    baseURL,
		BaseRouter: r,
	})
}

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

type ServerInterfaceWrapper struct {
	Handler            ServerInterface
	HandlerMiddlewares []MiddlewareFunc
	ErrorHandlerFunc   func(w http.ResponseWriter, r *http.Request, err error)
}

type ChiServerOptions struct {
	BaseURL          string
	BaseRouter       chi.Router
	Middlewares      []MiddlewareFunc
	ErrorHandlerFunc func(w http.ResponseWriter, r *http.Request, err error)
}

type MiddlewareFunc func(http.Handler) http.Handler
