package main

import (
	"context"
	"fmt"
	"testing"

	"github.com/function61/gokit/testing/assert"
	"github.com/function61/lambda-alertmanager/pkg/amstate"
)

func TestOneFails(t *testing.T) {
	failures := scanMonitors(context.Background(), []amstate.HTTPMonitor{
		{
			URL:  "http://example.com/frontpage",
			Find: "Welcome to",
		},
		{
			URL:  "http://example.com/contacts",
			Find: "bar@exmaple.com",
		},
	}, &testScanner{}, nil)

	assert.Equal(t, len(failures), 1)
	assert.Equal(
		t,
		failures[0].err.Error(),
		"string-to-find `bar@exmaple.com` NOT in body: Contact us by email at foo@example.com")
}

func TestAllSucceed(t *testing.T) {
	failures := scanMonitors(context.Background(), []amstate.HTTPMonitor{
		{
			URL:  "http://example.com/frontpage",
			Find: "Welcome to",
		},
		{
			URL:  "http://example.com/contacts",
			Find: "foo@example.com",
		},
	}, &testScanner{}, nil)

	assert.Equal(t, len(failures), 0)
}

func Test404(t *testing.T) {
	failures := scanMonitors(context.Background(), []amstate.HTTPMonitor{
		{
			URL:  "http://notfound.net/",
			Find: "doesntmatter",
		},
	}, &testScanner{}, nil)

	assert.Equal(t, len(failures), 1)
	assert.Equal(t, failures[0].err.Error(), "404: http://notfound.net/")
	assert.EqualJSON(t, failures[0].monitor, `{
  "id": "",
  "created": "0001-01-01T00:00:00Z",
  "enabled": false,
  "url": "http://notfound.net/",
  "find": "doesntmatter"
}`)
}

type testScanner struct{}

func (a *testScanner) Scan(ctx context.Context, monitor amstate.HTTPMonitor) error {
	pages := map[string]string{
		"http://example.com/frontpage": "Welcome to frontpage",
		"http://example.com/contacts":  "Contact us by email at foo@example.com",
	}

	page, found := pages[monitor.URL]
	if !found {
		return fmt.Errorf("404: %s", monitor.URL)
	}

	return mustFindStringInBody(page, monitor.Find)
}
