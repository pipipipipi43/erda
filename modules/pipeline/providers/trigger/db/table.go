package db

import (
	"time"

	"github.com/erda-project/erda/apistructs"
)

// PipelineBase represents `pipeline_triggers` table.
type PipelineTrigger struct {
	ID                   uint64                    `json:"id" xorm:"pk autoincr"`
	EventName            string                    `json:"eventName" xorm:"event_name"`
	PipelineSource       apistructs.PipelineSource `json:"pipelineSource" xorm:"pipeline_source"`
	PipelineYmlName      string                    `json:"pipelineYmlName" xorm:"pipeline_yml_name"`
	PipelineDefinitionID uint64                    `json:"pipelineDefinitionID" xorm:"pipeline_definition_id"`
	Filter               map[string]string         `json:"filter" xorm:"filter"`
	CreatedAt            *time.Time                `json:"createdAt,omitempty" xorm:"created_at created"`
	UpdatedAt            *time.Time                `json:"updatedAt,omitempty" xorm:"updated_at updated"`
}

func (*PipelineTrigger) TableName() string {
	return "pipeline_triggers"
}
