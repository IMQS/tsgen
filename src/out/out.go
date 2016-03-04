package out

import (
	"config"
	"encoding/csv"
	"fmt"
	"math"
	"math/rand"
	"os"
	"profile"
	"rabbit"
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
	Path      string               // Usually a file path when writing to disk
	Property  *config.TSProperties // Set of properties that fully describe set
	DataPoint []rest.TSWrite       // Structure that describes the DataPoint output in full
	Query     []rest.TSRead        // Structure that describes the Query output in full
	Rabbit    []rabbit.TSQueue

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
		switch dst.Property.Mode {
		case config.LOAD:
			dst.DataPoint = make([]rest.TSWrite, dst.Property.Spools)

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
				dst.DataPoint[id].DBase = dst.Property.DBase
				dst.DataPoint[id].Retry = dst.Property.Retry
				//Enable and initilaise REST latency calculations
				dst.DataPoint[id].Latency.Aggregates.Enable = true
				dst.DataPoint[id].Latency.Aggregates.Ignore = true
				dst.DataPoint[id].Latency.Aggregates.Reset()
				dst.SrcSite[id] = rand.NewSource(12359 % (id + 1))
				go dst.Spool(id)
			}
		case config.QUERY:
			dst.Query = make([]rest.TSRead, dst.Property.Spools)
			dst.SrcSite = make([]rand.Source, dst.Property.Spools)

			dst.Job = 0
			// Initiate the pipes that will control the concurrent workforce
			dst.Jobs = make(chan uint64)
			dst.Done = make(chan int64, dst.Property.Spools)
			// Create each Spool (worker) for concurrent processing of jobs
			var id int64
			for id = 0; id < dst.Property.Spools; id++ {
				dst.Query[id].DBase = dst.Property.DBase
				dst.Query[id].Retry = dst.Property.Retry
				//Enable and initilaise REST latency calculations
				dst.Query[id].Latency.Aggregates.Enable = true
				dst.Query[id].Latency.Aggregates.Ignore = true
				dst.Query[id].Latency.Aggregates.Reset()
				dst.SrcSite[id] = rand.NewSource(18457 % (id + 1))
				go dst.Spool(id)
			}
		default:
		}

	case config.RABBIT:
		for id, name := range dst.Property.Queues {
			dst.Rabbit = append(dst.Rabbit, rabbit.TSQueue{
				name,
				dst.Property.Subscribe[id],
				dst.Property.Enable[id],
				dst.Property.Ack[id],
				dst.Property.Host,
				dst.Property.Port,
				dst.Property.User,
				dst.Property.Pass,
				nil,
				nil,
				nil,
				nil})
			dst.Rabbit[id].Init()
		}

		dst.Job = 0
		// Initiate the pipes that will control the concurrent workforce
		dst.Jobs = make(chan uint64)
		dst.Done = make(chan int64, dst.Property.Spools)
		// Create each Spool (worker) for concurrent processing of jobs
		var id int64
		for id = 0; id < dst.Property.Spools; id++ {
			go dst.Spool(id)
		}
	default:
	}

	fmt.Println(dst.Hdr)

}

