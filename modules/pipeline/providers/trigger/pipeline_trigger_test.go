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
)

func TestPipelineSvc_GetTriggerState(t *testing.T) {
	type fields struct {
	}
	type args struct {
		newPipelineTrigger map[string]map[string]string
		oldPipelineTrigger map[string]map[string]string
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
				newPipelineTrigger: map[string]map[string]string{
					"sdf": {
						"asd": "123",
					},
				},
				oldPipelineTrigger: map[string]map[string]string{
					"sdf": {
						"asd": "123",
					},
				},
			},
			want: "update",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &TriggerService{}
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
