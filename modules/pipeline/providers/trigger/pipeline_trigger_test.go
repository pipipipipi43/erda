// Copyright (c) 2021 Terminus, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package trigger

import (
	"testing"

	"github.com/erda-project/erda/modules/pipeline/providers/trigger/db"
)

func TestPipelineSvc_GetTriggerState(t *testing.T) {
	type fields struct {
	}
	type args struct {
		newPipelineTrigger *db.PipelineTrigger
		oldPipelineTrigger *db.PipelineTrigger
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			args: args{
				newPipelineTrigger: &db.PipelineTrigger{
					ID:        2,
					EventName: "ceshi1",
					Filter: map[string]string{
						"appid":     "2",
						"branch":    "4324",
						"projectid": "adfdf",
					},
				},
				oldPipelineTrigger: &db.PipelineTrigger{
					ID:        2,
					EventName: "ceshi1",
					Filter: map[string]string{
						"appid":     "2",
						"branch":    "4324",
						"projectid": "adfdf",
					},
				},
			},
			want: "",
		},
		{
			args: args{
				oldPipelineTrigger: &db.PipelineTrigger{
					ID:        2,
					EventName: "ceshi1",
					Filter: map[string]string{
						"appid":     "2",
						"branch":    "4324",
						"projectid": "adfdf",
					},
				},
			},
			want: "delete",
		},
		{
			args: args{
				newPipelineTrigger: &db.PipelineTrigger{
					ID:        2,
					EventName: "ceshi1",
					Filter: map[string]string{
						"appid":     "2",
						"branch":    "4324",
						"projectid": "adfdf",
					},
				},
			},
			want: "create",
		},
		{
			args: args{
				newPipelineTrigger: &db.PipelineTrigger{
					ID:        2,
					EventName: "ceshi1",
					Filter: map[string]string{
						"appid":     "2",
						"projectid": "adfdf",
					},
				},
				oldPipelineTrigger: &db.PipelineTrigger{
					ID:        2,
					EventName: "ceshi1",
					Filter: map[string]string{
						"appid":     "2",
						"branch":    "4324",
						"projectid": "adfdf",
					},
				},
			},
			want: "update",
		},
		{
			args: args{
				newPipelineTrigger: &db.PipelineTrigger{
					ID:        2,
					EventName: "ceshi3",
					Filter: map[string]string{
						"appid":     "2",
						"branch":    "4324",
						"projectid": "adfdf",
					},
				},
				oldPipelineTrigger: &db.PipelineTrigger{
					ID:        2,
					EventName: "ceshi1",
					Filter: map[string]string{
						"appid":     "2",
						"branch":    "4324",
						"projectid": "adfdf",
					},
				},
			},
			want: "update",
		},
		{
			args: args{
				newPipelineTrigger: &db.PipelineTrigger{
					ID:        2,
					EventName: "cesh1",
					Filter: map[string]string{
						"appid":     "2",
						"branch":    "432",
						"projectid": "adfdf",
					},
				},
				oldPipelineTrigger: &db.PipelineTrigger{
					ID:        2,
					EventName: "ceshi1",
					Filter: map[string]string{
						"appid":     "2",
						"branch":    "4324",
						"projectid": "adfdf",
					},
				},
			},
			want: "update",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &pipelineCm{}
			got, err := s.GetTriggerState(tt.args.newPipelineTrigger, tt.args.oldPipelineTrigger)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetTriggerState() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetTriggerState() got = %v, want %v", got, tt.want)
			}
		})
	}
}
