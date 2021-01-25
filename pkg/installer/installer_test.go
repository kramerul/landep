package installer

import (
	"github.com/Masterminds/semver/v3"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.tools.sap/D001323/landep/pkg/landep"
)

var _ = Describe("landep", func() {
	var logs []string
	log := func(message string) {
		logs = append(logs, message)
	}
	landep.InitFakeTargetFactory(log)

	pkgManager := landep.NewPackageManager(landep.Repository)
	k8sConfig := &landep.K8sConfig{URL: "https://gardener.canary.hana-ondemand.com"}

	It("works with cluster-pkg installer", func() {
		target := landep.NewK8sTarget("default", k8sConfig)
		constraint, err := semver.NewConstraint(">= 1.0")
		Expect(err).To(Succeed())
		By("applies", func() {
			logs = nil
			parameter := landep.Parameter(nil)
			_, err = pkgManager.Apply(target, "docker.io/pkgs/cluster", constraint, parameter)
			Expect(err).To(Succeed())
			Expect(logs).To(HaveLen(1))
			Expect(logs[0]).To(ContainSubstring("helm upgrade"))
		})
		By("deletes", func() {
			logs = nil
			err = pkgManager.Delete(target, "docker.io/pkgs/cluster")
			Expect(err).To(Succeed())
			Expect(logs).To(HaveLen(1))
			Expect(logs[0]).To(ContainSubstring("helm delete"))
		})
	})
	It("works with dependencies", func() {
		target := landep.NewK8sTarget("default", k8sConfig)
		constraint, err := semver.NewConstraint(">= 1.0")
		Expect(err).To(Succeed())
		By("applies", func() {
			logs = nil
			parameter := landep.Parameter(nil)
			_, err = pkgManager.Apply(target, "docker.io/pkgs/cloud-foundry-environment", constraint, parameter)
			Expect(err).To(Succeed())
			Expect(logs).To(HaveLen(5))
			Expect(logs[0]).To(MatchRegexp("helm upgrade -i -n default --version 1.0.1 \\w* cluster"))
			Expect(logs[1]).To(MatchRegexp("helm upgrade -i -n istio-system --version 1.7.0 \\w* istio"))
			Expect(logs[2]).To(MatchRegexp("kapp deploy -n cf-system -a \\w* cf-for-k8s-scp"))
			Expect(logs[3]).To(MatchRegexp("(helm upgrade -i -n service-agent-manager --version 0.1.0 \\w* service-manager-agent|cf create org \\w*)"))
			Expect(logs[4]).To(MatchRegexp("(helm upgrade -i -n service-agent-manager --version 0.1.0 \\w* service-manager-agent|cf create org \\w*)"))
		})
		By("deletes", func() {
			logs = nil
			err = pkgManager.Delete(target, "docker.io/pkgs/cloud-foundry-environment")
			Expect(err).To(Succeed())
			Expect(logs).To(HaveLen(5))
			// Sequence for the first 2 iterms is not defined because they don't depend on each other
			Expect(logs[0]).To(MatchRegexp("(helm delete -n service-agent-manager|cf delete org) \\w*"))
			Expect(logs[1]).To(MatchRegexp("(helm delete -n service-agent-manager|cf delete org) \\w*"))
			Expect(logs[2]).To(MatchRegexp("kapp delete -n cf-system -a \\w*"))
			Expect(logs[3]).To(MatchRegexp("helm delete -n istio-system \\w*"))
			Expect(logs[4]).To(MatchRegexp("helm delete -n default \\w*"))
		})

	})
	It("deals with shared dependencies and conflicting parameters", func() {
		target := landep.NewK8sTarget("cf-system", k8sConfig)
		constraint, err := semver.NewConstraint(">= 1.0")
		Expect(err).To(Succeed())
		By("applies cloud-foundry", func() {
			logs = nil
			parameter := landep.Parameter(nil)
			_, err = pkgManager.Apply(target, "docker.io/pkgs/cloud-foundry", constraint, parameter)
			Expect(err).To(Succeed())
			Expect(logs).To(HaveLen(2))
			Expect(logs[0]).To(MatchRegexp(`helm upgrade -i -n istio-system --version 1.7.0 \w* istio \{"pilot":\{"instances":1\}\}`))
			Expect(logs[1]).To(MatchRegexp("kapp deploy -n cf-system -a \\w* cf-for-k8s-scp"))
		})
		By("apply kyma", func() {
			target := landep.NewK8sTarget("kyma-system", k8sConfig)
			logs = nil
			parameter := landep.Parameter(nil)
			_, err = pkgManager.Apply(target, "docker.io/pkgs/kyma", constraint, parameter)
			Expect(err).To(Succeed())
			Expect(logs).To(HaveLen(2))
			// Also update istio because of potential different parameters
			Expect(logs[0]).To(MatchRegexp(`helm upgrade -i -n istio-system --version 1.7.0 \w* istio \{"pilot":\{"instances":3\}\}`))
			Expect(logs[1]).To(MatchRegexp("helm upgrade -i -n kyma-system --version 1.17.0 \\w* kyma"))
		})
		By("deletes cloud-foundry and istio", func() {
			logs = nil
			err = pkgManager.Delete(target, "docker.io/pkgs/cloud-foundry")
			Expect(err).To(Succeed())
			Expect(logs).To(HaveLen(1))
			Expect(logs[0]).To(MatchRegexp("kapp delete -n cf-system -a \\w*"))
		})
		By("deletes kyma and istio", func() {
			logs = nil
			target := landep.NewK8sTarget("kyma-system", k8sConfig)
			err = pkgManager.Delete(target, "docker.io/pkgs/kyma")
			Expect(err).To(Succeed())
			Expect(logs).To(HaveLen(2))
			Expect(logs[0]).To(MatchRegexp("helm delete -n kyma-system \\w*"))
			Expect(logs[1]).To(MatchRegexp("helm delete -n istio-system \\w*"))
		})

	})
})
