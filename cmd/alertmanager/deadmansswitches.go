package main

import (
	"context"
	"fmt"
	"github.com/function61/eventhorizon/pkg/ehevent"
	"github.com/function61/gokit/ossignal"
	"github.com/function61/lambda-alertmanager/pkg/amdomain"
	"github.com/function61/lambda-alertmanager/pkg/amstate"
	"github.com/scylladb/termtables"
	"github.com/spf13/cobra"
	"time"
)

func deadMansSwitchEntry() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dms",
		Short: "Manage dead man´s switches",
	}

	expired := false

	ls := &cobra.Command{
		Use:   "ls",
		Short: "List dead man´s switches",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			exitIfError(deadmansswitchList(
				ossignal.InterruptOrTerminateBackgroundCtx(nil),
				expired))
		},
	}

	ls.Flags().BoolVarP(&expired, "expired", "e", expired, "List only expired")

	cmd.AddCommand(ls)

	cmd.AddCommand(&cobra.Command{
		Use:   "rm [id]",
		Short: "Remove a switch",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			exitIfError(deadmansswitchRemove(
				ossignal.InterruptOrTerminateBackgroundCtx(nil),
				args[0]))
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "checkin [subject] [ttl]",
		Short: "Make a checkin",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			ttl, err := parseTtlSpec(args[1], time.Now())
			exitIfError(err)

			_, err = deadmansswitchCheckin(
				ossignal.InterruptOrTerminateBackgroundCtx(nil),
				args[0],
				ttl)
			exitIfError(err)
		},
	})

	return cmd
}

func deadmansswitchList(ctx context.Context, expired bool) error {
	app, err := getApp(ctx)
	if err != nil {
		return err
	}

	dmss := app.State.DeadMansSwitches()
	if expired {
		dmss = amstate.GetExpiredDeadMansSwitches(dmss, time.Now())
	}

	view := termtables.CreateTable()
	view.AddHeaders("Subject", "TTL")

	for _, dms := range dmss {
		view.AddRow(dms.Subject, dms.Ttl.Format(time.RFC3339))
	}

	fmt.Println(view.Render())

	return nil
}

func deadmansswitchRemove(ctx context.Context, subject string) error {
	app, err := getApp(ctx)
	if err != nil {
		return err
	}

	return app.Reader.TransactWrite(ctx, func() error {
		if amstate.FindDeadMansSwitchWithSubject(subject, app.State.DeadMansSwitches()) == nil {
			return fmt.Errorf("switch to delete not found: %s", subject)
		}

		return app.AppendAfter(ctx, app.State.Version(), amdomain.NewDeadMansSwitchDeleted(
			subject,
			ehevent.MetaSystemUser(time.Now())))
	})
}

func deadmansswitchCheckin(ctx context.Context, subject string, ttl time.Time) (bool, error) {
	app, err := getApp(ctx)
	if err != nil {
		return false, err
	}

	now := time.Now()

	alertAcked := false

	checkin := amdomain.NewDeadMansSwitchCheckin(
		subject,
		ttl,
		ehevent.MetaSystemUser(now))

	if err := app.Reader.TransactWrite(ctx, func() error {
		events := []ehevent.Event{}

		// first time seeing this checkin => create said switch
		if amstate.FindDeadMansSwitchWithSubject(subject, app.State.DeadMansSwitches()) == nil {
			events = append(events, amdomain.NewDeadMansSwitchCreated(
				subject,
				ttl,
				ehevent.MetaSystemUser(now)))

			alertAcked = true
		}

		events = append(events, checkin)

		if alert := amstate.FindAlertWithSubject(subject, app.State.ActiveAlerts()); alert != nil {
			events = append(events, amdomain.NewAlertAcknowledged(
				alert.Id,
				ehevent.MetaSystemUser(now)))
		}

		return app.AppendAfter(ctx, app.State.Version(), events...)
	}); err != nil {
		return false, err
	}

	return alertAcked, nil
}
