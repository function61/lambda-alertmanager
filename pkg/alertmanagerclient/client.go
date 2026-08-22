package alertmanagerclient

import (
	"context"
	"os"
	"time"

	"github.com/function61/gokit/net/http/ezhttp"
	"github.com/function61/gokit/os/osutil"
	"github.com/function61/lambda-alertmanager/pkg/alertmanagertypes"
)

const (
	baseURLEnvVarName = "ALERTMANAGER_BASEURL"
)

type Client struct {
	baseURL string
}

func New(baseURL string) *Client {
	return &Client{baseURL}
}

func (c *Client) Alert(ctx context.Context, alert alertmanagertypes.Alert) error {
	_, err := ezhttp.Post(ctx, c.baseURL+"/alerts/ingest", ezhttp.SendJSON(&alert))
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
	_, err := ezhttp.Post(ctx, c.baseURL+"/deadmansswitch/checkin", ezhttp.SendJSON(&req))
	return err
}

// if ALERTMANAGER_BASEURL is set, returns client
func ClientFromEnvOptional() *Client {
	baseURL := os.Getenv(baseURLEnvVarName)
	if baseURL == "" {
		return nil
	}

	return New(baseURL)
}

// required version of ClientFromEnvOptional()
func ClientFromEnvRequired() (*Client, error) {
	baseURL, err := osutil.GetenvRequired(baseURLEnvVarName)
	if err != nil {
		return nil, err
	}

	return New(baseURL), nil
}
