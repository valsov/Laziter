package laziter

type ValuesProvider[T any] struct {
	valuesChannel chan T
	nextChannel   chan struct{}
}

func (vp *ValuesProvider[T]) Wait() bool {
	_, ok := <-vp.nextChannel
	return ok
}

func (vp *ValuesProvider[T]) Yield(value T) {
	vp.valuesChannel <- value
}

func (vp *ValuesProvider[T]) Close() {
	close(vp.valuesChannel)
}
