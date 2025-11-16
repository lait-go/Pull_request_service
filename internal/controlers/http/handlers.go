package http

import (
	"encoding/json"
	"net/http"
	"revivers/internal/domain"
	"revivers/internal/usecase"
	"revivers/pkg/logger"
	"revivers/pkg/render"
)

// Controller реализует ServerInterface и является адаптером между HTTP и usecase слоем
type Controller struct {
	useCase usecase.UseCase
}

// NewController создает новый HTTP контроллер
func NewController(useCase usecase.UseCase) *Controller {
	return &Controller{
		useCase: useCase,
	}
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

// PostPullRequestCreate создает PR и автоматически назначает до 2 ревьюверов из команды автора
// (POST /pullRequest/create)
func (c *Controller) PostPullRequestCreate(w http.ResponseWriter, r *http.Request) {
	var req domain.PostPullRequestCreateJSONBody
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		renderError(w, newErrorResponse(domain.NOTFOUND, "invalid request body: "+err.Error()), http.StatusBadRequest)
		return
	}

	pr, err := c.useCase.CreatePullRequest(r.Context(), req)
	if err != nil {
		handleUseCaseError(w, r, err)
		return
	}

	render.JSON(w, pr, http.StatusCreated)
}

// PostPullRequestMerge помечает PR как MERGED (идемпотентная операция)
// (POST /pullRequest/merge)
func (c *Controller) PostPullRequestMerge(w http.ResponseWriter, r *http.Request) {
	var req domain.PostPullRequestMergeJSONBody
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		renderError(w, newErrorResponse(domain.NOTFOUND, "invalid request body: "+err.Error()), http.StatusBadRequest)
		return
	}

	if err := c.useCase.MergePullRequest(r.Context(), req); err != nil {
		handleUseCaseError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// PostPullRequestReassign переназначает конкретного ревьювера на другого из его команды
// (POST /pullRequest/reassign)
func (c *Controller) PostPullRequestReassign(w http.ResponseWriter, r *http.Request) {
	var req domain.PostPullRequestReassignJSONBody
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		renderError(w, newErrorResponse(domain.NOTFOUND, "invalid request body: "+err.Error()), http.StatusBadRequest)
		return
	}

	if err := c.useCase.ReassignReviewer(r.Context(), req); err != nil {
		handleUseCaseError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// PostTeamAdd создает команду с участниками (создаёт/обновляет пользователей)
// (POST /team/add)
func (c *Controller) PostTeamAdd(w http.ResponseWriter, r *http.Request) {
	var team domain.Team
	if err := json.NewDecoder(r.Body).Decode(&team); err != nil {
		renderError(w, newErrorResponse(domain.NOTFOUND, "invalid request body: "+err.Error()), http.StatusBadRequest)
		return
	}

	logger.Info("PostTeamAdd called with team: ", team.TeamName)

	if err := c.useCase.CreateOrUpdateTeam(r.Context(), team); err != nil {
		handleUseCaseError(w, r, err)
		return
	}

	logger.Info("Team created or updated successfully: ", team.TeamName)

	w.WriteHeader(http.StatusOK)
}

// GetTeamGet возвращает команду с участниками
// (GET /team/get)
func (c *Controller) GetTeamGet(w http.ResponseWriter, r *http.Request, params domain.GetTeamGetParams) {
	team, err := c.useCase.GetTeam(r.Context(), params.TeamName)
	if err != nil {
		handleUseCaseError(w, r, err)
		return
	}

	render.JSON(w, team, http.StatusOK)
}

// GetUsersGetReview возвращает PR'ы, где пользователь назначен ревьювером
// (GET /users/getReview)
func (c *Controller) GetUsersGetReview(w http.ResponseWriter, r *http.Request, params domain.GetUsersGetReviewParams) {
	prs, err := c.useCase.GetPullRequestsByReviewer(r.Context(), params.UserId)
	if err != nil {
		handleUseCaseError(w, r, err)
		return
	}

	render.JSON(w, prs, http.StatusOK)
}

// PostUsersSetIsActive устанавливает флаг активности пользователя
// (POST /users/setIsActive)
func (c *Controller) PostUsersSetIsActive(w http.ResponseWriter, r *http.Request) {
	var req domain.PostUsersSetIsActiveJSONBody
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		renderError(w, newErrorResponse(domain.NOTFOUND, "invalid request body: "+err.Error()), http.StatusBadRequest)
		return
	}

	if err := c.useCase.SetUserActive(r.Context(), req.UserId, req.IsActive); err != nil {
		handleUseCaseError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}
