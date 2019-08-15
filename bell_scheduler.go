package main

/*
Programming by Kevin Huang
07/2019
kevin11206@gmail.com
web-based bell scheduler(relay timer)
Platform: Raspberry Pi/Debian/Linux
*/

import (
	"bell_scheduler/pkg/setting" //relative to GOPATH, set by: export GOPATH=/your_gopath
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stianeikeland/go-rpio"
	"gopkg.in/robfig/cron.v3"
)

/*
Customer's real ring schedule
--------------------------
	10
	10:10
	12
	12:40
	15
	15:10
	16:55


converted to cron specs
-----------------------------------
		Min	Hr	day	Mon	DayOfWeek
cron1	0   10	*	*	?
cron2	10	10	*	*	?
cron3	0	12	*	*	?
cron4	40	12	*	*	?
cron5	0	15	*	*	?
cron6	10	15	*	*	?
cron7	55	16	*	*	?



*/

/*
[App Structure Diagram]


     App start
        |
        V
   setting.init( )
load settings from config.ini
load cron specifications(string array)
load http server parameters
        |
        V
     main( )
        |
        V
   GPIO initialize
     run crons ----------------------------
	|                                 |
	V                                 V
 initializeRouters( )                 go cron()
         |
	 v
   run http server -----------------------------------------
	 |                          |                       |
         V                          v                       V
    index.html             /api/update_spec      /api/update_ring_duration

*/

//Global variables
var (
	c             *cron.Cron
	jobs          []cron.EntryID //job IDs for controlling
	mutex         sync.Mutex
	pin           rpio.Pin
	ring_duration int = 0
)

type CronSpec struct {
	Index  string `json:"index"` // variable name MUST BE started with a CAPITAL letter
	Hour   string `json:"hour"`
	Minute string `json:"minute"`
	Day    string `json:"day"`
}

func main() {

	//GPIO setup
	err := rpio.Open()
	if err != nil {
		log.Println(err)
	}

	// assign an availabe gpio pin 23
	pin = rpio.Pin(23)
	// Output mode
	//TestGPIO()
	defer rpio.Close()

	//run cron scheduler
	c = cron.New()

	go startCron()
	defer c.Stop()

	//Initialize gin engine, use default settings
	router := gin.Default()
	initializeRouters(router)

	s := &http.Server{
		Addr:           ":" + strconv.Itoa(setting.HTTPPort),
		Handler:        router,
		ReadTimeout:    setting.ReadTimeout,
		WriteTimeout:   setting.WriteTimeout,
		MaxHeaderBytes: 1 << 20,
	}
	s.ListenAndServe()

	router.Run()

	// blocks the app, keep running
	select {}

}

// Thread of each cron schedule, running in background
func startCron() {

	// stop all running schedules
	c.Stop()

	// lock the jobs ID before update it
	mutex.Lock()

	// remove all running jobs
	for _, j := range jobs {
		fmt.Printf("Job ID: %v is removed\n", j)
		c.Remove(j)
	}

	// clearn job IDs
	jobs = nil

	fmt.Println()
	fmt.Println("Cron started for: ")
	fmt.Println("*****************************************************")

	// assign each job
	for i, spec := range setting.Cron.Specs {
		fmt.Println("Cron", strconv.Itoa(i), ": ", spec)

		// pass the value to local variable
		spec := spec
		jobId, _ := c.AddFunc(spec, func() {

			// The schdule could be disabled by setting the duration to 0
			if setting.Cron.RingDuration == 0 {
				return
			}

			pin.Output() // Output mode
			pin.High()   // Set pin High
			log.Println("Ring started. :", spec)
			time.Sleep(time.Duration(setting.Cron.RingDuration) * time.Second)
			pin.Low() // Set pin Low
			log.Println("Ring stoped!  :", spec)

		})

		jobs = append(jobs, jobId)
	}

	// Unlock the job IDs
	mutex.Unlock()

	// Restart cron
	c.Start()

}

