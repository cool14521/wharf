package cmd

import (
	"github.com/spf13/cobra"

	"github.com/containerops/wharf/models"
)

// databasecmd is subcommand which migrate/backup/restore wharf's database.
var databaseCmd = &cobra.Command{
	Use:   "database",
	Short: "Database subcommand migrate/backup/restore wharf's database.",
	Long:  ``,
}

// migrateDatabaseCmd is subcommand migrate wharf's database.
var migrateDatabaseCmd = &cobra.Command{
	Use:   "migrate",
	Short: "migrate subcommand migrate wharf's database.",
	Long:  ``,
	Run:   migrateDatabase,
}

// backupDatabaseCmd is subcommand backup wharf's database.
var backupDatabaseCmd = &cobra.Command{
	Use:   "backup",
	Short: "backup subcommand backup wharf's database.",
	Long:  ``,
	Run:   backupDatabase,
}

// restoreDatabaseCmd is subcommand restore wharf's database.
var restoreDatabaseCmd = &cobra.Command{
	Use:   "restore",
	Short: "restore subcommand restore wharf's database.",
	Long:  ``,
	Run:   restoreDatabase,
}

// init()
func init() {
	RootCmd.AddCommand(databaseCmd)

	databaseCmd.AddCommand(migrateDatabaseCmd)
	databaseCmd.AddCommand(backupDatabaseCmd)
	databaseCmd.AddCommand(restoreDatabaseCmd)
}

// migrateDatabase is auto-migrate database of wharf.
func migrateDatabase(cmd *cobra.Command, args []string) {
	models.Migrate()
}

// backupDatabase is
func backupDatabase(cmd *cobra.Command, args []string) {

}

// restoreDatabase is
func restoreDatabase(cmd *cobra.Command, args []string) {

}
