package main

import (
	"js-centralized-wallet/pkg/model"
	"js-centralized-wallet/pkg/server"
	"log/slog"
	"os"
)

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			return a
		},
	})))

	model := model.NewModel()

	err := model.Setup()
	if err != nil {
		slog.Error("failed to setup model", "err", err)
		os.Exit(1)
	}
	server := server.NewServer(model)

	err = server.StartScheduler()
	if err != nil {
		slog.Error("failed to start schedule", "err", err)
		os.Exit(1)
	}

	err = server.Run()
	if err != nil {
		slog.Error("failed to run server", "err", err)
		os.Exit(1)
	}
}
