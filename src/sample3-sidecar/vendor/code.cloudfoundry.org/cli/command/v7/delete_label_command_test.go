package v7_test

import (
	"errors"
	"regexp"

	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command/commandfakes"
	. "code.cloudfoundry.org/cli/command/v7"
	"code.cloudfoundry.org/cli/command/v7/v7fakes"
	"code.cloudfoundry.org/cli/types"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("delete-label command", func() {
	var (
		cmd             DeleteLabelCommand
		fakeConfig      *commandfakes.FakeConfig
		testUI          *ui.UI
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v7fakes.FakeDeleteLabelActor
		executeErr      error
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v7fakes.FakeDeleteLabelActor)
		cmd = DeleteLabelCommand{
			UI:          testUI,
			Config:      fakeConfig,
			SharedActor: fakeSharedActor,
			Actor:       fakeActor,
		}
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	It("doesn't error", func() {
		Expect(executeErr).ToNot(HaveOccurred())
	})

	It("checks that the user is logged in and targeted to an org and space", func() {
		Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
		checkOrg, checkSpace := fakeSharedActor.CheckTargetArgsForCall(0)
		Expect(checkOrg).To(BeTrue())
		Expect(checkSpace).To(BeTrue())
	})

	When("checking the target fails", func() {
		BeforeEach(func() {
			fakeSharedActor.CheckTargetReturns(errors.New("Target not found"))
		})

		It("we expect an error to be returned", func() {
			Expect(executeErr).To(MatchError("Target not found"))
		})
	})

	When("checking the target succeeds", func() {
		var appName string

		BeforeEach(func() {
			fakeConfig.TargetedOrganizationReturns(configv3.Organization{Name: "fake-org"})
			fakeConfig.TargetedSpaceReturns(configv3.Space{Name: "fake-space", GUID: "some-space-guid"})
			appName = "some-app"
			cmd.RequiredArgs.ResourceName = appName
		})

		When("getting the current user succeeds", func() {
			BeforeEach(func() {
				fakeConfig.CurrentUserReturns(configv3.User{Name: "some-user"}, nil)
				cmd.RequiredArgs.LabelKeys = []string{"some-label", "some-other-key"}
			})

			It("informs the user that labels are being deleted", func() {
				Expect(testUI.Out).To(Say(regexp.QuoteMeta(`Deleting label(s) for app %s in org fake-org / space fake-space as some-user...`), appName))
			})

			When("updating the app labels succeeds", func() {
				BeforeEach(func() {
					fakeActor.UpdateApplicationLabelsByApplicationNameReturns(v7action.Warnings{"some-warning-1", "some-warning-2"},
						nil)
				})

				It("does not return an error", func() {
					Expect(executeErr).ToNot(HaveOccurred())
				})

				It("prints all warnings", func() {
					Expect(testUI.Err).To(Say("some-warning-1"))
					Expect(testUI.Err).To(Say("some-warning-2"))
				})

				It("passes the correct parameters into the actor", func() {
					expectedMaps := map[string]types.NullString{
						"some-label":     types.NewNullString(),
						"some-other-key": types.NewNullString()}

					Expect(fakeActor.UpdateApplicationLabelsByApplicationNameCallCount()).To(Equal(1))
					actualAppName, spaceGUID, labelsMap := fakeActor.UpdateApplicationLabelsByApplicationNameArgsForCall(0)
					Expect(actualAppName).To(Equal(appName))
					Expect(spaceGUID).To(Equal("some-space-guid"))
					Expect(labelsMap).To(Equal(expectedMaps))
				})
			})

			When("updating the app labels fails", func() {
				BeforeEach(func() {
					fakeActor.UpdateApplicationLabelsByApplicationNameReturns(v7action.Warnings{"some-warning-1", "some-warning-2"},
						errors.New("api call failed"))
				})

				It("prints all warnings", func() {
					Expect(testUI.Err).To(Say("some-warning-1"))
					Expect(testUI.Err).To(Say("some-warning-2"))
				})

				It("returns the error", func() {
					Expect(executeErr).To(MatchError("api call failed"))
				})
			})
		})
		When("getting the user fails", func() {
			BeforeEach(func() {
				fakeConfig.CurrentUserReturns(configv3.User{}, errors.New("could not get user"))
				cmd.RequiredArgs.LabelKeys = []string{"some-label", "some-other-key"}
			})

			It("returns the error", func() {
				Expect(executeErr).To(MatchError("could not get user"))
			})
		})
	})
})
