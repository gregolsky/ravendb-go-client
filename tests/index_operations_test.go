package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ravendb/ravendb-go-client"
)

func NewUsersInvalidIndex() *ravendb.AbstractIndexCreationTask {
	res := ravendb.NewAbstractIndexCreationTask("UsersInvalidIndex")
	res.Map = "from u in docs.Users select new { a = 5 / u.Age }"
	return res
}

func NewUsers_Index() *ravendb.AbstractIndexCreationTask {
	res := ravendb.NewAbstractIndexCreationTask("Users_Index")
	res.Map = "from u in docs.Users select new { u.name }"
	return res
}

func testIndexCanDeleteIndex(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	index := NewUsersIndex()
	err = index.Execute(store)
	assert.NoError(t, err)

	op := ravendb.NewGetIndexNamesOperation(0, 10)
	err = store.Maintenance().Send(op)
	assert.NoError(t, err)
	indexNames := op.Command.Result
	assert.True(t, ravendb.StringArrayContains(indexNames, "UsersIndex"))

	op2 := ravendb.NewDeleteIndexOperation("UsersIndex")
	err = store.Maintenance().Send(op2)
	assert.NoError(t, err)

	op3 := ravendb.NewGetIndexNamesOperation(0, 10)
	err = store.Maintenance().Send(op3)
	assert.NoError(t, err)
	indexNames = op3.Command.Result
	assert.Equal(t, len(indexNames), 0)
}

func testIndexCanDisableAndEnableIndex(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	index := NewUsersIndex()
	err = index.Execute(store)
	assert.NoError(t, err)

	{
		op := ravendb.NewDisableIndexOperation("UsersIndex")
		err = store.Maintenance().Send(op)
		assert.NoError(t, err)
	}

	{
		op := ravendb.NewGetIndexingStatusOperation()
		err = store.Maintenance().Send(op)
		assert.NoError(t, err)
		indexingStatus := op.Command.Result
		indexStatus := indexingStatus.GetIndexes()[0]
		assert.Equal(t, indexStatus.GetStatus(), ravendb.IndexRunningStatus_DISABLED)
	}

	{
		op := ravendb.NewEnableIndexOperation("UsersIndex")
		err = store.Maintenance().Send(op)
		assert.NoError(t, err)
	}

	{
		op := ravendb.NewGetIndexingStatusOperation()
		err = store.Maintenance().Send(op)
		assert.NoError(t, err)
		indexingStatus := op.Command.Result
		indexStatus := indexingStatus.GetIndexes()[0]
		assert.Equal(t, indexStatus.GetStatus(), ravendb.IndexRunningStatus_RUNNING)
	}
}

func testIndexGetCanIndexes(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	index := NewUsersIndex()
	err = index.Execute(store)
	assert.NoError(t, err)

	{
		op := ravendb.NewGetIndexesOperation(0, 10)
		err = store.Maintenance().Send(op)
		assert.NoError(t, err)
		indexDefinitions := op.Command.Result
		assert.Equal(t, len(indexDefinitions), 1)
	}
}

func testIndexGetCanIndexesStats(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	index := NewUsersIndex()
	err = index.Execute(store)
	assert.NoError(t, err)

	{
		op := ravendb.NewGetIndexesStatisticsOperation()
		err = store.Maintenance().Send(op)
		assert.NoError(t, err)
		indexStats := op.Command.Result
		assert.Equal(t, len(indexStats), 1)
	}
}

func testIndexGetTerms(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	index := NewUsersIndex()
	err = index.Execute(store)
	assert.NoError(t, err)

	{
		session := openSessionMust(t, store)
		user := NewUser()
		user.setName("Marcin")
		err = session.Store(user)
		assert.NoError(t, err)
		err = session.SaveChanges()
		assert.NoError(t, err)
		session.Close()
	}

	err = gRavenTestDriver.waitForIndexing(store, store.GetDatabase(), 0)
	assert.NoError(t, err)

	{
		op := ravendb.NewGetTermsOperation("UsersIndex", "name", "")
		err = store.Maintenance().Send(op)
		assert.NoError(t, err)
		terms := op.Command.Result
		assert.Equal(t, len(terms), 1)
		assert.Equal(t, terms[0], "marcin")
	}
}