func (dst *TSOutput) Finally() {
	switch dst.Property.Form {
	case config.CSV:
	case config.HTTP:
		switch dst.Property.Mode {
		case config.LOAD:
			for idx, value := range dst.DataPoint {
				fmt.Println("Data point aggregates")
				fmt.Print("Latency: ")
				fmt.Printf(
					"%v samples:%v sum:%v avg:%v minidx: %v min:%v maxidx:%v max:%v ms\r\n",
					idx,
					value.Latency.Aggregates.Samples,
					value.Latency.Aggregates.Sum/1e6,
					value.Latency.Aggregates.Avg/1e6,
					value.Latency.Aggregates.SampleMin,
					value.Latency.Aggregates.Min/1e6,
					value.Latency.Aggregates.SampleMax,
					value.Latency.Aggregates.Max/1e6)
				//fmt.Println(value.Latency.Data)

			}
		case config.QUERY:
			for idx, value := range dst.Query {
				fmt.Println("Query aggregates")
				fmt.Print("Latency: ")
				fmt.Printf(
					"%v samples:%v sum:%v avg:%v minidx: %v min:%v maxidx:%v max:%v ms\r\n",
					idx,
					value.Latency.Aggregates.Samples,
					value.Latency.Aggregates.Sum/1e6,
					value.Latency.Aggregates.Avg/1e6,
					value.Latency.Aggregates.SampleMin,
					value.Latency.Aggregates.Min/1e6,
					value.Latency.Aggregates.SampleMax,
					value.Latency.Aggregates.Max/1e6)
				//fmt.Println(value.Latency.Data)
			}
		default:
		}

	case config.RABBIT:
	default:
	}
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

		switch dst.Property.Form {
		case config.HTTP:
			switch dst.Property.Mode {
			case config.LOAD:
				// Do HTTP post for adding data points to tsdb
				dst.DataPoint[id].Add(dst.Property.Host, dst.Property.Port)
			case config.QUERY:
				// Do HTTP post for querying data points from tsdb
				dst.Query[id].Query(dst.Property.Host, dst.Property.Port)
			default:
			}
		case config.RABBIT:
			dst.Publish()
		default:
		}
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
			 * DataPoint does not utilise the Content array, it stores its
			 * commands internally
			 * Structured to POST for each value pair produced by the value
			 * pump
			 */
			switch dst.Property.Mode {
			case config.LOAD:
				dst.SpoolStamp[id] = make([]int64, 0)
				dst.SpoolValue[id] = make([]float64, 0)
				dst.SpoolStamp[id] = make([]int64, len(dst.Stamp))
				dst.SpoolValue[id] = make([]float64, len(dst.Value))
				copy(dst.SpoolStamp[id], dst.Stamp)
				copy(dst.SpoolValue[id], dst.Value)

				dst.DataPoint[id].Init(id, dst.Property.Post)

				for dst.DataPoint[id].IdxSeries = 0; dst.DataPoint[id].IdxSeries < len(dst.SpoolStamp[id]); dst.DataPoint[id].IdxSeries++ {
					if dst.Property.Distribute {
						dst.DataPoint[id].Val = int64((float64(dst.SrcSite[id].Int63()) / float64(math.MaxInt64)) * float64(dst.Property.Sites))
						dst.DataPoint[id].Site = strconv.FormatInt(dst.DataPoint[id].Val, 10)
					}
					dst.DataPoint[id].Tags = map[string]string{"site": "north", "alarm": "none"}
					dst.DataPoint[id].Create(dst.Property.Name, dst.DataPoint[id].Site, dst.SpoolStamp[id][dst.DataPoint[id].IdxSeries], dst.SpoolValue[id][dst.DataPoint[id].IdxSeries], dst.DataPoint[id].Tags)
				}
			case config.QUERY:
				dst.Query[id].Init(id, dst.Property.Post)
				if dst.Property.Distribute {
					dst.Query[id].Val = int64((float64(dst.SrcSite[id].Int63()) / float64(math.MaxInt64)) * float64(dst.Property.Sites))
					dst.Query[id].Site = strconv.FormatInt(dst.Query[id].Val, 10)
				}
				dst.Query[id].Tags = map[string]string{"site": "north", "alarm": "none"}
				dst.Query[id].Create(dst.Property.Name, dst.Query[id].Site, dst.Property.Start, dst.Property.End, dst.Query[id].Tags)
			default:
			}

		case config.RABBIT:

		default:
		}
	}
}

func (dst *TSOutput) Out() {
	switch dst.Property.Form {
	case config.CSV:
		dst.CSV()
	case config.HTTP:
		dst.HTTP()
	case config.RABBIT:
		dst.RabbitMQ()
	default:
		dst.Default()
	}
}

func (dst *TSOutput) CSV() {
	disk := dst.Open()
	defer dst.Close(disk)

	// Data already formatted to content, write to disk
	disk.Write(dst.Content)
	dst.Flush()
}

func (dst *TSOutput) AssignJob() {
	// Add a job for the first available spool to process
	dst.Jobs <- atomic.LoadUint64(&dst.Job)
	atomic.AddUint64(&dst.Job, 1)
}

func (dst *TSOutput) HTTP() {
	dst.AssignJob()
}

func (dst *TSOutput) RabbitMQ() {
	dst.AssignJob()
}

func (dst *TSOutput) Publish() {
	/*
	 * Publish on all of the queues that are subscribed
	 * on this time series
	 */
	for _, queue := range dst.Rabbit {
		for _, pub := range queue.Publish {
			pub.Do([]byte("HALO"), queue.Ack, queue.NAck)
		}
	}
}

func (dst *TSOutput) Default() {
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
