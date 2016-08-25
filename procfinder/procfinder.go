package procfinder

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
)

func readNetTCP() ([]string, error) {
	procTcp, err := ioutil.ReadFile("/proc/net/tcp")

	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(procTcp), "\n")

	return lines[1 : len(lines)-1], nil
}

// Process represents a unix based process. This provides the direct path to the exe that
// it was originally spawned with, along with the nicename and process ID.
type Process struct {
	PID         string
	Name        string
	Exe         string
	User        string
	IP          string
	Port        int64
	ForeignIP   string
	ForeignPort int64
}

func removeEmpty(array []string) []string {
	// remove empty data from line
	var new_array []string
	for _, i := range array {
		if i != "" {
			new_array = append(new_array, i)
		}
	}

	return new_array
}

// convert hexadecimal to decimal
func hexToDec(h string) (d int64) {
	d, _ = strconv.ParseInt(h, 16, 32)

	return d
}

// convert the ipv4 to decimal. would need to rearrange the ip because the default value
// is in little Endian order.
func ip(ip string) string {
	var out string

	// check ip size. if greater than 8, it is ipv6
	if len(ip) > 8 {
		i := []string{ip[30:32], ip[28:30], ip[26:28], ip[24:26], ip[22:24], ip[20:22], ip[18:20], ip[16:18], ip[14:16], ip[12:14], ip[10:12], ip[8:10], ip[6:8], ip[4:6], ip[2:4], ip[0:2]}
		out = fmt.Sprintf("%v%v:%v%v:%v%v:%v%v:%v%v:%v%v:%v%v:%v%v", i[14], i[15], i[13], i[12], i[10], i[11], i[8], i[9], i[6], i[7], i[4], i[5], i[2], i[3], i[0], i[1])
	} else {
		// ipv4
		i := []int64{hexToDec(ip[6:8]), hexToDec(ip[4:6]), hexToDec(ip[2:4]), hexToDec(ip[0:2])}

		out = fmt.Sprintf("%v.%v.%v.%v", i[0], i[1], i[2], i[3])
	}

	return out
}

// loop through all fd dirs of process on /proc to compare the inode and get the pid
func getPid(inode string) (pid string) {
	d, err := filepath.Glob("/proc/[0-9]*/fd/[0-9]*")
	if err != nil {
		return pid
	}

	for _, item := range d {
		path, _ := os.Readlink(item)
		if strings.Contains(path, inode) {
			pid = strings.Split(item, "/")[2]
		}
	}

	return pid
}

func getProcessExe(pid string) string {
	exe := fmt.Sprintf("/proc/%s/exe", pid)
	path, _ := os.Readlink(exe)
	return path
}

func getProcessName(pid string) (name string) {
	tmp, err := ioutil.ReadFile(fmt.Sprintf("/proc/%s/comm", pid))

	if err != nil {
		return ""
	}

	return strings.Split(string(tmp), "\n")[0]
}

func getUser(uid string) string {
	u, _ := user.LookupId(uid)
	return u.Username
}

// GetProcs crawls /proc/ for all pids that have bound ports
func GetProcs() (pl []*Process, err error) {
	tcp, err := readNetTCP()

	if err != nil {
		return nil, err
	}

	for _, line := range tcp {
		line_array := removeEmpty(strings.Split(strings.TrimSpace(line), " "))
		ip_port := strings.Split(line_array[1], ":")
		fip_port := strings.Split(line_array[2], ":")

		proc := &Process{
			IP:          ip(ip_port[0]),
			Port:        hexToDec(ip_port[1]),
			ForeignIP:   ip(fip_port[0]),
			ForeignPort: hexToDec(fip_port[1]),
			User:        getUser(line_array[7]),
			PID:         getPid(line_array[9]),
		}

		proc.Exe = getProcessExe(proc.PID)
		proc.Name = getProcessName(proc.PID)

		pl = append(pl, proc)
	}

	return pl, nil
}