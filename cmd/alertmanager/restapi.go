package main

import (
	"bytes"
	"context"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/function61/gokit/jsonfile"
	"github.com/function61/lambda-alertmanager/pkg/alertmanagertypes"
	"github.com/function61/lambda-alertmanager/pkg/amstate"
	"github.com/function61/lambda-alertmanager/pkg/lambdautils"
	"os"
	"time"
)

func handleRestCall(ctx context.Context, req events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	app, err := getApp(ctx)
	if err != nil {
		return lambdautils.InternalServerError(err.Error()), nil
	}

	synopsis := req.HTTPMethod + " " + req.Path

	switch synopsis {
	case "GET /alerts":
		return handleGetAlerts(ctx, req, app)
	case "GET /alerts/acknowledge":
		// this endpoint should really be a POST (since it mutates state), but we've to be
		// pragmatic here because we want acks to be ack-able from emails
		id := req.QueryStringParameters["id"]
		if id == "" {
			return lambdautils.BadRequest("id not specified"), nil
		}

		return handleAcknowledgeAlert(ctx, id)
	case "POST /alerts/ingest":
		alert := amstate.Alert{}
		if err := jsonfile.Unmarshal(bytes.NewBufferString(req.Body), &alert, true); err != nil {
			return lambdautils.BadRequest(err.Error()), nil
		}
		alert.Id = amstate.NewAlertId()

		created, err := ingestAlertsAndReturnCreatedFlag(ctx, []amstate.Alert{alert}, app)
		if err != nil {
			return lambdautils.InternalServerError(err.Error()), nil
		}

		if created {
			return lambdautils.Created(), nil
		} else {
			return lambdautils.NoContent(), nil
		}
	case "GET /deadmansswitch/checkin": // /deadmansswitch/checkin?subject=ubackup_done&ttl=24h30m
		// same semantic hack here as acknowledge endpoint
		return handleDeadMansSwitchCheckin(ctx, alertmanagertypes.DeadMansSwitchCheckinRequest{
			Subject: req.QueryStringParameters["subject"],
			TTL:     req.QueryStringParameters["ttl"],
		})
	case "POST /deadmansswitch/checkin": // {"subject":"ubackup_done","ttl":"24h30m"}
		checkin := alertmanagertypes.DeadMansSwitchCheckinRequest{}
		if err := jsonfile.Unmarshal(bytes.NewBufferString(req.Body), &checkin, true); err != nil {
			return lambdautils.BadRequest(err.Error()), nil
		}

		return handleDeadMansSwitchCheckin(ctx, checkin)
	case "GET /deadmansswitches":
		return handleGetDeadMansSwitches(ctx, app)
	case "POST /prometheus-alertmanager/api/v1/alerts":
		return lambdautils.InternalServerError("not implemented yet"), nil
	default:
		return lambdautils.BadRequest(fmt.Sprintf("unknown endpoint: %s", synopsis)), nil
	}
}

func handleGetAlerts(
	ctx context.Context,
	req events.APIGatewayProxyRequest,
	app *amstate.App,
) (*events.APIGatewayProxyResponse, error) {
	return lambdautils.RespondJson(app.State.ActiveAlerts())
}

func handleAcknowledgeAlert(ctx context.Context, id string) (*events.APIGatewayProxyResponse, error) {
	if err := alertAck(ctx, id); err != nil {
		return lambdautils.InternalServerError(err.Error()), nil
	}

	return lambdautils.OkText(fmt.Sprintf("Ack ok for %s", id))
}

func handleGetDeadMansSwitches(
	ctx context.Context,
	app *amstate.App,
) (*events.APIGatewayProxyResponse, error) {
	return lambdautils.RespondJson(app.State.DeadMansSwitches())
}

func handleDeadMansSwitchCheckin(ctx context.Context, raw alertmanagertypes.DeadMansSwitchCheckinRequest) (*events.APIGatewayProxyResponse, error) {
	if raw.Subject == "" || raw.TTL == "" {
		return lambdautils.BadRequest("subject or ttl empty"), nil
	}

	now := time.Now()

	ttl, err := parseTtlSpec(raw.TTL, now)
	if err != nil {
		return lambdautils.BadRequest(err.Error()), nil
	}

	alertAcked, err := deadmansswitchCheckin(ctx, raw.Subject, ttl)
	if err != nil {
		return lambdautils.InternalServerError(err.Error()), nil
	}

	if alertAcked {
		return lambdautils.OkText("Check-in noted; alert that was firing for this dead mans's switch was acked")
	}

	return lambdautils.OkText("Check-in noted")
}

func ackLink(alert amstate.Alert) string {
	return os.Getenv("API_ENDPOINT") + "/alerts/acknowledge?id=" + alert.Id
}
