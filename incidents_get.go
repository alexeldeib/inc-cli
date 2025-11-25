package main

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/alexeldeib/incli/client"
	kitlog "github.com/go-kit/log"
	"github.com/spf13/cobra"
)

type GetIncidentOptions struct {
	incidentReference int
	incidentID        string
	watchlive         bool
}

func (o *GetIncidentOptions) Run(ctx context.Context, logger kitlog.Logger, cl *client.ClientWithResponses) error {
	if o.incidentReference != -1 && o.incidentID != "" {
		return fmt.Errorf("only one of --id or --ref may be specified")
	}

	if o.incidentReference != -1 && o.incidentReference < 0 {
		return fmt.Errorf("incident --ref must be positive integer: %q", o.incidentReference)
	}

	if o.incidentReference > 0 {
		incident, err := ShowIncidentByReference(ctx, logger, cl, o.incidentReference)
		if err != nil {
			return fmt.Errorf("failed to list incidents: %s", err)
		}

		if err := serialize(incident); err != nil {
			return fmt.Errorf("failed to marshal json: %s", err)
		}

		return nil
	}

	if o.incidentID != "" {
		incident, err := ShowIncidentByID(ctx, logger, cl, o.incidentID)
		if err != nil {
			return fmt.Errorf("failed to list incidents: %s", err)
		}

		if err := serialize(incident); err != nil {
			return fmt.Errorf("failed to marshal json: %s", err)
		}

		return nil
	}

	if o.watchlive {
		for {
			incidents, err := ListAllIncidents(ctx, logger, cl)
			if err != nil {
				return fmt.Errorf("failed to list incidents: %s", err)
			}
			wincs := make([]client.IncidentV2, 0)
			for i, inc := range incidents {
				if inc.IncidentStatus.Name != "Closed" {
					wincs = append(wincs, incidents[i])
				}
			}
			// Sort by incident name
			sort.Slice(wincs, func(i, j int) bool {
				return wincs[i].Name < wincs[j].Name
			})
			// Clear terminal screen
			fmt.Print("\033[2J\033[H")
			println(time.Now().Format("2006-01-02 15:04:05"), "live incidents", len(wincs))
			org := ""
			if len(wincs) > 0 {
				org = strings.Split(*wincs[0].Permalink, "/")[3]
			}
			for _, inc := range wincs {
				println(
					inc.Reference,
					"https://app.incident.io/"+org+"/incidents/"+strings.ReplaceAll(inc.Reference, "INC-", ""),
					"...", inc.Name,
					"status", inc.IncidentStatus.Name,
				)
			}
			wincs = nil
			time.Sleep(3 * time.Second)
		}
	} else {
		incidents, err := ListAllIncidents(ctx, logger, cl)
		if err != nil {
			return fmt.Errorf("failed to list incidents: %s", err)
		}

		if err := serialize(incidents); err != nil {
			return fmt.Errorf("failed to marshal json: %s", err)
		}

	}

	return nil
}

func NewGetIncidentCommand() *cobra.Command {
	opts := &GetIncidentOptions{}
	cmd := &cobra.Command{
		Use:     "get",
		Short:   "get one or all incidents",
		Aliases: []string{"list"},
		Run: func(cmd *cobra.Command, args []string) {
			ctx, logger, cl, err := setup()
			if err != nil {
				fmt.Printf("failed to setup: %s", err)
				os.Exit(1)
			}

			if err := opts.Run(ctx, logger, cl); err != nil {
				logger.Log("msg", "failed to run", "error", err)
				os.Exit(1)
			}
		},
	}

	cmd.Flags().StringVar(&opts.incidentID, "id", "", "incident ID, e.g. 01HE6...")
	cmd.Flags().IntVar(&opts.incidentReference, "ref", -1, "incident reference number, e.g. 27 for INC-27")
	cmd.Flags().BoolVar(&opts.watchlive, "watch", false, "specify watching for non-closed incs")

	return cmd
}
