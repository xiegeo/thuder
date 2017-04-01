package thuder

import (
	"bytes"
	"os/exec"
)

func getDriveID() (string, error) {
	buf := bytes.NewBuffer(nil)
	cmd := exec.Command("cmd", "/c", "vol", "c:")
	cmd.Stdout = buf
	err := cmd.Run()
	if err != nil {
		return "", err
	}
	bs := buf.Bytes()
	index := bytes.LastIndexByte(bs, ' ')
	return string(bytes.Trim(bs[index:], " \r\n")), err
}
