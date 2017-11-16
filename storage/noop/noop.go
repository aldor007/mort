package noop

import (
	"net/url"

	"github.com/aldor007/stow"
)

// Kind represents the name of the location/storage type.
const Kind = "noop"

func init() {
	makefn := func(config stow.Config) (stow.Location, error) {

		// Create a location with given config and client (s3 session).
		loc := &location{
			config: config,
		}

		return loc, nil
	}

	kindfn := func(u *url.URL) bool {
		return u.Scheme == Kind
	}

	stow.Register(Kind, makefn, kindfn)
}
