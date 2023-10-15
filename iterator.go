package laziter

// Allows value retrieval at each iteration
type Iterator[T any] interface {
	// Try to request the next value
	NextValue() (T, bool)
	// Try to request the next value, if it succeeds it is stored for later retrieval
	Next() bool
	// Retrieve the current iterator internal value if it exists
	GetCurrentValue() (T, bool)
	// Reset the current persisted values index
	ResetIteratorPosition()
	// Get the ValuesProvider associated with the Iterator
	GetValuesProvider() *ValuesProvider[T]
	// Stop the iteration process. If the value provider is still able to produce values, it should stop doing so as the transport channel is now closed.
	//
	// This method should always be called after the iteration is no longer needed to prevent leaking goroutines. ('defer' is useful here)
	Close()
}

// Default iterator implementation. Allows persisting iterated values.
type persitableIterator[T any] struct {
	Provider      *ValuesProvider[T]
	InternalSlice []T
	currentIndex  int
	currentValue  T
	hasValue      bool
	persistValues bool
}

// Produce a configured iterator with a ValuesProvider
func New[T any](persistValues bool) Iterator[T] {
	return &persitableIterator[T]{
		InternalSlice: []T{},
		persistValues: persistValues,
		currentIndex:  -1,
		Provider: &ValuesProvider[T]{
			valuesChannel: make(chan T, 1),
			nextChannel:   make(chan struct{}),
		},
	}
}

func (i *persitableIterator[T]) NextValue() (T, bool) {
	if !i.Next() {
		var value T
		return value, false
	}
	return i.GetCurrentValue()
}

func (i *persitableIterator[T]) Next() bool {
	// Check if the currentIndex is within already available values (case of a currentIndex reset)
	if i.persistValues && len(i.InternalSlice) > i.currentIndex+1 {
		i.currentIndex++
		i.hasValue = true
		i.currentValue = i.InternalSlice[i.currentIndex]
		return true
	}

	var (
		value T
		ok    bool
	)

	select {
	case value, ok = <-i.Provider.valuesChannel:
		// Value already available or closed chan
	case i.Provider.nextChannel <- struct{}{}: // Send Next value signal
		value, ok = <-i.Provider.valuesChannel
	}

	i.hasValue = ok
	if ok {
		i.currentValue = value
		if i.persistValues {
			i.store(value)
			i.currentIndex++
		}
	}
	return ok
}

func (i *persitableIterator[T]) GetCurrentValue() (T, bool) {
	if i.hasValue {
		return i.currentValue, true
	}
	var defaultVal T
	return defaultVal, false
}

func (i *persitableIterator[T]) ResetIteratorPosition() {
	i.currentIndex = -1
}

func (i *persitableIterator[T]) GetValuesProvider() *ValuesProvider[T] {
	return i.Provider
}

func (i *persitableIterator[T]) Close() {
	close(i.Provider.nextChannel)
}

// Persist a value in the iterator's internal store
func (i *persitableIterator[T]) store(value T) {
	i.InternalSlice = append(i.InternalSlice, value)
}
