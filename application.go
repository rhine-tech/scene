package scene

type AppName string

type AppStatus int

const (
	AppStatusStopped AppStatus = iota
	AppStatusRunning
	AppStatusError
)

type Application interface {
	Name() AppName // return scene
	Status() AppStatus
	Error() error
}

type Repository interface {
	RepoImplName() string
	Status() error
}

type Service interface {
	SrvImplName() string
}
