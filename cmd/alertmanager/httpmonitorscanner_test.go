package main

import (
	"context"
	"fmt"
	"github.com/function61/gokit/assert"
	"github.com/function61/lambda-alertmanager/pkg/amstate"
	"testing"
)

func TestOneFails(t *testing.T) {
	failures := scanMonitors(context.Background(), []amstate.HttpMonitor{
		{
			Url:  "http://example.com/frontpage",
			Find: "Welcome to",
		},
		{
			Url:  "http://example.com/contacts",
			Find: "bar@exmaple.com",
		},
	}, &testScanner{}, nil)

	assert.Assert(t, len(failures) == 1)
	assert.EqualString(
		t,
		failures[0].err.Error(),
		"string-to-find `bar@exmaple.com` NOT in body: Contact us by email at foo@example.com")
}

func TestAllSucceed(t *testing.T) {
	failures := scanMonitors(context.Background(), []amstate.HttpMonitor{
		{
			Url:  "http://example.com/frontpage",
			Find: "Welcome to",
		},
		{
			Url:  "http://example.com/contacts",
			Find: "foo@example.com",
		},
	}, &testScanner{}, nil)

	assert.Assert(t, len(failures) == 0)
}

func Test404(t *testing.T) {
	failures := scanMonitors(context.Background(), []amstate.HttpMonitor{
		{
			Url:  "http://notfound.net/",
			Find: "doesntmatter",
		},
	}, &testScanner{}, nil)

	assert.Assert(t, len(failures) == 1)
	assert.EqualString(t, failures[0].err.Error(), "404: http://notfound.net/")
	assert.EqualJson(t, failures[0].monitor, `{
  "id": "",
  "created": "0001-01-01T00:00:00Z",
  "enabled": false,
  "url": "http://notfound.net/",
  "find": "doesntmatter"
}`)
}

type testScanner struct{}

func (a *testScanner) Scan(ctx context.Context, monitor amstate.HttpMonitor) error {
	pages := map[string]string{
		"http://example.com/frontpage": "Welcome to frontpage",
		"http://example.com/contacts":  "Contact us by email at foo@example.com",
	}

	page, found := pages[monitor.Url]
	if !found {
		return fmt.Errorf("404: %s", monitor.Url)
	}

	return mustFindStringInBody(page, monitor.Find)
}