func initializeRouters(router *gin.Engine) {

	// Load templates
	router.LoadHTMLGlob("templates/*")
	// Specifies static folder
	router.Static("static", "static")

	// favicon.ico
	router.StaticFile("/favicon.ico", "static/favicon.ico")

	//Authentication
	authorized := router.Group("/", gin.BasicAuth(gin.Accounts{
		"admin": "password",
	}))

	//route setting must be added before running http.Server
	//DEPRECATED： initialize options list for html, using go http/template. changed to jquery method
	/*
		var CronCount []string
		var Hours []string
		var Minutes []string
		var Day []string

		initializeHtmlList(&CronCount, &Hours, &Minutes, &Day)

		strJson := cronSpecsToJson(setting.Cron.Specs)

		authorized.GET("/", func(c *gin.Context) {
			c.HTML(http.StatusOK, "index.html", gin.H{
				"title": "Ring schedule", "Hours": &Hours, "Minutes": &Minutes, "Day": &Day,
				"CronCount": &CronCount, "CronSpecs": strJson,
			})
		})
	*/

	authorized.GET("/", func(c *gin.Context) {

		strJson := cronSpecsToJson(setting.Cron.Specs)
		c.HTML(http.StatusOK, "index.html", gin.H{
			"title":        "Ring schedule",
			"CronSpecs":    strJson,
			"RingDuration": `{"Duration":"` + strconv.Itoa(setting.Cron.RingDuration) + `"}`,
		})
	})

	// an api for testing
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	// Update schedule api
	authorized.POST("/api/update_spec", func(c *gin.Context) {
		/*DEMO:
		// read content body raw data
		raw, _ := c.GetRawData()
		log.Println("RAW: ", string(raw))

		log.Println("FromValue: ", c.Request.FormValue("Hour"))
		log.Println("FromValue: ", c.Request.FormValue("minute"))

		log.Println("PostForm: ", c.PostForm("Hour"))

		contentType := c.ContentType()
		log.Println("contentType: ", contentType)

		*/

		var cs []CronSpec
		err := c.Bind(&cs) // or use BindJson()
		if err != nil {
			log.Println("Binding err: ", err)

			c.JSON(200, gin.H{
				"message": "Data binding error: " + err.Error(),
			})
			return
		}

		setting.Cron.Specs = jsonToCronSpecs(cs)
		setting.UpdateCron(setting.Cron.Specs)

		startCron()
		//log.Println("POS: ", cs[3].Hour)

		c.JSON(200, gin.H{
			"message": "Update successful!",
		})
	})

	authorized.POST("/api/update_ring_duration", func(c *gin.Context) {
		var _json struct {
			Duration string `json:"duration"` // variable names MUST BE started with a CAPITAL letter
		}

		err := c.Bind(&_json)
		if err != nil {
			log.Println("Binding err: ", err)
			c.JSON(200, gin.H{
				"message": "Update failed!",
			})
			return
		}

		log.Println("Update ring duration to :", _json.Duration)
		// update config.ini
		setting.Cron.RingDuration, _ = strconv.Atoi(_json.Duration)
		setting.UpdateRingDuration(setting.Cron.RingDuration)

		// restart cron scheduler
		startCron()
		//log.Println("Get ring_duration: ", _json.Duration)

		c.JSON(200, gin.H{
			"message": "Update ring duration successful!",
		})
	})
}

