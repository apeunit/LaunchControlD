package cmdrunner

import (
	"errors"
	"os/exec"
	"strings"

	log "github.com/sirupsen/logrus"
)

// CommandRunner func type allows for mocking out RunCommand()
type CommandRunner func([]string, []string) (string, error)

// RunCommand runs a command
func RunCommand(command, envVars []string) (out string, err error) {
	cmd := exec.Command(command[0], command[1:]...)
	// add the binary folder to the exec path
	cmd.Env = envVars
	log.Debug("Running command ", command, cmd.Env)
	// execute the command
	o, err := cmd.CombinedOutput()
	if err != nil {
		log.Errorf("%s failed with %s, %s\n", command, err, string(o))
		return "", errors.New(strings.TrimSpace(string(o)))
	}
	out = strings.TrimSpace(string(o))
	log.Debug("Command stdout: ", out)
	return
}
