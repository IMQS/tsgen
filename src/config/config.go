package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"time"
)

type TSConfig struct {
	Property []TSProperties
}

type TSProperties struct {
	// Set/File
	Name   string // Identifier string for time series data
	Format string // Output file format
	Host   string // Host name in IP address format
	Port   int64  // Port number

	// Content
	Start    time.Time // specified in year, month etc
	Seed     int64     // unitless
	Samples  uint64    // unitless
	Duration float64   // seconds
	Type     []string  // data type
	Compound bool      // Combine different signals to form one

	Bias []float64
	// Type specifics
	// Sin/Cos/Clock
	Freq []float64 // Hz (Hertz) if applicable
	Amp  []float64 // unitless
	// Logic
	Toggles []uint64 // Number of toggles in logic
	State   string   // Start state for logic
	High    float64  // Factor to scale the logic HIGH signal level
	Low     float64  // Factor to scale the logic LOW signal level

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
