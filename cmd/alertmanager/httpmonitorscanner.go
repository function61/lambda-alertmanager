package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/function61/gokit/net/http/ezhttp"
	"github.com/function61/lambda-alertmanager/pkg/amstate"
)

type monitorFailure struct {
	err     error
	monitor amstate.HTTPMonitor
}

func httpMonitorScanAndAlertFailures(ctx context.Context, app *amstate.App) error {
	startOfScan := time.Now()

	failures := scanMonitors(
		ctx,
		amstate.EnabledHTTPMonitors(app.State.HTTPMonitors()),
		newRetryScanner(newScanner()),
		app.Logger.With("component", "httpscanner"))

	// convert monitor failures into alerts
	alerts := []amstate.Alert{}
	for _, failure := range failures {
		alerts = append(alerts, amstate.Alert{
			ID:        amstate.NewAlertID(),
			Subject:   failure.monitor.URL,
			Details:   failure.err.Error(),
			Timestamp: startOfScan,
		})
	}

	// ok with len(alerts) == 0
	return ingestAlerts(ctx, alerts, app)
}

// scans HTTP monitors and returns the ones that failed
func scanMonitors(
	ctx context.Context,
	monitors []amstate.HTTPMonitor,
	scanner HTTPMonitorScanner,
	logger *slog.Logger,
) []monitorFailure {
	if logger == nil {
		logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	}

	failed := []monitorFailure{}
	failedMu := sync.Mutex{}

	checkOne := func(monitor amstate.HTTPMonitor) {
		ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()

		started := time.Now()

		err := scanner.Scan(ctx, monitor)

		durationMs := time.Since(started).Milliseconds()

		if err != nil {
			failedMu.Lock()
			defer failedMu.Unlock()

			failed = append(failed, monitorFailure{
				err,
				monitor,
			})

			logger.ErrorContext(ctx, "HTTP monitor scan failed",
				"url", monitor.URL,
				"duration_ms", durationMs,
				"err", err)
		} else {
			logger.DebugContext(ctx, "HTTP monitor scan succeeded",
				"url", monitor.URL,
				"duration_ms", durationMs)
		}
	}

	work := make(chan amstate.HTTPMonitor)

	concurrently(3, func() {
		for monitor := range work {
			checkOne(monitor)
		}
	}, func() {
		for _, monitor := range monitors {
			work <- monitor
		}

		close(work)
	})

	return failed
}

type HTTPMonitorScanner interface {
	Scan(context.Context, amstate.HTTPMonitor) error
}

type retryScanner struct {
	actualScanner HTTPMonitorScanner
}

// retries once
func newRetryScanner(actual HTTPMonitorScanner) HTTPMonitorScanner {
	return &retryScanner{actual}
}

func (r *retryScanner) Scan(ctx context.Context, monitor amstate.HTTPMonitor) error {
	firstTryCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	if err := r.actualScanner.Scan(firstTryCtx, monitor); err != nil {
		time.Sleep(2 * time.Second)

		// it'd be hard to detect if we shouldn't retry this at all, since timeouts,
		// HTTP gateway errors, internal server errors etc. all can be transient

		// now use the longer context
		if err2 := r.actualScanner.Scan(ctx, monitor); err2 != nil {
			return fmt.Errorf("first error: %v; retry error: %v", err, err2)
		}

		return nil
	}

	return nil
}

type scanner struct {
	noRedirects *http.Client
}

func newScanner() HTTPMonitorScanner {
	return &scanner{
		&http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse // do not follow redirects
			},
		},
	}
}

func (s *scanner) Scan(ctx context.Context, monitor amstate.HTTPMonitor) error {
	resp, err := ezhttp.Get(
		ctx,
		monitor.URL,
		ezhttp.TolerateNon2xxResponse,
		ezhttp.Client(s.noRedirects)) // rationale: no much else than how previous one worked
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	buf, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	return mustFindStringInBody(string(buf), monitor.Find)
}

func mustFindStringInBody(body string, find string) error {
	if !strings.Contains(body, find) {
		return fmt.Errorf("string-to-find `%s` NOT in body: %s", find, body)
	}

	return nil
}

func concurrently(numWorkers int, worker func(), produceWork func()) {
	workersDone := sync.WaitGroup{}

	for range numWorkers {
		workersDone.Go(func() {

			worker()
		})
	}

	produceWork()

	workersDone.Wait()
}
