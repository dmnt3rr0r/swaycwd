package main

import (
	"bytes"
	"fmt"
	"strconv"
	//"io/ioutil"
	"encoding/json"
	"os/exec"
)

type SwayNode struct {
	Name    string     `json:"name"`
	Focused bool       `json:"focused"`
	Type    string     `json:"type"`
	Pid     *int       `json:"pid,omitempty"`
	Nodes   []SwayNode `json:"nodes"`
}

func FindFocusPid(node *SwayNode) (int, error) {
	if node.Focused {
		if node.Type == "con" {
			return *node.Pid, nil
		} else {
			return -1, fmt.Errorf("Focused node was not type 'con'")
		}
	}

	for _, n := range node.Nodes {
		p, err := FindFocusPid(&n)
		if err != nil {
			return -1, err
		}

		if p > -1 {
			return p, nil
		}
	}

	return -1, nil
}

func GetCWD(pid int) (string, error) {

	pgrep := exec.Command("pgrep", "-P", strconv.Itoa(pid))
	var out bytes.Buffer
	pgrep.Stdout = &out

	if err := pgrep.Run(); err != nil {
		// we are at the bottom?
		cwd := exec.Command("readlink", fmt.Sprintf("/proc/%v/cwd", pid))
		cwd_out, err := cwd.Output()
		if err != nil {
			return "", err
		}

		return string(cwd_out), nil
	}

	b := out.Bytes()
	cpid, err := strconv.ParseInt(string(b[:len(b)-1]), 10, 32)
	if err != nil {
		return "", err
	}

	return GetCWD(int(cpid))
}

func main() {
	get_tree := exec.Command("swaymsg", "-t", "get_tree")

	get_tree_out, err := get_tree.Output()
	if err != nil {
		panic(err)
	}

	root := SwayNode{}
	err = json.Unmarshal(get_tree_out, &root)

	pid, err := FindFocusPid(&root)
	if err != nil {
		panic(err)
	}

	if pid > 0 {
		cwd, err := GetCWD(pid)
		if err != nil {
			panic(err)
		}

		fmt.Printf(cwd)
	} else {
		fmt.Printf("/tmp")
	}
}
