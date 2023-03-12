package filters

import (
	"streams/store"
)

func MinFilter(min float64, in <-chan store.Entry) <-chan store.Entry {
	outCh := make(chan store.Entry)
	go func() {
		defer close(outCh)
		for entry := range in {
			if entry.Error != nil || entry.Value >= min {
				outCh <- entry
			}
		}
	}()
	return outCh
}
