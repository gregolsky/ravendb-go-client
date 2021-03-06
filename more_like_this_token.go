package ravendb

var _ QueryToken = &MoreLikeThisToken{}

type MoreLikeThisToken struct {
	documentParameterName string
	optionsParameterName  string
	whereTokens           []QueryToken
}

func NewMoreLikeThisToken() *MoreLikeThisToken {
	return &MoreLikeThisToken{}
}

func (t *MoreLikeThisToken) WriteTo(writer *StringBuilder) {
	writer.append("moreLikeThis(")

	if t.documentParameterName == "" {
		var prevToken QueryToken
		for _, whereToken := range t.whereTokens {
			DocumentQueryHelper_addSpaceIfNeeded(prevToken, whereToken, writer)
			whereToken.WriteTo(writer)
			prevToken = whereToken
		}
	} else {
		writer.append("$")
		writer.append(t.documentParameterName)
	}

	if t.optionsParameterName == "" {
		writer.append(")")
		return
	}

	writer.append(", $")
	writer.append(t.optionsParameterName)
	writer.append(")")

}
