package v6_test

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/cfnetworkingaction"
	"code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/command/commandfakes"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
	. "code.cloudfoundry.org/cli/command/v6"
	"code.cloudfoundry.org/cli/command/v6/v6fakes"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("add-network-policy Command", func() {
	var (
		cmd                    AddNetworkPolicyCommand
		testUI                 *ui.UI
		fakeConfig             *commandfakes.FakeConfig
		fakeSharedActor        *commandfakes.FakeSharedActor
		fakeNetworkPolicyActor *v6fakes.FakeAddNetworkPolicyActor
		fakeMembershipActor    *v6fakes.FakeMembershipActor
		binaryName             string
		executeErr             error
		srcApp                 string
		destApp                string
		protocol               string
		org                    string
		space                  string
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeNetworkPolicyActor = new(v6fakes.FakeAddNetworkPolicyActor)
		fakeMembershipActor = new(v6fakes.FakeMembershipActor)

		srcApp = "some-app"
		destApp = "some-other-app"
		protocol = "tcp"
		org = ""
		space = ""

		cmd = AddNetworkPolicyCommand{
			UI:                 testUI,
			Config:             fakeConfig,
			SharedActor:        fakeSharedActor,
			NetworkPolicyActor: fakeNetworkPolicyActor,
			MembershipActor:    fakeMembershipActor,
			RequiredArgs:       flag.AddNetworkPolicyArgs{SourceApp: srcApp},
			DestinationApp:     destApp,
			DestinationSpace:   space,
			DestinationOrg:     org,
		}

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	When("checking target fails", func() {
		BeforeEach(func() {
			fakeSharedActor.CheckTargetReturns(actionerror.NotLoggedInError{BinaryName: binaryName})
		})

		It("returns an error", func() {
			Expect(executeErr).To(MatchError(actionerror.NotLoggedInError{BinaryName: binaryName}))

			Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
			checkTargetedOrg, checkTargetedSpace := fakeSharedActor.CheckTargetArgsForCall(0)
			Expect(checkTargetedOrg).To(BeTrue())
			Expect(checkTargetedSpace).To(BeTrue())
		})
	})

	When("the user is logged in, an org is targeted, and a space is targeted", func() {
		BeforeEach(func() {
			fakeConfig.CurrentUserReturns(configv3.User{Name: "some-user"}, nil)
			fakeConfig.TargetedSpaceReturns(configv3.Space{Name: "some-space", GUID: "some-space-guid"})
			fakeConfig.TargetedOrganizationReturns(configv3.Organization{Name: "some-org"})
		})

		When("protocol is specified but port is not", func() {
			BeforeEach(func() {
				cmd.Protocol = flag.NetworkProtocol{Protocol: protocol}
			})

			It("returns an error", func() {
				Expect(executeErr).To(MatchError(translatableerror.NetworkPolicyProtocolOrPortNotProvidedError{}))
				Expect(testUI.Out).NotTo(Say(`Adding network policy`))
			})
		})

		When("port is specified but protocol is not", func() {
			BeforeEach(func() {
				cmd.Port = flag.NetworkPort{StartPort: 8080, EndPort: 8081}
			})

			It("returns an error", func() {
				Expect(executeErr).To(MatchError(translatableerror.NetworkPolicyProtocolOrPortNotProvidedError{}))
				Expect(testUI.Out).NotTo(Say(`Adding network policy`))
			})
		})

		When("both protocol and port are specificed", func() {
			BeforeEach(func() {
				cmd.Protocol = flag.NetworkProtocol{Protocol: protocol}
				cmd.Port = flag.NetworkPort{StartPort: 8080, EndPort: 8081}
			})

			When("the policy creation is successful", func() {
				BeforeEach(func() {
					fakeNetworkPolicyActor.AddNetworkPolicyReturns(cfnetworkingaction.Warnings{"some-warning-1", "some-warning-2"}, nil)
				})

				It("displays OK when no error occurs", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(fakeMembershipActor.GetOrganizationByNameCallCount()).To(Equal(0))
					Expect(fakeMembershipActor.GetSpaceByNameAndOrganizationCallCount()).To(Equal(0))
					Expect(fakeNetworkPolicyActor.AddNetworkPolicyCallCount()).To(Equal(1))
					passedSrcSpaceGuid, passedSrcAppName, passedDestSpaceGuid, passedDestAppName, passedProtocol, passedStartPort, passedEndPort := fakeNetworkPolicyActor.AddNetworkPolicyArgsForCall(0)
					Expect(passedSrcSpaceGuid).To(Equal("some-space-guid"))
					Expect(passedSrcAppName).To(Equal("some-app"))
					Expect(passedDestSpaceGuid).To(Equal("some-space-guid"))
					Expect(passedDestAppName).To(Equal("some-other-app"))
					Expect(passedProtocol).To(Equal("tcp"))
					Expect(passedStartPort).To(Equal(8080))
					Expect(passedEndPort).To(Equal(8081))

					Expect(testUI.Out).To(Say(`Adding network policy from app %s to app %s in org some-org / space some-space as some-user\.\.\.`, srcApp, destApp))
					Expect(testUI.Err).To(Say("some-warning-1"))
					Expect(testUI.Err).To(Say("some-warning-2"))
					Expect(testUI.Out).To(Say("OK"))
				})
			})

			When("the policy creation is not successful", func() {
				BeforeEach(func() {
					fakeNetworkPolicyActor.AddNetworkPolicyReturns(cfnetworkingaction.Warnings{"some-warning-1", "some-warning-2"}, actionerror.ApplicationNotFoundError{Name: srcApp})
				})

				It("does not display OK when an error occurs", func() {
					Expect(executeErr).To(MatchError(actionerror.ApplicationNotFoundError{Name: srcApp}))

					Expect(testUI.Out).To(Say(`Adding network policy from app %s to app %s in org some-org / space some-space as some-user\.\.\.`, srcApp, destApp))
					Expect(testUI.Err).To(Say("some-warning-1"))
					Expect(testUI.Err).To(Say("some-warning-2"))
					Expect(testUI.Out).ToNot(Say("OK"))
				})
			})
		})

		When("both protocol and port are not specified", func() {
			It("defaults protocol to 'tcp' and port to '8080'", func() {
				Expect(executeErr).ToNot(HaveOccurred())

				Expect(fakeNetworkPolicyActor.AddNetworkPolicyCallCount()).To(Equal(1))
				_, _, _, _, passedProtocol, passedStartPort, passedEndPort := fakeNetworkPolicyActor.AddNetworkPolicyArgsForCall(0)
				Expect(passedProtocol).To(Equal("tcp"))
				Expect(passedStartPort).To(Equal(8080))
				Expect(passedEndPort).To(Equal(8080))
			})
		})

		When("org is specified and space is not", func() {
			BeforeEach(func() {
				cmd.DestinationOrg = "bananarama"
				cmd.DestinationSpace = ""
				fakeMembershipActor.GetOrganizationByNameReturns(v3action.Organization{}, v3action.Warnings{}, actionerror.OrganizationNotFoundError{Name: "bananarama"})
			})

			It("returns an error", func() {
				Expect(executeErr).To(MatchError(translatableerror.NetworkPolicyDestinationOrgWithoutSpaceError{}))
				Expect(fakeNetworkPolicyActor.AddNetworkPolicyCallCount()).To(Equal(0))
			})
		})

		When("invalid org is specified", func() {
			BeforeEach(func() {
				cmd.DestinationOrg = "bananarama"
				cmd.DestinationSpace = "hamdinger"
				warnings := v3action.Warnings{"some-org-warning-1", "some-org-warning-2"}
				fakeMembershipActor.GetOrganizationByNameReturns(v3action.Organization{}, warnings, actionerror.OrganizationNotFoundError{Name: "bananarama"})
			})

			It("returns an error", func() {
				Expect(executeErr).To(MatchError(actionerror.OrganizationNotFoundError{Name: "bananarama"}))
				Expect(fakeNetworkPolicyActor.AddNetworkPolicyCallCount()).To(Equal(0))
			})

			It("prints the warnings", func() {
				Expect(testUI.Err).To(Say("some-org-warning-1"))
				Expect(testUI.Err).To(Say("some-org-warning-2"))
			})
		})

		When("a valid org but invalid space is specified", func() {
			BeforeEach(func() {
				cmd.DestinationOrg = "bananarama"
				cmd.DestinationSpace = "hamdinger"
				warnings := v3action.Warnings{"some-space-warning-1", "some-space-warning-2"}
				fakeMembershipActor.GetOrganizationByNameReturns(v3action.Organization{GUID: "some-org-guid"}, v3action.Warnings{}, nil)
				fakeMembershipActor.GetSpaceByNameAndOrganizationReturns(v3action.Space{}, warnings, actionerror.SpaceNotFoundError{Name: "bananarama"})
			})

			It("returns an error", func() {
				passedSpaceName, passedOrgGuid := fakeMembershipActor.GetSpaceByNameAndOrganizationArgsForCall(0)
				Expect(passedSpaceName).To(Equal("hamdinger"))
				Expect(passedOrgGuid).To(Equal("some-org-guid"))
				Expect(executeErr).To(MatchError(actionerror.SpaceNotFoundError{Name: "bananarama"}))
				Expect(fakeNetworkPolicyActor.AddNetworkPolicyCallCount()).To(Equal(0))
			})

			It("prints the warnings", func() {
				Expect(testUI.Err).To(Say("some-space-warning-1"))
				Expect(testUI.Err).To(Say("some-space-warning-2"))
			})
		})

		When("a destination space but no destination org is specified", func() {
			BeforeEach(func() {
				cmd.DestinationSpace = "hamdinger"
				warnings := v3action.Warnings{"some-warning-1", "some-warning-2"}
				fakeMembershipActor.GetSpaceByNameAndOrganizationReturns(v3action.Space{GUID: "some-other-space-guid"}, warnings, nil)
			})

			It("displays OK", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(fakeMembershipActor.GetOrganizationByNameCallCount()).To(Equal(0))
				Expect(fakeNetworkPolicyActor.AddNetworkPolicyCallCount()).To(Equal(1))
				passedSrcSpaceGuid, _, passedDestSpaceGuid, _, _, _, _ := fakeNetworkPolicyActor.AddNetworkPolicyArgsForCall(0)
				Expect(passedSrcSpaceGuid).To(Equal("some-space-guid"))
				Expect(passedDestSpaceGuid).To(Equal("some-other-space-guid"))

				Expect(testUI.Out).To(Say(`Adding network policy from app %s in org some-org / space some-space to app %s in org some-org / space hamdinger as some-user\.\.\.`, srcApp, destApp))
				Expect(testUI.Out).To(Say("OK"))
			})

			It("prints the warnings", func() {
				Expect(testUI.Err).To(Say("some-warning-1"))
				Expect(testUI.Err).To(Say("some-warning-2"))
			})
		})

		When("a destination org and space is specified for destination app", func() {
			BeforeEach(func() {
				cmd.DestinationOrg = "bananarama"
				cmd.DestinationSpace = "hamdinger"
				fakeMembershipActor.GetOrganizationByNameReturns(v3action.Organization{GUID: "some-org-guid"}, v3action.Warnings{"some-org-warning-1", "some-org-warning-2"}, nil)
				fakeMembershipActor.GetSpaceByNameAndOrganizationReturns(v3action.Space{GUID: "some-other-space-guid"}, v3action.Warnings{"some-space-warning-1", "some-space-warning-2"}, nil)
				fakeNetworkPolicyActor.AddNetworkPolicyReturns(cfnetworkingaction.Warnings{"some-add-warning-1", "some-add-warning-2"}, nil)
			})

			It("displays OK", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(fakeNetworkPolicyActor.AddNetworkPolicyCallCount()).To(Equal(1))
				passedSrcSpaceGuid, _, passedDestSpaceGuid, _, _, _, _ := fakeNetworkPolicyActor.AddNetworkPolicyArgsForCall(0)
				Expect(passedSrcSpaceGuid).To(Equal("some-space-guid"))
				Expect(passedDestSpaceGuid).To(Equal("some-other-space-guid"))

				Expect(testUI.Out).To(Say(`Adding network policy from app %s in org some-org / space some-space to app %s in org bananarama / space hamdinger as some-user\.\.\.`, srcApp, destApp))
				Expect(testUI.Err).To(Say("some-org-warning-1"))
				Expect(testUI.Err).To(Say("some-org-warning-2"))
				Expect(testUI.Err).To(Say("some-space-warning-1"))
				Expect(testUI.Err).To(Say("some-space-warning-2"))
				Expect(testUI.Err).To(Say("some-add-warning-1"))
				Expect(testUI.Err).To(Say("some-add-warning-2"))
				Expect(testUI.Out).To(Say("OK"))
			})
		})
	})
})
