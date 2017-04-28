/*
sample standalone app

*/
package main

import (
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
			//todo load default HostConfig
		} else {
			panic(err)
		}
	}
	//todo load HostConfig config file
	_, _ = hc, file
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
