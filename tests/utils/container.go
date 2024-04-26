package utils

import (
	"fmt"
	"os/exec"
)

var (
	containerRuntime string
	initialized      = false
)

func findContainerRuntime() string {
	for _, cmd := range []string{"podman", "docker"} {
		path, err := exec.LookPath(cmd)
		if err != nil {
			continue
		}
		err = exec.Command(path, "ps").Run()
		if err == nil {
			fmt.Printf("Found container runtime %s, path=%s\n", cmd, path)
			return path
		}
	}
	fmt.Println("Did not find any container runtimes")
	return ""
}

func HasContainerRuntimer() bool {
	if !initialized {
		containerRuntime = findContainerRuntime()
		initialized = true
	}
	return containerRuntime != ""
}

func GetCommand(args ...string) *exec.Cmd {
	return exec.Command(containerRuntime, args...)
}

func ExecCRI(args ...string) error {
	if !initialized {
		containerRuntime = findContainerRuntime()
		initialized = true
	}
	return GetCommand(args...).Run()
}
