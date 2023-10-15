package laziter

// Produces values and send then through a channel
type ValuesProvider[T any] struct {
	valuesChannel chan T
	nextChannel   chan struct{}
}

// Wait for next value request
//
// If the result of this method is false, the values production should stop and the function calling this method should consider exiting
func (vp *ValuesProvider[T]) Wait() bool {
	_, ok := <-vp.nextChannel
	return ok
}

// Send a value though a channel
func (vp *ValuesProvider[T]) Yield(value T) {
	vp.valuesChannel <- value
}

// Cleanup values channel
func (vp *ValuesProvider[T]) Close() {
	close(vp.valuesChannel)
}
