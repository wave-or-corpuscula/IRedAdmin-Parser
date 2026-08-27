package parser

type Server struct {
	Name string `db:"name"`
}

type Domain struct {
	Disabled        bool   `db:"disabled"`
	Name            string `db:"name"`
	DisplayName     string `db:"display_name"`
	QuotaBytes      int64  `db:"quota_bytes"`
	UsedMemoryBytes int64  `db:"used_memory_bytes"`
	UsersAmount     int
}

type Mailbox struct {
	Disabled        bool   `db:"disabled"`
	IsAdmin         bool   `db:"is_admin"`
	DisplayName     string `db:"display_name"`
	Address         string `db:"address"`
	QuotaBytes      int64  `db:"quota_bytes"`
	UsedMemoryBytes int64  `db:"used_memory_bytes"`
}

type ParseDomainResult struct {
	Domains []*Domain
	Total   int
	Errors  []error
}

type ParseMailboxesResult struct {
	Mailboxes []*Mailbox
	Total     int
	Errors    []error
}

func (r *ParseMailboxesResult) Extend(other *ParseMailboxesResult) {
	r.Mailboxes = append(r.Mailboxes, other.Mailboxes...)
	r.Total += other.Total
	r.Errors = append(r.Errors, other.Errors...)
}
