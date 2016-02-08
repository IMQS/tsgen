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
	Sin      string = "Sin"
	Cos      string = "Cos"
	Logic    string = "Logic"
	Clock    string = "Clock"
	Setpoint string = "Setpoint"
	Complex  string = "Complex"
)

// Package constant
const (
	pageCSV  int64 = 131072 //page size at which data gets dumped to file
	pageHTTP int64 = 100    //page size at which data gets batched for HTTP REST API
)

type EState string

const (
	UNDEFINED EState = "UNDEFINED"
	HIGH      EState = "HIGH"
	LOW       EState = "LOW"
	TRI       EState = "TRI"
)

/**
 * Structure describes a data set in full
 * Contains at most a page worth (number of pageSize samples)
 * of the generated series at a time
 */
type TSSet struct {
	Id       int64
	Property config.TSProperties
	Dest     out.TSDestination

	idx  int64
	idxP int64

	TimeStamp []int64   // Unix nanoseconds
	Value     []float64 // normalised before transform
	State     EState

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
	set.State = EState(set.Property.State)

	set.Dest.Type = out.EFormatType(set.Property.Format)
	set.Dest.Path = set.Property.Name + "." + set.Property.Format
	set.Dest.Name = set.Property.Name
	set.Dest.Host = set.Property.Host
	set.Dest.Port = set.Property.Port

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

	/**
	 * Create pipe(channel) through which the next time series event will
	 * arrive here, be transformed and routed to the desired form
	 * of output.
	 */
	set.idx = 0
	var e = make(chan ts.TSEvent)
	go set.event(e)
	/**
	 * Implement data transform on normalised time base series.
	 * In time transform on values as they are made available through
	 * the piping system that drills down to time series generation.
	 */
	for v := range e {
		set.stamp(v.Tn)
		if len(set.Property.Type) <= 0 {

		} else {
			if len(set.Property.Type) <= 1 {
				switch set.Property.Type[0] {
				case Sin:
					set.sin(0, v.Tn)
				case Cos:
					set.cos(0, v.Tn)
				case Logic:
					set.logic(&v)
				default:
				}
			} else {
				/**
				 * Use multiple definitions for type parameter to
				 * build a complex signal
				 */
				set.compound(&v)
			}
		}

		/**
		 * Select (and implement) the output format configured for the
		 * time series.
		 * NB: 	Take note that set type and destination type are not the same
		 *		Set Type is the type of data (transformation applied) in the set
		 *     	Destination type is the format of the output
		 */
		switch set.Dest.Type {
		case out.CSV:
			if set.idx%pageCSV == 0 {
				set.Process()
			}
		case out.HTTP:
			if set.idx%pageHTTP == 0 {
				set.Process()
				fmt.Printf("%d:%d|", set.Id, set.idx)

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
	if set.Dest.Type == out.CSV {
		set.Process()
	}

	// Stop moment measurement
	set.Profile.Execute.Stop()

	// Relay information on each generated output
	fmt.Println(set.Dest.Path)
	fmt.Println(set.Profile.Execute.Telapsed.Seconds(), "s")

	set.Done <- true
}

func (set *TSSet) Process() {
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

func (set *TSSet) event(x chan ts.TSEvent) {
	/**
	 * Create pipe(channel) through which the next time series event will
	 * be pumped to the surface
	 */
	var e = make(chan ts.TSEvent)
	// Start separate process(es) to construct time series event
	fmt.Println(set.Property.Toggles)
	go ts.EventSpreadInterval(&set.Property, e)

	set.TimeStamp = make([]int64, 0)
	set.Value = make([]float64, 0)
	for event := range e {
		// Send event up the pipe
		x <- event
	}
	close(x)
}

func (set *TSSet) clear() {
	set.TimeStamp = make([]int64, 0)
	set.Value = make([]float64, 0)
}

func (set *TSSet) stamp(v float64) {
	nano := int64(v * set.Property.Duration * 1e9)
	set.TimeStamp = append(set.TimeStamp, set.Property.Start.UnixNano()+nano)
}

func (set *TSSet) sin(idx int, v float64) float64 {
	var val float64 = set.Property.Amp[idx] * math.Sin((2*math.Pi)*(v*set.Property.Duration)*(set.Property.Freq[idx]))
	if set.Property.Compound {
		// No specific action as of yet
	} else {
		set.Value = append(set.Value, set.Property.Bias[idx]+val)
	}
	return val
}

func (set *TSSet) cos(idx int, v float64) float64 {
	var val float64 = set.Property.Amp[idx] * math.Cos((2*math.Pi)*(v*set.Property.Duration)*(set.Property.Freq[idx]))
	if set.Property.Compound {
		// No specific action as of yet
	} else {
		set.Value = append(set.Value, set.Property.Bias[idx]+val)
	}
	return val
}

func (set *TSSet) logic(event *ts.TSEvent) {
	if event.Type == ts.Toggle {
		switch set.State {
		case UNDEFINED:
			set.State = LOW
		case HIGH:
			set.State = LOW
		case LOW:
			set.State = HIGH
		case TRI:
			set.State = TRI
			// Need to find a way to represent this in data
			set.Value = append(set.Value, 0)
		default:
			set.State = LOW
		}
	}
	switch set.State {
	case UNDEFINED:
		set.Value = append(set.Value, 0)
	case HIGH:
		set.Value = append(set.Value, set.Property.High)
	case LOW:
		set.Value = append(set.Value, set.Property.Low)
	case TRI:
		// Need to find a way to represent this in data
		set.Value = append(set.Value, 0)
	default:
		set.Value = append(set.Value, set.Property.Low)
	}
}

func (set *TSSet) compound(event *ts.TSEvent) {
	var val float64 = float64(0)
	for i, v := range set.Property.Type {
		switch v {
		case Sin:
			val += set.Property.Bias[i] + set.sin(i, event.Tn)
		case Cos:
			val += set.Property.Bias[i] + set.cos(i, event.Tn)
		default:
		}
	}
	set.Value = append(set.Value, val)
}
