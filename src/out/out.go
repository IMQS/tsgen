package out

import (
	"encoding/csv"
	"fmt"
	//"math"
	"math/rand"
	"os"
	"profile"
	"rest"
	"strconv"
	"time"
)

const (
	csvContent int = 0
)

// Enumeration like declaration of output format types
type EFormatType string

const (
	CSV  EFormatType = "CSV"
	HTTP EFormatType = "HTTP"
)

func (format *EFormatType) String() string {
	return format.String()
}

// Structure completely defines data destination
type TSDestination struct {
	Type EFormatType // Type of output that is addressed in this instance
	Path string      // Usually a file path when writing to disk
	Name string      // The text identifier of the data set
	Host string      // IP address in the formet 192.168.4.194
	Port int64       // Port on which REST API communicates

	Sites      uint64
	Distribute bool
	Spools     int64          // Fully describes the spools
	REST       []rest.TSNewts // Structure that describes the REST output in full
	Job        int64
	Jobs       chan int64 // Manages jobs for spooling
	Done       chan int64 // Manages jobs for spooling

	Verbose bool // enable or disable verbose display during create

	Seed    int64
	srcSite rand.Source

	// CSV
	Hdr     []string // Coloumn headers for CSV output type
	Content []byte   // Formatted content for output
	// HTTP
	Stamp      []int64     // Local buffer for time stamp storage
	Value      []float64   // Local buffer for Value storage
	SpoolStamp [][]int64   // Local buffer for time stamp storage
	SpoolValue [][]float64 // Local buffer for Value storage

}

func (dst *TSDestination) Init() {
	dst.Seed = 100
	dst.srcSite = rand.NewSource(dst.Seed)
	switch dst.Type {
	case CSV:
		dst.Hdr = make([]string, 0)
		var b byte
		for _, b = range []byte("Time") {
			dst.Content = append(dst.Content, b)
		}
		dst.Content = append(dst.Content, 44) // comma
		//strconv.AppendQuoteToASCII(dst.Content, "Value")
		for _, b = range []byte(`Value`) {
			dst.Content = append(dst.Content, b)
		}
		dst.Content = append(dst.Content, 13) //CR
		dst.Content = append(dst.Content, 10) //LF

		//  Always Create the file here
		disk, err := os.Create(dst.Path)
		if err != nil {

		}
		// Close the file so that it may be opened in a different mode
		disk.Close()
	case HTTP:
		dst.REST = make([]rest.TSNewts, dst.Spools)
		dst.SpoolStamp = make([][]int64, dst.Spools)
		dst.SpoolValue = make([][]float64, dst.Spools)
		dst.Job = 0
		dst.Jobs = make(chan int64)
		dst.Done = make(chan int64, dst.Spools)
		var s int64
		for s = 0; s < dst.Spools; s++ {
			go dst.Spool(s)
		}

	default:
	}

	fmt.Println(dst.Hdr)

}

func (dst *TSDestination) Dump() {
	if dst.Verbose {
		fmt.Println("dst.Dump")
	}

	if dst.Type == CSV {
		dst.Format(0)
	}
	dst.Out()

}

func (dst *TSDestination) Spool(id int64) {
	for _ = range dst.Jobs {
		dst.Format(id)
		// Data already formatted to content, use REST API
		dst.REST[id].Add(dst.Host, dst.Port)
		fmt.Print("#", dst.Job, id)
		dst.Job++
	}
	fmt.Print("@", id)
	dst.Done <- id
}

func (dst *TSDestination) Format(id int64) {
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
		switch dst.Type {
		case CSV:
			for idx, v := range dst.Stamp {
				dst.Content = strconv.AppendInt(dst.Content, v, 10)
				dst.Content = append(dst.Content, 44) // comma
				dst.Content = strconv.AppendFloat(dst.Content, dst.Value[idx], 'f', -1, 64)
				dst.Content = append(dst.Content, 13) //CR
				dst.Content = append(dst.Content, 10) //LF
			}
		case HTTP:
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
			dst.REST[id].Init()
			//var metric string
			for idxSeries := 0; idxSeries < len(dst.SpoolStamp[id]); idxSeries++ {
				if dst.Distribute {
					//val := int64((float64(dst.srcSite.Int63()) / float64(math.MaxInt64)) * float64(dst.Sites))
					//metric = strconv.FormatInt(val, 10)
				}
				//dst.REST[id].Create(dst.Name, metric, dst.SpoolStamp[id][idxSeries], dst.SpoolValue[id][idxSeries])
				dst.REST[id].Create(int(id), dst.Name, dst.SpoolStamp[id][idxSeries], dst.SpoolValue[id][idxSeries])
			}
			//fmt.Println(string(dst.REST[id].Json.Bytes()))
		default:
		}
	}
}

func (dst *TSDestination) Out() {
	if dst.Verbose {
		fmt.Println("dst.Write")
	}

	switch dst.Type {
	case CSV:
		disk := dst.Open()
		defer dst.Close(disk)

		// Data already formatted to content, write to disk
		disk.Write(dst.Content)
		dst.Flush()
	case HTTP:
		// Add a job for the first available spool to process
		dst.Jobs <- dst.Job
	default:
		disk := dst.Open()
		defer dst.Close(disk)
		/**
		 * Default CSV format writer for data set implemented as
		 * first order solution but generic enough to use as
		 * default case and data dump if not properly defined in config
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

}

func (dst *TSDestination) Open() *os.File {
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

func (dst *TSDestination) Close(disk *os.File) {
	disk.Close()
}

func (dst *TSDestination) Flush() {
	// Flush the content of the output block and reset the buffers
	switch dst.Type {
	case CSV:
		dst.Content = make([]byte, 0)
		dst.Stamp = make([]int64, 0)
		dst.Value = make([]float64, 0)
	case HTTP:
	default:

	}

}

func (dst *TSDestination) Wait(Td float64, to chan bool) {
	var timeout profile.TSProfile
	timeout.Execute.Start(time.Duration(Td * 1e9))
	for {
		if timeout.Execute.IsTimeOut() {
			break
		}
	}
	//to<-true
}
