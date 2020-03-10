package main

import (
	"context"
	"fmt"
	"github.com/function61/gokit/ezhttp"
	"github.com/function61/gokit/logex"
	"github.com/function61/lambda-alertmanager/pkg/amstate"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

type monitorFailure struct {
	err     error
	monitor amstate.HttpMonitor
}

func httpMonitorScanAndAlertFailures(ctx context.Context, app *amstate.App) error {
	startOfScan := time.Now()

	failures := scanMonitors(
		ctx,
		app.State.HttpMonitors(),
		newRetryScanner(newScanner()),
		logex.Prefix("httpscanner", app.Logger))

	// convert monitor failures into alerts
	alerts := []amstate.Alert{}
	for _, failure := range failures {
		alerts = append(alerts, amstate.Alert{
			Id:        amstate.NewAlertId(),
			Subject:   failure.monitor.Url,
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
	monitors []amstate.HttpMonitor,
	scanner HttpMonitorScanner,
	logger *log.Logger,
) []monitorFailure {
	logl := logex.Levels(logger)

	failed := []monitorFailure{}
	failedMu := sync.Mutex{}

	checkOne := func(monitor amstate.HttpMonitor) {
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

			logl.Error.Printf("❌ %s @ %d ms => %v", monitor.Url, durationMs, err.Error())
		} else {
			logl.Debug.Printf("✔️ %s @ %d ms", monitor.Url, durationMs)
		}
	}

	work := make(chan amstate.HttpMonitor)

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

type HttpMonitorScanner interface {
	Scan(context.Context, amstate.HttpMonitor) error
}

type retryScanner struct {
	actualScanner HttpMonitorScanner
}

// retries once, but only if it looks retryable
func newRetryScanner(actual HttpMonitorScanner) HttpMonitorScanner {
	return &retryScanner{actual}
}

func (r *retryScanner) Scan(ctx context.Context, monitor amstate.HttpMonitor) error {
	firstTryCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	if err := r.actualScanner.Scan(firstTryCtx, monitor); err != nil {
		if err != context.DeadlineExceeded { // non-retryable error
			return err
		}

		// now use the longer context
		return r.actualScanner.Scan(ctx, monitor)
	}

	return nil
}

type scanner struct {
	noRedirects *http.Client
}

func newScanner() HttpMonitorScanner {
	return &scanner{
		&http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse // do not follow redirects
			},
		},
	}
}

func (s *scanner) Scan(ctx context.Context, monitor amstate.HttpMonitor) error {
	resp, err := ezhttp.Get(
		ctx,
		monitor.Url,
		ezhttp.TolerateNon2xxResponse,
		ezhttp.Client(s.noRedirects)) // rationale: no much else than how previous one worked
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	buf, err := ioutil.ReadAll(resp.Body)
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

	for i := 0; i < numWorkers; i++ {
		workersDone.Add(1)
		go func() {
			defer workersDone.Done()

			worker()
		}()
	}

	produceWork()

	workersDone.Wait()
}
