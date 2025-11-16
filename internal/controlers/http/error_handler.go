package http

import (
	"database/sql"
	"errors"
	"net/http"
	"revivers/internal/domain"
	"revivers/pkg/logger"
	"revivers/pkg/render"
)

// newErrorResponse создает новый ErrorResponse
func newErrorResponse(code domain.ErrorResponseErrorCode, message string) domain.ErrorResponse {
	var errResp domain.ErrorResponse
	errResp.Error.Code = code
	errResp.Error.Message = message
	return errResp
}

// handleUseCaseError обрабатывает ошибки из usecase слоя
// Маппит доменные ошибки на соответствующие HTTP статусы и коды ошибок
func handleUseCaseError(w http.ResponseWriter, r *http.Request, err error) {
	var errResp domain.ErrorResponse
	var status int

	// Логируем ошибку с request ID для трейсинга
	requestID := GetRequestID(r.Context())
	logger.Error("usecase error", err,
		"request_id", requestID,
		"path", r.URL.Path,
		"method", r.Method,
	)

	switch e := err.(type) {
	// Ошибки валидации параметров (400 Bad Request)
	// Все ошибки валидации параметров обрабатываются одинаково
	case *domain.RequiredParamError,
		*domain.InvalidParamFormatError,
		*domain.RequiredHeaderError,
		*domain.UnescapedCookieParamError,
		*domain.UnmarshalingParamError,
		*domain.TooManyValuesForParamError:

		errResp.Error.Code = domain.NOTFOUND
		errResp.Error.Message = e.Error()
		status = http.StatusBadRequest

	// Бизнес-ошибки (404 Not Found)
	case *domain.NotFoundError:
		errResp.Error.Code = domain.NOTFOUND
		errResp.Error.Message = e.Error()
		status = http.StatusNotFound

	// Бизнес-ошибки (409 Conflict)
	case *domain.TeamAlreadyExistsError:
		errResp.Error.Code = domain.TEAMEXISTS
		errResp.Error.Message = e.Error()
		status = http.StatusConflict

	case *domain.PullRequestExistsError:
		errResp.Error.Code = domain.PREXISTS
		errResp.Error.Message = e.Error()
		status = http.StatusConflict

	// Бизнес-ошибки (400 Bad Request)
	case *domain.PullRequestMergedError:
		errResp.Error.Code = domain.PRMERGED
		errResp.Error.Message = e.Error()
		status = http.StatusBadRequest

	case *domain.NotAssignedError:
		errResp.Error.Code = domain.NOTASSIGNED
		errResp.Error.Message = e.Error()
		status = http.StatusBadRequest

	case *domain.NoCandidateError:
		errResp.Error.Code = domain.NOCANDIDATE
		errResp.Error.Message = e.Error()
		status = http.StatusBadRequest

	default:
		// Для неизвестных ошибок проверяем стандартные ошибки
		// Проверяем обернутые ошибки через errors.Is
		if errors.Is(err, sql.ErrNoRows) {
			errResp.Error.Code = domain.NOTFOUND
			errResp.Error.Message = "resource not found"
			status = http.StatusNotFound
		} else {
			// Для всех остальных ошибок возвращаем внутреннюю ошибку сервера
			// Не возвращаем детали ошибки клиенту для безопасности
			errResp.Error.Code = domain.NOTFOUND
			errResp.Error.Message = "internal server error"
			status = http.StatusInternalServerError
		}
	}

	renderError(w, errResp, status)
}

// renderError отправляет ошибку в формате ErrorResponse
func renderError(w http.ResponseWriter, errResp domain.ErrorResponse, status int) {
	render.JSON(w, errResp, status)
}
