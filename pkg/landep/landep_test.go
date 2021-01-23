package landep

import (
	"github.com/Masterminds/semver/v3"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("landep", func() {
	pkgManager := &packageManager{}
	pkgManager.repository.Register("docker.io/pkgs/cluster", semver.MustParse("1.0.1"), ClusterInstallerFactory)
	pkgManager.repository.Register("docker.io/pkgs/cloud-foundry", semver.MustParse("2.0.0"), CloudFoundryInstallerFactory)
	pkgManager.repository.Register("docker.io/pkgs/cloud-foundry-environment", semver.MustParse("1.0.0"), CloudFoundryEnvironmentInstallerFactory)
	pkgManager.repository.Register("docker.io/pkgs/organization", semver.MustParse("1.0.0"), OrganizationInstallerFactory)

	It("works with cluster-pkg installer", func() {
		target := NewK8sFake("default", "https://gardener.canary.hana-ondemand.com")
		targets := Targets{"default": target}
		constraint, err := semver.NewConstraint(">= 1.0")
		Expect(err).To(Succeed())
		By("applies", func() {
			parameter := Parameter(nil)
			_, err = pkgManager.Apply(targets, "docker.io/pkgs/cluster", constraint, parameter,nil)
			Expect(err).To(Succeed())
			objs, err := target.List("cluster")
			Expect(err).To(Succeed())
			Expect(objs).To(HaveLen(1))
		})
		By("deletes", func() {
			err = pkgManager.Delete(targets, "docker.io/pkgs/cluster")
			Expect(err).To(Succeed())
			objs, err := target.List("cluster")
			Expect(err).To(Succeed())
			Expect(objs).To(HaveLen(0))
		})
	})
	It("works with dependencies", func() {
		target := NewK8sFake("default", "https://gardener.canary.hana-ondemand.com")
		targets := Targets{"default": target}
		constraint, err := semver.NewConstraint(">= 1.0")
		Expect(err).To(Succeed())
		By("applies", func() {
			parameter := Parameter(nil)
			_, err = pkgManager.Apply(targets, "docker.io/pkgs/cloud-foundry-environment", constraint, parameter,nil)
			Expect(err).To(Succeed())
			objs, err := target.List("cluster")
			Expect(err).To(Succeed())
			Expect(objs).To(HaveLen(1))
		})
		By("deletes", func() {
			err = pkgManager.Delete(targets, "docker.io/pkgs/cloud-foundry-environment")
			Expect(err).To(Succeed())
			objs, err := target.List("cluster")
			Expect(err).To(Succeed())
			Expect(objs).To(HaveLen(0))
		})

	})
})
