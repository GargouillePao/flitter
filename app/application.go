package app

//Application application
type Application interface {
	OnStart()
	OnEnd(interface{})
}
