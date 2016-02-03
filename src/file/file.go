package file

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
)

type EFormatType string

const (
	CSV EFormatType = "CSV"
)

func (format *EFormatType) String() string {
	return format.String()
}

type TSFile struct {
	Type EFormatType
	Path string
	Page chan bool
	Done chan bool

	Verbose bool // enable or disable verbose display during create

	TimeStamp []int64
	Value     []float64
	Content   []byte
}

func (file *TSFile) Dump() {
	if file.Verbose {
		fmt.Println("file.Dump")
	}
	file.Format()
	file.Write()
}

func (file *TSFile) Init() {
	//  Always Create the file here
	disk, err := os.Create(file.Path)
	if err != nil {
		fmt.Println("Problem with creating file")
	}
	defer disk.Close()
}

func (file *TSFile) Format() {
	switch file.Type {
	case CSV:
		for idx, v := range file.TimeStamp {
			file.Content = strconv.AppendInt(file.Content, v, 10)
			file.Content = append(file.Content, 44) // comma
			file.Content = strconv.AppendFloat(file.Content, file.Value[idx], 'f', -1, 64)
			file.Content = append(file.Content, 13) //CR
			file.Content = append(file.Content, 10) //LF
		}
	default:

	}
}

func (file *TSFile) Write() {
	if file.Verbose {
		fmt.Println("file.Write")
	}
	if _, err := os.Stat(file.Path); os.IsNotExist(err) {
		_, err = os.Create(file.Path)
	}
	disk, err := os.OpenFile(file.Path, os.O_APPEND, 'a')
	if err != nil {
		fmt.Println("Problem with creating file")
	}
	defer disk.Close()

	switch file.Type {
	case CSV:
		disk.Write(file.Content)
	default:
		writer := csv.NewWriter(disk)
		for idx, value := range file.Value {
			var line = make([]string, 0)
			line = append(line, strconv.FormatInt(file.TimeStamp[idx], 10),
				strconv.FormatFloat(value, 'f', -1, 64))
			err := writer.Write(line)
			if err != nil {
				fmt.Print("Cannot write to file ", err)
			}
		}
		writer.Flush()
		fmt.Println(file.Type)
	}

	file.Content = make([]byte, 0)
	file.TimeStamp = make([]int64, 0)
	file.Value = make([]float64, 0)
}
