package ravendb

type AfterSaveChangesEventArgs struct {
	_documentMetadata *IMetadataDictionary

	session    *InMemoryDocumentSessionOperations
	documentId string
	entity     Object
}

func NewAfterSaveChangesEventArgs(session *InMemoryDocumentSessionOperations, documentId string, entity Object) *AfterSaveChangesEventArgs {
	return &AfterSaveChangesEventArgs{
		session:    session,
		documentId: documentId,
		entity:     entity,
	}
}

func (a *AfterSaveChangesEventArgs) GetSession() *InMemoryDocumentSessionOperations {
	return a.session
}

func (a *AfterSaveChangesEventArgs) GetDocumentID() string {
	return a.documentId
}

func (a *AfterSaveChangesEventArgs) GetEntity() Object {
	return a.entity
}

func (a *AfterSaveChangesEventArgs) GetDocumentMetadata() *IMetadataDictionary {
	if a._documentMetadata == nil {
		a._documentMetadata, _ = a.session.GetMetadataFor(a.entity)
	}

	return a._documentMetadata
}
