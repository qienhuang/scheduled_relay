package setting

import (
	_ "fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/go-ini/ini"
)

var (
	Cfg *ini.File

	RunMode string

	HTTPPort     int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration

	PageSize  int
	JwtSecret string

	Cron struct {
		Count        int
		RingDuration int
		Specs        []string
	}
)

// Initialize function will be called before main( )
func init() {
	var err error
	Cfg, err = ini.Load("conf/config.ini")
	if err != nil {
		log.Fatal("Fail to parse 'conf/config.ini': %v", err)

	}

	LoadBase()
	LoadServer()
	LoadApp()
	LoadCron()
}

func LoadBase() {
	RunMode = Cfg.Section("").Key("RUN_MODE").MustString("debug")
}

// load web server parameters from config.ini
func LoadServer() {
	sec, err := Cfg.GetSection("server")
	if err != nil {
		log.Fatal("Fail to get section 'server': %v", err)
	}

	RunMode = Cfg.Section("").Key("RUN_MODE").MustString("debug")
	HTTPPort = sec.Key("HTTP_PORT").MustInt(8000)
	ReadTimeout = time.Duration(sec.Key("READ_TIMEOUT").MustInt(60)) * time.Second
	WriteTimeout = time.Duration(sec.Key("WRITE_TIMEOUT").MustInt(60)) * time.Second
}

func LoadApp() {
	sec, err := Cfg.GetSection("app")
	if err != nil {
		log.Fatalf("Fail to get section 'app': %v", err)
	}

	JwtSecret = sec.Key("JWT_SECRET").MustString("!@)*#)!@U#@*!@!)")
	PageSize = sec.Key("PAGE_SIZE").MustInt(10)
}

// load cron scheduler from config.ini
func LoadCron() {

	sec, err := Cfg.GetSection("cron")
	if err != nil {
		log.Panicln("Fail to get section 'cron': %v", err)
		return
	}

	// get ring duration setting
	Cron.RingDuration = sec.Key("ring_duration").MustInt()

	// get all saved cron specifications
	Cron.Specs = nil
	allKeyNames := sec.KeyStrings()
	for _, keyName := range allKeyNames {
		if strings.HasPrefix(keyName, "cron") {
			Cron.Specs = append(Cron.Specs, sec.Key(keyName).String())
		}
	}
}

func UpdateCron(specs []string) {
	// read config.ini section
	sec, err := Cfg.GetSection("cron")
	if err != nil {
		log.Println("Fail to get section 'cron': %v", err)
	}

	// remove all old keys with the prefix "cron"
	allKeyNames := sec.KeyStrings()
	for _, keyName := range allKeyNames {
		if strings.HasPrefix(keyName, "cron") {
			//fmt.Println("key.Name: ", keyName)
			Cfg.Section("cron").DeleteKey(keyName)
		}
	}

	// generate new keys
	for i, spec := range specs {
		sec.NewKey("cron"+strconv.Itoa(i), spec)
	}

	// save to file
	Cfg.SaveTo("conf/config.ini")

}

func UpdateRingDuration(ringDuration int) {
	// read config.ini section
	sec, err := Cfg.GetSection("cron")
	if err != nil {
		log.Println("Fail to get section 'cron': %v", err)
	}
	sec.Key("ring_duration").SetValue(strconv.Itoa(ringDuration))

	// save to file
	Cfg.SaveTo("conf/config.ini")
}
