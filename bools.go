package otk

// Tribools
var (
	True    = BoolPtr(true)
	False   = BoolPtr(false)
	Nilbool *bool
)

func BoolPtr(b bool) *bool {
	return &b
}

func BoolVal(b *bool) bool {
	if b == nil {
		return false
	}
	return *b
}
