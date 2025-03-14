package output_generator

import (
	"time"

	sbomTypes "github.com/CodeClarityCE/plugin-sbom-javascript/src/types/sbom/js"
	codeclarity "github.com/CodeClarityCE/utility-types/codeclarity_db"
	exceptionManager "github.com/CodeClarityCE/utility-types/exceptions"
)

// GetAnalysisTiming calculates the start time, end time, and elapsed time of the analysis.
//
// It takes the start time as a parameter and returns the start time, end time, and elapsed time in seconds.
func GetAnalysisTiming(start time.Time) (string, string, float64) {
	// Record the current time to calculate the end time of the analysis
	end := time.Now()

	// Calculate the elapsed time since the start of the analysis
	elapsed := time.Since(start)

	// Return the formatted start and end times, as well as the elapsed time in seconds
	return start.Local().String(), end.Local().String(), elapsed.Seconds()
}

// WriteFailureOutput writes the failure output for the analysis.
//
// It sets the status of the output to codeclarity.FAILURE and updates the analysis timing information.
// It also retrieves and sets the private and public errors from the exception manager.
// The updated output is then returned.
func WriteFailureOutput(output sbomTypes.Output, start time.Time) sbomTypes.Output {
	// Set the status of the output to failure
	output.AnalysisInfo.Status = codeclarity.FAILURE

	// Calculate and update the analysis timing information
	formattedStart, formattedEnd, delta := GetAnalysisTiming(start)
	output.AnalysisInfo.Time.AnalysisStartTime = formattedStart
	output.AnalysisInfo.Time.AnalysisEndTime = formattedEnd
	output.AnalysisInfo.Time.AnalysisDeltaTime = delta

	// Retrieve and set the errors from the exception manager
	output.AnalysisInfo.Errors = exceptionManager.GetErrors()

	// Return the updated output
	return output
}
