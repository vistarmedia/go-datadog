// Simple client to the [Datadog API](http://docs.datadoghq.com/api/).
package datadog

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/rcrowley/go-metrics"
	"io"
	"net/http"
)

const (
	ENDPOINT        = "https://app.datadoghq.com/api"
	SERIES_ENDPIONT = "/v1/series"
	EVENT_ENDPOINT  = "/v1/events"
	CONTENT_TYPE    = "application/json"
)

type Client struct {
	Host        string
	ApiKey      string
	Environment string
}

type seriesMessage struct {
	Series []*Series `json:"series,omitempty"`
}

type Series struct {
	Metric string           `json:"metric"`
	Points [][2]interface{} `json:"points"`
	Type   string           `json:"type"`
	Host   string           `json:"host"`
	Tags   []string         `json:"tags,omitempty"`
}

type Event struct {
	Title     string   `json:"title"`
	Text      string   `json:"text"`
	Priority  string   `json:"priority"`
	Tags      []string `json:"tags"`
	AlertType string   `json:"alert_type"`
}

// Create a new Datadog client. In EC2, datadog expects the hostname to be the
// instance ID rather than `gethostname(2)`. However, that value can be obtained
// with `os.Hostname()`.
func New(host, apiKey string) *Client {
	return &Client{
		Host:   host,
		ApiKey: apiKey,
	}
}

func NewWithEnv(host, apiKey, environment string) *Client {
	return &Client{
		Host:        host,
		ApiKey:      apiKey,
		Environment: environment,
	}
}

// Gets an authenticated URL to POST series data to. In Datadog's examples, this
// value is 'https://app.datadoghq.com/api/v1/series?api_key=9775a026f1ca7d1...'
func (c *Client) SeriesUrl() string {
	return ENDPOINT + SERIES_ENDPIONT + "?api_key=" + c.ApiKey
}

// Gets an authenticate URL to POST events.
func (c *Client) EventUrl() string {
	return ENDPOINT + EVENT_ENDPOINT + "?api_key=" + c.ApiKey
}

func (c *Client) PostEvent(event *Event) (err error) {
	bs, err := json.Marshal(event)
	if err != nil {
		return err
	}

	resp, err := http.Post(c.EventUrl(), CONTENT_TYPE, bytes.NewBuffer(bs))
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	if !(resp.StatusCode == 200 || resp.StatusCode == 202) {
		return fmt.Errorf("Bad Datadog response: '%s'", resp.Status)
	}
	return
}

// Posts an array of series data to the Datadog API. The API expects an object,
// not an array, so it will be wrapped in a `seriesMessage` with a single
// `series` field.
func (c *Client) PostSeries(series []*Series) (err error) {
	body, err := c.seriesReader(series)
	if err != nil {
		return err
	}
	resp, err := http.Post(c.SeriesUrl(), CONTENT_TYPE, body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if !(resp.StatusCode == 200 || resp.StatusCode == 202) {
		return fmt.Errorf("Bad Datadog response: '%s'", resp.Status)
	}
	return
}

// Serializes an array of `Series` to JSON. The array will be wrapped in a
// `seriesMessage`, changing the serialized type from an array to an object with
// a single `series` field.
func (c *Client) seriesReader(series []*Series) (io.Reader, error) {
	msg := &seriesMessage{series}
	bs, err := json.Marshal(msg)
	if err != nil {
		return nil, err
	}
	return bytes.NewBuffer(bs), nil
}

// Create a `MetricsReporter` for the given metrics reporter. The returned
// reporter will not be started.
func (c *Client) Reporter(reg metrics.Registry) *MetricsReporter {
	return Reporter(c, reg)
}

// Create a `MetricsReporter` configured to use metric's default registry. This
// reporter will not be started.
func (c *Client) DefaultReporter() *MetricsReporter {
	return Reporter(c, metrics.DefaultRegistry)
}
