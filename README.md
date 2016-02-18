# tsgen
Time Series Generator
##Overview
The `tsgen` time series generator 

##Configuration file
The configuration file is a JSON data structure that enables the definition of data sets.  Provided is an example of a configuration file that defines both *CSV* and *HTTP* output.

```json
 {"Property" : 
 	[
 		{
 			"Name": "alarm",
 			"Format" : "HTTP",
 			"Host" : "192.168.4.209",
 			"Port" : 4242,
 			"Mode" : "LOAD",
 			"Batch" : 50,
 			"Sites" : 1,
 			"Distribute" : true,
 			"Spools" : 8,

			"Start" : "2016-02-01T14:00:00+02:00",
			"Now" : true,
			"Seed": 3369,
			"Samples": 10000,
			"Duration" : 86400,
			"Bias" : [12],
			"Type" : ["LOGIC"],

			"Freq" : [10],
			"Amp" : [2],

			"Toggles" : [148],
			"State"	: "LOW",
			"High" : 5,
			"Low" : 0
		},
		{
			"Name": "compound",
			"Format" : "CSV",
			"Mode" : "STORE",
			"Batch" : 131072,

			"Start" : "2016-02-01T14:00:00+02:00",
			"Now" : true,
			"Seed": 3369,
			"Samples": 10000,
			"Duration" : 60,
			"Bias" : [12, 0, 0, 0, 0],
			"Type" : ["SIN", "SIN", "SIN","SIN", "SIN"],

			"Freq": [50,2500, 5000, 15000, 20000, 50000],
			"Amp" : [0.3, 0.1, 0.1, 0.1, 0.05, 0.05]
		}
	]
}
```
The following sections describe the various properties seen in the example configuration file.

###Basic 
The table lists the basic parameters of the configuration file.  Note that their required status is referenced within the GLOBAL scope.

|Parameter|Type|Relevance|Description|Detail|Example|Required
|:---|:---|:---|:---|:---|:---|:---|
|Name|*string*|GLOBAL|Identifier for time series data set||"Name" : "voltage"|Yes|
|Format|*string*|GLOBAL|Output type|*CSV* or *HTTP*|"Format" : "HTTP"|Yes|
|Seed|*int64*|GLOBAL|Seed on which the random source for the time series is based||"Seed" : 3629|Yes|
|Samples|*uint64*|GLOBAL|Number of absolute points in time created for time series||"Samples" : 100000|Yes|
|Duration|*float64*|GLOBAL|Number of seconds over which the time series is spread, starting at **Start** point in time (take into account the **Now** flag)||"Duration" : 60|Yes|
|Start|*time.Time*|GLOBAL|IANA time specifier.  Note the go specific format|Number of nano seconds from 1 Jan 1970|"Start" : "2016-02-01T14:00:00+02:00"|Yes|
|Now|*bool*|GLOBAL|If set (true) the **Start** time is replaced by time.Now()|"Now" : true|Optional|
|Type|*[]ESignal*|GLOBAL|If more than one type is defined in the slice, a compound signal type is inferred.  Requires that the **Bias**, **Freq** and **Amp** properties match the number of elements in the **Type** slice.  ESignal is of type *string*.|SIN and COS and RANDOM currently supported for simple and compound type.  LOGIC is another simple type supported.|"Type" : ["SIN"] or "Type" : ["SIN", "SIN", "COS"].|Optional, default values provided|
|Bias|*[]float64*|GLOBAL|Bias value around which signal transform is to take place.  Signal transform is the sum of the **Bias* and particular signal transform **Type** on **Amp**||"Bias" : [12] or "Bias" : [12, 0.5, 0.2]|Yes|
|Batch|*uint64*|GLOBAL|Number of data points to collect before processing||"Batch" : 50|Optional, default values provided|

###Output **Format**
##CSV
There are no specific configuration items surrounding the *CSV* output format yet.

##HTTP
The table lists the parameters pertaining to a **Format** of the *HTTP* type.  Note that their required status is referenced within the *HTTP* scope, this if *HTTP* scope not defined they may be ommitted completely.

|Parameter|Type|Relevance|Description|Detail|Example|Required|
|:---|:---|:---|:---|:---|:---|:---|
|Host|*string*|**Format** *HTTP*|IP address of target||"Host" : "127.0.0.1"|Yes|
|Port|*int64*|**Format** *HTTP*|Port on which to address target||"Port" : 8080|Yes|
|Mode|*string*|**Format** *HTTP*|Determines the time base on which HTTP requests are released to output|*REAL* , *LOAD* or *STORE*|"Mode" : "LOAD"|Optional|
|Distribute|*bool*|**Format** *HTTP*|Set if time series data points are to be distributed amongst sites ||"Distribute" : true|Optional|
|Sites|*uint64*|**Format** *HTTP*|If **Distribute** is true, time series is distributed between indicated number of **Sites** ||"Sites" : 50000|Optional|
|Spools|*int64*|**Format** *HTTP*|Number of 'concurrent' workers that accept and process jobs for the HTTP output||"Spools" : 8|Yes|

