package bookkeeper

// Notes on tags: use an array column and a GIN-index in Postgres is proven to
// be faster than table join
type Account struct {
	Id   int      `json:"id"`
	Name string   `json:"name"`
	Desc string   `json:"desc_"`
	Tags []string `json:"tags"`
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
