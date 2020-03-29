package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/function61/gokit/httputils"
	"github.com/function61/gokit/jsonfile"
	"github.com/function61/gokit/logex"
	"github.com/function61/gokit/ossignal"
	"github.com/function61/gokit/taskrunner"
	"github.com/function61/lambda-alertmanager/pkg/alertmanagertypes"
	"github.com/function61/lambda-alertmanager/pkg/amstate"
	"github.com/spf13/cobra"
)

func newRestApi(ctx context.Context) http.Handler {
	app, err := getApp(ctx)
	if err != nil {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		})
	}

	mux := httputils.NewMethodMux()

	mux.GET.HandleFunc("/alerts", func(w http.ResponseWriter, r *http.Request) {
		noCacheHeaders(w)

		handleJsonOutput(w, app.State.ActiveAlerts())
	})

	mux.POST.HandleFunc("/alerts/ingest", func(w http.ResponseWriter, r *http.Request) {
		alert := amstate.Alert{}
		if err := jsonfile.Unmarshal(r.Body, &alert, true); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		alert.Id = amstate.NewAlertId() // FIXME: bad design

		created, err := ingestAlertsAndReturnCreatedFlag(r.Context(), []amstate.Alert{alert}, app)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if created {
			w.WriteHeader(http.StatusCreated)
		} else {
			w.WriteHeader(http.StatusNoContent)
		}
	})

	mux.GET.HandleFunc("/alerts/acknowledge", func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")

		noCacheHeaders(w)

		if err := alertAck(r.Context(), id); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, "Ack ok for %s", id)
	})

	mux.GET.HandleFunc("/deadmansswitches", func(w http.ResponseWriter, r *http.Request) {
		noCacheHeaders(w)

		handleJsonOutput(w, app.State.DeadMansSwitches())
	})

	// /deadmansswitch/checkin?subject=ubackup_done&ttl=24h30m
	mux.GET.HandleFunc("/deadmansswitch/checkin", func(w http.ResponseWriter, r *http.Request) {
		// same semantic hack here as acknowledge endpoint

		noCacheHeaders(w)

		// handles validation
		handleDeadMansSwitchCheckin(w, r, alertmanagertypes.DeadMansSwitchCheckinRequest{
			Subject: r.URL.Query().Get("subject"),
			TTL:     r.URL.Query().Get("ttl"),
		}, app)
	})

	mux.POST.HandleFunc("/deadmansswitch/checkin", func(w http.ResponseWriter, r *http.Request) {
		checkin := alertmanagertypes.DeadMansSwitchCheckinRequest{}
		if err := jsonfile.Unmarshal(r.Body, &checkin, true); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// handles validation
		handleDeadMansSwitchCheckin(w, r, checkin, app)
	})

	mux.POST.HandleFunc("/prometheus-alertmanager/api/v1/alerts", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "not implemented yet", http.StatusInternalServerError)
	})

	return mux
}

func handleDeadMansSwitchCheckin(
	w http.ResponseWriter,
	r *http.Request,
	raw alertmanagertypes.DeadMansSwitchCheckinRequest,
	app *amstate.App,
) {
	if raw.Subject == "" || raw.TTL == "" {
		http.Error(w, "subject or ttl empty", http.StatusBadRequest)
		return
	}

	now := time.Now()

	ttl, err := parseTtlSpec(raw.TTL, now)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	alertAcked, err := deadmansswitchCheckin(r.Context(), raw.Subject, ttl, app, time.Now())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if alertAcked {
		fmt.Fprintln(w, "Check-in noted; alert that was firing for this dead mans's switch was acked")
	} else {
		fmt.Fprintln(w, "Check-in noted")
	}
}

func handleJsonOutput(w http.ResponseWriter, output interface{}) {
	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(output); err != nil {
		panic(err)
	}
}

func restApiCliEntry() *cobra.Command {
	return &cobra.Command{
		Use:   "restapi",
		Short: "Start REST API (used mainly for dev/testing)",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			logger := logex.StandardLogger()

			exitIfError(runStandaloneRestApi(
				ossignal.InterruptOrTerminateBackgroundCtx(logger),
				logger))
		},
	}
}

func runStandaloneRestApi(ctx context.Context, logger *log.Logger) error {
	srv := &http.Server{
		Addr:    ":80",
		Handler: newRestApi(ctx),
	}

	tasks := taskrunner.New(ctx, logger)

	tasks.Start("listener "+srv.Addr, func(_ context.Context, _ string) error {
		return httputils.RemoveGracefulServerClosedError(srv.ListenAndServe())
	})

	tasks.Start("listenershutdowner", httputils.ServerShutdownTask(srv))

	return tasks.Wait()
}

func ackLink(alert amstate.Alert) string {
	return os.Getenv("API_ENDPOINT") + "/alerts/acknowledge?id=" + alert.Id
}

func noCacheHeaders(w http.ResponseWriter) {
	w.Header().Set("Cache-Control", "no-store, must-revalidate")
}
