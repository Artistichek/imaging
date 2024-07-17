package types

import "fmt"

type Option func(k *ObjectKey)

func WithBase(base string) Option {
	return func(key *ObjectKey) {
		key.Base = base
	}
}

func WithFormat(format string) Option {
	return func(key *ObjectKey) {
		key.Postfix += "." + format
	}
}

type ObjectKey struct {
	Base    string
	Prefix  string
	Postfix string
}

func NewObjectKey(prefix, postfix string, opts ...Option) *ObjectKey {
	key := &ObjectKey{
		Prefix:  prefix,
		Postfix: postfix,
	}

	for _, opt := range opts {
		opt(key)
	}

	return key
}

func (key ObjectKey) String() string {
	str := fmt.Sprintf("%s/%s/%s", key.Base, key.Prefix, key.Postfix)
	return str
}
