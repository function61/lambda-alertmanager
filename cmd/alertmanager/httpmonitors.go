package main

import (
	"context"
	"fmt"
	"time"

	"github.com/function61/eventhorizon/pkg/ehevent"
	"github.com/function61/gokit/strings/stringutils"
	"github.com/function61/lambda-alertmanager/pkg/amdomain"
	"github.com/function61/lambda-alertmanager/pkg/amstate"
	"github.com/scylladb/termtables"
	"github.com/spf13/cobra"
)

func httpMonitorEntry() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "hm",
		Short: "Manage HTTP monitors",
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "ls",
		Short: "List monitors",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return httpMonitorList(cmd.Context())
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "rm [id]",
		Short: "Remove HTTP monitor",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return httpMonitorDelete(cmd.Context(), args[0])
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "enable [id]",
		Short: "Enable disabled HTTP monitor",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return httpMonitorEnableOrDisable(
				cmd.Context(),
				args[0],
				true)
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "disable [id]",
		Short: "Temporarily disable a HTTP monitor",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return httpMonitorEnableOrDisable(
				cmd.Context(),
				args[0],
				false)
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "mk [url] [find]",
		Short: "Create HTTP monitor",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return httpMonitorCreate(
				cmd.Context(),
				args[0],
				args[1])
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "scan",
		Short: "Runs all enabled monitors and raises alerts if appropriate",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			app, err := getApp(cmd.Context())
			if err != nil {
				return err
			}

			return httpMonitorScanAndAlertFailures(cmd.Context(), app)
		},
	})

	return cmd
}

func httpMonitorList(ctx context.Context) error {
	app, err := getApp(ctx)
	if err != nil {
		return err
	}

	view := termtables.CreateTable()
	view.AddHeaders("Id", "Enabled", "Url", "Find")

	for _, alert := range app.State.HTTPMonitors() {
		view.AddRow(
			alert.ID,
			boolToCheckmark(alert.Enabled),
			stringutils.Truncate(alert.URL, 44),
			alert.Find)
	}

	fmt.Println(view.Render())

	return nil
}

func httpMonitorCreate(ctx context.Context, url string, find string) error {
	app, err := getApp(ctx)
	if err != nil {
		return err
	}

	monitorCreated := amdomain.NewHTTPMonitorCreated(
		amstate.NewHTTPMonitorID(),
		true,
		url,
		find,
		ehevent.MetaSystemUser(time.Now()))

	ver := app.State.Version()

	_, err = app.Writer.Append(ctx, ver.Stream(), []string{
		ehevent.Serialize(monitorCreated),
	})
	return err
}

func httpMonitorDelete(ctx context.Context, id string) error {
	app, err := getApp(ctx)
	if err != nil {
		return err
	}

	return app.Reader.TransactWrite(ctx, func() error {
		if amstate.FindHTTPMonitorWithID(id, app.State.HTTPMonitors()) == nil {
			return fmt.Errorf("monitor to delete not found: %s", id)
		}

		return app.AppendAfter(ctx, app.State.Version(), amdomain.NewHTTPMonitorDeleted(
			id,
			ehevent.MetaSystemUser(time.Now())))
	})
}

func httpMonitorEnableOrDisable(ctx context.Context, id string, newState bool) error {
	app, err := getApp(ctx)
	if err != nil {
		return err
	}

	return app.Reader.TransactWrite(ctx, func() error {
		monitorToEdit := amstate.FindHTTPMonitorWithID(id, app.State.HTTPMonitors())
		if monitorToEdit == nil {
			return fmt.Errorf("monitor not found: %s", id)
		}

		if monitorToEdit.Enabled == newState {
			return fmt.Errorf("monitor left unchanged: %s", id)
		}

		return app.AppendAfter(ctx, app.State.Version(), amdomain.NewHTTPMonitorEnabledUpdated(
			id,
			newState,
			ehevent.MetaSystemUser(time.Now())))
	})
}

func boolToCheckmark(input bool) string {
	if input {
		return "✓"
	} else {
		return "✗"
	}
}
