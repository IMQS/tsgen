package profile

import (
	"math"
	"time"
)

type TSAggregate struct {
	Enable    bool
	Ignore    bool
	Samples   int64
	Sum       float64
	Avg       float64
	Max       float64
	SampleMax int64
	Min       float64
	SampleMin int64
}

type TSProfile struct {
	Execute    TSMoment
	Data       []int64
	Aggregates TSAggregate
}

type TSMoment struct {
	T0       time.Time       // Start time of measurement
	Telapsed time.Duration   // Elapsed time since measurement
	Tremain  time.Duration   // Remaining time in countdown/timeout
	Tlap     time.Time       // Stores the elapsed time on call to calculate lap time
	Tout     time.Time       // Timeout trigger value
	Laps     []time.Duration // Stores an array of lap times
}

func (agg *TSAggregate) Reset() {
	agg.Sum = 0
	agg.Avg = 0
	agg.Max = 0
	agg.Min = math.MaxFloat64
}

func (agg *TSAggregate) Calculate(idx int, value float64) {
	agg.Samples += 1
	agg.Sum += value
	agg.Avg = agg.Sum / float64(agg.Samples)
	if value > agg.Max {
		agg.Max = value
		agg.SampleMax = int64(idx)
	}
	if value < agg.Min {
		agg.Min = value
		agg.SampleMin = int64(idx)
	}
}

func (pro *TSProfile) Reset() {
	pro.Aggregates.Reset()
}

func (pro *TSProfile) Start(value int64) {

}

func (pro *TSProfile) Append(value int64) {
	pro.Data = append(pro.Data, value)
	if pro.Aggregates.Enable {
		pro.Aggregate()
	}
}

func (pro *TSProfile) Aggregate() {
	var idx int = len(pro.Data) - 1
	if pro.Aggregates.Ignore {
		if idx == 0 {

		} else {
			pro.Aggregates.Calculate(idx, float64(pro.Data[idx]))
		}
	} else {
		pro.Aggregates.Calculate(idx, float64(pro.Data[idx]))
	}
}

func (pro *TSProfile) Sum() int64 {
	var sum int64
	for _, value := range pro.Data {
		sum += value
	}
	return sum
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

func (m *TSMoment) TimeOut(d time.Duration) {
	m.Start(d)
	for {
		if m.IsTimeOut() {
			break
		} else {
		}
	}
	m.Stop()
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
