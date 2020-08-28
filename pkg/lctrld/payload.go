package lctrld

import (
	"fmt"
	"os/exec"

	"github.com/apeunit/LaunchControlD/pkg/config"
	log "github.com/sirupsen/logrus"
)

func GenerateKeys(settings config.Schema, eventID string) (err error) {
	fmt.Println("Inside GenerateKEys")
	evt, err := loadEvent(settings, eventID)
	if err != nil {
		return
	}
	fmt.Println("Event", evt)

	for email, state := range evt.State {
		fmt.Println("Node owner is", email)
		fmt.Println("Node IP is", state.Instance.IPAddress)

	}
	// TODO: get launch payload path somewehre else
	fmt.Println("Executing launchpayloadd")
	cmd := exec.Command("/tmp/workspace/bin/launchpayloadd")
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Errorf("launchpayloadd cmd failed with %s, %s\n", err, out)
		return err
	}

	err = storeEvent(settings, evt)
	if err != nil {
		return
	}
	return
}
