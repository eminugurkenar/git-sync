/*
Copyright 2019 The Kubernetes Authors All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package hook

import (
	"testing"
	"time"

	"k8s.io/git-sync/pkg/log"
)

func TestWebhookDo(t *testing.T) {
	t.Run("test invalid urls are handled", func(t *testing.T) {
		wh := Webhook{
			URL:     ":http://localhost:601426/hooks/webhook",
			Method:  "POST",
			Success: 200,
			Timeout: time.Second,
			Logger:  log.NewLogger("", ""),
		}
		err := wh.Do("hash")
		if err == nil {
			t.Fatalf("expected error for invalid url but got none")
		}
	})
}
