package data

import (
	"config"
	"data/ts"
	"file"
	"fmt"
	"math"
)

const (
	Sin   string = "Sin"
	Cos   string = "Cos"
	Logic string = "Logic"
	Block string = "Block"
)

const (
	pageSize int64 = 131072
)

type TSSet struct {
	Property config.TSProperties
	File     file.TSFile

	idx       int64
	idxP      int64
	TimeStamp []int64   // Unix nanoseconds
	Value     []float64 // normalised before transform

	Done     chan bool
	Pause    chan bool
	Continue chan bool
}

func radians(deg float64) float64 {
	return deg * (math.Pi / 180.0)
}

func degrees(rad float64) float64 {
	return rad * (180.0 / math.Pi)
}

func (set *TSSet) Create() {
	// Initiate File IO
	if set.Property.Verbose {
		fmt.Println("Create")
	}
	set.File.Type = file.EFormatType(set.Property.Format)
	set.File.Path = set.Property.Name + "." + set.Property.Format
	set.File.Init()
	if set.Property.Verbose {
		fmt.Println(set.File.Type)
	}
	fmt.Println(set.File.Path)

	set.idx = 0
	set.idxP = 0
	var x = make(chan float64)
	go set.time(x)
	for v := range x {
		switch set.Property.Type {
		case Sin:
			set.sin(v)
		case Cos:

		default:
		}

		if set.idx%pageSize == 0 {
			set.Store()
		}
		set.idx++

	}
	set.Store()
	set.Done <- true
}

func (set *TSSet) Store() {
	if set.Property.Verbose {
		fmt.Println("data.Store")
	}

	// Create space in the file buffer
	set.File.TimeStamp = make([]int64, len(set.TimeStamp))
	set.File.Value = make([]float64, len(set.Value))

	// Transfer available data to the file buffer
	copy(set.File.TimeStamp, set.TimeStamp)
	copy(set.File.Value, set.Value)

	// Clear the source buffer that sits within
	set.TimeStamp = make([]int64, 0)
	set.Value = make([]float64, 0)

	set.idxP = set.idx
	set.File.Dump()

}

func (set *TSSet) time(x chan float64) {
	// Here you can switch between different methods of generating X
	var c = make(chan float64)
	go ts.SpreadInterval(set.Property.Seed, set.Property.Samples, c)
	set.TimeStamp = make([]int64, 0)
	set.Value = make([]float64, 0)
	for val := range c {
		x <- val
	}
	close(x)
}

func (set *TSSet) clear() {
	set.TimeStamp = make([]int64, set.Property.Samples)
	set.Value = make([]float64, set.Property.Samples)
}

func (set *TSSet) display(nano int64) {
	if set.Property.Verbose {
		fmt.Println(set.idx, set.TimeStamp[(set.idx-set.idxP)], float64(nano)/1e9, set.Value[(set.idx-set.idxP)])
	}
}

func (set *TSSet) stamp(idx int64, v float64) int64 {
	nano := int64(v * set.Property.Duration * 1e9)
	set.TimeStamp = append(set.TimeStamp, set.Property.Start.UnixNano()+nano)
	return nano
}

func (set *TSSet) sin(v float64) {
	nano := set.stamp(int64(set.idx), v)
	set.Value = append(set.Value, set.Property.Amp*math.Sin((2*math.Pi)*(v*set.Property.Duration)*(set.Property.Freq)))
	set.display(nano)
}