###Transform **Type**
####SIN and COS
|Parameter|Type|Relevance|Description|Detail|Example|Required|
|:---|:---|:---|:---|:---|:---|:---|
|Freq|*[]float64*|**Type** *SIN*, *COS*|Frequency of the generated wave transform||"Freq" : [50] or "Freq" : [50,50,50]|Yes|
|Amp|*[]float64*|**Type** *SIN*, *COS*|Port on which to address target||"Amp" : [24] or "Amp" : [24,12, 230]|Yes|

####LOGIC
|Parameter|Type|Relevance|Description|Detail|Example|Required|
|:---|:---|:---|:---|:---|:---|:---|
|Toggles|*[]uint64*|**Type** *LOGIC*|Number of state toggles that will be imbedded at random in the signal transform.  **Note** that this transform can currently not be used as a **Type** in a compound signal transform, although the toggles are presented in the configuration file as a slice||"Toggles" : [148]|Yes|
|State|*EState*|**Type** *LOGIC*|Starting state of the logic signal.  EState is a *string*.|*UNDEFINED*, *LOW*, and *HIGH* are currently supported|"State" : "LOW"|Yes|
|Low|*float64*|**Type** *LOGIC*|Not to be confused with the *State* *LOW*.  Value of a logic LOW.||"Low" : 0|Yes|
|High|*float64*|**Type** *LOGIC*|Not to be confused with the *State* *HIGH*.  Value of a logic HIGH.||"HIGH" : 3.3|Yes|

####RANDOM
There are no specific configuration items surrounding a random value time series yet.
TODO: Allow for a configurable seed value for the time series transform (currently fixed value)

####Compound
The compound signal is not a specific transform **Type**.  It is an implied type. This is achieved when more than one entry in the **Type** slice is specified.  **Note** that the number of items in the **Bias**, **Freq** and **Amp** properties has to match the number of entries in the **Type** property slice.
The resulting transform is a summation of each of the transforms defined by the aligned entries of the properties identified above.  
It is thus possible to sum a variety of frequency signals on top of different biased offsets, each with a different amplitude and transform type to result in a compound signal.

##Manual switch between target dbase
TODO: Try and find an alternative to having ot manually recompile code to target different data base. 

Currently, unfortunately, some manual intervention is required when targetting different data base from `tsgen` (this should obviously be eradicated ASAP).
To change between target data base, do the following:

1.  Edit the structure type in `out.go` that is associated with the REST membe of the `TSDestination` structure.
  * There are currently three options
    * `TSKairos` for KairosDB
    * `TSOpen` for OpenTSDB
    * `TSNewts`for NewTSDB
  *  Change the following line of code in `out.go` to reflect the target database that is to be used
    *  `REST []rest.TSOpen`

##Deployment
This paragraph aims to describe the process involved in deploying the tsgen and using it to generate time series data that is available on whichever output (**Format**) was opted for.

A procedure for compiling and executing the code:

1.  Clone `IMQS\tsgen` repository to a folder on your system (e.g. `c:\local\go\imqs\tsgen`).
2.  Navigate to the folder to which to clone was made (e.g. `cd c:\local\go\imqs\tsgen`).
3.  Run the `env.bat` file (Windows) to set up environment variables accordingly (e.g. `$ env`).
  *  Alternatively set your `GOPATH` to the repository root of your system folder (e.g. `$ set GOPATH=c:\local\go\imqs\tsgen`).
4.  Navigate to the `src\main` folder (e.g. `$ cd src\main`).
5.  Type `go build` to build the code (e.g. `$ go build`).
6.  The `exe` will be run from the `src\main` folder, thus ensure that the configuration file `conf.JSON` resides in this folder.
7.  Type `go run main.go` to execute code (e.g. `$ go run main.go`).
8.  Alternatively run `go install` in the repository root on your system (e.g. `$ go install`).
  *  If it preferable to use the same configuration file `conf.JSON` that resides within the current working directory, copy the file to the `bin` folder (e.g. `$ copy conf.JSON c:\local\go\imqs\tsgen\bin\`).
  *  Navigate to the `bin` folder that was created in the repository root of your system folder during the install (e.g. `$ cd c:\local\go\imqs\tsgen\bin`).
  *  Execute the `exe`, ensure that the configuration file `conf.JSON` resides within the bin folder (e.g. `$ main`). 
