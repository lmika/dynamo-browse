package scriptmanager

//go:generate mockery --with-expecter --name UIService

type Ifaces struct {
	UI UIService
}

type UIService interface {
	PrintMessage(msg string)
}
