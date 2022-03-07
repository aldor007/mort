package tengo

import (
	"regexp"

	tengoLib "github.com/d5/tengo/v2"
	"github.com/d5/tengo/v2/token"
)

type Regexp struct {
	tengoLib.ObjectImpl
	Value *regexp.Regexp
}

func (o *Regexp) String() string {
	return o.Value.String()
}

func (o *Regexp) BinaryOp(op token.Token, rhs tengoLib.Object) (tengoLib.Object, error) {
	return nil, tengoLib.ErrInvalidOperator
}

func (o *Regexp) IsFalsy() bool {
	return false
}

func (o *Regexp) Equals(x tengoLib.Object) bool {
	return false
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

func (o *Regexp) Call(args ...tengoLib.Object) (ret tengoLib.Object, err error) {
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
