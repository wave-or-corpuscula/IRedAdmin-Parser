// Package database provides database interaction controller
package database

import (
	"fmt"
	"log"

	"iredparser/internal/parser"

	"github.com/jmoiron/sqlx"

	_ "modernc.org/sqlite"
)

type ServerModel struct {
	ID int64 `db:"id"`
	parser.Server
}

type DomainModel struct {
	ID       int64 `db:"id"`
	ServerID int64 `db:"server_id"`
	parser.Domain
}

type MailboxModel struct {
	ID       int64 `db:"id"`
	DomainID int64 `db:"domain_id"`
	parser.Mailbox
}

type Database struct {
	db *sqlx.DB
}

func Connect(dsn string) (*Database, error) {
	db, err := sqlx.Connect("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to sqlite: %w", err)
	}

	storage := &Database{db: db}

	if err := storage.initSchema(); err != nil {
		db.Close()
		return nil, err
	}

	return storage, nil
}

func (d *Database) initSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS "Servers" (
			"id" INTEGER,
			"name" TEXT NOT NULL UNIQUE,
			PRIMARY KEY("id" AUTOINCREMENT)
	);

	CREATE TABLE IF NOT EXISTS "Domains" (
			"id" INTEGER,
			"server_id" INTEGER NOT NULL,
			"disabled" BLOB NOT NULL,
			"name" TEXT NOT NULL,
			"display_name" TEXT,
			"quota_bytes" INTEGER NOT NULL,
			"used_memory_bytes" INTEGER NOT NULL,
			PRIMARY KEY("id" AUTOINCREMENT),
			UNIQUE("server_id", "name"),
			FOREIGN KEY("server_id") REFERENCES "Servers"("id") ON DELETE CASCADE ON UPDATE CASCADE
	);

	CREATE TABLE IF NOT EXISTS "Mailboxes" (
			"id" INTEGER,
			"domain_id" INTEGER NOT NULL,
			"address" TEXT NOT NULL,
			"display_name" TEXT,
			"disabled" BLOB NOT NULL,
			"is_admin" BLOB NOT NULL,
			"quota_bytes" INTEGER NOT NULL,
			"used_memory_bytes" INTEGER NOT NULL,
			PRIMARY KEY("id" AUTOINCREMENT),
			UNIQUE("domain_id", "address"),
			FOREIGN KEY("domain_id") REFERENCES "Domains"("id") ON DELETE CASCADE ON UPDATE CASCADE
	);	
	PRAGMA foreign_keys = ON;
	`

	_, err := d.db.Exec(schema)
	log.Println(err)
	return err
}

func (d *Database) Close() error {
	return d.db.Close()
}

func (d *Database) UpsertServer(server *parser.Server) (*ServerModel, error) {
	query := `INSERT INTO Servers (name) VALUES (:name) 

	ON CONFLICT(name) DO UPDATE SET name = name

	RETURNING id;`
	model := ServerModel{Server: *server}
	rows, err := d.db.NamedQuery(query, model)
	if err != nil {
		return nil, fmt.Errorf("failed to insert server: %w", err)
	}
	defer rows.Close()

	if rows.Next() {
		err = rows.StructScan(&model)
		if err != nil {
			return nil, fmt.Errorf("cannot read new server id: %w", err)
		}
	}
	return &model, nil
}

func (d *Database) GetServers() ([]ServerModel, error) {
	serverModels := []ServerModel{}

	query := `SELECT * FROM Servers;`

	err := d.db.Select(&serverModels, query)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch servers: %w", err)
	}

	return serverModels, nil
}

func (d *Database) DeleteServer(id int64) error {
	query := `DELETE FROM Servers WHERE id = ?;`

	_, err := d.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed delete server: %w", err)
	}
	return nil
}

func (d *Database) GetServerID(name string) (int64, error) {
	query := `SELECT id FROM Servers WHERE name = ?;`

	var ID int64
	err := d.db.Get(&ID, query, name)
	if err != nil {
		return -1, fmt.Errorf("failed to get server id: %w", err)
	}

	return ID, nil
}

func (d *Database) UpsertDomain(domain *parser.Domain, serverID int64) (*DomainModel, error) {
	model, err := d.UpsertDomainMany([]*parser.Domain{domain}, serverID)
	if err != nil {
		return nil, err
	}

	return model[0], nil
}

func (d *Database) UpsertDomainMany(domains []*parser.Domain, serverID int64) ([]*DomainModel, error) {
	tx, err := d.db.Beginx()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction in domains: %w", err)
	}
	defer tx.Rollback()

	domainModels := []*DomainModel{}

	for _, domain := range domains {
		domainModel, err := d.upsertDomainTx(tx, domain, serverID)
		if err != nil {
			return nil, fmt.Errorf("failed to upsert domain %q: %w", domain.Name, err)
		}
		domainModels = append(domainModels, domainModel)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit domains transaction: %w", err)
	}

	return domainModels, err
}

func (d *Database) upsertDomainTx(tx *sqlx.Tx, domain *parser.Domain, serverID int64) (*DomainModel, error) {
	query := `
	INSERT INTO Domains (server_id, disabled, name, display_name, quota_bytes, used_memory_bytes) 
	VALUES (:server_id, :disabled, :name, :display_name, :quota_bytes, :used_memory_bytes)

	ON CONFLICT(server_id, name) DO UPDATE SET
		disabled = EXCLUDED.disabled,
		display_name = EXCLUDED.display_name,
		quota_bytes = EXCLUDED.quota_bytes,
		used_memory_bytes = EXCLUDED.used_memory_bytes

	RETURNING id;`

	domainModel := &DomainModel{ServerID: serverID, Domain: *domain}
	rows, err := tx.NamedQuery(query, domainModel)
	if err != nil {
		return nil, fmt.Errorf("failed to upsert domain: %w", err)
	}
	defer rows.Close()

	if rows.Next() {
		err = rows.StructScan(domainModel)
		if err != nil {
			return nil, fmt.Errorf("cannot scan domain id: %w", err)
		}
	}

	return domainModel, nil
}

func (d *Database) GetDomains() ([]*DomainModel, error) {
	query := `SELECT * FROM Domains;`
	domains := []*DomainModel{}

	err := d.db.Select(&domains, query)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch domains: %w", err)
	}

	return domains, nil
}

func (d *Database) DeleteDomain(id int64) error {
	query := `DELETE FROM Domains WHERE id = ?;`

	_, err := d.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete domain: %w", err)
	}
	return nil
}

func (d *Database) UpsertMailbox(mailbox *parser.Mailbox, DomainID int64) (*MailboxModel, error) {
	model, err := d.UpsertMailboxMany([]*parser.Mailbox{mailbox}, DomainID)
	if err != nil {
		return nil, err
	}

	return model[0], nil
}

func (d *Database) UpsertMailboxMany(mailboxes []*parser.Mailbox, DomainID int64) ([]*MailboxModel, error) {
	tx, err := d.db.Beginx()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction in mailboxes: %w", err)
	}
	defer tx.Rollback()

	mailboxModels := []*MailboxModel{}

	for _, mailbox := range mailboxes {
		model, err := d.upsertMailboxTx(tx, mailbox, DomainID)
		if err != nil {
			return nil, fmt.Errorf("failed to upsert mailbox %q: %w", mailbox.Address, err)
		}

		mailboxModels = append(mailboxModels, model)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit mailboxes: %w", err)
	}

	return mailboxModels, nil
}

func (d *Database) upsertMailboxTx(tx *sqlx.Tx, mailbox *parser.Mailbox, DomainID int64) (*MailboxModel, error) {
	query := `
	INSERT INTO Mailboxes (domain_id, address, display_name, disabled, is_admin, quota_bytes, used_memory_bytes)
	VALUES (:domain_id, :address, :display_name, :disabled, :is_admin, :quota_bytes, :used_memory_bytes) 

	ON CONFLICT(domain_id, address) DO UPDATE SET
		display_name      = EXCLUDED.display_name,
		disabled          = EXCLUDED.disabled,
		is_admin          = EXCLUDED.is_admin,
		quota_bytes       = EXCLUDED.quota_bytes,
		used_memory_bytes = EXCLUDED.used_memory_bytes

	RETURNING id;`

	model := &MailboxModel{DomainID: DomainID, Mailbox: *mailbox}
	rows, err := tx.NamedQuery(query, model)
	if err != nil {
		return nil, fmt.Errorf("failed to insert user: %w", err)
	}
	defer rows.Close()

	if rows.Next() {
		err = rows.StructScan(model)
		if err != nil {
			return nil, fmt.Errorf("cannot scan mailbox id: %w", err)
		}
	}

	return model, nil
}

func (d *Database) GetMailboxes() ([]*MailboxModel, error) {
	query := `SELECT * FROM Mailboxes;`

	mailboxes := []*MailboxModel{}
	err := d.db.Select(&mailboxes, query)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch mailboxes: %w", err)
	}

	return mailboxes, nil
}

func (d *Database) DeleteMailbox(id int64) error {
	query := `DELETE FROM Mailboxes WHERE id = ?;`

	_, err := d.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete mailbox: %w", err)
	}

	return nil
}
