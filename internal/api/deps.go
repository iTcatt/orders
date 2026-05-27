package api

import "net/http"

type productHandler interface {
	Get(w http.ResponseWriter, r *http.Request)
	GetByID(w http.ResponseWriter, r *http.Request)
	Create(w http.ResponseWriter, r *http.Request)
	Update(w http.ResponseWriter, r *http.Request)
	Delete(w http.ResponseWriter, r *http.Request)
}

type imageHandler interface {
	Upload(w http.ResponseWriter, r *http.Request)
	Delete(w http.ResponseWriter, r *http.Request)
	Reorder(w http.ResponseWriter, r *http.Request)
}

type categoryHandler interface {
	Get(w http.ResponseWriter, r *http.Request)
}
