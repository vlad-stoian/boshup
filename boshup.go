package boshup

import (
	"strings"

	"github.com/cloudfoundry/bosh-cli/director/template"
	"github.com/cloudfoundry/bosh-utils/errors"
	"github.com/cppforlife/go-patch/patch"
	"github.com/pivotal-cf/on-demand-services-sdk/bosh"
	"github.com/pivotal-cf/on-demand-services-sdk/serviceadapter"
	"gopkg.in/yaml.v2"
)

func getOpsFromBytes(opsBytes []byte) (patch.Ops, error) {
	var opDefinitions []patch.OpDefinition

	err := yaml.Unmarshal(opsBytes, &opDefinitions)
	if err != nil {
		return nil, errors.WrapError(err, "failed to unmarshal ops")
	}

	ops, err := patch.NewOpsFromDefinitions(opDefinitions)
	if err != nil {
		return nil, errors.WrapError(err, "failed to create ops from definitions")
	}

	return ops, nil
}

func getStaticVariablesFromMap(variables map[string]interface{}) (template.Variables, error) {
	staticVariables := template.StaticVariables(variables)

	return staticVariables, nil
}

func GetPath(manifestBytes []byte, path string) (string, error) {
	tpl := template.NewTemplate(manifestBytes)

	pathPointer, err := patch.NewPointerFromString(path)
	if err != nil {
		return "", errors.WrapError(err, "failed to create pointer from path")
	}

	evaluated, err := tpl.Evaluate(template.StaticVariables{}, patch.Ops{}, template.EvaluateOpts{
		PostVarSubstitutionOp: patch.FindOp{Path: pathPointer},
		UnescapedMultiline:    true,
	})
	if err != nil {
		return "", errors.WrapError(err, "failed to evaluate get path")
	}

	trimmedEvaluated := strings.TrimSpace(string(evaluated))

	return trimmedEvaluated, nil
}

func SetPath(manifestBytes []byte, path string, value interface{}) ([]byte, error) {
	tpl := template.NewTemplate(manifestBytes)

	pathPointer, err := patch.NewPointerFromString(path)
	if err != nil {
		return nil, errors.WrapError(err, "failed to create pointer from path")
	}

	evaluated, err := tpl.Evaluate(template.StaticVariables{}, patch.Ops{}, template.EvaluateOpts{
		PostVarSubstitutionOp: patch.ReplaceOp{
			Path:  pathPointer,
			Value: value,
		},
		UnescapedMultiline: true,
	})
	if err != nil {
		return nil, errors.WrapError(err, "failed to evaluate get path")
	}

	return evaluated, nil
}

func Interpolate(manifestBytes []byte, opsBytes []byte, variables map[string]interface{}) ([]byte, error) {
	tpl := template.NewTemplate(manifestBytes)

	ops, err := getOpsFromBytes(opsBytes)
	if err != nil {
		return nil, err
	}

	staticVariables, err := getStaticVariablesFromMap(variables)
	if err != nil {
		return nil, err
	}

	evaluated, err := tpl.Evaluate(staticVariables, ops, template.EvaluateOpts{})
	if err != nil {
		return nil, errors.WrapError(err, "failed to evaluate template")
	}

	return evaluated, nil
}

func UpdateFromServiceDeployment(manifestBytes []byte, serviceDeployment serviceadapter.ServiceDeployment) ([]byte, error) {
	var boshManifest bosh.BoshManifest

	err := yaml.Unmarshal(manifestBytes, &boshManifest)
	if err != nil {
		return nil, errors.WrapError(err, "failed to unmarshal boshManifest")
	}

	boshManifest.Name = serviceDeployment.DeploymentName

	var stemcell bosh.Stemcell

	if len(boshManifest.Stemcells) == 1 {
		stemcell = boshManifest.Stemcells[0]
	}

	stemcell.Version = serviceDeployment.Stemcell.Version
	stemcell.OS = serviceDeployment.Stemcell.OS

	boshManifest.Stemcells = []bosh.Stemcell{stemcell}

	var releases []bosh.Release
	for _, serviceDeploymentRelease := range serviceDeployment.Releases {
		releases = append(releases, bosh.Release{
			Name:    serviceDeploymentRelease.Name,
			Version: serviceDeploymentRelease.Version,
		})
	}

	boshManifest.Releases = releases

	updatedManifestBytes, _ := yaml.Marshal(boshManifest)

	return updatedManifestBytes, nil
}
