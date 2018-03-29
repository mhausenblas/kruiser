package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func kubectl(withstderr bool, cmd string, args ...string) (string, error) {
	kubectlbin, err := shellout(withstderr, "which", "kubectl")
	if err != nil {
		return "", err
	}
	all := append([]string{cmd}, args...)
	result, err := shellout(withstderr, kubectlbin, all...)
	if err != nil {
		return "", err
	}
	return result, nil
}

func shellout(withstderr bool, cmd string, args ...string) (string, error) {
	result := ""
	var out bytes.Buffer
	c := exec.Command(cmd, args...)
	c.Env = os.Environ()
	if withstderr {
		c.Stderr = os.Stderr
	}
	c.Stdout = &out
	if debug {
		fmt.Println(append([]string{cmd}, args...))
	}
	err := c.Run()
	if err != nil {
		return result, err
	}
	result = strings.TrimSpace(out.String())
	return result, nil
}
