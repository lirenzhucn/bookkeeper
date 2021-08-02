package bookkeeper

// Notes on tags: use an array column and a GIN-index in Postgres is proven to
// be faster than table join
type Account struct {
	Id   int      `json:"id" header:"Id"`
	Name string   `json:"name" header:"Name"`
	Desc string   `json:"desc_" header:"Desc"`
	Tags []string `json:"tags" header:"Tags"`
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
