package laziter

type Iterator[T any] interface {
	NextValue() (T, bool)
	Next() bool
	GetCurrentValue() (T, bool)
	ResetIteratorPosition()
	GetValuesProvider() ValuesProvider[T]
	Close()
}

type ValuesProvider[T any] interface {
	Wait() bool
	Yield(value T)
	Close()
}

type PersitableIterator[T any] struct {
	Provider      *BaseValuesProvider[T]
	InternalSlice []T
	currentIndex  int
	currentValue  T
	hasValue      bool
	persistValues bool
}

func New[T any](persistValues bool) Iterator[T] {
	return &PersitableIterator[T]{
		InternalSlice: []T{},
		persistValues: persistValues,
		currentIndex:  -1,
		Provider: &BaseValuesProvider[T]{
			valuesChannel: make(chan T, 1),
			nextChannel:   make(chan struct{}),
		},
	}
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
	case value, ok = <-i.Provider.valuesChannel:
		// Value already available or closed chan
	case i.Provider.nextChannel <- struct{}{}: // Send Next value signal
		value, ok = <-i.Provider.valuesChannel
	}

	i.hasValue = true
	if ok {
		i.currentValue = value
		if i.persistValues {
			i.store(value)
			i.currentIndex++
		}
	} else if i.persistValues && len(i.InternalSlice) > i.currentIndex+1 {
		// In the case of a reset with stored values
		i.currentIndex++
		i.currentValue = i.InternalSlice[i.currentIndex]
		ok = true
	} else {
		i.hasValue = false
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

func (i *PersitableIterator[T]) ResetIteratorPosition() {
	i.currentIndex = -1
}

func (i *PersitableIterator[T]) GetValuesProvider() ValuesProvider[T] {
	return i.Provider
}

func (i *PersitableIterator[T]) Close() {
	close(i.Provider.nextChannel)
}

func (i *PersitableIterator[T]) store(value T) {
	i.InternalSlice = append(i.InternalSlice, value)
}

type BaseValuesProvider[T any] struct {
	valuesChannel chan T
	nextChannel   chan struct{}
}

func (vp *BaseValuesProvider[T]) Wait() bool {
	_, ok := <-vp.nextChannel
	return ok
}

func (vp *BaseValuesProvider[T]) Yield(value T) {
	vp.valuesChannel <- value
}

func (vp *BaseValuesProvider[T]) Close() {
	close(vp.valuesChannel)
}
