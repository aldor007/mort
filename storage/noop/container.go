package noop

import (
	"github.com/aldor007/stow"
	"github.com/pkg/errors"
	"io"
)

type container struct {
	name string
}

// ID returns a string value which represents the name of the container.
func (c *container) ID() string {
	return c.name
}

// Name returns a string value which represents the name of the container.
func (c *container) Name() string {
	return c.name
}

// Item returns a stow.Item instance of a container based on the
// name of the container and the key representing
func (c *container) Item(id string) (stow.Item, error) {
	return &item{name: id}, stow.ErrNotFound
}

// Items sends a request to retrieve a list of items that are prepended with
// the prefix argument. The 'cursor' variable facilitates pagination.
func (c *container) Items(prefix, cursor string, count int) ([]stow.Item, string, error) {
	return nil, "", errors.New("not implemented")
}

func (c *container) RemoveItem(id string) error {
	return nil
}

func (c *container) Put(name string, r io.Reader, size int64, metadata map[string]interface{}) (stow.Item, error) {
	return &item{name: name}, nil
}
