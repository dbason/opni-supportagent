package commands

var (
	password string
)

type Distribution string

const (
	RKE  Distribution = "rke"
	RKE2 Distribution = "rke2"
	K3S  Distribution = "k3s"
)