// coverter Cron Specification from string to json for jquery
func cronSpecsToJson(cronSpecs []string) string {

	var JsonCronSpecs string
	var cronSpec CronSpec
	var arrayCronSpec []CronSpec
	/* EXAMPLE:
			Sec	Min	Hr	day	Mon	DayOfWeek
	cron1	0   0   10	*	*	?
	cron2	0	10	10	*	*	?
	*/
	for i, spec := range cronSpecs {
		fields := strings.Fields(spec)
		if len(fields) != 5 {
			continue //Invalid fields
		}

		// Index
		if i < 10 {
			cronSpec.Index = "0" + strconv.Itoa(i)
		} else {
			cronSpec.Index = strconv.Itoa(i)
		}

		// Minute
		if len(fields[0]) == 1 {
			cronSpec.Minute = "0" + fields[0]
		} else {
			cronSpec.Minute = fields[0]
		}

		// Hour
		if len(fields[1]) == 1 {
			cronSpec.Hour = "0" + fields[1]
		} else {
			cronSpec.Hour = fields[1]
		}

		// Day
		switch fields[4] {
		case "1-5":
			cronSpec.Day = "Monday To Friday"
		case "6":
			cronSpec.Day = "Saturday"
		case "0":
			cronSpec.Day = "Sunday"
		case "0-6":
			cronSpec.Day = "All days"

		}

		arrayCronSpec = append(arrayCronSpec, cronSpec)

		//log.Printf("minute: %v hour: %v day: %v", cronSpec.Minute, cronSpec.Hour, cronSpec.Day)

	}

	jsonObj, _ := json.Marshal(arrayCronSpec)

	JsonCronSpecs = string(jsonObj)
	//log.Println("JsonCronSpecs:", JsonCronSpecs)

	return JsonCronSpecs

}

// converter json string to gocron specs
func jsonToCronSpecs(json []CronSpec) []string {
	/* EXAMPLE OF GOCRON SPECS:
			Sec	Min	Hr	day	Mon	DayOfWeek
	cron1	0   0   10	*	*	?
	cron2	0	10	10	*	*	?
	*/
	// clear slice
	var specs []string
	var min, hr, day string
	for _, schedule := range json {
		//log.Println("spec: ", schedule.Index, "hour: ", schedule.Hour)
		if len(schedule.Minute) == 1 {
			min = schedule.Minute
		} else {

			min = strings.TrimPrefix(schedule.Minute, "0")
		}

		if len(schedule.Hour) == 1 {
			hr = schedule.Hour
		} else {
			hr = strings.TrimPrefix(schedule.Hour, "0")
		}

		switch schedule.Day {
		case "Monday To Friday":
			day = "1-5"
		case "Saturday":
			day = "6"
		case "Sunday":
			day = "0"
		case "All days":
			day = "0-6"
		}

		specs = append(specs, min+"	"+hr+"	*	*	"+day)

	}
	fmt.Println()
	fmt.Println("Got update from web browser: ")
	fmt.Println("-----------------------------------------------")
	for _, str := range specs {
		fmt.Println(str)
	}

	return specs
}

func TestGPIO() {

	/* For high-trigger relay */
	pin.Output() // Output mode
	pin.High()   // Set pin High
	time.Sleep(time.Duration(2) * time.Second)
	pin.Low() // Set pin Low
	//pin.Toggle()

	/* For low-trigger relay */
	pin.Output() // Output mode
	pin.Low()    // Set pin High
	log.Println("Ring start")
	time.Sleep(time.Duration(1) * time.Second)
	pin.High() // Set pin Low
	log.Println("Ring stop")
	//pin.Toggle()
}

// DEPRECATED: Generating data for html droplist
func initializeHtmlList(cronCount *[]string, hours *[]string, minutes *[]string, days *[]string) {

	for i := 0; i < len(setting.Cron.Specs); i++ {
		*cronCount = append(*cronCount, strconv.Itoa(i))
	}

	for i := 0; i <= 23; i++ {
		if i < 10 {
			*hours = append(*hours, "0"+strconv.Itoa(i))
		} else {
			*hours = append(*hours, strconv.Itoa(i))
		}
	}

	for i := 0; i <= 59; i++ {
		if i < 10 {
			*minutes = append(*minutes, "0"+strconv.Itoa(i))
		} else {
			*minutes = append(*minutes, strconv.Itoa(i))
		}
	}
	*days = []string{"Monday To Friday", "Saturday", "Sunday", "All days"}
}
