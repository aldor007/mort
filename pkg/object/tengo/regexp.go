package tengo

import (
	"regexp"

	tengoLib "github.com/d5/tengo/v2"
	"github.com/d5/tengo/v2/token"
)

// Regexp struct create regexp in tengo VM
type Regexp struct {
	tengoLib.ObjectImpl
	Value *regexp.Regexp
}

// String return regexp in string
func (o *Regexp) String() string {
	return o.Value.String()
}

// BinaryOp not implemented
func (o *Regexp) BinaryOp(op token.Token, rhs tengoLib.Object) (tengoLib.Object, error) {
	return nil, tengoLib.ErrInvalidOperator
}

// IsFalsy return true if regexp is nil
func (o *Regexp) IsFalsy() bool {
	return o.Value == nil
}

// Equals return true if regexp string are equal
func (o *Regexp) Equals(x tengoLib.Object) bool {
	return o.Value.String() == o.Value.String()
}

func (o *Regexp) Copy() tengoLib.Object {
	return &Regexp{
		Value: o.Value,
	}
}

func (o *Regexp) TypeName() string {
	return "Regexp-object"
}

func (o *Regexp) CanCall() bool {
	return true
}

// Call can executed regexp on given value and return immutable map with matches
func (o *Regexp) Call(args ...tengoLib.Object) (ret tengoLib.Object, err error) {
	if len(args) != 1 {
		err = tengoLib.ErrWrongNumArguments
		return
	}
	val := args[0].(*tengoLib.String)
	matches := o.Value.FindStringSubmatch(val.Value)
	if matches == nil {
		return &tengoLib.Map{}, nil
	}
	subMatchMap := make(map[string]tengoLib.Object)

	for i, name := range o.Value.SubexpNames() {
		if i != 0 && name != "" {
			subMatchMap[name] = &tengoLib.String{Value: matches[i]}
		}
	}

	ret = &tengoLib.ImmutableMap{Value: subMatchMap}
	return
}
