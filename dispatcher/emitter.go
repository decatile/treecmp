package dispatcher

type Emitter interface {
	Emit(pathA, pathB string) (shouldContinue bool)
	Close() error
}

type noopEmitter struct{}

func (noopEmitter) Emit(pathA string, pathB string) (shouldContinue bool) {
	return true
}

func (noopEmitter) Close() error {
	return nil
}
