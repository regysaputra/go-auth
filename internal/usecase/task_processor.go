package usecase

// TaskProcessor is the interface for processing tasks
type TaskProcessor interface {
	Start() error
	Shutdown()
}
