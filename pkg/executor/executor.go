package executor

import (
	"context"
	"github.com/MR5356/jietan/pkg/executor/api"
)

type Executor interface {
	Execute(context context.Context, params *api.ExecuteParams) *api.ExecuteResult
	GetResult(field api.ResultField, params interface{}) interface{}
}
