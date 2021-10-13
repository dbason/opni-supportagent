package input

type ComponentInput interface {
	Publish(endpoint string) error // Publish should read the contents of component logs and publish them to the payload endpoint.
}
