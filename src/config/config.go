package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"rest"
	"time"
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

type EMode string

const (
	REAL  EMode = "REAL"
	LOAD  EMode = "LOAD"
	STORE EMode = "STORE"
)

// Defines the constants for the different data types supported
type ESignal string

const (
	SIN    ESignal = "SIN"
	COS    ESignal = "COS"
	LOGIC  ESignal = "LOGIC"
	RANDOM ESignal = "RANDOM"
)

type EState string

const (
	UNDEFINED EState = "UNDEFINED"
	HIGH      EState = "HIGH"
	LOW       EState = "LOW"
	TRI       EState = "TRI"
)

type TSConfig struct {
	Property []TSProperties
}

type TSProperties struct {
	// Set/File
	DBase      rest.EDBaseType
	Name       string      // Identifier string for time series data
	Form       EFormatType // Output format
	Host       string      // Host name in IP address format
	Port       int64       // Port number
	Mode       EMode       // Time based or load
	Batch      uint64      // Number of samples to batch for an HTTP post
	Sites      uint64      // Number of sites to simulate data for
	Distribute bool        // Distribute points between Sites at random
	Spools     int64       // Number of pools to sping up
	Post       bool

	// Content
	Start    time.Time // specified in year, month etc
	Now      bool      // When true overrides the Start time with Now()
	SeedX    int64     // unitless
	Samples  uint64    // unitless
	Duration float64   // seconds
	Type     []ESignal // data type
	Compound bool      // Combine different signals to form one

	Bias []float64
	// Type specifics
	// Sin/Cos/Clock
	Freq []float64 // Hz (Hertz) if applicable
	Amp  []float64 // unitless
	// Logic
	Toggles []uint64 // Number of toggles in logic
	State   EState   // Start state for logic
	High    float64  // Factor to scale the logic HIGH signal level
	Low     float64  // Factor to scale the logic LOW signal level

	SeedY int64 // Seed used to generate the RANDOM type data set

	// Clock
	Duty float64 // Duty cycle of the clock signal

	// Control
	Verbose bool // enable or disable verbose display during create
	Valid   bool // Indicates whether the JSON completely defines a set
}

func Get(url string) TSConfig {
	// Initialise
	configs := TSConfig{}

	file, e := ioutil.ReadFile(url)
	json.Unmarshal(file, &configs)
	if e != nil {
		fmt.Println("error:", e)
	}

	for i := 0; i < len(configs.Property); i++ {
		configs.Property[i].IsCompound()

		configs.Property[i].IsValid()
	}

	return configs
}

func (prop *TSProperties) IsValid() {
	if len(prop.Type) == len(prop.Freq) {
		if len(prop.Type) == len(prop.Amp) {
		}
	}
}

func (prop *TSProperties) IsCompound() {
	if len(prop.Type) > 1 {
		prop.Compound = true
	}
}
