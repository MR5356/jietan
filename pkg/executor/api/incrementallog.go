package api

type IncrementalLog struct {
	Log   []string `json:"log"`
	Start int      `json:"start"`
	End   int      `json:"end"`
	More  bool     `json:"more"`
}
