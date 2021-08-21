package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/lirenzhucn/bookkeeper/internal/pkg/bookkeeper"
)

func getTransactionById(transId int, trans *bookkeeper.Transaction_) (err error) {
	url_ := fmt.Sprintf("%stransactions/%d", BASE_URL, transId)
	resp, err := http.Get(url_)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf(
			"failed to get transaction %d; response status: %s",
			transId, resp.Status,
		)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}
	err = json.Unmarshal(body, trans)
	return
}

func postAccounts(accountMap *map[string]bookkeeper.Account) error {
	var newAccount bookkeeper.Account
	url_ := BASE_URL + "accounts"
	for key, account := range *accountMap {
		buffer := new(bytes.Buffer)
		json.NewEncoder(buffer).Encode(account)
		resp, err := http.Post(url_, "application/json", buffer)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			return fmt.Errorf(
				"failed to insert account with name %s", account.Name)
		}
		json.NewDecoder(resp.Body).Decode(&newAccount)
		(*accountMap)[key] = newAccount
	}
	return nil
}

func getAllAccounts(accounts *[]bookkeeper.Account) error {
	url_ := BASE_URL + "accounts"
	resp, err := http.Get(url_)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("failed to get accounts; response status: %s", resp.Status)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	err = json.Unmarshal(body, accounts)
	if err != nil {
		return err
	}
	return nil
}

func postSingleTransaction(trans bookkeeper.Transaction) (bookkeeper.Transaction, error) {
	var newTrans bookkeeper.Transaction
	url_ := BASE_URL + "transactions"
	buffer := new(bytes.Buffer)
	json.NewEncoder(buffer).Encode(trans)
	resp, err := http.Post(url_, "application/json", buffer)
	if err != nil {
		return trans, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return trans, fmt.Errorf("failed to insert transaction")
	}
	json.NewDecoder(resp.Body).Decode(&newTrans)
	return newTrans, nil
}