func testIndexHasIndexChanged(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	index := NewUsersIndex()
	indexDef := index.CreateIndexDefinition()
	op := ravendb.NewPutIndexesOperation(indexDef)
	err = store.Maintenance().Send(op)
	assert.NoError(t, err)
	op2 := ravendb.NewIndexHasChangedOperation(indexDef)
	err = store.Maintenance().Send(op2)
	assert.NoError(t, err)
	{
		cmd := op2.Command
		assert.False(t, cmd.Result)
	}
	m := ravendb.NewStringSetFromStrings("from users")
	indexDef.SetMaps(m)

	op3 := ravendb.NewIndexHasChangedOperation(indexDef)
	err = store.Maintenance().Send(op3)
	assert.NoError(t, err)
	{
		cmd := op3.Command
		assert.True(t, cmd.Result)
	}
}

func testIndexCanStopStartIndexing(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	index := NewUsersIndex()
	indexDef := index.CreateIndexDefinition()
	{
		op := ravendb.NewPutIndexesOperation(indexDef)
		err = store.Maintenance().Send(op)
		assert.NoError(t, err)
	}

	{
		op := ravendb.NewStopIndexingOperation()
		err = store.Maintenance().Send(op)
		assert.NoError(t, err)
	}

	{
		op := ravendb.NewGetIndexingStatusOperation()
		err = store.Maintenance().Send(op)
		assert.NoError(t, err)
		indexingStatus := op.Command.Result
		indexStatus := indexingStatus.GetIndexes()[0]
		assert.Equal(t, indexStatus.GetStatus(), ravendb.IndexRunningStatus_PAUSED)
	}

	{
		op := ravendb.NewStartIndexingOperation()
		err = store.Maintenance().Send(op)
		assert.NoError(t, err)
	}

	{
		op := ravendb.NewGetIndexingStatusOperation()
		err = store.Maintenance().Send(op)
		assert.NoError(t, err)
		indexingStatus := op.Command.Result
		indexStatus := indexingStatus.GetIndexes()[0]
		assert.Equal(t, indexStatus.GetStatus(), ravendb.IndexRunningStatus_RUNNING)
	}
}

func testIndexCanStopStartIndex(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	index := NewUsersIndex()
	indexDef := index.CreateIndexDefinition()
	{
		op := ravendb.NewPutIndexesOperation(indexDef)
		err = store.Maintenance().Send(op)
		assert.NoError(t, err)
	}

	{
		op := ravendb.NewStopIndexOperation(indexDef.GetName())
		err = store.Maintenance().Send(op)
		assert.NoError(t, err)
	}

	{
		op := ravendb.NewGetIndexingStatusOperation()
		err = store.Maintenance().Send(op)
		assert.NoError(t, err)
		indexingStatus := op.Command.Result
		assert.Equal(t, indexingStatus.GetStatus(), ravendb.IndexRunningStatus_RUNNING)
		indexStatus := indexingStatus.GetIndexes()[0]
		assert.Equal(t, indexStatus.GetStatus(), ravendb.IndexRunningStatus_PAUSED)
	}

	{
		op := ravendb.NewStartIndexOperation(indexDef.GetName())
		err = store.Maintenance().Send(op)
		assert.NoError(t, err)
	}

	{
		op := ravendb.NewGetIndexingStatusOperation()
		err = store.Maintenance().Send(op)
		assert.NoError(t, err)
		indexingStatus := op.Command.Result
		assert.Equal(t, indexingStatus.GetStatus(), ravendb.IndexRunningStatus_RUNNING)
		indexStatus := indexingStatus.GetIndexes()[0]
		assert.Equal(t, indexStatus.GetStatus(), ravendb.IndexRunningStatus_RUNNING)
	}
}

