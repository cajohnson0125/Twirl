package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/cajohnson0125/Twirl/internal/agent"
	"github.com/cajohnson0125/Twirl/internal/config"
	"github.com/cajohnson0125/Twirl/internal/orchestrator"
	"github.com/cajohnson0125/Twirl/internal/pubsub"
	"github.com/cajohnson0125/Twirl/internal/state"
	"github.com/cajohnson0125/Twirl/internal/tui"
	"github.com/cajohnson0125/Twirl/internal/workflow"
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

			g := workflow.DefaultGraph()
			if err := workflow.Validate(g); err != nil {
				log.Error("graph invalid", "err", err)
				return err
			}

			regs := agent.NewRegistry()
			registerAgents(regs)

			bus := pubsub.NewBus(64)
			hitlCh := make(chan state.HITLResponse, 8)

			projectDir, err := os.Getwd()
			if err != nil {
				return err
			}
			store := state.NewStore(projectDir)

			ctx, cancel := context.WithCancel(
				context.Background(),
			)
			defer cancel()

			e := orchestrator.NewEngine(
				"twirl-project",
				g, store, regs, bus, hitlCh,
			)

			// Start engine in background; send result to TUI.
			engineDone := make(chan error, 1)
			go func() {
				err := e.Run(ctx)
				if err != nil {
					log.Error("engine", "err", err)
				}
				engineDone <- err
			}()

			// Forward OS signals to engine cancel.
			sigCh := make(chan os.Signal, 1)
			signal.Notify(sigCh,
				syscall.SIGINT, syscall.SIGTERM)
			go func() {
				<-sigCh
				cancel()
			}()

			return tui.Run(tui.Opts{
				CursorShape: cfg.Cursor,
				CursorBlink: cfg.Blink,
				LLM:         cfg.LLM,
				Bus:         bus,
				HITLOut:     hitlCh,
				Engine:      &engineCancel{cancel: cancel},
				EngineDone:  engineDone,
			})
		},
	}

	if err := fang.Execute(context.Background(), root); err != nil {
		log.Error("fatal", "err", err)
		os.Exit(1)
	}
}

// engineCancel implements tui.engineController.
type engineCancel struct {
	cancel context.CancelFunc
}

func (e *engineCancel) Cancel() { e.cancel() }

// registerAgents registers stub agents for all roles.
// Real agents wrapping llm.Client will replace these
// as they are implemented.
func registerAgents(regs *agent.Registry) {
	roles := []agent.Role{
		agent.Brainstorm, agent.Research, agent.Report,
		agent.Plan, agent.PlanReview, agent.Execution,
		agent.CodeReview, agent.Triage, agent.Assessment,
		agent.Scribe,
	}
	for _, r := range roles {
		regs.Register(r, func() agent.Agent {
			return agent.NewStubAgent(r,
				&state.Result{
					Status: state.StatusCompleted,
				},
			)
		})
	}
}
