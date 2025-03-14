package fastqc

import (
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"time"

	codeclarity "github.com/CodeClarityCE/utility-types/codeclarity_db"
	exceptionManager "github.com/CodeClarityCE/utility-types/exceptions"
	"github.com/uptrace/bun"

	"github.com/parithera/plugin-fastqc/src/types"
	"github.com/parithera/plugin-fastqc/src/utils/output_generator"
)

// Start is a function that analyzes the source code directory and generates a software bill of materials (SBOM) output.
// It returns an sbomTypes.Output struct containing the analysis results.
func Start(sourceCodeDir string, codeclarityDB *bun.DB) types.Output {
	return ExecuteScript(sourceCodeDir)
}

// ExecuteScript executes a script on the provided source code directory and returns the output.
// The function searches for .fastq.gz files in the directory, runs fastqc on them, and generates an output based on the results.
func ExecuteScript(sourceCodeDir string) types.Output {
	// Record the start time of the analysis
	start := time.Now()

	// Search for .fastq.gz files in the source code directory
	files, err := filepath.Glob(sourceCodeDir + "/*.fastq.gz")
	if err != nil {
		// Log and exit if an error occurs while searching for files
		log.Fatal(err)
	}

	// Check if any .fastq.gz files were found
	if len(files) == 0 {
		// If no files are found, return an output indicating success with a message
		return generate_output(start, "no fastq file", codeclarity.SUCCESS, []exceptionManager.Error{})
	}

	// Create the output directory for fastqc results
	outputPath := path.Join(sourceCodeDir, "fastqc")
	err = os.MkdirAll(outputPath, os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}

	// Prepare arguments for the fastqc command
	args := append([]string{"-o", outputPath, "-t", "1"}, files...)

	// Run the fastqc command with the prepared arguments
	cmd := exec.Command("fastqc", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Create an error object if the fastqc command fails
		codeclarity_error := exceptionManager.Error{
			Private: exceptionManager.ErrorContent{
				Description: string(output),
				Type:        exceptionManager.GENERIC_ERROR,
			},
			Public: exceptionManager.ErrorContent{
				Description: "The script failed to execute",
				Type:        exceptionManager.GENERIC_ERROR,
			},
		}
		// Return an output indicating failure with the error object
		return generate_output(start, nil, codeclarity.FAILURE, []exceptionManager.Error{codeclarity_error})
	}

	// If the fastqc command succeeds, return an output indicating success
	return generate_output(start, "done", codeclarity.SUCCESS, []exceptionManager.Error{})
}

// generate_output creates a types.Output object based on the provided parameters.
// The function takes into account the start time of the analysis, any data to be included in the output, the status of the analysis, and any errors that occurred.
func generate_output(start time.Time, data any, status codeclarity.AnalysisStatus, errors []exceptionManager.Error) types.Output {
	// Calculate the timing information for the analysis
	formattedStart, formattedEnd, delta := output_generator.GetAnalysisTiming(start)

	// Create a new types.Output object with the calculated timing information and provided parameters
	output := types.Output{
		Result: types.Result{
			Data: data,
		},
		AnalysisInfo: types.AnalysisInfo{
			Errors: errors,
			Time: types.Time{
				AnalysisStartTime: formattedStart,
				AnalysisEndTime:   formattedEnd,
				AnalysisDeltaTime: delta,
			},
			Status: status,
		},
	}
	return output
}
