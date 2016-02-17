package ts

import (
	"config"
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
	Idx  uint64  // Sample index as a cumulative counter (max: Property.Samples)
	Tn   float64 // Current point in time
	Tp   float64 // Previous point in time
	Td   float64 // Duration between points in time
	Type EEventType
}

func EventFixedInterval(config *config.TSProperties, e chan TSEvent) {
	if config.Samples <= 0 {

	} else {
		interval := float64(1.00) / float64(config.Samples)
		if interval <= 0 {

		} else {

		}
	}

}

func EventSpreadInterval(config *config.TSProperties, e chan TSEvent) {
	/**
	 * Generate equispaced timebase framework that defines
	 * the next sample time as a random interval within
	 * the allowed fixed interval, creating a spread interval
	 * with a jitter of +-interval
	 */
	interval := float64(1.00) / float64(config.Samples)
	baseSpread := rand.NewSource(config.Seed)
	nodeSpread := rand.NewSource(math.MaxInt64 - config.Seed)
	// To force type (type mismatch in for if assumed)
	var idx uint64
	var idxEvent uint64
	var Events = make([]uint64, 0)

	if len(config.Toggles) <= 0 {
		// Index out of range
	} else {
		if len(config.Toggles) <= 1 {
			for idx = 0; idx < config.Toggles[0]; idx++ {
				Events = append(Events, uint64((float64(nodeSpread.Int63())/float64(math.MaxInt64))*float64(config.Samples)))
			}
		} else {
			// Determine whether it would ever be necessary using multiple toggles
		}
	}

	idxEvent = 0
	var pevent TSEvent
	var event TSEvent
	// Call on utility function to sort Events that they can be used in chronologic order
	Events = util.QSortU64(Events)
	for idx = 0; idx < config.Samples; idx++ {
		reach := float64(baseSpread.Int63()) / float64(math.MaxInt64)

		if reach < 0.1 {
			reach = 0.1
		}
		if reach > 0.9 {
			reach = 0.9
		}
		Tn := float64((float64(idx) * interval) + (interval * reach))

		// Create event at this point in the time series, with default values
		event = TSEvent{Idx: 0, Tn: 0.00, Tp: 0.00, Td: 0.00, Type: None}
		event.Idx = idx
		// Has to be assigned to before calculating the duration between points
		event.Tn = Tn
		// Calculate duration (Td)
		if (&pevent) == nil {
			event.Tp = Tn
		} else {
			/**
			 * Use the previous point's absolute time as the previous time of the
			 * current event
			 */
			event.Tp = pevent.Tn
			event.Td = event.Tn - event.Tp
			/** NB NOTE: Store a copy of the last event in order to be able to
			 * calculate the diffrence in absolute times between points
			 */
			pevent = TSEvent{Idx: event.Idx, Tn: event.Tn, Tp: event.Tp, Td: event.Td, Type: event.Type}
		}

		if idxEvent < uint64(len(Events)) {
			if len(Events) <= 0 {
				// No events need be linked to any point in time
			} else {
				// Define event specific detail here
				if Events[idxEvent] == idx {
					event.Type = Toggle
					idxEvent++
				}
			}
		}
		// Send event up the pipe
		e <- event
	}

	close(e)
}
