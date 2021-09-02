package apistructs

type PipelineTriggerRequest struct {
	EventName string            `json:"eventName"`
	Label     map[string]string `json:"lable"`
}

type PipelineTriggerResponse struct {
	PipelineId []uint64
}
