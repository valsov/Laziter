# Laziter
Lazy iterator experimentation. The aim of this lib is to enable lazy loading an iterator's content. The lazy loading is achieved using a goroutine and two channels per iterator.

## Example

### Base use
```go
func main() {
    iter := laziter.New[int](false) // No persistence needed here
    defer iter.Close()

    // Start values provider goroutine
    go sampleValuesGenerator(iter.GetValuesProvider(), 5)

    // Iterate over values
    for iter.Next() {
		value, _ := iter.GetCurrentValue()
	}
}

func sampleValuesGenerator(vp *laziter.ValuesProvider[int], valuesCount int) {
	defer vp.Close()
	for i := 0; i < valuesCount; i++ {
		if !vp.Wait() {
            // Iterator stopped iteration, quit
			break
		}
		vp.Yield(i) // Produce value
	}
}
```

### Partial iteration
The following example shows a partial iteration with the lazy iterator. Once `sampleFunction()` returns, `sampleValuesGenerator()` goroutine will end, skipping the last 2 values that won't be evaluated (3 & 4).

```go
func sampleFunction() {
    iter := laziter.New[int](false) // No persistence needed here
    defer iter.Close()

    // Start values provider goroutine
    go sampleValuesGenerator(iter.GetValuesProvider(), 5)

    // Get 3 values
    for i := 0; i < 4; i++ {
		value, _ := iter.GetCurrentValue()
	}
}

func sampleValuesGenerator(vp *laziter.ValuesProvider[int], valuesCount int) {
	// [...] (same as above)
}
```

### Values persistence
The current `Iterator` implementation allows values persistence which enables, for example, multiple `for iter.Next() { [...] }` statements.

```go
func main() {
    iter := laziter.New[int](true) // Persistence enabled
    defer iter.Close()

    go sampleValuesGenerator(iter.GetValuesProvider(), 5)

    // Iterate over values once
    for iter.Next() {
		value, _ := iter.GetCurrentValue()
	}

    // Reset the iterator position
    iter.ResetIteratorPosition()
    // From now on, values are fetched from the iterator's persisted storage.
    // sampleValuesGenerator() goroutine already exited.
    
    // Iterate over values again
    for iter.Next() {
		value, _ := iter.GetCurrentValue()
	}
}

func sampleValuesGenerator(vp *laziter.ValuesProvider[int], valuesCount int) {
	// [...] (same as above)
}
```
