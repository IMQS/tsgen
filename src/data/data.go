package data

import (
	"config"
	"data/ts"
	"fmt"
	"math"
	"out"
	"profile"
)

// Package constant
const (
	pageCSV  int64 = 131072 //page size at which data gets dumped to file
	pageHTTP int64 = 1      //page size at which data gets batched for HTTP REST API
)

/**
 * Structure describes a data set in full
 * Contains at most a page worth (number of pageSize samples)
 * of the generated series at a time
 */
type TSSet struct {
	Id       int64
	Property config.TSProperties
	Output   out.TSDestination
	Profile  profile.TSProfile

	idxSample int64
	idxStore  int64

	TimeStamp []int64   // Unix nanoseconds
	Value     []float64 // normalised before transform
	State     config.EState

	Done chan bool
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
	set.State = set.Property.State

	set.Output.Type = out.EFormatType(set.Property.Format)
	set.Output.Path = set.Property.Name + "." + set.Property.Format
	set.Output.Name = set.Property.Name
	set.Output.Host = set.Property.Host
	set.Output.Port = set.Property.Port

	// Initialise the destination
	set.Output.Init()
	if set.Property.Verbose {
		fmt.Println(set.Output.Type)
	}

	// Initialise the data set buffer indices
	set.idxSample = 0
	set.idxStore = 0
}

func (set *TSSet) Create() {
	// Initialise the data set
	set.init()

	/**
	 * Create pipe(channel) through which the next time series event will
	 * arrive here, be transformed and routed to the desired form
	 * of output.
	 */
	var feedback profile.TSProfile
	feedback.Execute.Start(1e9)

	set.idxSample = 0
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
				case config.SIN:
					set.sin(0, v.Tn)
				case config.COS:
					set.cos(0, v.Tn)
				case config.LOGIC:
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

			switch set.Property.Mode {
			case config.REAL:
				switch set.Property.Format {
				case "HTTP":
					var to chan bool
					set.Output.Wait(v.Td*set.Property.Duration, to)
				default:
				}
			default:
			}
		}

		if feedback.Execute.IsTimeOut() {
			fmt.Print(".")
			feedback.Execute.Reset(1e9)
		}
		/**
		 * Select (and implement) the output format configured for the
		 * time series.
		 * NB: 	Take note that set type and destination type are not the same
		 *		Set Type is the type of data (transformation applied) in the set
		 *     	Destination type is the format of the output
		 */
		switch set.Output.Type {
		case out.CSV:
			if set.idxSample%pageCSV == 0 {
				set.Process()
			}
		case out.HTTP:
			if set.idxSample%pageHTTP == 0 {
				set.Process()
			}
		default:
		}
		set.idxSample++
	}

	/**
	 * Call on the process routine to write the remainder of the data
	 * that was not handled on a modulus event on page* size
	 * that may still reside within the buffer to disk
	 */
	set.Process()

	feedback.Execute.Stop()
	// Stop moment measurement
	set.Profile.Execute.Stop()

	// Relay information on each generated output
	fmt.Println(set.Output.Path)
	fmt.Println(set.Profile.Execute.Telapsed.Seconds(), "s")

	set.Done <- true
}

func (set *TSSet) Process() {
	// Create space in the file buffer
	set.Output.TimeStamp = make([]int64, len(set.TimeStamp))
	set.Output.Value = make([]float64, len(set.Value))

	// Transfer available data to the file buffer
	copy(set.Output.TimeStamp, set.TimeStamp)
	copy(set.Output.Value, set.Value)

	// Clear the source buffer that sits within
	set.TimeStamp = make([]int64, 0)
	set.Value = make([]float64, 0)

	set.idxStore = set.idxSample
	set.Output.Dump()

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
		if set.Property.Verbose {
			fmt.Print("|", event.Td)
		}
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
	/**
	 * Change the state of the signal only on an identified Toggle event
	 * associated with an absolute tim eni the time series
	 */
	if event.Type == ts.Toggle {
		switch set.State {
		case config.UNDEFINED:
			set.State = config.LOW
		case config.HIGH:
			set.State = config.LOW
		case config.LOW:
			set.State = config.HIGH
		case config.TRI:
			set.State = config.TRI
			// Need to find a way to represent this in data
			set.Value = append(set.Value, 0)
		default:
			set.State = config.LOW
		}
	}
	/**
	 * Set the value according to the current state of the
	 * signal. As long as a state persists, the value for that state should
	 * persist.
	 */
	switch set.State {
	case config.UNDEFINED:
		set.Value = append(set.Value, 0)
	case config.HIGH:
		set.Value = append(set.Value, set.Property.High)
	case config.LOW:
		set.Value = append(set.Value, set.Property.Low)
	case config.TRI:
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
		case config.SIN:
			val += set.Property.Bias[i] + set.sin(i, event.Tn)
		case config.COS:
			val += set.Property.Bias[i] + set.cos(i, event.Tn)
		default:
		}
	}
	set.Value = append(set.Value, val)
}
