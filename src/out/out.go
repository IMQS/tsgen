package out

import (
	"config"
	"encoding/csv"
	"fmt"
	"math"
	"math/rand"
	"os"
	"profile"
	"rest"
	"strconv"
	"sync/atomic"
	"time"
)

const (
	csvContent int = 0
)

// Structure completely defines data destination
type TSOutput struct {
	Path     string               // Usually a file path when writing to disk
	Property *config.TSProperties // Set of properties that fully describe set
	REST     []rest.TSDBase       // Structure that describes the REST output in full

	Job  uint64
	Jobs chan uint64 // Manages jobs for spooling
	Done chan int64  // Manages jobs for spooling

	Verbose bool // enable or disable verbose display during create

	SrcSite []rand.Source

	// CSV
	Hdr     []string // Coloumn headers for CSV output type
	Content []byte   // Formatted content for output
	// HTTP
	Stamp      []int64     // Local buffer for time stamp storage
	Value      []float64   // Local buffer for Value storage
	SpoolStamp [][]int64   // Local buffer for time stamp storage
	SpoolValue [][]float64 // Local buffer for Value storage

}

func (dst *TSOutput) Init() {
	switch dst.Property.Form {
	case config.CSV:
		// Create header row for CSV file
		dst.Hdr = make([]string, 0)
		var b byte
		for _, b = range []byte("Time") {
			dst.Content = append(dst.Content, b)
		}
		dst.Content = append(dst.Content, 44) // comma

		for _, b = range []byte(`Value`) {
			dst.Content = append(dst.Content, b)
		}
		dst.Content = append(dst.Content, 13) //CR
		dst.Content = append(dst.Content, 10) //LF

		//  Always Create the file here
		disk, err := os.Create(dst.Path)
		if err != nil {
			fmt.Println("File not created.")
			os.Exit(3)
		}
		// Close the file so that it may be opened in a different mode
		disk.Close()
	case config.HTTP:

		// Initialise REST request/query
		dst.REST = make([]rest.TSDBase, dst.Property.Spools)

		/**
		 * Create random source for each spool from different seed
		 * Initially made the mistake to regenerate the source for eacg
		 * batch process, thus there were never more than batch number of
		 * different sites
		 */
		dst.SrcSite = make([]rand.Source, dst.Property.Spools)

		// Create buffers that recevie data from signal transform pipe
		dst.SpoolStamp = make([][]int64, dst.Property.Spools)
		dst.SpoolValue = make([][]float64, dst.Property.Spools)
		// Initiate the job counter as one of the verification metrics
		dst.Job = 0
		// Initiate the pipes that will control the concurrent workforce
		dst.Jobs = make(chan uint64)
		dst.Done = make(chan int64, dst.Property.Spools)
		// Create each Spool (worker) for concurrent processing of jobs
		var id int64
		for id = 0; id < dst.Property.Spools; id++ {
			dst.REST[id].DBase = dst.Property.DBase
			dst.SrcSite[id] = rand.NewSource(12359 % (id + 1))
			go dst.Spool(id)
		}

	default:
	}

	fmt.Println(dst.Hdr)

}

func (dst *TSOutput) Dump() {
	switch dst.Property.Form {
	case config.CSV:
		/**
		 * No concurrent workers for writing to CSV file implemented.
		 * Files are created concurrently to one another, but the complexity
		 * of managing order of writes is just not worth the effort
		 */
		dst.Format(0)
	default:
	}
	// Create the output
	dst.Out()

}

func (dst *TSOutput) Spool(id int64) {
	for _ = range dst.Jobs {
		// Prepare samples for output
		dst.Format(id)
		// Do HTTP request for adding data points to tsdb
		dst.REST[id].Add(dst.Property.Host, dst.Property.Port)
		// Atomically increase job counter for verification
		atomic.AddUint64(&dst.Job, 1)
	}
	fmt.Print("@", id)
	dst.Done <- id
}

