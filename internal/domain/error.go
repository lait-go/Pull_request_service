package domain

import "fmt"

const (
	NOCANDIDATE ErrorResponseErrorCode = "NO_CANDIDATE"
	NOTASSIGNED ErrorResponseErrorCode = "NOT_ASSIGNED"
	NOTFOUND    ErrorResponseErrorCode = "NOT_FOUND"
	PREXISTS    ErrorResponseErrorCode = "PR_EXISTS"
	PRMERGED    ErrorResponseErrorCode = "PR_MERGED"
	TEAMEXISTS  ErrorResponseErrorCode = "TEAM_EXISTS"
)

// Defines values for PullRequestStatus.
const (
	PullRequestStatusMERGED PullRequestStatus = "MERGED"
	PullRequestStatusOPEN   PullRequestStatus = "OPEN"
)

// Defines values for PullRequestShortStatus.
const (
	PullRequestShortStatusMERGED PullRequestShortStatus = "MERGED"
	PullRequestShortStatusOPEN   PullRequestShortStatus = "OPEN"
)

// ErrorResponse defines model for ErrorResponse.
type ErrorResponse struct {
	Error struct {
		Code    ErrorResponseErrorCode `json:"code"`
		Message string                 `json:"message"`
	} `json:"error"`
}

// ErrorResponseErrorCode defines model for ErrorResponse.Error.Code.
type ErrorResponseErrorCode string

// UnescapedCookieParamError — ошибка при декодировании (unescape) cookie-параметра.
type UnescapedCookieParamError struct {
	ParamName string
	Err       error
}

func (e *UnescapedCookieParamError) Error() string {
	return fmt.Sprintf("error unescaping cookie parameter '%s'", e.ParamName)
}

func (e *UnescapedCookieParamError) Unwrap() error {
	return e.Err
}

type UnmarshalingParamError struct {
	ParamName string
	Err       error
}

// UnmarshalingParamError — ошибка десериализации параметра из JSON.

func (e *UnmarshalingParamError) Error() string {
	return fmt.Sprintf("Error unmarshaling parameter %s as JSON: %s", e.ParamName, e.Err.Error())
}

func (e *UnmarshalingParamError) Unwrap() error {
	return e.Err
}

type RequiredParamError struct {
	ParamName string
}

// RequiredParamError — отсутствует обязательный query-параметр.

func (e *RequiredParamError) Error() string {
	return fmt.Sprintf("Query argument %s is required, but not found", e.ParamName)
}

type RequiredHeaderError struct {
	ParamName string
	Err       error
}

// RequiredHeaderError — отсутствует обязательный заголовок запроса.

func (e *RequiredHeaderError) Error() string {
	return fmt.Sprintf("Header parameter %s is required, but not found", e.ParamName)
}

func (e *RequiredHeaderError) Unwrap() error {
	return e.Err
}

type InvalidParamFormatError struct {
	ParamName string
	Err       error
}

// InvalidParamFormatError — неверный формат параметра.

func (e *InvalidParamFormatError) Error() string {
	return fmt.Sprintf("Invalid format for parameter %s: %s", e.ParamName, e.Err.Error())
}

func (e *InvalidParamFormatError) Unwrap() error {
	return e.Err
}

type TooManyValuesForParamError struct {
	ParamName string
	Count     int
}

// TooManyValuesForParamError — передано несколько значений для параметра, ожидается одно.

func (e *TooManyValuesForParamError) Error() string {
	return fmt.Sprintf("Expected one value for %s, got %d", e.ParamName, e.Count)
}

// Бизнес-ошибки

// NotFoundError — ресурс не найден
type NotFoundError struct {
	Resource string
	ID       string
}

func (e *NotFoundError) Error() string {
	if e.ID != "" {
		return fmt.Sprintf("%s with id %s not found", e.Resource, e.ID)
	}
	return fmt.Sprintf("%s not found", e.Resource)
}

// TeamAlreadyExistsError — команда уже существует
type TeamAlreadyExistsError struct {
	TeamName string
}

func (e *TeamAlreadyExistsError) Error() string {
	return fmt.Sprintf("team already exists: %s", e.TeamName)
}

// PullRequestExistsError — Pull Request уже существует
type PullRequestExistsError struct {
	PullRequestID string
}

func (e *PullRequestExistsError) Error() string {
	return fmt.Sprintf("pull request already exists: %s", e.PullRequestID)
}

// PullRequestMergedError — Pull Request уже смержен
type PullRequestMergedError struct {
	PullRequestID string
}

func (e *PullRequestMergedError) Error() string {
	return fmt.Sprintf("pull request already merged: %s", e.PullRequestID)
}

// NotAssignedError — ревьювер не назначен
type NotAssignedError struct {
	PullRequestID string
	UserID        string
}

func (e *NotAssignedError) Error() string {
	if e.UserID != "" {
		return fmt.Sprintf("user %s is not assigned as reviewer for pull request %s", e.UserID, e.PullRequestID)
	}
	return fmt.Sprintf("reviewer not assigned for pull request %s", e.PullRequestID)
}

// NoCandidateError — нет доступного кандидата для назначения
type NoCandidateError struct {
	TeamName string
}

func (e *NoCandidateError) Error() string {
	if e.TeamName != "" {
		return fmt.Sprintf("no active candidate found in team %s", e.TeamName)
	}
	return "no active candidate found"
}
