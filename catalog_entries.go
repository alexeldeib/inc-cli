package main

import (
	"context"
	"fmt"
	"os"

	"github.com/alexeldeib/incli/client"
	kitlog "github.com/go-kit/log"
	"github.com/spf13/cobra"
)

func NewGetCatalogEntriesCommand() *cobra.Command {
	opts := &GetCatalogEntriesOptions{}
	cmd := &cobra.Command{
		Use:   "get",
		Short: "get one, many, or all catalog entries, by name/id with or without type name/id",
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

	cmd.Flags().StringVarP(&opts.catalogTypeID, "type-id", "t", "", "catalog type id, e.g. 01HE6...")
	cmd.Flags().StringVar(&opts.catalogTypeName, "type-name", "", "catalog type name, e.g. PagerdutyService")
	cmd.Flags().StringVarP(&opts.catalogEntryName, "name", "n", "", "name or alias of custom catalog entry, e.g. Serving Infra Default")
	cmd.Flags().StringVar(&opts.catalogEntryID, "id", "", "custom field to patch, e.g. --field foo=bar --field baz=qux. --field foo=bar=baz sets field `foo` to `bar=baz`")

	return cmd
}

type GetCatalogEntriesOptions struct {
	catalogTypeName  string
	catalogTypeID    string
	catalogEntryName string
	catalogEntryID   string
}

func (o *GetCatalogEntriesOptions) Run(ctx context.Context, logger kitlog.Logger, cl *client.ClientWithResponses) error {
	if o.catalogTypeName != "" && o.catalogTypeID != "" {
		return fmt.Errorf("exactly one of --type-name or --type-id may be specified")
	}

	if o.catalogEntryName != "" && o.catalogEntryID != "" {
		return fmt.Errorf("exactly one of --entry-name or --entry-id may be specified")
	}

	if o.catalogEntryID != "" && (o.catalogTypeID != "" || o.catalogTypeName != "") {
		return fmt.Errorf("--entry-id is mutually exclusive with both --type-id and --type-name")
	}

	if o.catalogEntryID != "" {
		res, err := FindCatalogEntryByID(ctx, logger, cl, o.catalogEntryID)
		if err != nil {
			return fmt.Errorf("failed to find catalog entry: %s", err)
		}

		if err := serialize(res); err != nil {
			return fmt.Errorf("failed to marshal json: %q", err)
		}

		return nil
	}

	if o.catalogTypeID != "" {
		if o.catalogEntryName != "" {
			res, err := FindCatalogEntryByNameWithTypeID(ctx, logger, cl, o.catalogEntryName, o.catalogTypeID)
			if err != nil {
				return fmt.Errorf("failed to find catalog entry: %s", err)
			}

			if err := serialize(res); err != nil {
				return fmt.Errorf("failed to marshal json: %q", err)
			}

			return nil
		} else {
			res, err := ListAllCatalogEntriesByTypeID(ctx, logger, cl, o.catalogTypeID)
			if err != nil {
				return fmt.Errorf("failed to find catalog entry: %s", err)
			}

			if err := serialize(res); err != nil {
				return fmt.Errorf("failed to marshal json: %q", err)
			}

			return nil
		}
	} else if o.catalogTypeName != "" {
		if o.catalogEntryName != "" {
			res, err := FindCatalogEntryByNameWithTypeName(ctx, logger, cl, o.catalogEntryName, o.catalogTypeName)
			if err != nil {
				return fmt.Errorf("failed to find catalog entry: %s", err)
			}

			if err := serialize(res); err != nil {
				return fmt.Errorf("failed to marshal json: %q", err)
			}
		} else {
			res, err := ListAllCatalogEntriesByTypeName(ctx, logger, cl, o.catalogTypeName)
			if err != nil {
				return fmt.Errorf("failed to find catalog entry: %s", err)
			}

			if err := serialize(res); err != nil {
				return fmt.Errorf("failed to marshal json: %q", err)
			}

			return nil
		}
	} else {
		res, err := ListAllCatalogEntries(ctx, logger, cl)
		if err != nil {
			return fmt.Errorf("failed to list all catalog entries: %s", err)
		}

		if o.catalogEntryName != "" {
			n := 0
			for _, v := range res {
				if v.Name == o.catalogEntryName {
					res[n] = v
					n++
				}
			}
			res = res[:n]
		} else if o.catalogEntryID != "" {
			n := 0
			for _, v := range res {
				if v.Id == o.catalogEntryID {
					res[n] = v
					n++
				}
			}
			res = res[:n]
		}

		if err := serialize(res); err != nil {
			return fmt.Errorf("failed to marshal json: %q", err)
		}
	}
	return nil
}
