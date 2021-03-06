package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ravendb/ravendb-go-client"
)

func deleteDocumentCommandTest_canDeleteDocument(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	{
		session := openSessionMust(t, store)
		user := NewUser()
		user.setName("Marcin")
		err = session.StoreWithID(user, "users/1")
		assert.NoError(t, err)
		err = session.SaveChanges()
		assert.NoError(t, err)
		session.Close()
	}
	command := ravendb.NewDeleteDocumentCommand("users/1", nil)
	err = store.GetRequestExecutor().ExecuteCommand(command)
	assert.NoError(t, err)
	{
		session := openSessionMust(t, store)
		loadedUserI, err := session.Load(ravendb.GetTypeOf(&User{}), "users/1")
		assert.NoError(t, err)
		loadedUser := loadedUserI.(*User)
		assert.Nil(t, loadedUser)
		session.Close()
	}
}

func deleteDocumentCommandTest_canDeleteDocumentByEtag(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	var changeVector *string
	{
		session := openSessionMust(t, store)
		user := NewUser()
		user.setName("Marcin")
		err = session.StoreWithID(user, "users/1")
		assert.NoError(t, err)
		err = session.SaveChanges()
		assert.NoError(t, err)
		changeVector, err = session.Advanced().GetChangeVectorFor(user)
		assert.NoError(t, err)
		session.Close()
	}

	{
		session := openSessionMust(t, store)
		loadedUserI, err := session.Load(ravendb.GetTypeOf(&User{}), "users/1")
		assert.NoError(t, err)
		loadedUser := loadedUserI.(*User)
		assert.NotNil(t, loadedUser)
		loadedUser.setAge(5)
		err = session.SaveChanges()
		assert.NoError(t, err)
		session.Close()
	}

	command := ravendb.NewDeleteDocumentCommand("users/1", changeVector)
	err = store.GetRequestExecutor().ExecuteCommand(command)
	assert.Error(t, err)
	_ = err.(*ravendb.ConcurrencyException)
}

func TestDeleteDocumentCommand(t *testing.T) {
	if dbTestsDisabled() {
		return
	}

	destroyDriver := createTestDriver(t)
	defer recoverTest(t, destroyDriver)

	// follows execution order of java tests
	deleteDocumentCommandTest_canDeleteDocument(t)
	deleteDocumentCommandTest_canDeleteDocumentByEtag(t)
}
