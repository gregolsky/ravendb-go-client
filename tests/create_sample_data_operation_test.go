package tests

import (
	"net/http"
	"github.com/ravendb/ravendb-go-client"
)

var (
	_ ravendb.IVoidMaintenanceOperation = &CreateSampleDataOperation{}
)

type CreateSampleDataOperation struct {
	Command *CreateSampleDataCommand
}

func NewCreateSampleDataOperation() *CreateSampleDataOperation {
	return &CreateSampleDataOperation{}
}

func (o *CreateSampleDataOperation) GetCommand(conventions *ravendb.DocumentConventions) ravendb.RavenCommand {
	o.Command = NewCreateSampleDataCommand(conventions)
	return o.Command
}

var _ ravendb.RavenCommand = &CreateSampleDataCommand{}

type CreateSampleDataCommand struct {
	*ravendb.RavenCommandBase
}

func NewCreateSampleDataCommand(conventions *ravendb.DocumentConventions) *CreateSampleDataCommand {
	cmd := &CreateSampleDataCommand{
		RavenCommandBase: ravendb.NewRavenCommandBase(),
	}
	cmd.RavenCommandBase.ResponseType = ravendb.RavenCommandResponseType_EMPTY
	return cmd
}

func (c *CreateSampleDataCommand) CreateRequest(node *ravendb.ServerNode) (*http.Request, error) {
	url := node.GetUrl() + "/databases/" + node.GetDatabase() + "/studio/sample-data"

	return ravendb.NewHttpPost(url, nil)
}
