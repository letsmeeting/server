package system

import (
	"fmt"
	ps "github.com/mitchellh/go-ps"
	"github.com/shirou/gopsutil/process"
	"os"
	"strings"
)

func GetPID(pName string) []int {
	var pidList []int
	processList, err := ps.Processes()
	if err != nil {
		return nil
	}
	for x := range processList {
		var process ps.Process
		process = processList[x]
		if process.Executable() == pName {
			//fmt.Println(process.Pid(), process.Executable())
			pidList = append(pidList, process.Pid())
		}
	}
	return pidList
}

func PrintProcess() {
	processList, err := ps.Processes()
	if err != nil {
		return
	}
	for x := range processList {
		var process ps.Process
		process = processList[x]
		fmt.Println(process.Pid(), process.Executable())
	}
}

func GetPid(pName string) []*process.Process {
	var pidList []*process.Process
	var count int
	psList, _ := process.Processes()
	mypid := os.Getpid()
	for _, p := range psList {
		//name, _ := p.Name()
		cmdLine, _ := p.Cmdline()
		if strings.Contains(cmdLine, pName) && p.Pid != int32(mypid) {
			count++
			pidList = append(pidList, p)
			//fmt.Printf("Found %d, pid(%d): Name[%s], Cmd[%s]\n", count, p.Pid, name, cmdLine)
		}
	}
	return pidList
}