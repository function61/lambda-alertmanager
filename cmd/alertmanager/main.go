package main

import (
	"context"
	"log/slog"
	"time"

	"github.com/function61/eventhorizon/pkg/ehcli"
	"github.com/function61/eventhorizon/pkg/ehreader"
	"github.com/function61/gokit/app/aws/lambdautils"
	"github.com/function61/gokit/app/cli"
	"github.com/function61/lambda-alertmanager/pkg/amstate"
	"github.com/spf13/cobra"
)

func main() {
	// AWS Lambda doesn't support giving argv, so we use an ugly hack to detect when
	// we're in Lambda
	if lambdautils.InLambda() {
		lambdaHandler()
		return
	}

	app := &cobra.Command{
		Short: "Alert manager",
	}

	app.AddCommand(alertEntry())

	app.AddCommand(deadMansSwitchEntry())

	app.AddCommand(httpMonitorEntry())

	app.AddCommand(ehcli.Entrypoint())

	app.AddCommand(restAPICliEntry())

	app.AddCommand(&cobra.Command{
		Use:   "lambda-scheduler",
		Short: "Run what Lambda would invoke in response to scheduler event",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return handleCloudwatchScheduledEvent(cmd.Context(), time.Now())
		},
	})

	cli.Execute(app)
}

func getApp(ctx context.Context) (*amstate.App, error) {
	tenantCtx, err := ehreader.TenantCtxWithSnapshotsFrom(ehreader.ConfigFromEnv, "am:v1")
	if err != nil {
		return nil, err
	}

	return amstate.LoadUntilRealtime(
		ctx,
		tenantCtx,
		slog.Default())
}
