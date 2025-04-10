package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"js-centralized-wallet/pkg/model"
	"js-centralized-wallet/pkg/trace"
	"js-centralized-wallet/pkg/utils/middlewares"
	"net"
	"net/http"
	"os"
)

type Server struct {
	model   *model.Model
	jobChan chan model.TransferJob
}

func NewServer(m *model.Model) *Server {
	jobChan := make(chan model.TransferJob, 100)

	transferPool := model.NewTransferWorkerPool(m, jobChan,
		model.WithNumWorkers(10),
	)

	transferPool.Start()

	return &Server{
		model:   m,
		jobChan: jobChan,
	}
}

func (s *Server) Run() error {
	l, err := net.Listen("tcp", os.Getenv("LISTEN_ADDR"))
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	return http.Serve(l, middlewares.ComposeMiddlewares(
		middlewares.AccessLog,
		middlewares.AllowAllOrigins,
		middlewares.GzipMiddleware,
		s.apiRoutes,
	)(http.NewServeMux().ServeHTTP))
}

func (s *Server) ping(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "pong"}`))
}

func respondJSON(w http.ResponseWriter, r *http.Request, data interface{}) {
	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(data); err != nil {
		respondErr(w, r, err)
	}
}

func respondErr(w http.ResponseWriter, r *http.Request, err error) {
	_, l := trace.Logger(r.Context())

	clientErr := &model.ClientError{}
	if errors.As(err, &clientErr) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)

		_ = json.NewEncoder(w).Encode(struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		}{clientErr.Code, err.Error()})

		l.Warn("client error", "code", clientErr.Code, "err", err)
		return
	}

	http.Error(w, err.Error(), http.StatusInternalServerError)

	l.Error("internal error", "err", err)
}
