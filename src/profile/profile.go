package profile

import (
	"time"
)

type TSProfile struct {
	Execute TSMoment
}

type TSMoment struct {
	T0       time.Time       // Start time of measurement
	Telapsed time.Duration   // Elapsed time since measurement
	Tremain  time.Duration   // Remaining time in countdown/timeout
	Tlap     time.Time       // Stores the elapsed time on call to calculate lap time
	Tout     time.Time       // Timeout trigger value
	Laps     []time.Duration // Stores an array of lap times
}

func (m *TSMoment) Start(d time.Duration) {
	// Initialise laps array
	m.Laps = make([]time.Duration, 0)

	// Initialise moment timer
	m.T0 = time.Now()
	m.Tout = time.Date(m.T0.Year(), m.T0.Month(), m.T0.Day(), m.T0.Hour(), m.T0.Minute(), m.T0.Second(), m.T0.Nanosecond(), m.T0.Location())
	m.Tremain = d
	m.Tout = m.Tout.Add(m.Tremain)

	m.Tlap = time.Date(m.T0.Year(), m.T0.Month(), m.T0.Day(), m.T0.Hour(), m.T0.Minute(), m.T0.Second(), m.T0.Nanosecond(), m.T0.Location())
}

func (m *TSMoment) Lap() {
	m.Laps = append(m.Laps, time.Now().Sub((m.Tlap)))
	// There might be some loss of accuracy right here
	m.Tlap = time.Now()
}

func (m *TSMoment) Elapsed() int64 {
	m.Telapsed = time.Now().Sub((m.T0))
	return m.Telapsed.Nanoseconds()
}

func (m *TSMoment) IsTimeOut() bool {
	return time.Now().After(m.Tout)
}

func (m *TSMoment) Stop() {
	m.Elapsed()
	m.Tremain = 0
}

func (m *TSMoment) Reset(d time.Duration) {
	// Reset existing durations
	m.Telapsed = 0
	m.Tremain = 0
	m.Start(d)
}
