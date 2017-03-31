// +build !windows

package thuder

import (
	"bytes"
	"os/exec"
)

func getDriveID() string {
	buf := bytes.NewBuffer(nil)
	cmd := exec.Command("cmd", "/c", "vol", "c:")
	cmd.Stdout = buf
	err := cmd.Run()
	if err != nil {
		return err.Error()
	}
	bs := buf.Bytes()
	index := bytes.LastIndexByte(bs, ' ')
	return string(bytes.Trim(bs[index:], " \r\n"))
}
