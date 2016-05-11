package datadog

import (
	"testing"
)

func TestName(t *testing.T) {
	var testData = []struct {
		prefix string
		name   string
		tags   string
		wanted string
	}{
		{"", "meter", "", "meter"},
		{"prefix", "meter", "", "prefix.meter"},
		{"prefix", "meter", "tag", "prefix.meter[tag]"},
		{"", "meter", "tag", "meter[tag]"},
		{"", "meter", "tag:a,tag:b", "meter[tag:a,tag:b]"},
	}
	for _, metricNameParts := range testData {
		name, _ := NewMetricName(metricNameParts.prefix, metricNameParts.name, metricNameParts.tags)
		if want, have := metricNameParts.wanted, name; want != have {
			t.Errorf("%s [wanted] != %s [have]", want, have)
		}
	}
}

func TestMissingName(t *testing.T) {
	if _, err := NewMetricName("", "", ""); err == nil {
		t.Error("An empty name must return an error.")
	}
}
