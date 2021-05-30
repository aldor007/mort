package cloudinary

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/aldor007/mort/pkg/transforms"
)

type (
	tokens struct {
		entries      []token
		currentIndex int
	}

	notationParser struct {
		tokens *tokens
	}

	// token holds parsed cloudinary command
	token struct {
		Name                string
		PositionalArguments []string
	}

	qualifiers struct {
		Width  uint
		Height uint
	}

	notImplementedError struct {
		Message string
	}
)

func (e notImplementedError) Error() string {
	return e.Message
}

func newNotationParser(source string) (*notationParser, error) {
	plainTokens := strings.Split(source, ",")
	parsedTokens := make([]token, len(plainTokens))
	for i := 0; i < len(plainTokens); i++ {
		parsedToken, err := parseToken(plainTokens[i])
		if err != nil {
			return nil, err
		}
		parsedTokens[i] = parsedToken
	}
	return &notationParser{
		tokens: &tokens{entries: parsedTokens},
	}, nil
}

func parseToken(plainToken string) (token, error) {
	// todo: Do not allocate new strings.
	// As the original token is not modified it is safe to reference original string in a response.
	splitted := strings.SplitN(plainToken, "_", 2)
	if len(splitted) < 2 {
		return token{}, errors.New("non parsable token")
	}
	result := token{Name: splitted[0]}
	index := 1
	result.PositionalArguments = strings.Split(splitted[index], ":")
	return result, nil
}

var errNoToken = errors.New("no token")

func (t *tokens) Token() (token, bool) {
	if t.currentIndex > len(t.entries)-1 {
		return token{}, false
	}
	// log.Printf("TOKEN: %+v\n", t.entries[t.currentIndex])
	return t.entries[t.currentIndex], true
}

func (t *tokens) Next() {
	t.currentIndex++
}

func (t *tokens) HasNext() bool {
	return t.currentIndex < len(t.entries)-1
}

func (c *notationParser) NextTransform() (transforms.Transforms, error) {
	token, exists := c.tokens.Token()
	if !exists {
		return transforms.Transforms{}, errNoToken
	}
	trans := transforms.New()
	switch token.Name {
	case "c":
		t, err := c.parseCrop(trans, token)
		if err != nil {
			return transforms.Transforms{}, err
		}
		return t, nil
	default:
		return transforms.Transforms{}, notImplementedError{Message: fmt.Sprintf("unknown action %s", token.Name)}
	}

}

func (c *notationParser) HasNext() bool {
	return c.tokens.HasNext()
}

func parseUint(t token) (uint, error) {
	if len(t.PositionalArguments) != 1 {
		return 0, fmt.Errorf("no integer value provided for '%s'", t.Name)
	}
	v, err := strconv.Atoi(t.PositionalArguments[0])
	if err != nil {
		return 0, fmt.Errorf("value '%s' is not an integer but expected for '%s'", t.PositionalArguments[0], t.Name)
	}
	if v < 0 {
		return 0, fmt.Errorf("value '%s' cannot be negative integer for '%s'", t.PositionalArguments[0], t.Name)
	}
	return uint(v), nil
}

func (c *notationParser) parseQualifiers() (qualifiers, error) {
	result := qualifiers{}
	for {
		c.tokens.Next()
		token, exists := c.tokens.Token()
		if !exists {
			return result, nil
		}
		switch token.Name {
		case "w":
			v, err := parseUint(token)
			if err != nil {
				return result, err
			}
			result.Width = v
		case "h":
			v, err := parseUint(token)
			if err != nil {
				return result, err
			}
			result.Height = v
		case "b", "ar", "g", "bo", "co", "x", "y":
			return result, notImplementedError{Message: fmt.Sprintf("'%s' qualifier is not implemented", token.Name)}
		default:

		}
	}
}

func (c *notationParser) parseCrop(result transforms.Transforms, token token) (transforms.Transforms, error) {
	if len(token.PositionalArguments) != 1 {
		return result, errors.New("crop requires mode")
	}
	switch token.PositionalArguments[0] {
	case "fit":
		qualifiers, err := c.parseQualifiers()
		if err != nil {
			return result, err
		}
		err = result.Resize(int(qualifiers.Width), int(qualifiers.Height), false, true, false)
		return result, err
	case "fill":
		qualifiers, err := c.parseQualifiers()
		if err != nil {
			return result, err
		}
		err = result.Resize(int(qualifiers.Width), int(qualifiers.Height), true, true, true)
		return result, err
	case "crop":
		qualifiers, err := c.parseQualifiers()
		if err != nil {
			return result, err
		}
		err = result.Crop(int(qualifiers.Width), int(qualifiers.Height), "", false, false)
		return result, err
	default:
		return result, notImplementedError{Message: fmt.Sprintf("'%s' is not implemented for crop action", token.PositionalArguments[0])}
	}
}
