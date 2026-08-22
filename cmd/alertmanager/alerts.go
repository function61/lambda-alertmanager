package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/function61/eventhorizon/pkg/ehevent"
	"github.com/function61/gokit/strings/stringutils"
	"github.com/function61/lambda-alertmanager/pkg/amdomain"
	"github.com/function61/lambda-alertmanager/pkg/amstate"
	"github.com/scylladb/termtables"
	"github.com/spf13/cobra"
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
		RunE: func(cmd *cobra.Command, args []string) error {
			return alertRaise(
				cmd.Context(),
				args[0],
				args[1])
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "ls",
		Short: "List active alerts",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return alertList(cmd.Context())
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "ack [id]",
		Short: "Acknowledge an alert",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return alertAck(cmd.Context(), args[0])
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
		amstate.NewAlertID(),
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
			alert.ID,
			alert.Timestamp.Format(time.RFC3339),
			alert.Subject,
			stringutils.Truncate(removeLinebreaks(alert.Details), 50))
	}

	fmt.Println(view.Render())

	return nil
}

func alertAck(ctx context.Context, alertID string) error {
	app, err := getApp(ctx)
	if err != nil {
		return err
	}

	acked := amdomain.NewAlertAcknowledged(
		alertID,
		ehevent.MetaSystemUser(time.Now()))

	return app.Reader.TransactWrite(ctx, func() error {
		if !amstate.HasAlertWithID(alertID, app.State.ActiveAlerts()) {
			return fmt.Errorf("no alert: %s", alertID)
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
