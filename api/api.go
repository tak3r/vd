package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/e9ctrl/vd/log"
	"github.com/go-chi/chi/v5"
)

type Device interface {
	GetParameter(param string) (any, error)
	SetParameter(param string, val any) error
	GetGlobDel(typ string) (time.Duration, error)
	SetGlobDel(typ string, val string) error
	GetDel(typ string, param string) (time.Duration, error)
	SetDel(typ string, param string, val string) error
	GetMismatch() []byte
	SetMismatch(string) error
	TrigParam(string) error
}

type api struct {
	d Device
}

func NewAPI(d Device) *api {
	return &api{d: d}
}

func (a *api) routes() http.Handler {
	r := chi.NewRouter()

	r.Route("/", func(r chi.Router) {
		r.Get("/{param}", a.getParameter)
		r.Post("/{param}/{value}", a.setParameter)
		r.Get("/delay/{type}", a.getGlobDel)
		r.Post("/delay/{type}/{value}", a.setGlobDel)
		r.Get("/delay/{type}/{param}", a.getDel)
		r.Post("/delay/{type}/{param}/{value}", a.setDel)
		r.Get("/mismatch", a.getMismatch)
		r.Post("/mismatch/{value}", a.setMismatch)
		r.Post("/trigger/{param}", a.triggerParameter)
	})

	return r
}

func (a *api) getMismatch(w http.ResponseWriter, r *http.Request) {
	value := a.d.GetMismatch()

	log.API("get mismatch")
	w.Header().Set("Content-Type", "text/plain")
	w.Write(value)
}

func (a *api) setMismatch(w http.ResponseWriter, r *http.Request) {
	value := chi.URLParam(r, "value")

	err := a.d.SetMismatch(value)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.API("set mismatch to", value)
	w.Write([]byte("Mismatch set successfully"))
}

func (a *api) getParameter(w http.ResponseWriter, r *http.Request) {
	param := chi.URLParam(r, "param")

	value, err := a.d.GetParameter(param)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error: %s", err)
		return
	}

	log.API("get", param)
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(fmt.Sprintf("%v", value)))
}

func (a *api) setParameter(w http.ResponseWriter, r *http.Request) {
	param := chi.URLParam(r, "param")
	value := chi.URLParam(r, "value")

	err := a.d.SetParameter(param, value)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error: %s", err)
		return
	}

	log.API("set", param, "to", value)
	w.Write([]byte("Parameter set successfully"))
}

func (a *api) getGlobDel(w http.ResponseWriter, r *http.Request) {
	typ := chi.URLParam(r, "type")

	del, err := a.d.GetGlobDel(typ)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error: %s", err)
		return
	}

	log.API("get delay", typ)
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(del.String()))
}

func (a *api) setGlobDel(w http.ResponseWriter, r *http.Request) {
	typ := chi.URLParam(r, "type")
	value := chi.URLParam(r, "value")

	err := a.d.SetGlobDel(typ, value)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error: %s", err)
		return
	}

	log.API("set delay", typ, "to", value)
	w.Write([]byte("Delay set successfully"))
}

func (a *api) getDel(w http.ResponseWriter, r *http.Request) {
	typ := chi.URLParam(r, "type")
	param := chi.URLParam(r, "param")

	del, err := a.d.GetDel(typ, param)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error: %s", err)
		return
	}

	log.API("get delay", typ, "of", param)
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(del.String()))
}

func (a *api) setDel(w http.ResponseWriter, r *http.Request) {
	typ := chi.URLParam(r, "type")
	param := chi.URLParam(r, "param")
	value := chi.URLParam(r, "value")

	err := a.d.SetDel(typ, param, value)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error: %s", err)
		return
	}

	log.API("set delay", typ, "of", param, "to", value)
	w.Write([]byte("Delay set successfully"))
}

func (a *api) triggerParameter(w http.ResponseWriter, r *http.Request) {
	param := chi.URLParam(r, "param")

	err := a.d.TrigParam(param)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error: %s", err)
		return
	}

	log.API("trigger parameter", param)
	w.Write([]byte("Parameter triggered successfully"))
}
