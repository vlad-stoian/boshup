package boshup_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/pivotal-cf/on-demand-services-sdk/bosh"
	"github.com/pivotal-cf/on-demand-services-sdk/serviceadapter"
	"github.com/vlad-stoian/boshup"
	"gopkg.in/yaml.v2"
)

var _ = Describe("Boshup Interpolate", func() {

	var manifest string
	var ops string
	var variables map[string]interface{}

	var interpolatedManifestBytes []byte
	var generateErr error
	var interpolatedManifest string

	BeforeEach(func() {
		manifest = `---
key: value`
		ops = ""
		variables = map[string]interface{}{}
	})

	JustBeforeEach(func() {
		manifestBytes := []byte(manifest)
		opsBytes := []byte(ops)

		interpolatedManifestBytes, generateErr = boshup.Interpolate(manifestBytes, opsBytes, variables)
		Expect(generateErr).ToNot(HaveOccurred())

		interpolatedManifest = string(interpolatedManifestBytes)
	})

	Context("when just the manifest is passed", func() {

		It("returns back the manifest", func() {
			Expect(interpolatedManifest).To(Equal("key: value\n"))
		})
	})

	Context("when ops files are passed", func() {
		BeforeEach(func() {
			ops = `
- type: replace
  path: /key
  value: 10`
		})
		It("modifies the manifest", func() {
			Expect(interpolatedManifest).To(Equal("key: 10\n"))
		})
	})

	Context("when variables are in manifest", func() {
		BeforeEach(func() {
			manifest = `
---
key: ((variable))`
		})

		Context("and no variable value is provided", func() {
			It("the variable remains there", func() {
				interpolatedManifest := string(interpolatedManifestBytes)
				Expect(interpolatedManifest).To(Equal("key: ((variable))\n"))
			})
		})

		Context("and a variable value is provided", func() {
			BeforeEach(func() {
				variables["variable"] = "value"
			})
			It("replaces the variable with the value", func() {
				interpolatedManifest := string(interpolatedManifestBytes)
				Expect(interpolatedManifest).To(Equal("key: value\n"))
			})
		})

		Context("and variable has multiple nested levels", func() {
			BeforeEach(func() {
				variables["variable"] = map[string]map[string]string{
					"level1": {
						"level2": "level3",
					},
				}
			})

			It("replaces the variable with the value", func() {
				interpolatedManifest := string(interpolatedManifestBytes)
				Expect(interpolatedManifest).To(Equal("key:\n  level1:\n    level2: level3\n"))
			})
		})

		//XContext("and variable has multiple nested levels", func() {
		//	BeforeEach(func() {
		//		data := struct {
		//			firstLevel string `yaml:"firstLevel"`
		//		}{
		//			firstLevel: "value",
		//		}
		//		variables["variable"] = data
		//	})
		//
		//	It("replaces the variable with the value", func() {
		//		interpolatedManifest := string(interpolatedManifestBytes)
		//		Expect(interpolatedManifest).To(Equal("key:\n  level1:\n    level2: level3\n"))
		//	})
		//})

	})
})

var _ = Describe("Boshup UpdateFromServiceDeployment", func() {
	var boshManifest bosh.BoshManifest
	var boshManifestBytes []byte

	var serviceDeployment serviceadapter.ServiceDeployment

	var updatedManifestBytes []byte
	var updateErr error
	var updatedManifest string

	BeforeEach(func() {
		boshManifest = bosh.BoshManifest{
			Name: "bosh-manifest-name",
			Releases: []bosh.Release{
				{
					Name:    "original-release-name",
					Version: "original-release-version",
				},
			},
			Stemcells: []bosh.Stemcell{
				{
					Alias:   "original-stemcell-alias",
					Version: "original-stemcell-version",
					OS:      "original-stemcell-os",
				},
			},
		}

		boshManifestBytes, _ = yaml.Marshal(boshManifest)

		serviceDeployment = serviceadapter.ServiceDeployment{}
	})

	JustBeforeEach(func() {

		updatedManifestBytes, updateErr = boshup.UpdateFromServiceDeployment(boshManifestBytes, serviceDeployment)
		Expect(updateErr).ToNot(HaveOccurred())

		updatedManifest = string(updatedManifestBytes)
	})

	Context("when the manifest is not a bosh manifest", func() {
		BeforeEach(func() {
			boshManifestBytes = []byte("---\nkey: value\notherkey: othervalue\n")
		})

		It("returns empty bosh manifest", func() {
			emptyBoshManifest, _ := yaml.Marshal(bosh.BoshManifest{Stemcells: []bosh.Stemcell{{Alias: "", Version: "", OS: ""}}})

			Expect(updatedManifest).To(Equal(string(emptyBoshManifest)))
		})

	})

	Context("when the service deployment contains stemcell", func() {
		BeforeEach(func() {
			serviceDeployment = serviceadapter.ServiceDeployment{
				Stemcell: serviceadapter.Stemcell{
					Version: "service-deployment-stemcell-version",
					OS:      "service-deployment-stemcell-os",
				},
			}
		})
		It("does update the stemcell", func() {
			var updatedBoshManifest bosh.BoshManifest
			yaml.Unmarshal(updatedManifestBytes, &updatedBoshManifest)

			Expect(updatedBoshManifest.Stemcells).To(HaveLen(1))
			Expect(updatedBoshManifest.Stemcells[0].Alias).To(Equal("original-stemcell-alias"))
			Expect(updatedBoshManifest.Stemcells[0].Version).To(Equal("service-deployment-stemcell-version"))
			Expect(updatedBoshManifest.Stemcells[0].OS).To(Equal("service-deployment-stemcell-os"))
		})
	})

	Context("when the service deployment contains releases", func() {
		BeforeEach(func() {
			serviceDeployment = serviceadapter.ServiceDeployment{
				Releases: serviceadapter.ServiceReleases{
					{
						Name:    "service-deployment-release1-name",
						Version: "service-deployment-release1-version",
						Jobs:    []string{"service-deployment-release1-job"},
					},
					{
						Name:    "service-deployment-release2-name",
						Version: "service-deployment-release2-version",
						Jobs:    []string{"service-deployment-release2-job"},
					},
				},
			}
		})
		It("does update the stemcell", func() {
			var updatedBoshManifest bosh.BoshManifest
			yaml.Unmarshal(updatedManifestBytes, &updatedBoshManifest)

			Expect(updatedBoshManifest.Releases).To(HaveLen(2))

			Expect(updatedBoshManifest.Releases[0].Name).To(Equal("service-deployment-release1-name"))
			Expect(updatedBoshManifest.Releases[0].Version).To(Equal("service-deployment-release1-version"))

			Expect(updatedBoshManifest.Releases[1].Name).To(Equal("service-deployment-release2-name"))
			Expect(updatedBoshManifest.Releases[1].Version).To(Equal("service-deployment-release2-version"))
		})
	})
})
