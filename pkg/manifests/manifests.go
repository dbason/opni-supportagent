package manifests

import _ "embed" // embed should be a blank import

//go:embed certmanager.yaml
var CertManagerYaml string

//go:embed operator.yaml
var OperatorYaml string
