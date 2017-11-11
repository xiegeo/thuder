/*
sample standalone app.

create pswd.go with something like the following:

	func init() {
		pswd = "yourpassword3292390"
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

var monitor = flag.Bool("monitor", false, "enables monitoring for new mounts and runs pull and push automatically")

var sleep = time.Second * 5

var logE = logLab.New(os.Stderr, "[thuder err]", logLab.LstdFlags)

// optional build time customizations
var filters []thuder.Filter //set this to set default host filters
var postScript string       //set this to run after pull/push

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

func hostConfig() (*thuder.HostConfig, error) {
	uhn, err := thuder.GenerateUniqueHostname()
	if err != nil {
		return nil, err
	}

	hc := &thuder.HostConfig{}
	fn := filepath.Join(hostConfigPath(), "thuder_host_config.json")
	file, err := os.Open(fn)
	if err != nil {
		if os.IsNotExist(err) {
			//load default HostConfig
			hc.MediaLocation = mediaLocation()
			hc.UniqueHostName = uhn
			hc.Filters = filters
			hc.Group = groupName()
			err = saveFile(fn, hc)
			if err != nil {
				logE.Println(err)
			}
		} else {
			return nil, err
		}
	} else {
		dec := json.NewDecoder(file)
		err = dec.Decode(hc)
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
	}
	return hc, nil
}

func runOnce(hc *thuder.HostConfig) error {
	defer func(a, b io.Writer, c *logLab.Logger) {
		thuder.LogErrorOut = a
		thuder.LogVerbosOut = b
		logE = c
	}(thuder.LogErrorOut, thuder.LogVerbosOut, logE)
	lw := logger(hc)
	thuder.LogErrorOut = lw
	thuder.LogVerbosOut = lw
	logE = logLab.New(lw, "[thuder err]", logLab.LstdFlags)
	fmt.Fprintln(lw, "start thuder ", time.Now())
	defer fmt.Fprintln(lw, "end thuder")

	hc.Authorization = authorize
	mc, err := hc.MediaConfig()
	if err != nil {
		logE.Println("Can not load Media Config", err)
		return err
	}
	if postScript != "" {
		defer func() {
			cmd := exec.Command(postScript)
			cmd.Stdout = lw
			cmd.Stderr = lw
			err := cmd.Run()
			if err != nil {
				logE.Println(err)
			}
		}()
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

//hostConfigPath uses location of executable or current working directory,
//change this to where you want the configuration file to be.
func hostConfigPath() string {
	path, err := os.Executable()
	if err == nil {
		return filepath.Dir(path)
	}

	logE.Println("path name for the executable not supported: ", err)
	path, err = os.Getwd()
	if err != nil {
		panic(err)
	}
	return path
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
