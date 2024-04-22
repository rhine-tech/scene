package scene

type AppStatus int

const (
	AppStatusStopped AppStatus = iota
	AppStatusRunning
	AppStatusError
)
