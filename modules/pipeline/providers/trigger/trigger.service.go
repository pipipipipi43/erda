package trigger

import (
	context "context"
	"fmt"
	"github.com/erda-project/erda-infra/providers/mysqlxorm"
	pb "github.com/erda-project/erda-proto-go/core/pipeline/trigger/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/providers/trigger/db"
	"github.com/erda-project/erda/modules/pipeline/services/pipelinesvc"
)

type triggerService struct {
	p *provider

	cm          ConfigManager
	db          mysqlxorm.Interface
	dbClient    *db.Client
	pipelineSvc *pipelinesvc.PipelineSvc
}

func (s *triggerService) RunPipelinesByTrigger(ctx context.Context, req *pb.PipelineTriggerRequest) (*pb.PipelineTriggerResponse, error) {
	// TODO .

	pipelineTriggers, err := s.dbClient.ListPipelineTriggersBy(req)
	if err != nil {
		return nil, err
	}

	var pipelineIDs []uint64
	for _, trigger := range pipelineTriggers {
		fmt.Println(trigger)
		// 需要查询 pipleine 定义, pipeline 定义中包含 pipelineCreatev2

		lable := map[string]string{
			"diceWorkspace": "TEST",
			"branch":        "develop",
			"orgID":         "1",
			"projectID":     "1",
			"appID":         "3",
		}
		for k, v := range trigger.Filter {
			lable[fmt.Sprintf("trigger.%s", k)] = v
		}
		pipeline, err := s.pipelineSvc.CreateV2(&apistructs.PipelineCreateRequestV2{
			PipelineYml:     "version: \"1.1\"\nstages:\n  - stage:\n      - testplan-run:\n          alias: testplan-run\n          description: 根据自动化测试计划启动测试计划并等待完成\n          version: \"1.0\"\n          params:\n            cms: autotest^scope-project-autotest-testcase^scopeid-1^369136180207301828\n            is_continue_execution: \"false\"\n            test_plan: 1\n  - stage:\n      - git-checkout:\n          alias: git-checkout\n          description: 代码仓库克隆\n      - custom-script:\n          alias: custom-script1\n          description: 运行自定义命令\n          version: \"1.0\"\n          commands:\n            - ls\n      - custom-script:\n          alias: custom-script2\n          description: 运行自定义命令\n          version: \"1.0\"\n          commands:\n            - ls\n",
			ClusterName:     "terminus-dev",
			PipelineYmlName: "3/TEST/develop/pipeline.yml", // 字段换成 定义表里的
			PipelineSource:  "dice",                        // 定义
			Labels:          lable,
			NormalLabels: map[string]string{
				"appName":     "test13",
				"projectName": "ss",
				"orgName":     "terminus",
			},
			ConfigManageNamespaces: []string{
				"pipeline-secrets-app-3-default",
				"pipeline-secrets-app-3-develop",
				"app-3-default",
				"app-3-test",
			},
			AutoRunAtOnce: true,
			AutoStartCron: false,
			ForceRun:      true,
		})
		if err != nil {
			return nil, err
		}
		pipelineIDs = append(pipelineIDs, pipeline.ID)
	}

	return &pb.PipelineTriggerResponse{PipelineIds: pipelineIDs}, nil
}
