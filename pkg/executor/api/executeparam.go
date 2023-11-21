package api

const (
	ResultFieldLog     = "log"
	ResultFieldErr     = "error"
	ResultFieldIncrLog = "incrLog"

	ExecuteParamScript = "script"
	ExecuteParamParams = "params"
	ExecuteParamHosts  = "hosts"
)

type ExecuteParam string

type ExecuteParams map[ExecuteParam]interface{}

func (ps ExecuteParams) SetParam(param ExecuteParam, value interface{}) ExecuteParams {
	ps[param] = value
	return ps
}

func (ps ExecuteParams) SetParams(params map[ExecuteParam]interface{}) ExecuteParams {
	for k, v := range params {
		ps[k] = v
	}
	return ps
}

func (ps ExecuteParams) GetParam(param ExecuteParam) interface{} {
	if v, ok := ps[param]; ok {
		return v
	} else {
		return nil
	}
}

func (ps ExecuteParams) GetScript() string {
	if v, ok := ps[ExecuteParamScript]; ok {
		return v.(string)
	} else {
		return ""
	}
}

func (ps ExecuteParams) GetParams() string {
	if v, ok := ps[ExecuteParamParams]; ok {
		return v.(string)
	} else {
		return ""
	}
}
