package datadog

import (
	. "github.com/go-check/check"
	"testing"
)

func Test(t *testing.T) { TestingT(t) }

type ClientSuite struct{}

var _ = Suite(&ClientSuite{})
var client *Client

func (s *ClientSuite) SetUpTest(c *C) {
	client = &Client{}
}

func (s *ClientSuite) TestSeriesEndpoint(c *C) {
	client.ApiKey = "secret"
	c.Check(client.SeriesUrl(), Equals,
		"https://app.datadoghq.com/api/v1/series?api_key=secret")
}

func (s *ClientSuite) TestEventsEndpoint(c *C) {
	client.ApiKey = "secret"
	c.Check(client.EventUrl(), Equals,
		"https://app.datadoghq.com/api/v1/events?api_key=secret")
}
