// Simple client to the [Datadog API](http://docs.datadoghq.com/api/).
package datadog

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httputil"

	"bytes"
	"github.com/rcrowley/go-metrics"
	"io"
	"io/ioutil"
)

const (
	ENDPOINT        = "https://app.datadoghq.com/api"
	SERIES_ENDPIONT = "/v1/series"
	EVENT_ENDPOINT  = "/v1/events"
	CONTENT_TYPE    = "application/json"
)

type Client struct {
	Host   string
	ApiKey string
}

type seriesMessage struct {
	Series []json.RawMessage `json:"series,omitempty"`
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
	body, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return c.doRequest(c.EventUrl(), body)
}

const messageChunkSize = 2 * 1024 * 1024

// Posts an array of series data to the Datadog API. The API expects an object,
// not an array, so it will be wrapped in a `seriesMessage` with a single
// `series` field.
//
// If the slice contains too many series, the message will be split into
// multiple chunks of around 2mb each.
//
func (c *Client) PostSeries(series []Series, reg metrics.Registry) error {
	var approxTotalSize int
	var encodedSeries []json.RawMessage

	for _, serie := range series {
		// encode series to json
		jsonSerie, err := json.Marshal(serie)
		if err != nil {
			return err
		}

		encoded := json.RawMessage(jsonSerie)

		// count bytes of this message
		approxTotalSize += len(encoded)
		encodedSeries = append(encodedSeries, encoded)

		if approxTotalSize > messageChunkSize {
			if err := c.sendEncodedSeries(encodedSeries); err != nil {
				return err
			}

			// reset and start to collect the next chunk
			encodedSeries = encodedSeries[:0]
			approxTotalSize = 0
		}
	}

	if len(encodedSeries) > 0 {
		return c.sendEncodedSeries(encodedSeries)
	}

	return nil
}

func (c *Client) sendEncodedSeries(series []json.RawMessage) error {
	body, err := json.Marshal(seriesMessage{series})
	if err != nil {
		return err
	}

	return c.doRequest(c.SeriesUrl(), body)
}

// Create a `MetricsReporter` for the given metrics reporter. The returned
// reporter will not be started.
func (c *Client) Reporter(reg metrics.Registry, tags []string) *MetricsReporter {
	return Reporter(c, reg, tags)
}

// Create a `MetricsReporter` configured to use metric's default registry. This
// reporter will not be started.
func (c *Client) DefaultReporter() *MetricsReporter {
	return Reporter(c, metrics.DefaultRegistry, nil)
}

func (c *Client) doRequest(url string, body []byte) (err error) {
	req, err := http.NewRequest("POST", c.SeriesUrl(), bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("building request: %s", err)
	}

	req.ContentLength = int64(len(body))
	req.Header.Set("Content-Type", "application/json")

	// now execute the request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	// for keep-alive we ensure that the response-body is read.
	defer io.Copy(ioutil.Discard, resp.Body)
	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		dumpReq, _ := httputil.DumpRequest(req, false)
		dumpRes, _ := httputil.DumpResponse(resp, true)
		return fmt.Errorf("bad datadog request and response:\n%s\n%s", string(dumpReq), string(dumpRes))
	}

	return nil
}
