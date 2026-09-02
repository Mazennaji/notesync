package ipc

type Request struct {
	Command string         `json:"command"`
	Args    map[string]any `json:"args"`
	Config  map[string]any `json:"config"`
}

type Response struct {
	OK    bool   `json:"ok"`
	Data  any    `json:"data"`
	Error string `json:"error,omitempty"`
}
