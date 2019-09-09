package alertmanagerclient

import (
	"context"
	"github.com/function61/gokit/ezhttp"
	"github.com/function61/lambda-alertmanager/alertmanager/pkg/alertmanagertypes"
)

type Client struct {
	baseUrl string
}

func New(baseUrl string) *Client {
	return &Client{baseUrl}
}

func (c *Client) DeadMansSwitchCheckin(ctx context.Context, req alertmanagertypes.DeadMansSwitchCheckinRequest) error {
	_, err := ezhttp.Post(ctx, c.baseUrl+"/deadmansswitches/checkin", ezhttp.SendJson(&req))
	return err
}
