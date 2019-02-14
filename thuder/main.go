/*
Sample standalone app.

Create pswd.go with something like the following:

	func init() {
		pswd = "yourpassword3292390"
		thuder.SetPinID(17) //to use pin 17 as light indicator
		filters = []thuder.Filter{...} //filters for what oprations are allowed by the host
		postScript = "..." //a commad to run after files are synchronized.
	}

*/
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	logLab "log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"github.com/xiegeo/thuder"
)

var monitor = flag.Bool("monitor", false, "Enables monitoring for new mounts and runs pull and push automatically.")

var hostConfigName = flag.String("host_config", "", "Set the path to the read and write host config. "+
	"For security purpurses, this file should not be on the same storage device that thuder is backing up to, "+
	"as such is equivalent to allowing all operations listed in that device. "+
	"Default value is empty, which disables using config file from overwriting build time settings.")

var sleep = time.Second * 5 //How often to pull mediaLocation to detect new devices.

var logE = logLab.New(os.Stderr, "[thuder err]", logLab.LstdFlags)

// optional build time customizations
var filters []thuder.Filter //set this to set default host filters
var postScripts []string    //set this to run after pull/push

func main() {
	flag.Parse()
	if !*monitor {
		hc, err := hostConfig()
		if err != nil {
			panic(err)
		}
		runOnce(hc)
		return
	}
	thuder.FlashLED() //flash once for monitoring on
	for ; ; time.Sleep(sleep) {
		hc, err := hostConfig()
		if err != nil {
			continue
		}
		_, err = os.Open(hc.DefaultDirectory())
		if err != nil {
			//fmt.Println(err)
			if !os.IsNotExist(err) {
				logE.Println(err)
			}
			continue
		}
		runOnce(hc)
		fmt.Println("waiting for media to be removed")
		for err == nil {
			time.Sleep(time.Second)
			_, err = os.Open(hc.DefaultDirectory())
		}
		fmt.Println("removed: ", err)
	}
}

//loadDefault loads the default HostConfig
func loadDefault() (*thuder.HostConfig, error) {
	hc := &thuder.HostConfig{}
	hc.MediaLocation = mediaLocation()
	uhn, err := thuder.GenerateUniqueHostname()
	if err != nil {
		return nil, err
	}
	hc.UniqueHostName = uhn
	hc.Filters = filters
	hc.Group = groupName()
	return hc, nil
}

func hostConfig() (*thuder.HostConfig, error) {
	fn := *hostConfigName
	if fn == "" {
		return loadDefault()
	}
	file, err := os.Open(fn)
	if err != nil {
		if os.IsNotExist(err) {
			//load and save default HostConfig
			hc, err := loadDefault()
			if err != nil {
				return nil, err
			}
			err = saveFile(fn, hc)
			if err != nil {
				logE.Println(err)
			}
			return hc, nil
		}
		return nil, err
	}
	dec := json.NewDecoder(file)
	hc := &thuder.HostConfig{}
	err = dec.Decode(hc)
	if err != nil {
		return nil, err
	}
	//UniqueHostName does not match expected, the file could have been copied from
	//a different system. Fix this to avoid name collision.
	uhn, err := thuder.GenerateUniqueHostname()
	if err != nil {
		return nil, err
	}
	if hc.UniqueHostName != uhn {
		hc.UniqueHostName = uhn
		err = saveFile(fn, hc)
		if err != nil {
			logE.Println(err)
		}
	}
	return hc, nil
}

func runOnce(hc *thuder.HostConfig) error {
	defer func(a, b io.Writer, c *logLab.Logger) {
		thuder.LogErrorOut = a
		thuder.LogVerboseOut = b
		logE = c
	}(thuder.LogErrorOut, thuder.LogVerboseOut, logE)
	lw := logger(hc)
	thuder.LogErrorOut = lw
	thuder.LogVerboseOut = lw
	logE = logLab.New(lw, "[thuder err]", logLab.LstdFlags)
	fmt.Fprintln(lw, "start thuder ", time.Now())
	defer fmt.Fprintln(lw, "end thuder")

	hc.Authorization = authorize
	mc, err := hc.MediaConfig()
	if err != nil {
		logE.Println("Can not load Media Config", err)
		return err
	}
	for i := range postScripts {
		defer func(postScript string) {
			cmd := exec.Command(postScript)
			cmd.Stdout = lw
			cmd.Stderr = lw
			err := cmd.Run()
			if err != nil {
				logE.Println(err)
			}
		}(postScripts[i])
	}
	fmt.Fprintln(lw, mc)
	err = thuder.PullAndPush(hc, mc)
	if err != nil {
		logE.Println("Failed ", err)
		return err
	}
	return nil
}

func saveFile(fn string, v interface{}) error {
	data, err := json.MarshalIndent(v, "", "\t")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(fn, data, 0644)
}

//groupName is set here based on os and arch, so that different pathes and
//binaries can be used for cross platform support. groupName can be changed to
//use environment values for group based specializations.
func groupName() string {
	return runtime.GOOS + "-" + runtime.GOARCH
}

//mediaLocation is where removable device is mounted, it could be replaced by
//a command-line flag if using a launcher with more intelligence.
func mediaLocation() string {
	if os.PathSeparator == '/' {
		return "/media/usb" //by usbmount
	}
	return "E:\\" //windows
}

var pswd = ""

//authorize your removable device. You must customize this function
func authorize(hc *thuder.HostConfig) bool {
	if pswd == "" {
		panic("please init pswd in a new pswd.go file," +
			" or rewite authorize to use a different method")
	}
	p, err := ioutil.ReadFile(filepath.Join(hc.DefaultDirectory(), "pswd"))
	if err != nil {
		logE.Println(err)
		return false
	}
	return (string)(p) == pswd
}
