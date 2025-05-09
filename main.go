package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"path/filepath"
	"time"

	amqp_helper "github.com/CodeClarityCE/utility-amqp-helper"
	dbhelper "github.com/CodeClarityCE/utility-dbhelper/helper"
	types_amqp "github.com/CodeClarityCE/utility-types/amqp"
	codeclarity "github.com/CodeClarityCE/utility-types/codeclarity_db"
	plugin_db "github.com/CodeClarityCE/utility-types/plugin_db"
	plugin "github.com/parithera/plugin-fastqc/src"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
)

// Arguments struct to pass dependencies to the callback function.
type Arguments struct {
	codeclarity *bun.DB // Database connection.
}

// main is the entry point of the program.
// It reads the configuration, initializes the necessary databases and graph,
// and starts listening on the queue.
func main() {
	config, err := readConfig()
	if err != nil {
		log.Printf("%v", err) // Log the error if configuration reading fails.
		return
	}

	host := os.Getenv("PG_DB_HOST")
	if host == "" {
		log.Printf("PG_DB_HOST is not set") // Log if the environment variable is not set.
		return
	}
	port := os.Getenv("PG_DB_PORT")
	if port == "" {
		log.Printf("PG_DB_PORT is not set") // Log if the environment variable is not set.
		return
	}
	user := os.Getenv("PG_DB_USER")
	if user == "" {
		log.Printf("PG_DB_USER is not set") // Log if the environment variable is not set.
		return
	}
	password := os.Getenv("PG_DB_PASSWORD")
	if password == "" {
		log.Printf("PG_DB_PASSWORD is not set") // Log if the environment variable is not set.
		return
	}

	// Construct the database connection string.
	dsn := "postgres://" + user + ":" + password + "@" + host + ":" + port + "/" + dbhelper.Config.Database.Results + "?sslmode=disable"

	// Open a database connection.
	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn), pgdriver.WithTimeout(50*time.Second)))

	// Create a Bun database connection.
	db_codeclarity := bun.NewDB(sqldb, pgdialect.New())
	defer db_codeclarity.Close() // Ensure the database connection is closed when the function exits.

	// Create an Arguments struct to pass to the callback function.
	args := Arguments{
		codeclarity: db_codeclarity,
	}

	// Start listening on the queue.
	amqp_helper.Listen("dispatcher_"+config.Name, callback, args, config)
}

// startAnalysis performs the analysis using the specified plugin.
func startAnalysis(args Arguments, dispatcherMessage types_amqp.DispatcherPluginMessage, config plugin_db.Plugin, analysis_document codeclarity.Analysis) (map[string]any, codeclarity.AnalysisStatus, error) {

	// Get analysis config from the analysis document.
	messageData := analysis_document.Config[config.Name].(map[string]any)

	// Get the download path from the environment variables.
	path := os.Getenv("DOWNLOAD_PATH")

	// Prepare the sample path for the plugin.
	sample := filepath.Join(path, dispatcherMessage.OrganizationId.String(), "samples", messageData["sample"].(string))

	// Start the plugin and get the output.
	rOutput := plugin.Start(sample, args.codeclarity)

	// Create a result object to store the plugin output.
	result := codeclarity.Result{
		Result:     rOutput,
		AnalysisId: dispatcherMessage.AnalysisId,
		Plugin:     config.Name,
	}

	// Insert the result into the database.
	_, err := args.codeclarity.NewInsert().Model(&result).Exec(context.Background())
	if err != nil {
		panic(err) // Handle the error appropriately in a production environment.
	}

	// Prepare the result to store in step.
	// In this case we only store the sbomKey.
	// The other plugins will use this key to get the sbom.
	res := make(map[string]any)
	res["rKey"] = result.Id

	// The output is always a map[string]any.
	return res, rOutput.AnalysisInfo.Status, nil
}
