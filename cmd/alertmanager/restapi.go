package main

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"net/http"
	"os"
	"time"

	"github.com/function61/gokit/encoding/jsonfile"
	"github.com/function61/gokit/net/http/httputils"
	"github.com/function61/lambda-alertmanager/pkg/alertmanagertypes"
	"github.com/function61/lambda-alertmanager/pkg/amstate"
	"github.com/spf13/cobra"
)

func newRestAPI(ctx context.Context) http.Handler {
	app, err := getApp(ctx)
	if err != nil {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		})
	}

	routes := http.NewServeMux()

	routes.HandleFunc("GET /alerts", func(w http.ResponseWriter, r *http.Request) {
		noCacheHeaders(w)

		handleJSONOutput(w, app.State.ActiveAlerts())
	})

	routes.HandleFunc("POST /alerts/ingest", func(w http.ResponseWriter, r *http.Request) {
		alert := amstate.Alert{}
		if err := jsonfile.UnmarshalDisallowUnknownFields(r.Body, &alert); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		alert.ID = amstate.NewAlertID() // FIXME: bad design

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

	routes.HandleFunc("GET /alerts/acknowledge", func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")

		noCacheHeaders(w)

		if err := alertAck(r.Context(), id); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		_, _ = fmt.Fprintf(w, "Ack ok for %s", html.EscapeString(id))
	})

	routes.HandleFunc("GET /deadmansswitches", func(w http.ResponseWriter, r *http.Request) {
		noCacheHeaders(w)

		handleJSONOutput(w, app.State.DeadMansSwitches())
	})

	// /deadmansswitch/checkin?subject=ubackup_done&ttl=24h30m
	routes.HandleFunc("GET /deadmansswitch/checkin", func(w http.ResponseWriter, r *http.Request) {
		// same semantic hack here as acknowledge endpoint

		noCacheHeaders(w)

		// handles validation
		handleDeadMansSwitchCheckin(w, r, alertmanagertypes.DeadMansSwitchCheckinRequest{
			Subject: r.URL.Query().Get("subject"),
			TTL:     r.URL.Query().Get("ttl"),
		}, app)
	})

	routes.HandleFunc("POST /deadmansswitch/checkin", func(w http.ResponseWriter, r *http.Request) {
		checkin := alertmanagertypes.DeadMansSwitchCheckinRequest{}
		if err := jsonfile.UnmarshalDisallowUnknownFields(r.Body, &checkin); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// handles validation
		handleDeadMansSwitchCheckin(w, r, checkin, app)
	})

	routes.HandleFunc("POST /prometheus-alertmanager/api/v1/alerts", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "not implemented yet", http.StatusInternalServerError)
	})

	return routes
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

	ttl, err := parseTTLSpec(raw.TTL, now)
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
		_, _ = fmt.Fprintln(w, "Check-in noted; alert that was firing for this dead mans's switch was acked")
	} else {
		_, _ = fmt.Fprintln(w, "Check-in noted")
	}
}

func handleJSONOutput(w http.ResponseWriter, output any) {
	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(output); err != nil {
		panic(err)
	}
}

func restAPICliEntry() *cobra.Command {
	return &cobra.Command{
		Use:   "restapi",
		Short: "Start REST API (used mainly for dev/testing)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runStandaloneRestAPI(cmd.Context())
		},
	}
}

func runStandaloneRestAPI(ctx context.Context) error {
	srv := &http.Server{
		Addr:    ":80",
		Handler: newRestAPI(ctx),

		ReadHeaderTimeout: httputils.DefaultReadHeaderTimeout,
	}

	return httputils.CancelableServer(ctx, srv, srv.ListenAndServe)
}

func ackLink(alert amstate.Alert) string {
	return os.Getenv("API_ENDPOINT") + "/alerts/acknowledge?id=" + alert.ID
}

func noCacheHeaders(w http.ResponseWriter) {
	w.Header().Set("Cache-Control", "no-store, must-revalidate")
}
