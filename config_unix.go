package thuder

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func getDriveID() string {
	buf := bytes.NewBuffer(nil)
	cmd := exec.Command("lsblk", "--nodeps", "-o", "name,rm", "-n")
	cmd.Stdout = buf
	err := cmd.Run()
	if err != nil {
		fmt.Println(err)
		//buf.WriteString("sda      1\nmmcblk0  0\n") //sample data
		return "lsblk-err"
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
		return "no-disk"
	}
	serial, err := os.Open("/sys/block/" + name + "/device/serial")
	if err != nil {
		fmt.Println(err)
		return "open-err"
	}
	defer serial.Close()

	b := make([]byte, 12)
	n, err := serial.Read(b)
	if err != nil {
		fmt.Println(err)
		return "read-err"
	}
	if n > 2 {
		return string(b[2:n])
	}
	return "unknown"
}
