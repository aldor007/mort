package http

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/aldor007/stow"
	"github.com/pkg/errors"
)

// A location contains a client + the configurations used to create the client.
type location struct {
	config   stow.Config
	endpoint string
	client   *http.Client
	headers  map[string]string
}

func (l *location) CreateContainer(containerName string) (stow.Container, error) {
	return nil, errors.New("not implemented")
}

func (l *location) Containers(prefix, cursor string, count int) ([]stow.Container, string, error) {
	var containers []stow.Container

	return containers, "", errors.New("not implemented")
}

// Close simply satisfies the Location interface. There's nothing that
// needs to be done in order to satisfy the interface.
func (l *location) Close() error {
	return nil // nothing to close
}

// Container retrieves a stow.Container based on its name which must be
// exact.
func (l *location) Container(id string) (stow.Container, error) {
	endpoint := strings.Replace(l.endpoint, "<container>", id, 1)
	return &container{
		client:   l.client,
		endpoint: endpoint,
		headers:  l.headers,
	}, nil

}

// RemoveContainer removes a container simply by name.
func (l *location) RemoveContainer(id string) error {
	return errors.New("not implemeted")
}

// ItemByURL retrieves a stow.Item by parsing the URL, in this
// case an item is an object.
func (l *location) ItemByURL(url *url.URL) (stow.Item, error) {
	return nil, errors.New("not implemeted")
}
