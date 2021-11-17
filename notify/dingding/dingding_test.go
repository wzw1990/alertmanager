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
	"testing"

	"github.com/go-kit/log"
	commoncfg "github.com/prometheus/common/config"
	"github.com/stretchr/testify/require"

	"github.com/prometheus/alertmanager/config"
	"github.com/prometheus/alertmanager/notify/test"
)

func TestDingDingNotifier(t *testing.T) {
	ctx, u, fn := test.GetContextWithCancelingURL()
	defer fn()

	ddUrl, err := u.Parse("https://oapi.dingtalk.com")
	require.NoError(t, err)
	secret := "4cc8c9f63242c81a74418573f26222d298199654d5b047e83eb59de0754493ff"
	notifier, err := New(
		&config.DingDingConfig{
			APIURL:     &config.URL{URL: ddUrl},
			HTTPConfig: &commoncfg.HTTPClientConfig{},
			Message: `{
				'msgtype': 'link',
				 'link': {
					'text':'Prometheus AlertManager',
					'title': '【监控报警】测试',
					'picUrl': 'http://img.25pp.com/uploadfile/soft/images/2014/1216/20141216112732330.jpg',
					'messageUrl': 'http://www.baidu.com'
				}
			}`,
			AccessToken: config.Secret(secret),
		},
		test.CreateTmpl(t),
		log.NewNopLogger(),
	)
	require.NoError(t, err)

	test.AssertNotifyLeaksNoSecret(t, ctx, notifier, secret)
}
