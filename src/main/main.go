package main

import (
	"config"
	"data"
	"fmt"
	"profile"
	"report"
	"strconv"
	"time"
)

func initilaise() {

}

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
	for idxProps, v := range configs.Property {
		/**
		 * Set each data set's properties from the configuration
		 * information pulled from the JSON configuration file
		 */
		sets[idxProps].Id = int64(idxProps)
		sets[idxProps].Property = v
		if sets[idxProps].Property.Now {
			sets[idxProps].Property.Start = time.Now()
		}
		sets[idxProps].Property.Verbose = false
		sets[idxProps].Output.Verbose = false

		// Create a channel for each data set to indicate when it is done
		sets[idxProps].Done = make(chan bool)

		/**
		 * Start the creation of each dataset as a separate go concurrent
		 * process(es)
		 */
		//go sets[idxProps].Create()
		// Open Report with config settings
		sets[idxProps].Report = new(report.TSReport)
		sets[idxProps].Report.Name = "Report_" + strconv.Itoa(idxProps+1)
		sets[idxProps].Report.AddString("Report_" + strconv.Itoa(idxProps+1) + " of " + strconv.Itoa(len(configs.Property)))
		sets[idxProps].Report.AddString("***Config***")
		sets[idxProps].Report.AddStruct(v)

		go sets[idxProps].Create()

	}

	// Wait for each of the data sets to complete before exit
	for idxProps := 0; idxProps < len(configs.Property); idxProps++ {
		<-sets[idxProps].Done
	}

	fmt.Println("Done")
	fmt.Println("Total execution time:", float64(ex.Execute.Elapsed())/1e9, "s")
}
