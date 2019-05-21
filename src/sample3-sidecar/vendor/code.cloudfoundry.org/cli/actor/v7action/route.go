package v7action

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
)

type Route ccv3.Route

func (actor Actor) CreateRoute(spaceName string, domainName string) (Warnings, error) {
	allWarnings := Warnings{}
	domain, warnings, err := actor.GetDomainByName(domainName)
	allWarnings = append(allWarnings, warnings...)

	if err != nil {
		return allWarnings, err
	}

	space, warnings, err := actor.GetSpaceByNameAndOrganization(spaceName, domain.OrganizationGUID)
	allWarnings = append(allWarnings, warnings...)

	if err != nil {
		return allWarnings, err
	}
	_, apiWarnings, err := actor.CloudControllerClient.CreateRoute(ccv3.Route{
		SpaceGUID: space.GUID, DomainGUID: domain.GUID,
	})

	actorWarnings := Warnings(apiWarnings)
	allWarnings = append(allWarnings, actorWarnings...)

	if _, ok := err.(ccerror.RouteNotUniqueError); ok {
		return allWarnings, actionerror.RouteAlreadyExistsError{Route: domainName}
	}

	return allWarnings, err
}
