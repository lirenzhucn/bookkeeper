package bookkeeper

type Account struct {
	Id       int    `json:"Id"`
	Name     string `json:"Name"`
	Type     string `json:"Type"`
	Category string `json:"Category"`
}

var VALID_ACCOUNT_TYPES = []string{"Credit", "Debit"}
var VALID_ACCOUNT_CATEGORIES = []string{"Asset", "Liability", "Revenue", "Expense"}

func stringInList(s string, l []string) bool {
	for _, ss := range l {
		if s == ss {
			return true
		}
	}
	return false
}

func (account *Account) Validate() bool {
	return stringInList(account.Type, VALID_ACCOUNT_TYPES) &&
		stringInList(account.Category, VALID_ACCOUNT_CATEGORIES)
}
