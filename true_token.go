package ravendb

var _ QueryToken = &TrueToken{}

var TrueToken_INSTANCE = NewTrueToken()

type TrueToken struct {
}

func NewTrueToken() *TrueToken {
	return &TrueToken{}
}

func (t *TrueToken) WriteTo(writer *StringBuilder) {
	writer.append("true")
}
