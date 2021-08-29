package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

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

func getAccountByName(accountName string) (account bookkeeper.Account, err error) {
	url_ := fmt.Sprintf("%saccounts?accountName=%s", BASE_URL,
		url.QueryEscape(accountName))
	resp, err := http.Get(url_)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}
	err = json.Unmarshal(body, &account)
	return
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

func patchSingleTransaction(trans bookkeeper.Transaction) (bookkeeper.Transaction, error) {
	var newTrans bookkeeper.Transaction
	url_ := fmt.Sprintf("%stransactions/%d", BASE_URL, trans.Id)
	buffer := new(bytes.Buffer)
	json.NewEncoder(buffer).Encode(trans)

	// prepare a PATCH
	client := &http.Client{}
	req, err := http.NewRequest(http.MethodPatch, url_, bytes.NewBuffer(buffer.Bytes()))
	if err != nil {
		return trans, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return trans, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return trans, fmt.Errorf("failed to update transaction; status: %s", resp.Status)
	}
	json.NewDecoder(resp.Body).Decode(&newTrans)
	return newTrans, nil
}

func deleteSingleTransaction(transId int) (err error) {
	url_ := fmt.Sprintf("%stransactions/%d", BASE_URL, transId)
	// prepare a DELETE
	client := &http.Client{}
	req, err := http.NewRequest(http.MethodDelete, url_, strings.NewReader(""))
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if resp.StatusCode != 200 {
		err = fmt.Errorf("failed to delete transaction %d; status: %s", transId, resp.Status)
		return
	}
	return
}

func getTransactionsByQuery(queryStr string) (transactions []bookkeeper.Transaction_, err error) {
	url_ := fmt.Sprintf("%stransactions?queryString=%s", BASE_URL,
		url.QueryEscape(queryStr))
	resp, err := http.Get(url_)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}
	err = json.Unmarshal(body, &transactions)
	if err != nil {
		return
	}
	return
}
