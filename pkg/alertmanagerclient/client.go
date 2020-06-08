package alertmanagerclient

import (
	"context"
	"os"
	"time"

	"github.com/function61/gokit/envvar"
	"github.com/function61/gokit/ezhttp"
	"github.com/function61/lambda-alertmanager/pkg/alertmanagertypes"
)

const (
	baseUrlEnvVarName = "ALERTMANAGER_BASEURL"
)

type Client struct {
	baseUrl string
}

func New(baseUrl string) *Client {
	return &Client{baseUrl}
}

func (c *Client) Alert(ctx context.Context, alert alertmanagertypes.Alert) error {
	_, err := ezhttp.Post(ctx, c.baseUrl+"/alerts/ingest", ezhttp.SendJson(&alert))
	return err
}

// simpler version of DeadMansSwitchCheckinCustom()
func (c *Client) DeadMansSwitchCheckin(
	ctx context.Context,
	subject string,
	ttl time.Duration,
) error {
	return c.DeadMansSwitchCheckinCustom(
		ctx,
		alertmanagertypes.NewDeadMansSwitchCheckinRequest(
			subject,
			"+"+ttl.String()))
}

func (c *Client) DeadMansSwitchCheckinCustom(
	ctx context.Context,
	req alertmanagertypes.DeadMansSwitchCheckinRequest,
) error {
	_, err := ezhttp.Post(ctx, c.baseUrl+"/deadmansswitch/checkin", ezhttp.SendJson(&req))
	return err
}

// if ALERTMANAGER_BASEURL is set, returns client
func ClientFromEnvOptional() *Client {
	baseUrl := os.Getenv(baseUrlEnvVarName)
	if baseUrl == "" {
		return nil
	}

	return New(baseUrl)
}

// required version of ClientFromEnvOptional()
func ClientFromEnvRequired() (*Client, error) {
	baseUrl, err := envvar.Required(baseUrlEnvVarName)
	if err != nil {
		return nil, err
	}

	return New(baseUrl), nil
}
