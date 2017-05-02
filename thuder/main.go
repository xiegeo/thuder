/*
sample standalone app

*/
package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	logLab "log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/xiegeo/thuder"
)

var logE = logLab.New(os.Stderr, "[thuder]", logLab.LstdFlags)

func main() {
	hc := &thuder.HostConfig{}
	fn := filepath.Join(hostConfigPath(), "thuder_host_config.json")
	file, err := os.Open(fn)
	if err != nil {
		if os.IsNotExist(err) {
			//load default HostConfig
			hc.MediaLocation = mediaLocation()
			hc.UniqueHostName, err = thuder.GenerateUniqueHostname()
			if err != nil {
				panic(err)
			}
			hc.Group = groupName()
			err = saveFile(fn, hc)
			if err != nil {
				logE.Println(err)
			}
		} else {
			panic(err)
		}
	} else {
		dec := json.NewDecoder(file)
		err = dec.Decode(hc)
		if err != nil {
			panic(err)
		}
	}
	hc.Authorization = authorize
	mc, err := hc.MediaConfig()
	if err != nil {
		logE.Println("Can not load Media Config", err)
		return
	}
	fmt.Println(mc)
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

//authorize your removable device. You must customize this function
func authorize(hc *thuder.HostConfig) bool {
	p, err := ioutil.ReadFile(filepath.Join(hc.DefaultDirectory(), "pswd"))
	if err != nil {
		logE.Println(err)
		return false
	}
	return (string)(p) == pswd //please define pswd in a new pswd.go file,
	// or rewite authorize to use a different method
}
