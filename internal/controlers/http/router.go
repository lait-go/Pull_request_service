package http

import (
	"net/http"

	"github.com/go-chi/chi"
)

type ChiServerOptions struct {
	BaseURL          string
	BaseRouter       chi.Router
	Middlewares      []MiddlewareFunc
	ErrorHandlerFunc func(w http.ResponseWriter, r *http.Request, err error)
}

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

// изначально была функция HandlerFromMuxWithBaseURL, пришлось переделать под свои нужды
func HandlerWithBaseURL(si ServerInterface, baseURL string) http.Handler {
	return HandlerWithOptions(si, ChiServerOptions{
		BaseURL:    baseURL,
	})
}

// HandlerWithOptions creates http.Handler with additional options
func HandlerWithOptions(si ServerInterface, options ChiServerOptions) http.Handler {
	r := options.BaseRouter

	if r == nil {
		r = chi.NewRouter()
	}
	if options.ErrorHandlerFunc == nil {
		options.ErrorHandlerFunc = func(w http.ResponseWriter, r *http.Request, err error) {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
	}
	wrapper := ServerInterfaceWrapper{
		Handler:            si,
		HandlerMiddlewares: options.Middlewares,
		ErrorHandlerFunc:   options.ErrorHandlerFunc,
	}

	r.Group(func(r chi.Router) {
		r.Post(options.BaseURL+"/pullRequest/create", wrapper.PostPullRequestCreate)
	})
	r.Group(func(r chi.Router) {
		r.Post(options.BaseURL+"/pullRequest/merge", wrapper.PostPullRequestMerge)
	})
	r.Group(func(r chi.Router) {
		r.Post(options.BaseURL+"/pullRequest/reassign", wrapper.PostPullRequestReassign)
	})
	r.Group(func(r chi.Router) {
		r.Post(options.BaseURL+"/team/add", wrapper.PostTeamAdd)
	})
	r.Group(func(r chi.Router) {
		r.Get(options.BaseURL+"/team/get", wrapper.GetTeamGet)
	})
	r.Group(func(r chi.Router) {
		r.Get(options.BaseURL+"/users/getReview", wrapper.GetUsersGetReview)
	})
	r.Group(func(r chi.Router) {
		r.Post(options.BaseURL+"/users/setIsActive", wrapper.PostUsersSetIsActive)
	})

	return r
}
