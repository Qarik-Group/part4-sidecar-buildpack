package isolated

import (
	"regexp"

	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("bind-route-service command", func() {
	Describe("help", func() {
		It("includes a description of the options", func() {
			session := helpers.CF("help", "bind-route-service")
			Eventually(session).Should(Say("NAME:"))
			Eventually(session).Should(Say("bind-route-service - Bind a service instance to an HTTP route"))
			Eventually(session).Should(Say("USAGE:"))
			Eventually(session).Should(Say(regexp.QuoteMeta("cf bind-route-service DOMAIN [--hostname HOSTNAME] [--path PATH] SERVICE_INSTANCE [-c PARAMETERS_AS_JSON]")))
			Eventually(session).Should(Say("EXAMPLES:"))
			Eventually(session).Should(Say(regexp.QuoteMeta("cf bind-route-service example.com --hostname myapp --path foo myratelimiter")))
			Eventually(session).Should(Say(regexp.QuoteMeta("cf bind-route-service example.com myratelimiter -c file.json")))
			Eventually(session).Should(Say(regexp.QuoteMeta(`cf bind-route-service example.com myratelimiter -c '{"valid":"json"}'`)))
			Eventually(session).Should(Say(regexp.QuoteMeta(`In Windows PowerShell use double-quoted, escaped JSON: "{\"valid\":\"json\"}"`)))
			Eventually(session).Should(Say(regexp.QuoteMeta(`In Windows Command Line use single-quoted, escaped JSON: '{\"valid\":\"json\"}'`)))
			Eventually(session).Should(Say("ALIAS:"))
			Eventually(session).Should(Say("brs"))
			Eventually(session).Should(Say("OPTIONS:"))
			Eventually(session).Should(Say(`-c\s+Valid JSON object containing service-specific configuration parameters, provided inline or in a file\. For a list of supported configuration parameters, see documentation for the particular service offering.`))
			Eventually(session).Should(Say(`--hostname, -n\s+Hostname used in combination with DOMAIN to specify the route to bind`))
			Eventually(session).Should(Say(`--path\s+Path used in combination with HOSTNAME and DOMAIN to specify the route to bind`))
			Eventually(session).Should(Say(`SEE ALSO:`))
			Eventually(session).Should(Say(`routes, services`))
			Eventually(session).Should(Exit(0))
		})
	})

	Describe("Targeted space requirements", func() {
		Context("when an org or space is not targeted", func() {
			It("display an error that no org or space is targeted", func() {
				helpers.LoginCF()
				session := helpers.CF("bind-route-service", "www.example.org", "myInstance")
				Eventually(session).Should(Say("No org and space targeted, use 'cf target -o ORG -s SPACE' to target an org and space"))
			})
		})

		Context("when an org or space is targeted", func() {
			var (
				orgName             string
				spaceName           string
				domain              string
				host                string
				serviceInstanceName string
			)

			BeforeEach(func() {
				orgName = helpers.NewOrgName()
				spaceName = helpers.NewSpaceName()
				host = helpers.PrefixedRandomName("host")

				helpers.SetupCF(orgName, spaceName)
				domain = helpers.DefaultSharedDomain()

				serviceInstanceName = helpers.PrefixedRandomName("instance")
				servicePlanName := "fake-plan"
				serviceName := "fake-service"
				broker := helpers.NewServiceBroker(
					helpers.NewServiceBrokerName(),
					helpers.NewAssets().ServiceBroker,
					helpers.DefaultSharedDomain(),
					serviceName,
					servicePlanName,
				)
				broker.Push()
				broker.Service.Requires = `["route_forwarding"]`
				broker.Configure(true)
				broker.Create()

				Eventually(helpers.CF("enable-service-access", serviceName)).Should(Exit(0))
				Eventually(helpers.CF("create-service", serviceName, servicePlanName, serviceInstanceName)).Should(Exit(0))
				Eventually(helpers.CF("create-route", spaceName, domain, "--hostname", host)).Should(Exit(0))
			})

			AfterEach(func() {
				helpers.QuickDeleteOrg(orgName)
			})

			It("succeeds", func() {
				session := helpers.CF("bind-route-service", domain, serviceInstanceName, "--hostname", host)
				Eventually(session).Should(Say("OK"))
			})
		})
	})
})
