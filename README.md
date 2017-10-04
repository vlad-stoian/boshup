# How to use

Example
-------

###### boshup.Interpolate(\[\]byte, \[\]byte, map\[string\]interface{}) (\[\]byte, error)
Code:
```go
manifestBytes := []byte(`---
key: value`)

opsBytes := []byte(`
- type: replace
  path: /key
  value: ((variable))`)

variables := map[string]interface{}{
    "variable": map[string]map[string]string{
        "level1": {
            "level2": "level3",
        },
    }
}

result, _ := boshup.Interpolate(manifestBytes, opsBytes, variables)

fmt.Println(string(result))
```
Output:
```yaml
key:
  level1:
    level2: level3
```

###### boshup.GetPath(\[\]byte, string) (string, error)

Code:
```go
manifest := `---
key:
  second_key:
  - name: first_array_element
    value: get-me-please
  - this: is-multi-line
    value: |
        ok
        this
        is
        weird`
path := "/key/second_key/name=first_array_element/value"

result, _ := boshup.GetPath(manifestBytes, path)

fmt.Println(result)

```

Output
```yaml
get-me-please
```

###### boshup.SetPath(\[\]byte, string, interface{}) (\[\]byte, error)

Code:
```go
manifest := `---
key:
  second_key:
  - name: first_array_element
    value: get-me-please`

path := "/key/second_key/name=first_array_element/value"

valueToBeSet := map[interface{}]interface{}{
    "some-random-key": map[interface{}]interface{}{
        "level-2-random-key": "finally-value",
    },
}

result, _ := boshup.SetPath(manifestBytes, path, valueToBeSet)

fmt.Println(result)
```

Output:
```yaml
key:
  second_key:
  - name: first_array_element
    value:
    some-random-key:
      level-2-random-key: finally-value
```


## Bonus:

##### boshup.UpdateFromServiceDeployment(\[\]byte, serviceadapter.ServiceDeployment) (\[\]byte, error)

Code:
```go
boshManifest := bosh.BoshManifest{
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

serviceDeployment := serviceadapter.ServiceDeployment{
    Stemcell: serviceadapter.Stemcell{
        Version: "service-deployment-stemcell-version",
        OS:      "service-deployment-stemcell-os",
    },
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

result, _ := boshup.UpdateFromServiceDeployment(boshManifestBytes, serviceDeployment)

fmt.Println(string(result))
```

Output:
```yaml
name: service-deployment-name
releases:
- name: service-deployment-release1-name
  version: service-deployment-release1-version
- name: service-deployment-release2-name
  version: service-deployment-release2-version
stemcells:
- alias: original-stemcell-alias
  os: service-deployment-stemcell-os
  version: service-deployment-stemcell-version
instance_groups: []
update:
  canaries: 0
  canary_watch_time: ""
  update_watch_time: ""
  max_in_flight: 0
```
