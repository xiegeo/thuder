package thuder

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"

	"github.com/bmatcuk/doublestar"
)

var (
	//ErrNeedAuthorizationFunction : the action required Authorization to be set
	ErrNeedAuthorizationFunction = errors.New("authorization function required")
	//ErrFailedAuthorization : authorization function returned false
	ErrFailedAuthorization = errors.New("authorization function returned false")
)

//HostConfig is configuration data of the host. Autherizaiton is required before
//a removable media is trusted. Authorization can also modify HostConfig.
type HostConfig struct {
	MediaLocation  string                   //ie: /media/usb, E:\ ...
	UniqueHostName string                   //ie: hostname-hardward-ids
	Authorization  func(h *HostConfig) bool `json:"-"`
	AllowPulls     []string                 //approved Pull/backup paths on the host device
	AllowPushes    []string                 //approved Push/update paths on the host device
	Group          string                   //used to select different default configeration files on the removable media
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

//MediaConfig runs the Authorization delegate and loads MediaConfig from
func (h *HostConfig) MediaConfig() (*MediaConfig, error) {
	if !filepath.IsAbs(h.MediaLocation) {
		return nil, ErrBadPath
	}
	fi, err := fs.Stat(h.MediaLocation)
	if err != nil {
		return nil, err
	}
	if !fi.IsDir() {
		return nil, ErrNeedDir
	}

	fn := h.Group + ".MediaConfig.json"
	if h.Authorization == nil {
		return nil, ErrNeedAuthorizationFunction
	}
	if !h.Authorization(h) {
		return nil, ErrFailedAuthorization
	}

	if len(h.UniqueHostName) == 0 {
		h.UniqueHostName, err = GenerateUniqueHostname()
		if err != nil {
			return nil, err
		}
	}

	mc, err := LoadMediaConfig(filepath.Join(h.UniqueDirectory(), fn))
	if err != nil && os.IsNotExist(err) {
		mc, err = LoadMediaConfig(filepath.Join(h.DefaultDirectory(), fn))
	}
	if err != nil {
		return nil, err
	}
	var errs, errs2 []error
	mc.Pulls, errs = filterPathes(mc.Pulls, h.AllowPulls)
	mc.Pushes, errs2 = filterPathes(mc.Pushes, h.AllowPushes)
	if len(errs) != 0 || len(errs2) != 0 {
		b := bytes.NewBuffer(nil)
		for _, e := range append(errs, errs2...) {
			fmt.Fprintln(b, e)
		}
		return mc, errors.New(b.String())
	}
	return mc, nil
}

//filterPathes returns any path allowed, and any allows partterns with ErrBadPattern as error
func filterPathes(pathes, allows []string) ([]string, []error) {
	var out []string
	errmap := make(map[int]error)
	var errs []error
	for _, path := range pathes {
		for i, allow := range allows {
			if errmap[i] != nil {
				continue
			}
			m, err := doublestar.PathMatch(allow, path)
			if err != nil {
				if err == doublestar.ErrBadPattern {
					err = errors.New(allow)
					errmap[i] = err
				}
				errs = append(errs, err)
				continue
			}
			if m {
				out = append(out, path)
				break
			}
		}
	}
	return out, errs
}

//MediaConfig stores configation data for a removable media
type MediaConfig struct {
	Pulls  []string //paths to Pull/backup from
	Pushes []string //paths to Push/update to the host device
}

//LoadMediaConfig loads MediaConfig from path
func LoadMediaConfig(path string) (*MediaConfig, error) {
	f, err := fs.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	dec := json.NewDecoder(f)
	mc := &MediaConfig{}
	err = dec.Decode(mc)
	if err != nil {
		return nil, err
	}
	return mc, nil
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
