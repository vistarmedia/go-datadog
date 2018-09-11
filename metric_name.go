package datadog

import (
	"errors"
	"strings"
)

func NewCheckedMetricName(prefix string, name string, tags string) (string, error) {
	if len(name) == 0 {
		return "", errors.New("Metric name cannot be empty.")
	}

	return NewMetricName(prefix, name, tags), nil
}

func NewMetricName(prefix string, name string, tags string) string {
	var buffer strings.Builder
	if len(prefix) > 0 {
		buffer.WriteString(prefix)
		buffer.WriteString(".")
	}
	buffer.WriteString(name)
	if len(tags) > 0 {
		buffer.WriteString("[")
		buffer.WriteString(tags)
		buffer.WriteString("]")
	}
	return buffer.String()
}
