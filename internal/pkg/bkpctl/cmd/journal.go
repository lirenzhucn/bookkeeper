package cmd

import (
	"encoding/json"
	"io/ioutil"

	"github.com/lirenzhucn/bookkeeper/internal/pkg/bookkeeper"
	"github.com/spf13/cobra"
)

var journalCmd = &cobra.Command{
	Use:   "journal",
	Short: "Record a journal entry",
	Run:   recordActivity,
}

func initRecordCmd(rootCmd *cobra.Command) {
	journalCmd.Flags().StringP("template", "t", "", "template file")
	journalCmd.MarkFlagRequired("template")
	rootCmd.AddCommand(journalCmd)
}

func recordActivity(cmd *cobra.Command, args []string) {
	var entry bookkeeper.JournalEntry
	entry.Init(1)
	tplFile, err := cmd.Flags().GetString("template")
	cobra.CheckErr(err)
	buf, err := ioutil.ReadFile(tplFile)
	cobra.CheckErr(err)
	err = json.Unmarshal(buf, &entry)
	cobra.CheckErr(err)
}
