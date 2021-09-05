package trigger

import (
	context "context"
	"fmt"
	"github.com/erda-project/erda-infra/providers/mysqlxorm"
	pb "github.com/erda-project/erda-proto-go/core/pipeline/trigger/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/openapi/component-protocol/pkg/component_key"
	"github.com/erda-project/erda/modules/pipeline/providers/definition"
	definitionDb "github.com/erda-project/erda/modules/pipeline/providers/definition/db"
	triggerDb "github.com/erda-project/erda/modules/pipeline/providers/trigger/db"
	"github.com/erda-project/erda/modules/pipeline/services/apierrors"
	"github.com/erda-project/erda/modules/pipeline/services/pipelinesvc"
	"github.com/erda-project/erda/pkg/common/apis"
	"github.com/erda-project/erda/pkg/parser/pipelineyml"
	"github.com/pkg/errors"
)

type TriggerService struct {
	p *provider

	cm                 ConfigManager
	db                 mysqlxorm.Interface
	triggerDbClient    *triggerDb.Client
	definitionDbClient *definitionDb.Client
	pipelineSvc        *pipelinesvc.PipelineSvc
}

func (s *TriggerService) SetPipelineSvc(pipelineSvc *pipelinesvc.PipelineSvc) {
	s.pipelineSvc = pipelineSvc
}

func (s *TriggerService) RunPipelineByTriggerRequest(ctx context.Context, req *pb.PipelineTriggerRequest) (*pb.PipelineTriggerResponse, error) {
	// TODO .
	err := s.checkPermission(ctx)
	if err != nil {
		return nil, apierrors.ErrCheckPermission.AccessDenied()
	}

	pipelineTriggers, err := s.triggerDbClient.ListPipelineTriggers(req)
	if err != nil {
		return nil, err
	}

	// Get trigger label
	triggerLabel := make(map[string]string)
	for key, val := range req.Label {
		triggerLabel[fmt.Sprintf("triggers.%s", key)] = val
	}

	// De-duplication
	pipelineTriggersMap := make(map[uint64]triggerDb.PipelineTrigger)
	for _, trigger := range pipelineTriggers {
		if _, ok := pipelineTriggersMap[trigger.PipelineDefinitionID]; !ok {
			pipelineTriggersMap[trigger.PipelineDefinitionID] = trigger
		}
	}

	var pipelineIDs []uint64
	for _, trigger := range pipelineTriggersMap {

		pipelineDefinition, err := s.definitionDbClient.GetPipelineDefinitionByNameAndSource(trigger.PipelineSource, trigger.PipelineYmlName)
		if err != nil {
			return nil, err
		}

		if pipelineDefinition.PipelineYmlName != "" && pipelineDefinition.PipelineYml != "" && pipelineDefinition.PipelineSource != "" {
			if pipelineDefinition.Extra.CreateRequest == nil {
				break
			}

			label := make(map[string]string)
			if pipelineDefinition.Extra.CreateRequest != nil && pipelineDefinition.Extra.CreateRequest.Labels != nil {
				label = pipelineDefinition.Extra.CreateRequest.Labels
			}
			for key, val := range triggerLabel {
				label[key] = val
			}

			pipeline, err := s.pipelineSvc.CreateV2(&apistructs.PipelineCreateRequestV2{
				PipelineYml:            pipelineDefinition.PipelineYml,
				ClusterName:            pipelineDefinition.Extra.CreateRequest.ClusterName,
				PipelineYmlName:        pipelineDefinition.PipelineYmlName,
				PipelineSource:         pipelineDefinition.PipelineSource,
				Labels:                 label,
				NormalLabels:           pipelineDefinition.Extra.CreateRequest.NormalLabels,
				ConfigManageNamespaces: pipelineDefinition.Extra.CreateRequest.ConfigManageNamespaces,
				AutoRunAtOnce:          true,
				AutoStartCron:          false,
				ForceRun:               true,
				IdentityInfo: apistructs.IdentityInfo{
					UserID: apis.GetUserID(ctx),
				},
			})
			if err != nil {
				return nil, err
			}
			pipelineIDs = append(pipelineIDs, pipeline.ID)
		}
	}

	return &pb.PipelineTriggerResponse{PipelineIDs: pipelineIDs}, nil
}

func (s *TriggerService) RegisterTriggerEvent(definition definition.PipelineDefinitionProcess, yml pipelineyml.PipelineYml) error {
	newPipelineTriggerMap := make(map[string]map[string]string)
	if yml.Spec() != nil && yml.Spec().Trigger != nil {
		if yml.Spec().Trigger != nil {
			for _, trigger := range yml.Spec().Trigger {
				if trigger.Filter != nil {
					newPipelineTriggerMap[trigger.On] = trigger.Filter
				}
			}
		}
	}

	oldPipelineTriggers, err := s.triggerDbClient.GetPipelineTriggerByID(definition.ID)
	if err != nil {
		return err
	}
	oldPipelineTriggerMap := make(map[string]map[string]string)
	for _, pipelineTrigger := range oldPipelineTriggers {
		if pipelineTrigger.Filter != nil {
			oldPipelineTriggerMap[pipelineTrigger.EventName] = pipelineTrigger.Filter
		}
	}

	state, err := s.GetTriggerState(newPipelineTriggerMap, oldPipelineTriggerMap)
	if err != nil {
		return err
	}

	if definition.IsDelete {
		state = "delete"
	}

	switch state {
	case "create":
		for eventName, newPipelineTrigger := range newPipelineTriggerMap {
			err := s.triggerDbClient.CreatePipelineTrigger(&triggerDb.PipelineTrigger{
				EventName:            eventName,
				PipelineSource:       definition.PipelineSource,
				PipelineYmlName:      definition.PipelineYmlName,
				PipelineDefinitionID: definition.ID,
				Filter:               newPipelineTrigger,
			})
			if err != nil {
				return err
			}
		}
	case "update":
		for _, oldPipelineTrigger := range oldPipelineTriggers {
			err := s.triggerDbClient.DeletePipelineTrigger(oldPipelineTrigger.ID)
			if err != nil {
				return err
			}
		}
		for eventName, newPipelineTrigger := range newPipelineTriggerMap {
			err := s.triggerDbClient.CreatePipelineTrigger(&triggerDb.PipelineTrigger{
				ID:                   component_key.GetKey(0),
				EventName:            eventName,
				PipelineSource:       definition.PipelineSource,
				PipelineYmlName:      definition.PipelineYmlName,
				PipelineDefinitionID: definition.ID,
				Filter:               newPipelineTrigger,
			})
			if err != nil {
				return err
			}
		}
	case "delete":
		for _, oldPipelineTrigger := range oldPipelineTriggers {
			err := s.triggerDbClient.DeletePipelineTrigger(oldPipelineTrigger.ID)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *TriggerService) GetTriggerState(newPipelineTrigger map[string]map[string]string, oldPipelineTrigger map[string]map[string]string) (string, error) {
	if len(newPipelineTrigger) == 0 {
		return "delete", nil
	}
	if len(oldPipelineTrigger) == 0 {
		return "create", nil
	}
	return "update", nil
}

func (s *TriggerService) checkPermission(ctx context.Context) error {
	userID := apis.GetUserID(ctx)
	if userID == "" {
		return errors.Errorf("failed to get permission(User-ID is empty)")
	}
	ordID := apis.GetOrgID(ctx)
	if ordID == "" {
		return errors.Errorf("failed to get permission(org-ID is empty)")
	}
	return nil
}
