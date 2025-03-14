package fastqc

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	codeclarity "github.com/CodeClarityCE/utility-types/codeclarity_db"
	exceptionManager "github.com/CodeClarityCE/utility-types/exceptions"
	"github.com/uptrace/bun"

	"github.com/parithera/plugin-fastqc/src/types"
	"github.com/parithera/plugin-fastqc/src/utils/output_generator"
)

// Start analyzes the source code directory and generates a FastQC report.
// It returns a types.Output struct containing the analysis results.
func Start(sourceCodeDir string, codeclarityDB *bun.DB) types.Output {
	return ExecuteScript(sourceCodeDir)
}

// ExecuteScript runs FastQC on the provided source code directory and returns the output.
// It searches for .fastq.gz files, executes FastQC, and generates an output based on the results.
func ExecuteScript(sourceCodeDir string) types.Output {
	// Record the start time of the analysis.
	startTime := time.Now()

	// Search for .fastq.gz files in the source code directory.
	fastqFiles, err := filepath.Glob(filepath.Join(sourceCodeDir, "*.fastq.gz"))
	if err != nil {
		// Log the error and return a failure output if file searching fails.
		log.Fatal(err)
		return generate_output(startTime, nil, codeclarity.FAILURE, []exceptionManager.Error{{
			Private: exceptionManager.ErrorContent{
				Description: "Error while searching for fastq files",
				Type:        exceptionManager.GENERIC_ERROR,
			},
			Public: exceptionManager.ErrorContent{
				Description: "Error while searching for fastq files",
				Type:        exceptionManager.GENERIC_ERROR,
			},
		}})
	}

	// Check if any .fastq.gz files were found.
	if len(fastqFiles) == 0 {
		// If no files are found, return a success output with a message.
		return generate_output(startTime, "no fastq file found", codeclarity.SUCCESS, []exceptionManager.Error{})
	}

	// Create the output directory for FastQC results.
	outputPath := filepath.Join(sourceCodeDir, "fastqc")
	err = os.MkdirAll(outputPath, os.ModePerm)
	if err != nil {
		log.Fatal(err)
		return generate_output(startTime, nil, codeclarity.FAILURE, []exceptionManager.Error{{
			Private: exceptionManager.ErrorContent{
				Description: "Error creating output directory",
				Type:        exceptionManager.GENERIC_ERROR,
			},
			Public: exceptionManager.ErrorContent{
				Description: "Error creating output directory",
				Type:        exceptionManager.GENERIC_ERROR,
			},
		}})
	}

	// Prepare arguments for the FastQC command.
	args := []string{"-o", outputPath, "-t", "1"}
	args = append(args, fastqFiles...)

	// Run the FastQC command with the prepared arguments.
	cmd := exec.Command("fastqc", args...)
	outputBytes, err := cmd.CombinedOutput()
	if err != nil {
		// Create an error object if the FastQC command fails.
		codeclarityError := exceptionManager.Error{
			Private: exceptionManager.ErrorContent{
				Description: string(outputBytes),
				Type:        exceptionManager.GENERIC_ERROR,
			},
			Public: exceptionManager.ErrorContent{
				Description: "The FastQC script failed to execute",
				Type:        exceptionManager.GENERIC_ERROR,
			},
		}
		// Return an output indicating failure with the error object.
		return generate_output(startTime, nil, codeclarity.FAILURE, []exceptionManager.Error{codeclarityError})
	}

	// If the FastQC command succeeds, return an output indicating success.
	return generate_output(startTime, "done", codeclarity.SUCCESS, []exceptionManager.Error{})
}

// generate_output creates a types.Output object based on the provided parameters.
// It takes into account the start time of the analysis, any data to be included in the output, the status of the analysis, and any errors that occurred.
func generate_output(startTime time.Time, data any, status codeclarity.AnalysisStatus, errors []exceptionManager.Error) types.Output {
	// Calculate the timing information for the analysis.
	formattedStart, formattedEnd, delta := output_generator.GetAnalysisTiming(startTime)

	// Create a new types.Output object with the calculated timing information and provided parameters.
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
