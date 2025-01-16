package server

import (
	"encoding/json"
	"github.com/hansels/coda-payments-routing-api/src/loadbalancer"
	"github.com/hansels/coda-payments-routing-api/src/model"
	"log"

	"net/http"
)

type Opts struct {
	ListenAddress string
}

type Handler struct {
	options     *Opts
	listenErrCh chan error
}

func New(o *Opts) *Handler {
	handler := &Handler{options: o}
	return handler
}

func (h *Handler) Run() {
	log.Printf("Listening on %s", h.options.ListenAddress)

	servers := []string{
		"http://localhost:3000",
		"http://localhost:3001", // Non Active
		"http://localhost:3002",
		"http://localhost:3003", // Non Active
		"http://localhost:3004",
	}

	lb := loadbalancer.NewLoadBalancer(servers)

	http.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		h.RegisterServer(w, r, lb)
	})

	http.HandleFunc("/unregister", func(w http.ResponseWriter, r *http.Request) {
		h.UnregisterServer(w, r, lb)
	})

	http.Handle("/", lb)

	h.listenErrCh <- http.ListenAndServe(h.options.ListenAddress, nil)
}

func (h *Handler) ListenError() <-chan error {
	return h.listenErrCh
}

func (h *Handler) RegisterServer(w http.ResponseWriter, r *http.Request, lb *loadbalancer.LoadBalancer) {
	req := model.ServerRegisterRequest{}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid Request", http.StatusBadRequest)
		return
	}

	lb.RegisterServer(req.URL)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (h *Handler) UnregisterServer(w http.ResponseWriter, r *http.Request, lb *loadbalancer.LoadBalancer) {
	req := model.ServerUnregisterRequest{}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	lb.UnregisterServer(req.URL)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
