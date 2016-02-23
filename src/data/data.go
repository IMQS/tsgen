package data

import (
	"config"
	"data/ts"
	"fmt"
	"math"
	"math/rand"
	"out"
	"profile"
	"report"
	"strconv"
	"sync/atomic"
)

// Package constant
const (
	pageCSV  uint64 = 131072 //page size at which data gets dumped to file
	pageHTTP uint64 = 1      //page size at which data gets batched for HTTP REST API
)

/**
 * Structure describes a data set in full
 * Contains at most a page worth (number of pageSize samples)
 * of the generated series at a time
 */
type TSSet struct {
	Id       int64               // Unique identifier of data set
	Property config.TSProperties // Set of properties that fully describe set
	Output   out.TSOutput        // Fully describes the type of output for set
	Profile  profile.TSProfile   // Tool to do simple profiling on code

	idxSample int64 // Index of the current sample being processed in set
	idxStore  int64 // Index of previous sample processed in set
	srcBase   rand.Source
	srcSign   rand.Source

	Stamp []int64       // Unix nanoseconds
	Value []float64     // normalised before transform
	State config.EState // Start up state for LOGIC signals

	Report *report.TSReport // Report of test results

	Done chan bool // Each set acknowledges when it has completed
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
	set.Output.Property = &set.Property

	set.Output.Path = set.Property.Name + "." + string(set.Property.Form)

	set.srcBase = rand.NewSource(set.Property.SeedY)
	set.srcSign = rand.NewSource(math.MaxInt64 - set.Property.SeedY)

	// Initialise the destination
	set.Output.Init()
	if set.Property.Verbose {
		fmt.Println(set.Output.Property.Form)
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
				case config.RANDOM:
					set.random(0)
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
				switch set.Property.Form {
				case "HTTP":
					var to chan bool
					set.Output.Wait((v.Td*1.0)*set.Property.Duration, to)
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
		switch set.Output.Property.Form {
		case config.CSV:
			if set.Property.Batch <= 0 {
				set.Property.Batch = pageCSV
			}
		case config.HTTP:
			if set.Property.Batch <= 0 {
				set.Property.Batch = pageHTTP
			}
		default:
		}

		if set.idxSample%int64(set.Property.Batch) == 0 {
			set.Process()
		}
		set.idxSample++
	}

	/**
	 * Call on the process routine to write the remainder of the data
	 * that was not handled on a modulus event on page* size
	 * that may still reside within the buffer to disk
	 */
	set.Process()

	switch set.Output.Property.Form {
	case config.CSV:
	case config.HTTP:
		// No more jobs, all samples have been processed
		close(set.Output.Jobs)

		var s int64
		// Wait on all Spools to spin down
		for s = 0; s < set.Property.Spools; s++ {
			<-set.Output.Done
		}
		fmt.Println("Spools")

		var j uint64
		var jobs uint64
		jobs = (set.Property.Samples / set.Property.Batch) + 1
		// Wait on all Jobs to complete
		for j = 0; j < jobs; j++ {
			<-set.Output.Jobs
		}

	default:
	}

	feedback.Execute.Stop()
	// Stop moment measurement
	set.Profile.Execute.Stop()

	// Relay information on each generated output
	fmt.Println(set.Output.Path)
	fmt.Println(set.Profile.Execute.Telapsed.Seconds(), "s")

	fmt.Println("Jobs:", atomic.LoadUint64(&set.Output.Job))
	set.Report.AddString("\n" + strconv.FormatFloat(set.Profile.Execute.Telapsed.Seconds(), 'f', 6, 64) + " s")
	set.Report.AddString("Jobs:" + strconv.FormatUint(atomic.LoadUint64(&set.Output.Job), 16))
	set.Report.Create()
	set.Done <- true
}

func (set *TSSet) Process() {
	// Create space in the file buffer
	set.Output.Stamp = make([]int64, len(set.Stamp))
	set.Output.Value = make([]float64, len(set.Value))

	copy(set.Output.Stamp, set.Stamp)
	copy(set.Output.Value, set.Value)

	//fmt.Println(set.Output.Stamp)
	//fmt.Println(set.Output.Value)

	// Clear the source buffer that sits within
	set.Stamp = make([]int64, 0)
	set.Value = make([]float64, 0)

	// Reset index offset
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

	set.Stamp = make([]int64, 0)
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
	set.Stamp = make([]int64, 0)
	set.Value = make([]float64, 0)
}

func (set *TSSet) stamp(v float64) {
	nano := int64(v * set.Property.Duration * 1e9)
	set.Stamp = append(set.Stamp, set.Property.Start.UnixNano()+nano)

}

func (set *TSSet) random(idx int) float64 {
	reach := float64(set.srcBase.Int63()) / float64(math.MaxInt64)
	sign := float64(set.srcSign.Int63()) / float64(math.MaxInt64)
	var val float64
	if sign <= 0.5 {
		val = -1.00 * reach * set.Property.Amp[idx]
	} else {
		val = reach * set.Property.Amp[idx]
	}

	if set.Property.Compound {
		// No specific action as of yet
	} else {
		set.Value = append(set.Value, set.Property.Bias[idx]+val)
	}
	return val
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
		case config.RANDOM:
			val += set.Property.Bias[i] + set.random(i)
		default:
		}
	}
	set.Value = append(set.Value, val)
}
