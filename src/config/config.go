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

	// Content
	Start    time.Time // specified in year, month etc
	Seed     int64     // unitless
	Samples  uint64    // unitless
	Duration float64   // seconds
	Type     string    // data type

	// Type specifics
	// Sin/Cos/Clock
	Freq float64 // Hz (Hertz) if applicable
	Amp  float64 // unitless
	// Logic
	Toggle int64 // number of toggles in logic

	// Control
	Verbose bool // enable or disable verbose display during create
}

func Get(url string) TSConfig {
	// Initialise
	configs := TSConfig{}

	file, e := ioutil.ReadFile(url)
	json.Unmarshal(file, &configs)
	if e != nil {
		fmt.Println("error:", e)
	}

	return configs
}
