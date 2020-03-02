package main

import (
	"context"
	"fmt"
	"github.com/function61/eventhorizon/pkg/ehcli"
	"github.com/function61/eventhorizon/pkg/ehreader"
	"github.com/function61/gokit/dynversion"
	"github.com/function61/gokit/logex"
	"github.com/function61/gokit/ossignal"
	"github.com/function61/lambda-alertmanager/pkg/amstate"
	"github.com/spf13/cobra"
	"os"
	"strings"
	"time"
)

func main() {
	// AWS Lambda doesn't support giving argv, so we use an ugly hack to detect when
	// we're in Lambda
	if strings.Contains(os.Args[0], "_lambda") {
		lambdaHandler()
		return
	}

	app := &cobra.Command{
		Use:     os.Args[0],
		Short:   "Alert manager",
		Version: dynversion.Version,
	}

	app.AddCommand(alertEntry())

	app.AddCommand(deadMansSwitchEntry())

	app.AddCommand(httpMonitorEntry())

	app.AddCommand(ehcli.Entrypoint())

	app.AddCommand(restApiCliEntry())

	app.AddCommand(&cobra.Command{
		Use:   "lambda-scheduler",
		Short: "Run what Lambda would invoke in response to scheduler event",
		Run: func(*cobra.Command, []string) {
			exitIfError(handleCloudwatchScheduledEvent(
				ossignal.InterruptOrTerminateBackgroundCtx(nil),
				time.Now()))
		},
	})

	exitIfError(app.Execute())
}

func getApp(ctx context.Context) (*amstate.App, error) {
	tenantCtx, err := ehreader.TenantCtxWithSnapshotsFrom(ehreader.ConfigFromEnv, "am:v1")
	if err != nil {
		return nil, err
	}

	logger := logex.StandardLogger()

	return amstate.LoadUntilRealtime(
		ctx,
		tenantCtx,
		logger)
}

func exitIfError(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
