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
}

type Mailbox struct {
	Disabled        bool   `db:"disabled"`
	IsAdmin         bool   `db:"is_admin"`
	DisplayName     string `db:"display_name"`
	Address         string `db:"address"`
	QuotaBytes      int64  `db:"quota_bytes"`
	UsedMemoryBytes int64  `db:"used_memory_bytes"`
}
