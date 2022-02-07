package otk

import "go.oneofone.dev/otk/sets"

type (
	Set      = sets.Set
	SafeSet  = sets.SafeSet
	MultiSet = sets.MultiSet
)

func NewSet(keys ...string) Set {
	return sets.SetOf(keys...)
}

func NewSafeSet(keys ...string) *SafeSet {
	return sets.SafeSetOf(keys...)
}

func NewMultiSet() MultiSet {
	return sets.MultiSet{}
}
