package v7pushaction

import (
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	log "github.com/sirupsen/logrus"
)

func (actor Actor) StopApplication(pushPlan PushPlan, eventStream chan<- Event, progressBar ProgressBar) (PushPlan, Warnings, error) {
	var warnings v7action.Warnings
	var err error

	if pushPlan.Application.State == constant.ApplicationStarted {
		log.Info("Stopping Application")
		eventStream <- StoppingApplication
		warnings, err = actor.V7Actor.StopApplication(pushPlan.Application.GUID)
		if err != nil {
			return pushPlan, Warnings(warnings), err
		}
		eventStream <- StoppingApplicationComplete
	}

	return pushPlan, Warnings(warnings), nil
}
