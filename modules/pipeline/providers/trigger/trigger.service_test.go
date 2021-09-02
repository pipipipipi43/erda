package trigger

import (
	context "context"
	servicehub "github.com/erda-project/erda-infra/base/servicehub"
	pb "github.com/erda-project/erda-proto-go/core/pipeline/trigger/pb"
	reflect "reflect"
	testing "testing"
)

func Test_triggerService_RunPipelinesByTrigger(t *testing.T) {
	type args struct {
		ctx context.Context
		req *pb.PipelineTriggerRequest
	}
	tests := []struct {
		name     string
		service  string
		config   string
		args     args
		wantResp *pb.PipelineTriggerResponse
		wantErr  bool
	}{
		// TODO: Add test cases.
		{
			"case 1",
			"erda.core.pipeline.trigger.TriggerService",
			`
erda.core.pipeline.trigger:
`,
			args{
				context.TODO(),
				&pb.PipelineTriggerRequest{
					// TODO: setup fields
				},
			},
			&pb.PipelineTriggerResponse{
				// TODO: setup fields.
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hub := servicehub.New()
			events := hub.Events()
			go func() {
				hub.RunWithOptions(&servicehub.RunOptions{Content: tt.config})
			}()
			err := <-events.Started()
			if err != nil {
				t.Error(err)
				return
			}
			srv := hub.Service(tt.service).(pb.TriggerServiceServer)
			got, err := srv.RunPipelinesByTrigger(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("triggerService.RunPipelinesByTrigger() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.wantResp) {
				t.Errorf("triggerService.RunPipelinesByTrigger() = %v, want %v", got, tt.wantResp)
			}
		})
	}
}
