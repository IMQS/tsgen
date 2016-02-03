package data

import (
	"config"
	"data/ts"
	"fmt"
	"math"
	"out"
	"profile"
)

// Defines the constants for the different data types supported
const (
	Sin   string = "Sin"
	Cos   string = "Cos"
	Logic string = "Logic"
	Clock string = "Clock"
)

// Package constant
const (
	pageCSV int64 = 131072 //page size at which data gets dumped to file
)

/**
 * Structure describes a data set in full
 * Contains at most a page worth (number of pageSize samples)
 * of the generated series at a time
 */
type TSSet struct {
	Property config.TSProperties
	Dest     out.TSDestination

	idx       int64
	idxP      int64
	Event     []ts.TSEvent // Unix nanoseconds
	TimeStamp []int64      // Unix nanoseconds
	Value     []float64    // normalised before transform

	Done     chan bool
	Pause    chan bool
	Continue chan bool

	Profile profile.TSProfile
}

func radians(deg float64) float64 {
	return deg * (math.Pi / 180.0)
}

func degrees(rad float64) float64 {
	return rad * (180.0 / math.Pi)
}

func (set *TSSet) init() {
	// Start moment measurement
	set.Profile.Execute.Start(0)

	set.Dest.Type = out.EFormatType(set.Property.Format)
	set.Dest.Path = set.Property.Name + "." + set.Property.Format

	// Initialise the destination
	set.Dest.Init()
	if set.Property.Verbose {
		fmt.Println(set.Dest.Type)
	}

	// Initialise the data set buffer indices
	set.idx = 0
	set.idxP = 0
}

func (set *TSSet) Create() {

	// Initiate File IO
	if set.Property.Verbose {
		fmt.Println("Create")
	}

	// Initialise the data set
	set.init()

	var e = make(chan ts.TSEvent)
	go set.event(e)
	/**
	 * Create pipe(channel) through which the next time series event will
	 * arrive here, be transformed and routed to the desired form
	 * of output.
	 */
	var x = make(chan float64)
	// Start separate process(es) that generates the exact point in time
	go set.time(x)
	/**
	 * Implement data transform on normalised time base series.
	 * In time transform on values as they are made available through
	 * the piping system that drills down to time series generation.
	 */
	for v := range x {
		set.stamp(v)
		switch set.Property.Type {
		case Sin:
			set.sin(v)
		case Cos:
			set.cos(v)
		default:
		}

		/**
		 * Select (and implement) the output format configured for the
		 * time series.
		 */
		switch set.Dest.Type {
		case out.CSV:
			if set.idx%pageCSV == 0 {
				set.Store()
			}
		default:
		}
		set.idx++
	}
	/**
	 * Call on the store routine to write the remainder of the data
	 * that was not handled on a modulus event on page* size
	 * that may still reside within the buffer to disk
	 */
	set.Store()

	// Stop moment measurement
	set.Profile.Execute.Stop()

	// Relay information on each generated output
	fmt.Println(set.Dest.Path)
	fmt.Println(set.Profile.Execute.Telapsed.Seconds(), "s")

	set.Done <- true
}

func (set *TSSet) Store() {
	if set.Property.Verbose {
		fmt.Println("data.Store")
	}

	// Create space in the file buffer
	set.Dest.TimeStamp = make([]int64, len(set.TimeStamp))
	set.Dest.Value = make([]float64, len(set.Value))

	// Transfer available data to the file buffer
	copy(set.Dest.TimeStamp, set.TimeStamp)
	copy(set.Dest.Value, set.Value)

	// Clear the source buffer that sits within
	set.TimeStamp = make([]int64, 0)
	set.Value = make([]float64, 0)

	set.idxP = set.idx
	set.Dest.Dump()

}

func (set *TSSet) time(x chan float64) {
	/**
	 * Create pipe(channel) through which the next time series value will
	 * be pumped to the surface
	 */
	var t = make(chan float64)
	// Start separate process(es) to construct time series value
	go ts.SpreadInterval(set.Property.Seed, set.Property.Samples, t)

	set.TimeStamp = make([]int64, 0)
	set.Value = make([]float64, 0)
	for val := range t {
		x <- val
	}
	close(x)
}

func (set *TSSet) event(x chan ts.TSEvent) {
	/**
	 * Create pipe(channel) through which the next time series event will
	 * be pumped to the surface
	 */
	var e = make(chan ts.TSEvent)
	// Start separate process(es) to construct time series event
	go ts.EventSpreadInterval(set.Property.Seed, set.Property.Samples, 410, e)

	set.clear()
	for event := range e {
		// Send event up the pipe
		x <- event
	}
	close(x)
}

func (set *TSSet) clear() {
	set.TimeStamp = make([]int64, 0)
	set.Value = make([]float64, 0)
	set.Event = make([]ts.TSEvent, 0)
}

func (set *TSSet) stamp(v float64) {
	nano := int64(v * set.Property.Duration * 1e9)
	set.TimeStamp = append(set.TimeStamp, set.Property.Start.UnixNano()+nano)
}

func (set *TSSet) sin(v float64) {
	set.Value = append(set.Value, set.Property.Amp*math.Sin((2*math.Pi)*(v*set.Property.Duration)*(set.Property.Freq)))
}

func (set *TSSet) cos(v float64) {
	set.Value = append(set.Value, set.Property.Amp*math.Cos((2*math.Pi)*(v*set.Property.Duration)*(set.Property.Freq)))
}
