package cmd

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/lirenzhucn/bookkeeper/internal/pkg/bookkeeper"
	"github.com/spf13/cobra"
)

var importCmd = &cobra.Command{
	Use:   "import",
	Short: "Import data from various sources",
	Args:  cobra.NoArgs,
	Run:   importData,
}

func initImportCmd(rootCmd *cobra.Command) {
	importCmd.Flags().StringP("source", "s", "sui",
		"specify data source (default: sui)")
	importCmd.Flags().StringP(
		"config", "c", "",
		"config file that defines how import data are mapped to data model",
	)
	importCmd.MarkFlagRequired("config")
	importCmd.Flags().StringP("data", "d", "", "data file")
	importCmd.MarkFlagRequired("data")
	rootCmd.AddCommand(importCmd)
}

type ImportConfig struct {
	SourceType       string
	Accounts         map[string]bookkeeper.Account `json:"Accounts"`
	CategoryMap      map[string]string             `json:"CategoryMap"`
	SubCategoryMap   map[string]string             `json:"SubCategoryMap"`
	TransactionTypes map[string]string             `json:"TransactionTypes"`
	DateFormatter    string                        `json:"DateFormatter"`
}

func (c ImportConfig) Validate() bool {
	return c.SourceType == "sui"
}

func importData(cmd *cobra.Command, args []string) {
	sourceType, err := cmd.Flags().GetString("source")
	cobra.CheckErr(err)
	configPath, err := cmd.Flags().GetString("config")
	cobra.CheckErr(err)
	dataPath, err := cmd.Flags().GetString("data")
	cobra.CheckErr(err)

	config, err := readConfig(configPath, sourceType)
	cobra.CheckErr(err)
	err = postAccounts(&config.Accounts)
	cobra.CheckErr(err)

	err = readAndPostTransactions(dataPath, config)
	cobra.CheckErr(err)
}

func readConfig(configPath string, sourceType string) (ImportConfig, error) {
	var (
		importConfig ImportConfig
		err          error
	)
	configFile, err := os.Open(configPath)
	cobra.CheckErr(err)
	defer configFile.Close()
	err = json.NewDecoder(configFile).Decode(&importConfig)
	importConfig.SourceType = sourceType
	if !importConfig.Validate() {
		err = fmt.Errorf("failed to validate config loaded from %s", configPath)
	}
	return importConfig, err
}

func postAccounts(accountMap *map[string]bookkeeper.Account) error {
	var newAccount bookkeeper.Account
	url := BASE_URL + "accounts"
	for key, account := range *accountMap {
		buffer := new(bytes.Buffer)
		json.NewEncoder(buffer).Encode(account)
		resp, err := http.Post(url, "application/json", buffer)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		json.NewDecoder(resp.Body).Decode(&newAccount)
		(*accountMap)[key] = newAccount
	}
	return nil
}

func readAndPostTransactions(dataPath string, config ImportConfig) error {
	dataFile, err := os.Open(dataPath)
	if err != nil {
		return err
	}
	defer dataFile.Close()

	// Skip first row (line)
	row1, err := bufio.NewReader(dataFile).ReadSlice('\n')
	if err != nil {
		return err
	}
	_, err = dataFile.Seek(int64(len(row1)), io.SeekStart)
	if err != nil {
		return err
	}

	reader := csv.NewReader(dataFile)

	// read headers
	var (
		keys   []string
		record []string
	)
	keys, err = reader.Read()
	if err != nil {
		return err
	}
	for {
		record, err = reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		trans, err := createTransactionFromRowForSui(record, keys, &config)
		if err != nil {
			return err
		}
		_, err = postSingleTransaction(trans)
		if err != nil {
			return err
		}
	}

	return nil
}

func postSingleTransaction(trans bookkeeper.Transaction) (bookkeeper.Transaction, error) {
	var newTrans bookkeeper.Transaction
	url := BASE_URL + "transactions"
	buffer := new(bytes.Buffer)
	json.NewEncoder(buffer).Encode(trans)
	resp, err := http.Post(url, "application/json", buffer)
	if err != nil {
		return trans, err
	}
	defer resp.Body.Close()
	json.NewDecoder(resp.Body).Decode(&newTrans)
	return newTrans, nil
}

func createTransactionFromRowForSui(
	record []string, keys []string, config *ImportConfig,
) (bookkeeper.Transaction, error) {
	var (
		account bookkeeper.Account
		trans   bookkeeper.Transaction
		err     error
		ok      bool
	)
	for i, key := range keys {
		value := record[i]
		var notes []string
		switch key {
		case "交易类型":
			trans.Type, ok = config.TransactionTypes[value]
			if !ok {
				return trans, fmt.Errorf("invalid transaction type %s", value)
			}
		case "日期":
			trans.Date, err = time.Parse(config.DateFormatter, value)
			if err != nil {
				return trans, err
			}
		case "类别":
			if value == "" {
				trans.Category = ""
			} else {
				trans.Category, ok = config.CategoryMap[value]
				if !ok {
					return trans, fmt.Errorf("invalid transaction category %s", value)
				}
			}
		case "子类别":
			if value == "" {
				trans.SubCategory = ""
			} else {
				trans.SubCategory, ok = config.SubCategoryMap[value]
				if !ok {
					return trans, fmt.Errorf("invalid transaction sub-category %s", value)
				}
			}
		case "账户":
			account, ok = config.Accounts[value]
			if !ok {
				return trans, fmt.Errorf("invalid account name %s", value)
			}
			trans.AccountId = account.Id
		case "金额":
			amountFloat, err := strconv.ParseFloat(value, 64)
			if err != nil {
				return trans, err
			}
			trans.Amount = int64(amountFloat * 100)
		case "关联Id":
			trans.AssociationId = value
		case "备注":
			if value != "" {
				notes = append(notes, value)
			}
		case "商家":
			if value != "" {
				notes = append(notes, value)
			}
		}
		if len(notes) > 0 {
			trans.Notes = strings.Join(notes, "; ")
		}
	}
	return trans, nil
}
