package database

import (
	"fmt"
	"iredparser/internal/parser"
	"testing"

	"github.com/stretchr/testify/assert"
)

func getTestMailbox(num int) *parser.Mailbox {
	return &parser.Mailbox{
		Address:         fmt.Sprintf("mailbox %d", num),
		DisplayName:     fmt.Sprintf("mailbox %d", num),
		Disabled:        false,
		IsAdmin:         false,
		QuotaBytes:      int64(num * 1024),
		UsedMemoryBytes: int64(num*1024 - num*100),
	}
}

func getTestDomain(num int) *parser.Domain {
	return &parser.Domain{
		Name:            fmt.Sprintf("domain %d", num),
		DisplayName:     fmt.Sprintf("Domain %d", num),
		Disabled:        true,
		QuotaBytes:      int64(2048 * num),
		UsedMemoryBytes: int64(1024 * num),
	}
}

func TestInitDB(t *testing.T) {
	db, err := GetTestDB(":memory:")
	assert.NoError(t, err)

	defer db.Close()
}

func TestServerCRUD(t *testing.T) {
	server := GetTestServer(1)

	db, err := GetTestDB(":memory:")
	assert.NoError(t, err)

	_, err = db.UpsertServer(server)
	assert.NoError(t, err)
	newServer, err := db.UpsertServer(server)
	assert.NoError(t, err)

	servers, err := db.GetServers()
	assert.NoError(t, err)

	for _, server := range servers {
		if server.ID == newServer.ID {
			assert.Equal(t, server, *newServer)
			break
		}
	}
	err = db.DeleteServer(newServer.ID)
	assert.NoError(t, err)

	servers, err = db.GetServers()
	assert.NoError(t, err)

	for _, s := range servers {
		assert.NotEqual(t, s.Name, newServer)
	}
}

func TestDomainsCRUD(t *testing.T) {
	db, err := GetTestDB(":memory:")
	assert.NoError(t, err)

	serv := GetTestServer(1)
	server, err := db.UpsertServer(serv)
	assert.NoError(t, err)

	domain := getTestDomain(1)
	model, err := db.UpsertDomain(domain, server.ID)
	assert.NoError(t, err)

	domains, err := db.GetDomains()
	assert.NoError(t, err)
	found := false
	for _, domain := range domains {
		if domain.ID == model.ID {
			assert.Equal(t, domain, model)
			found = true
			break
		}
	}
	assert.True(t, found)

	err = db.DeleteDomain(model.ID)
	assert.NoError(t, err)

	domains, err = db.GetDomains()
	assert.NoError(t, err)

	found = false
	for _, domain := range domains {
		if domain.ID == model.ID {
			assert.Equal(t, domain, model)
			found = true
			break
		}
	}
	assert.False(t, found)
}

func TestMailboxCRUD(t *testing.T) {
	db, err := GetTestDB(":memory:")
	assert.NoError(t, err)

	ser := GetTestServer(1)
	server, err := db.UpsertServer(ser)
	assert.NoError(t, err)

	domain := getTestDomain(1)
	newDomain, err := db.UpsertDomain(domain, server.ID)
	assert.NoError(t, err)

	boxesN := 10
	boxes := []*parser.Mailbox{}

	for i := range boxesN {
		boxes = append(boxes, getTestMailbox(i+1))
	}

	upsertedBoxes := []*MailboxModel{}
	for _, box := range boxes {
		model, err := db.UpsertMailbox(box, newDomain.ID)
		assert.NoError(t, err)

		upsertedBoxes = append(upsertedBoxes, model)
	}

	dbBoxes, err := db.GetMailboxes()
	assert.NoError(t, err)
	assert.Len(t, dbBoxes, boxesN)

	assert.Equal(t, upsertedBoxes, dbBoxes)

	for _, box := range dbBoxes {
		err = db.DeleteMailbox(box.ID)
		assert.NoError(t, err)
	}

	dbBoxes, err = db.GetMailboxes()
	assert.NoError(t, err)
	assert.Len(t, dbBoxes, 0)
}

func TestGetServerID(t *testing.T) {
	db, server, err := GetTestDBWithServer(1)
	assert.NoError(t, err)

	srv, err := db.GetServer(server.Name)
	assert.NoError(t, err)

	assert.Equal(t, srv, server)
}
