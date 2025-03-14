package output_generator

import (
	"time"

	sbomTypes "github.com/CodeClarityCE/plugin-sbom-javascript/src/types/sbom/js"
	codeclarity "github.com/CodeClarityCE/utility-types/codeclarity_db"
	exceptionManager "github.com/CodeClarityCE/utility-types/exceptions"
)

// GetAnalysisTiming calculates and returns the start time, end time, and elapsed time of the analysis.
//
// Parameters:
//
//	start: The time when the analysis started.
//
// Returns:
//
//	A tuple containing the formatted start time, formatted end time, and elapsed time in seconds.
func GetAnalysisTiming(start time.Time) (string, string, float64) {
	// Record the current time to determine the analysis end time.
	end := time.Now()

	// Calculate the duration of the analysis.
	elapsed := time.Since(start)

	// Return the formatted start and end times, along with the elapsed time in seconds.
	return start.Local().String(), end.Local().String(), elapsed.Seconds()
}

// WriteFailureOutput constructs a failure output for the analysis.
//
// Parameters:
//
//	output: The initial output object to be updated.
//	start: The time when the analysis started.
//
// Returns:
//
//	The updated output object with failure status, timing information, and errors.
func WriteFailureOutput(output sbomTypes.Output, start time.Time) sbomTypes.Output {
	// Set the analysis status to failure.
	output.AnalysisInfo.Status = codeclarity.FAILURE

	// Calculate and update the analysis timing information.
	formattedStart, formattedEnd, delta := GetAnalysisTiming(start)
	output.AnalysisInfo.Time.AnalysisStartTime = formattedStart
	output.AnalysisInfo.Time.AnalysisEndTime = formattedEnd
	output.AnalysisInfo.Time.AnalysisDeltaTime = delta

	// Retrieve and set any errors that occurred during the analysis.
	output.AnalysisInfo.Errors = exceptionManager.GetErrors()

	// Return the updated output object.
	return output
}
