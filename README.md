# tsgen
Time Series Generator
##Configuration file
The configuration file is a JSON data structure that enables the definition of data sets
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
There is cirrently no specific configuration 

##HTTP
The table lists the parameters pertaining to a **Format** of the *HTTP* type.  Note that their required status is referenced within the *HTTP* scope, this if *HTTP* scope not defined they may be ommitted completely.

|Parameter|Type|Relevance|Description|Detail|Example|Required|
|:---|:---|:---|:---|:---|:---|:---|
|Host|*string*|**Format** *HTTP*|IP address of target||"Host" : "127.0.0.1"|Yes|
|Port|*int64*|**Format** *HTTP*|Port on which to address target||"Port" : 8080|Yes|
|Mode|*string*|**Format** *HTTP*|Determines the time base on which HTTP requests are released to output|*REAL* , *LOAD* or *STORE*|"Mode" : "LOAD"|Optional|
|Distribute|*bool*|**Format** *HTTP*|Set if time series data points are to be distributed amongst sites ||"Sites" : 50000|Optional|
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

##Deployment
This paragraph aims to describe the process involved in deploying the tsgen and using it to generate time series data that is available on whichever output (**Format**) was opted for.
