package ts

import (
	"fmt"
	"math"
	"math/rand"

	"util"
)

type TSSeries struct {
	Event TSEvent
}

// Enumeration like declaration of output format types
type EEventType int

const (
	None   EEventType = 0
	Toggle EEventType = 1
)

// Complete defnition of a exact point in time for series
type TSEvent struct {
	Tn   float64
	Type EEventType
}

func FixedInterval(seed int64, n uint64, x chan float64) {
	// Generate equispaced timebase
	interval := float64(1.00) / float64(n)
	// To force type (type mismatch in for if assumed)
	var idx uint64
	for idx = 0; idx < n; idx++ {
		val := float64((float64(idx) * interval))
		// Send up the pipe
		x <- val
	}
	close(x)
}

func SpreadInterval(seed int64, n uint64, x chan float64) {
	/**
	 * Generate equispaced timebase framework that defines
	 * the next sample time as a random interval within
	 * the allowed fixed interval, creating a spread interval
	 * with a jitter of +-interval
	 */
	interval := float64(1.00) / float64(n)
	src := rand.NewSource(seed)
	// To force type (type mismatch in for if assumed)
	var idx uint64
	for idx = 0; idx < n; idx++ {
		reach := float64(src.Int63()) / float64(math.MaxInt64)
		val := float64((float64(idx) * interval) + (interval * reach))
		// Send up the pipe
		x <- val
	}
	close(x)
}

func EventSpreadInterval(seed int64, n uint64, ntog uint64, e chan TSEvent) {
	/**
	 * Generate equispaced timebase framework that defines
	 * the next sample time as a random interval within
	 * the allowed fixed interval, creating a spread interval
	 * with a jitter of +-interval
	 */
	interval := float64(1.00) / float64(n)
	baseSpread := rand.NewSource(seed)
	eventSpread := rand.NewSource(math.MaxInt64 - seed)
	// To force type (type mismatch in for if assumed)
	var idx uint64
	var idxEvent uint64
	var onEvents = make([]uint64, 0)
	for idx = 0; idx < ntog; idx++ {
		onEvents = append(onEvents, uint64((float64(eventSpread.Int63())/float64(math.MaxInt64))*float64(n)))
	}
	idxEvent = 0

	//Sort the events so that they retain chronologic order
	onEvents = util.QSortU64(onEvents)
	fmt.Println(onEvents)

	for idx = 0; idx < n; idx++ {
		reach := float64(baseSpread.Int63()) / float64(math.MaxInt64)
		Tn := float64((float64(idx) * interval) + (interval * reach))
		// Create event at this point in the time series
		var event TSEvent
		event.Tn = Tn
		// Define event specific detail here
		if onEvents[idxEvent] == idx {
			event.Type = Toggle
			idxEvent++
		}
		// Send event up the pipe
		e <- event
	}
	close(e)
}
