package thuder

import (
	"bytes"
	"fmt"
	"net"
	"os"
	"path/filepath"
)

//HostConfig is configuration data of the host. Autherizaiton is required before
//a removable media is trusted. Authorization can also modify HostConfig.
type HostConfig struct {
	MediaLocation  string //ie: /media/usb, E:\ ...
	UniqueHostName string //ie: hostname-hardward-ids
	Authorization  func(h *HostConfig) bool
	AllowPulls     []string //approved Pull/backup paths on the host device
	AllowPushes    []string //approved Push/update paths on the host device
}

//UniqueDirectory returns the path to the directory holding data on the
//removable media for this host
func (h *HostConfig) UniqueDirectory() string {
	return filepath.Join(h.MediaLocation, h.UniqueHostName)
}

//DefaultDirectory returns the path to the directory holding data on the
//removable media sharing shared data for all hosts
func (h *HostConfig) DefaultDirectory() string {
	return filepath.Join(h.MediaLocation, "thuder-default")
}

//PullTarget return where data should be saved on the removable media
func (h *HostConfig) PullTarget() string {
	return filepath.Join(h.UniqueDirectory(), "pull")
}

//MediaConfig stores configation data for a removable media
type MediaConfig struct {
	Pulls []string //paths to Pull/backup from
}

// GenerateUniqueHostname generates a human readable, unique name for the current host
func GenerateUniqueHostname() (string, error) {
	var name bytes.Buffer
	hn, err := os.Hostname()
	if err != nil {
		return "", err
	}
	if len(hn) > 12 {
		hn = hn[:12]
	}
	name.WriteString(hn)
	ifs, _ := net.Interfaces()
	for _, v := range ifs {
		fmt.Println(v)
		mac := v.HardwareAddr
		if v.Flags&net.FlagLoopback != 0 {
			continue // no loop backs
		}
		if len(mac) == 0 {
			continue // no macless pseudo-interfaces
		}
		end := len(mac)
		fmt.Fprintf(&name, "_%x", []byte(mac[end-3:end]))
		break //only use the first one
	}
	id, err := GetDriveID()
	if err != nil {
		//skips drive id
		fmt.Println(err)
	} else {
		name.WriteString("_" + id)
	}

	return name.String(), nil
}

//GetDriveID returns the serial number of the local disk. On raspberry pi, it is
// /sys/block/mmcblk0/device/serial. On windows it is returned by "vol c:".
//On unsupperted systems, it returns an error.
func GetDriveID() (string, error) {
	return getDriveID()
}
