package out

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
)

// Enumeration like declaration of output format types
type EFormatType string

const (
	CSV EFormatType = "CSV"
)

func (format *EFormatType) String() string {
	return format.String()
}

// Structure completely defines data destination
type TSDestination struct {
	Type EFormatType
	Path string

	Page chan bool
	Done chan bool

	Verbose bool // enable or disable verbose display during create

	TimeStamp []int64
	Value     []float64
	Content   []byte
}

func (dst *TSDestination) Dump() {
	if dst.Verbose {
		fmt.Println("dst.Dump")
	}
	dst.Format()
	dst.Write()
}

func (dst *TSDestination) Init() {
	//  Always Create the file here
	disk, err := os.Create(dst.Path)
	if err != nil {
		fmt.Println("Problem with creating file")
	}
	defer disk.Close()
}

func (dst *TSDestination) Format() {
	/*
	 * Implement formatting for set of data made available to
	 * the output according to the format type specifier config item
	 */
	if len(dst.TimeStamp) != len(dst.Value) {
		/**
		 * There is no corresponding y value for each
		 * independant x value and thus the series
		 * has not been transformed correctly
		 * and can not be sent to any form of output.
		 */
	} else {
		switch dst.Type {
		case CSV:
			for idx, v := range dst.TimeStamp {
				dst.Content = strconv.AppendInt(dst.Content, v, 10)
				dst.Content = append(dst.Content, 44) // comma
				dst.Content = strconv.AppendFloat(dst.Content, dst.Value[idx], 'f', -1, 64)
				dst.Content = append(dst.Content, 13) //CR
				dst.Content = append(dst.Content, 10) //LF
			}
		default:
		}
	}
}

func (dst *TSDestination) Write() {
	if dst.Verbose {
		fmt.Println("dst.Write")
	}

	// Test whether file already exist
	if _, err := os.Stat(dst.Path); os.IsNotExist(err) {
		_, err = os.Create(dst.Path)
	}

	// Append time series data to destination
	disk, err := os.OpenFile(dst.Path, os.O_APPEND, 'a')
	if err != nil {
		fmt.Println("Problem with creating file")
	}
	defer disk.Close()

	switch dst.Type {
	case CSV:
		// Data already formatted to content, write to disk
		disk.Write(dst.Content)
	default:
		/**
		 * Default CSV format writer for data set implemented as
		 * first order solution but generic enough to use as
		 * default case and data dump if not properly defined in config
		 */
		writer := csv.NewWriter(disk)
		for idx, value := range dst.Value {
			var line = make([]string, 0)
			line = append(line, strconv.FormatInt(dst.TimeStamp[idx], 10),
				strconv.FormatFloat(value, 'f', -1, 64))
			err := writer.Write(line)
			if err != nil {
				fmt.Print("Cannot write to file ", err)
			}
		}
		writer.Flush()
	}
	dst.Flush()
}

func (dst *TSDestination) Flush() {
	// Flush the content of the output block and reset the buffers
	dst.Content = make([]byte, 0)
	dst.TimeStamp = make([]int64, 0)
	dst.Value = make([]float64, 0)
}
