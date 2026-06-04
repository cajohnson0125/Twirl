package main

import (
	"context"
	"os"

	"github.com/cajohnson0125/Twirl/internal/config"
	"github.com/cajohnson0125/Twirl/internal/engine"
	"github.com/cajohnson0125/Twirl/internal/tui"
	"github.com/charmbracelet/fang"
	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
)

func main() {
	var debug bool

	root := &cobra.Command{
		Use:   "twirl",
		Short: "AI-assisted development orchestrator",
		Long: "Twirl orchestrates specialized AI agents " +
			"through a non-linear workflow " +
			"with human-in-the-loop gates.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if debug {
				log.SetLevel(log.DebugLevel)
				log.SetReportCaller(true)
			}
			cfg := config.Load()
			eng := engine.New()

			if !cfg.LLM.IsZero() {
				apiKey, err := cfg.LLM.ResolveAPIKey()
				if err != nil {
					log.Warn("LLM API key not resolved",
						"err", err)
				} else {
					fcfg := engine.FantasyConfig{
						BaseURL: cfg.LLM.BaseURL,
						APIKey:  apiKey,
						Model:   cfg.LLM.Model,
					}
					if err := eng.Configure(fcfg); err != nil {
						log.Warn("LLM provider setup failed",
							"err", err)
					}
				}
			}

			go eng.Start(context.Background())
			defer eng.Stop()
			return tui.Run(eng, cfg.Cursor, cfg.Blink)
		},
	}

	root.Flags().BoolVar(&debug, "debug", false,
		"enable verbose logging with source locations")

	if err := fang.Execute(context.Background(), root); err != nil {
		log.Error("fatal", "err", err)
		os.Exit(1)
	}
}
