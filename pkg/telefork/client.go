// Copyright 2022 Outreach Corporation. All Rights Reserved.

// Description: Implements a telefork client
package telefork

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"strings"

	"github.com/grevych/gobox/pkg/log"
	"go.opentelemetry.io/otel/attribute"
)

type Event map[string]interface{}

type Client interface {
	SendEvent(attributes []attribute.KeyValue)

	AddField(key string, val interface{})
	AddInfo(args ...log.Marshaler)

	Close()
}

func NewClient(appName, apiKey string) Client {
	c := &http.Client{}
	return NewClientWithHTTPClient(appName, apiKey, c)
}

func NewClientWithHTTPClient(appName, apiKey string, httpClient *http.Client) Client {
	baseURL := "https://telefork.outreach.io/"
	if os.Getenv("OUTREACH_TELEFORK_ENDPOINT") != "" {
		baseURL = os.Getenv("OUTREACH_TELEFORK_ENDPOINT")
	}
	return &client{
		http: httpClient,

		appName: appName,
		baseURL: baseURL,
		apiKey:  apiKey,
	}
}

type client struct {
	http *http.Client

	appName string
	apiKey  string
	baseURL string
	events  []Event

	commonProps map[string]interface{}
}

func (c *client) SendEvent(attributes []attribute.KeyValue) {
	e := make(Event)

	for k, v := range c.commonProps {
		e[k] = v
	}
	for _, a := range attributes {
		e[string(a.Key)] = a.Value.AsString()
	}

	c.events = append(c.events, e)
}

func (c *client) Close() {
	if c.apiKey == "" || c.apiKey == "NOTSET" {
		return
	}

	if len(c.events) == 0 {
		return
	}

	b, err := json.Marshal(c.events)
	if err != nil {
		return
	}

	r, err := http.NewRequest(http.MethodPost, strings.TrimSuffix(c.baseURL, "/")+"/", bytes.NewReader(b))
	if err != nil {
		return
	}

	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("X-OUTREACH-CLIENT-LOGGING", c.apiKey)
	r.Header.Set("X-OUTREACH-CLIENT-APP-ID", c.appName)

	res, err := c.http.Do(r)
	if err != nil {
		return
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusCreated {
		return
	}
}

func (c *client) AddInfo(args ...log.Marshaler) {
	for _, arg := range args {
		arg.MarshalLog(c.AddField)
	}
}

func (c *client) AddField(key string, val interface{}) {
	if c.commonProps == nil {
		c.commonProps = map[string]interface{}{}
	}
	c.commonProps[key] = val
}
