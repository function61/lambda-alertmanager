package main

import (
	"context"
	"fmt"
	"github.com/function61/eventhorizon/pkg/ehevent"
	"github.com/function61/gokit/ossignal"
	"github.com/function61/gokit/stringutils"
	"github.com/function61/lambda-alertmanager/pkg/amdomain"
	"github.com/function61/lambda-alertmanager/pkg/amstate"
	"github.com/scylladb/termtables"
	"github.com/spf13/cobra"
	"time"
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
		Run: func(cmd *cobra.Command, args []string) {
			exitIfError(httpMonitorList(
				ossignal.InterruptOrTerminateBackgroundCtx(nil)))
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "rm [id]",
		Short: "Remove HTTP monitor",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			exitIfError(httpMonitorDelete(
				ossignal.InterruptOrTerminateBackgroundCtx(nil),
				args[0]))
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "enable [id]",
		Short: "Enable disabled HTTP monitor",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			exitIfError(httpMonitorEnableOrDisable(
				ossignal.InterruptOrTerminateBackgroundCtx(nil),
				args[0],
				true))
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "disable [id]",
		Short: "Temporarily disable a HTTP monitor",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			exitIfError(httpMonitorEnableOrDisable(
				ossignal.InterruptOrTerminateBackgroundCtx(nil),
				args[0],
				false))
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "mk [url] [find]",
		Short: "Create HTTP monitor",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			exitIfError(httpMonitorCreate(
				ossignal.InterruptOrTerminateBackgroundCtx(nil),
				args[0],
				args[1]))
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "scan",
		Short: "Runs all enabled monitors and raises alerts if appropriate",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			ctx := ossignal.InterruptOrTerminateBackgroundCtx(nil)

			app, err := getApp(ctx)
			exitIfError(err)

			exitIfError(httpMonitorScanAndAlertFailures(ctx, app))
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

	for _, alert := range app.State.HttpMonitors() {
		view.AddRow(
			alert.Id,
			boolToCheckmark(alert.Enabled),
			stringutils.Truncate(alert.Url, 44),
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

	monitorCreated := amdomain.NewHttpMonitorCreated(
		amstate.NewHttpMonitorId(),
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
		if amstate.FindHttpMonitorWithId(id, app.State.HttpMonitors()) == nil {
			return fmt.Errorf("monitor to delete not found: %s", id)
		}

		return app.AppendAfter(ctx, app.State.Version(), amdomain.NewHttpMonitorDeleted(
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
		monitorToEdit := amstate.FindHttpMonitorWithId(id, app.State.HttpMonitors())
		if monitorToEdit == nil {
			return fmt.Errorf("monitor not found: %s", id)
		}

		if monitorToEdit.Enabled == newState {
			return fmt.Errorf("monitor left unchanged: %s", id)
		}

		return app.AppendAfter(ctx, app.State.Version(), amdomain.NewHttpMonitorEnabledUpdated(
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
