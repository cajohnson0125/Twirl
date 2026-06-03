package main

import (
	"context"
	"os"

	"github.com/cajohnson0125/Twirl/internal/config"
	"github.com/cajohnson0125/Twirl/internal/tui"
	"github.com/charmbracelet/fang"
	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
)

func main() {
	root := &cobra.Command{
		Use:   "twirl",
		Short: "AI-assisted development orchestrator",
		Long: "Twirl orchestrates specialized AI agents " +
			"through a non-linear workflow " +
			"with human-in-the-loop gates.",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := config.Load()
			return tui.Run(cfg.Cursor, cfg.Blink)
		},
	}

	if err := fang.Execute(context.Background(), root); err != nil {
		log.Error("fatal", "err", err)
		os.Exit(1)
	}
}
