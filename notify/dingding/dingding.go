// Copyright 2019 Prometheus Team
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package dingding

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/pkg/errors"
	commoncfg "github.com/prometheus/common/config"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/prometheus/alertmanager/config"
	"github.com/prometheus/alertmanager/notify"
	"github.com/prometheus/alertmanager/template"
	"github.com/prometheus/alertmanager/types"
)

// Notifier implements a Notifier for DingDing notifications.
type Notifier struct {
	conf   *config.DingDingConfig
	tmpl   *template.Template
	logger log.Logger
	client *http.Client
}

type dingDingResponse struct {
	Code  int    `json:"errcode"`
	Error string `json:"errmsg"`
}

// New returns a new DingDing notifier.
func New(c *config.DingDingConfig, t *template.Template, l log.Logger, httpOpts ...commoncfg.HTTPClientOption) (*Notifier, error) {
	client, err := commoncfg.NewClientFromConfig(*c.HTTPConfig, "dingding", httpOpts...)
	if err != nil {
		return nil, err
	}

	return &Notifier{conf: c, tmpl: t, logger: l, client: client}, nil
}

// Notify implements the Notifier interface.
func (n *Notifier) Notify(ctx context.Context, as ...*types.Alert) (bool, error) {
	key, err := notify.ExtractGroupKey(ctx)
	if err != nil {
		return false, err
	}

	level.Debug(n.logger).Log("incident", key)
	data := notify.GetTemplateData(ctx, n.tmpl, as, n.logger)

	tmpl := notify.TmplText(n.tmpl, data, &err)
	if err != nil {
		return false, err
	}

	msg := tmpl(n.conf.Message)

	postMessageURL := n.conf.APIURL.Copy()
	postMessageURL.Path += "robot/send"
	q := postMessageURL.Query()
	q.Set("access_token", string(n.conf.AccessToken))
	postMessageURL.RawQuery = q.Encode()

	resp, err := notify.PostJSON(ctx, n.client, postMessageURL.String(), strings.NewReader(msg))
	if err != nil {
		return true, notify.RedactURL(err)
	}
	defer notify.Drain(resp)

	if resp.StatusCode != 200 {
		return true, fmt.Errorf("unexpected status code %v", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return true, err
	}
	level.Debug(n.logger).Log("response", string(body), "incident", key)

	var ddResp dingDingResponse
	if err := json.Unmarshal(body, &ddResp); err != nil {
		return true, err
	}

	if ddResp.Code == 0 {
		return false, nil
	}

	return false, errors.New(ddResp.Error)
}
