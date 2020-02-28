package main

import (
	"context"
	"fmt"
	"github.com/function61/eventhorizon/pkg/ehevent"
	"github.com/function61/gokit/ossignal"
	"github.com/function61/lambda-alertmanager/pkg/amdomain"
	"github.com/function61/lambda-alertmanager/pkg/amstate"
	"github.com/function61/lambda-alertmanager/pkg/wtfgo"
	"github.com/scylladb/termtables"
	"github.com/spf13/cobra"
	"strings"
	"time"
)

func alertEntry() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "alert",
		Short: "Manage alerts",
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "mk [subject] [details]",
		Short: "Raise an alert",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			exitIfError(alertRaise(
				ossignal.InterruptOrTerminateBackgroundCtx(nil),
				args[0],
				args[1]))
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "ls",
		Short: "List active alerts",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			exitIfError(alertList(
				ossignal.InterruptOrTerminateBackgroundCtx(nil)))
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "ack [id]",
		Short: "Acknowledge an alert",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			exitIfError(alertAck(
				ossignal.InterruptOrTerminateBackgroundCtx(nil),
				args[0]))
		},
	})

	return cmd
}

func alertRaise(ctx context.Context, subject string, details string) error {
	app, err := getApp(ctx)
	if err != nil {
		return err
	}

	raised := amdomain.NewAlertRaised(
		amstate.NewAlertId(),
		subject,
		details,
		ehevent.MetaSystemUser(time.Now()))

	return app.Reader.TransactWrite(ctx, func() error {
		if amstate.FindAlertWithSubject(subject, app.State.ActiveAlerts()) != nil {
			return fmt.Errorf("already active have alert: %s", subject)
		}

		return app.AppendAfter(ctx, app.State.Version(), raised)
	})
}

func alertList(ctx context.Context) error {
	app, err := getApp(ctx)
	if err != nil {
		return err
	}

	view := termtables.CreateTable()
	view.AddHeaders("Id", "Raised", "Subject", "Details")

	for _, alert := range app.State.ActiveAlerts() {
		view.AddRow(
			alert.Id,
			alert.Timestamp.Format(time.RFC3339),
			alert.Subject,
			wtfgo.Truncate(removeLinebreaks(alert.Details), 50))
	}

	fmt.Println(view.Render())

	return nil
}

func alertAck(ctx context.Context, alertId string) error {
	app, err := getApp(ctx)
	if err != nil {
		return err
	}

	acked := amdomain.NewAlertAcknowledged(
		alertId,
		ehevent.MetaSystemUser(time.Now()))

	return app.Reader.TransactWrite(ctx, func() error {
		if !amstate.HasAlertWithId(alertId, app.State.ActiveAlerts()) {
			return fmt.Errorf("no alert: %s", alertId)
		}

		return app.AppendAfter(ctx, app.State.Version(), acked)
	})
}

func removeLinebreaks(input string) string {
	return strings.ReplaceAll(
		strings.ReplaceAll(
			input,
			"\r",
			`\r`),
		"\n",
		`\n`)
}
