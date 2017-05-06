// +build !windows

package thuder

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func getDriveID() (string, error) {
	buf := bytes.NewBuffer(nil)
	cmd := exec.Command("lsblk", "--nodeps", "-o", "name,rm", "-n")
	cmd.Stdout = buf
	err := cmd.Run()
	if err != nil {
		//buf.WriteString("sda      1\nmmcblk0  0\n") //sample data
		return "", err
	}
	var name string
	for {
		line, err := buf.ReadString('\n')
		if err != nil {
			break
		}
		rmi := strings.LastIndex(line, " ")
		rm := strings.TrimSpace(line[rmi:])
		if rm == "0" {
			name = strings.TrimSpace(line[:rmi])
			break
		}
	}
	if len(name) == 0 {
		return "", fmt.Errorf("no acceptable block listed")
	}
	serial, err := os.Open("/sys/block/" + name + "/device/serial")
	if err != nil {
		return "", err
	}
	defer serial.Close()

	b := make([]byte, 12)
	n, err := serial.Read(b)
	if err != nil {
		return "", err
	}
	b = bytes.TrimSpace(b[:n])
	if n > 2 {
		return string(b[2:]), nil
	}
	return "", fmt.Errorf("read serial \"%s\" too short", b)
}