func (dst *TSOutput) Format(id int64) {
	/*
	 * Implement formatting for set of data made available to
	 * the output according to the format type specifier config item
	 */
	if len(dst.Stamp) != len(dst.Value) {
		/**
		 * There is no corresponding y value for each
		 * independant x value and thus the series
		 * has not been transformed correctly
		 * and can not be sent to any form of output.
		 */
	} else {
		switch dst.Property.Form {
		case config.CSV:
			for idx, v := range dst.Stamp {
				dst.Content = strconv.AppendInt(dst.Content, v, 10)
				dst.Content = append(dst.Content, 44) // comma
				dst.Content = strconv.AppendFloat(dst.Content, dst.Value[idx], 'f', -1, 64)
				dst.Content = append(dst.Content, 13) //CR
				dst.Content = append(dst.Content, 10) //LF
			}
		case config.HTTP:
			/**
			 * REST does not utilise the Content array, it stores its
			 * commands internally
			 * Structured to POST for each value pair produced by the value
			 * pump
			 */
			dst.SpoolStamp[id] = make([]int64, 0)
			dst.SpoolValue[id] = make([]float64, 0)
			dst.SpoolStamp[id] = make([]int64, len(dst.Stamp))
			dst.SpoolValue[id] = make([]float64, len(dst.Value))
			copy(dst.SpoolStamp[id], dst.Stamp)
			copy(dst.SpoolValue[id], dst.Value)
			//dst.Flush()
			dst.REST[id].Init(id, dst.Property.Post)

			for dst.REST[id].IdxSeries = 0; dst.REST[id].IdxSeries < len(dst.SpoolStamp[id]); dst.REST[id].IdxSeries++ {
				if dst.Property.Distribute {
					dst.REST[id].Val = int64((float64(dst.SrcSite[id].Int63()) / float64(math.MaxInt64)) * float64(dst.Property.Sites))
					dst.REST[id].Site = strconv.FormatInt(dst.REST[id].Val, 10)
				}
				dst.REST[id].Tags = map[string]string{"site": "north", "alarm": "none"}
				// Kairos DB & OpenTS DB
				dst.REST[id].Create(dst.Property.Name, dst.REST[id].Site, dst.SpoolStamp[id][dst.REST[id].IdxSeries], dst.SpoolValue[id][dst.REST[id].IdxSeries], dst.REST[id].Tags)
			}
			//fmt.Println(string(dst.REST[id].Json.Bytes()))
		default:
		}
	}
}

func (dst *TSOutput) Out() {
	switch dst.Property.Form {
	case config.CSV:
		dst.CSV()
	case config.HTTP:
		// Add a job for the first available spool to process
		dst.Jobs <- dst.Job
	default:
		dst.defaultCSV()
	}
}

func (dst *TSOutput) defaultCSV() {
	disk := dst.Open()
	defer dst.Close(disk)
	/**
	 * Default CSV format writer for data set implemented as
	 * first order solution but generic enough to use as
	 * default case and data dump if not properly defined in config.
	 * Note that for each record a write is initiated which makes
	 * this a very slow way of creating the output.
	 */
	w := csv.NewWriter(disk)
	for idx, value := range dst.Value {
		var line = make([]string, 0)
		line = append(line, strconv.FormatInt(dst.Stamp[idx], 10),
			strconv.FormatFloat(value, 'f', -1, 64))
		err := w.Write(line)
		if err != nil {
			fmt.Print("Cannot write to file ", err)
		}
	}
	w.Flush()
}

func (dst *TSOutput) CSV() {
	disk := dst.Open()
	defer dst.Close(disk)

	// Data already formatted to content, write to disk
	disk.Write(dst.Content)
	dst.Flush()
}

func (dst *TSOutput) Open() *os.File {
	// Test whether file already exist
	if _, err := os.Stat(dst.Path); os.IsNotExist(err) {
		_, err = os.Create(dst.Path)
	}

	// Append time series data to destination
	disk, err := os.OpenFile(dst.Path, os.O_APPEND, 'a')
	if err != nil {
		fmt.Println("Problem with creating file")
	}
	return disk
}

func (dst *TSOutput) Close(disk *os.File) {
	disk.Close()
}

func (dst *TSOutput) Flush() {
	// Flush the content of the output block and reset the buffers
	switch dst.Property.Form {
	case config.CSV:
		dst.Content = make([]byte, 0)
		dst.Stamp = make([]int64, 0)
		dst.Value = make([]float64, 0)
	case config.HTTP:
	default:

	}
}

func (dst *TSOutput) Wait(Td float64, to chan bool) {
	var timeout profile.TSProfile
	timeout.Execute.Start(time.Duration(Td * 1e9))
	for {
		if timeout.Execute.IsTimeOut() {
			break
		}
	}
	//to<-true
}
