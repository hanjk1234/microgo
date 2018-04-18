package worker

type Worker interface {
	Start() error
	Stop() error
}
