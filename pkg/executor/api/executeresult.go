package api

const (
	Success executeResultCode = 0
	Failed  executeResultCode = 1
)

type ResultField string

type executeResultCode int

type ExecuteResult struct {
	Status  executeResultCode      `json:"status" yaml:"status"`
	Message string                 `json:"message" yaml:"message"`
	Data    map[string]interface{} `json:"data" yaml:"data"`
}

func (r *ExecuteResult) IsSuccess() bool {
	return r.Status == Success
}