func testIndexCanSetIndexLockMode(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	index := NewUsersIndex()
	indexDef := index.CreateIndexDefinition()
	{
		op := ravendb.NewPutIndexesOperation(indexDef)
		err = store.Maintenance().Send(op)
		assert.NoError(t, err)
	}

	{
		op := ravendb.NewSetIndexesLockOperation(indexDef.GetName(), ravendb.IndexLockMode_LOCKED_ERROR)
		err = store.Maintenance().Send(op)
		assert.NoError(t, err)
	}

	{
		op := ravendb.NewGetIndexOperation(indexDef.GetName())
		err = store.Maintenance().Send(op)
		newIndexDef := op.Command.Result
		assert.Equal(t, *newIndexDef.GetLockMode(), ravendb.IndexLockMode_LOCKED_ERROR)
	}
}

func testIndexCanSetIndexPriority(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	index := NewUsersIndex()
	indexDef := index.CreateIndexDefinition()
	op := ravendb.NewPutIndexesOperation(indexDef)
	err = store.Maintenance().Send(op)
	assert.NoError(t, err)

	op2 := ravendb.NewSetIndexesPriorityOperation(indexDef.GetName(), ravendb.IndexPriority_HIGH)
	err = store.Maintenance().Send(op2)
	assert.NoError(t, err)

	op3 := ravendb.NewGetIndexOperation(indexDef.GetName())
	err = store.Maintenance().Send(op3)
	newIndexDef := op3.Command.Result
	assert.Equal(t, *newIndexDef.GetPriority(), ravendb.IndexPriority_HIGH)
}

func testIndexCanListErrors(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	index := NewUsersInvalidIndex()
	indexDef := index.CreateIndexDefinition()
	op := ravendb.NewPutIndexesOperation(indexDef)
	err = store.Maintenance().Send(op)
	assert.NoError(t, err)

	{
		session := openSessionMust(t, store)
		user := NewUser()
		//user.setName(null)
		user.setAge(0)
		err = session.Store(user)
		assert.NoError(t, err)
		err = session.SaveChanges()
		assert.NoError(t, err)
		session.Close()
	}

	err = gRavenTestDriver.waitForIndexing(store, store.GetDatabase(), 0)
	assert.NoError(t, err)

	op2 := ravendb.NewGetIndexErrorsOperation(nil)
	err = store.Maintenance().Send(op2)
	assert.NoError(t, err)
	indexErrors := op2.Command.Result
	assert.Equal(t, len(indexErrors), 1)

	op3 := ravendb.NewGetIndexErrorsOperation([]string{indexDef.GetName()})
	err = store.Maintenance().Send(op3)
	assert.NoError(t, err)
	perIndexErrors := op3.Command.Result
	assert.Equal(t, len(perIndexErrors), 1)
}

func testIndexCanGetIndexStatistics(t *testing.T) {

	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	userIndex := NewUsers_Index()
	err = userIndex.Execute(store)
	assert.NoError(t, err)

	op := ravendb.NewGetIndexesStatisticsOperation()
	err = store.Maintenance().Send(op)
	assert.NoError(t, err)

	indexStats := op.Command.Result
	assert.Equal(t, len(indexStats), 1)
}

func TestIndexOperations(t *testing.T) {
	if dbTestsDisabled() {
		return
	}

	destroyDriver := createTestDriver(t)
	defer recoverTest(t, destroyDriver)

	// order matches Java tests
	testIndexHasIndexChanged(t)
	testIndexCanListErrors(t)
	testIndexCanGetIndexStatistics(t)
	testIndexCanSetIndexPriority(t)
	testIndexCanDisableAndEnableIndex(t)
	testIndexGetCanIndexes(t)
	testIndexCanDeleteIndex(t)
	testIndexCanStopStartIndexing(t)
	testIndexCanGetIndexStatistics(t)
	testIndexCanStopStartIndex(t)
	testIndexCanSetIndexLockMode(t)
	testIndexGetTerms(t)
}
