package bookkeeper

// Notes on tags: use an array column and a GIN-index in Postgres is proven to
// be faster than table join
type Account struct {
	Id   int      `json:"Id"`
	Name string   `json:"Name"`
	Desc string   `json:"Desc"`
	Tags []string `json:"Tags"`
}

func (account *Account) Validate() bool {
	return true
}

func GetSqlCreateAccounts() string {
	return `create table accounts (
		id   serial,
		name text,
		desc_ text,
		tags text[],
		primary key(id)
	);`
}
