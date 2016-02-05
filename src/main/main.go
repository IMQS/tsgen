package main

import (
	"config"
	"data"
	"fmt"
	"profile"
)

var voltage data.TSSet

func main() {

	var ex profile.TSProfile
	ex.Execute.Start(0)
	/**
	 * Get the configuration information and set the properties
	 * of each data set
	 */
	var configs = config.Get("conf.json")
	for _, config := range configs.Property {
		fmt.Println(config)
	}

	// Initialise each data set
	var sets = make([]data.TSSet, len(configs.Property))
	// Iterate through each data set as defined by its properties
	for idx, v := range configs.Property {
		/**
		 * Set each data set's properties from the configuration
		 * information pulled from the JSON configuration file
		 */
		sets[idx].Id = int64(idx)
		sets[idx].Property = v
		sets[idx].Property.Verbose = false
		sets[idx].Dest.Verbose = false

		// Create a channel for each data set to indicate when it is done
		sets[idx].Done = make(chan bool)

		/**
		 * Start the creation of each dataset as a separate go concurrent
		 * process(es)
		 */
		go sets[idx].Create()
	}

	// Wait for each of the data sets to complete before exit
	for idx := 0; idx < len(configs.Property); idx++ {
		<-sets[idx].Done
	}

	fmt.Println("Done")
	fmt.Println(float64(ex.Execute.Elapsed())/1e9, "s")
}
