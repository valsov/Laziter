package main

type Iterator[T any] interface {
	NextValue() (T, bool)
	Next() bool
	GetCurrentValue() (T, bool)
	Reset() // todo: need index
	GetValuesProvider() ValuesProvider[T]
}

type ValuesProvider[T any] interface {
	Yield(value T) bool
	Close()
}

func New[T any](persistValues bool) Iterator[T] {
	return &PersitableIterator[T]{
		InternalSlice: []T{},
		persistValues: persistValues,
		Provider: &BaseValuesProvider[T]{
			valuesChannel: make(chan T),
			nextChannel:   make(chan struct{}),
			quit:          make(chan struct{}),
		},
	}
}

type PersitableIterator[T any] struct {
	Provider      *BaseValuesProvider[T]
	InternalSlice []T
	persistValues bool
	currentValue  T
	hasValue      bool
}

func (i *PersitableIterator[T]) NextValue() (T, bool) {
	if !i.Next() {
		var value T
		return value, false
	}
	return i.GetCurrentValue()
}

func (i *PersitableIterator[T]) Next() bool {
	var value T
	ok := true

	select {
	case <-i.Provider.quit:
		return false
	case value, ok = <-i.Provider.valuesChannel:
		// Value already available or closed chan
	case i.Provider.nextChannel <- struct{}{}: // Send Next value signal
		value, ok = <-i.Provider.valuesChannel
	}

	if ok {
		i.hasValue = true
		i.currentValue = value
		if i.persistValues {
			i.store(value)
		}
	}
	return ok
}

func (i *PersitableIterator[T]) GetCurrentValue() (T, bool) {
	if i.hasValue {
		return i.currentValue, true
	}
	var defaultVal T
	return defaultVal, false
}

func (i *PersitableIterator[T]) GetValuesProvider() ValuesProvider[T] {
	return i.Provider
}

func (i *PersitableIterator[T]) store(value T) {
	i.InternalSlice = append(i.InternalSlice, value)
}

type BaseValuesProvider[T any] struct {
	valuesChannel     chan T
	nextChannel, quit chan struct{}
}

func (vp *BaseValuesProvider[T]) Yield(value T) bool {
	select {
	case <-vp.nextChannel:
		vp.valuesChannel <- value
		return true
	case <-vp.quit:
		return false
	}
}

func (vp *BaseValuesProvider[T]) Close() {
	close(vp.valuesChannel)
	// todo: Quit ?
}
