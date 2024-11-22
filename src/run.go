package js

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"time"

	codeclarity "github.com/CodeClarityCE/utility-types/codeclarity_db"
	exceptionManager "github.com/CodeClarityCE/utility-types/exceptions"
	"github.com/google/uuid"
	"github.com/uptrace/bun"

	"github.com/parithera/plugin-fastqc/src/types"
	"github.com/parithera/plugin-fastqc/src/utils/output_generator"
)

// Start is a function that analyzes the source code directory and generates a software bill of materials (SBOM) output.
// It returns an sbomTypes.Output struct containing the analysis results.
func Start(sourceCodeDir string, analysisId uuid.UUID, codeclarityDB *bun.DB) types.Output {
	scriptPath := path.Join(sourceCodeDir, "script.R")
	analysis := codeclarity.Analysis{
		Id: analysisId,
	}
	err := codeclarityDB.NewSelect().Model(&analysis).WherePK().Scan(context.Background())
	if err != nil {
		panic(fmt.Sprintf("Failed to fetch analysis by id: %s", err.Error()))
	}

	r_config, ok := analysis.Config["r"].(map[string]interface{})
	if !ok {
		panic("Failed to fetch analysis config")
	}

	projectId := r_config["project"].(string)

	var chat types.Chat
	err = codeclarityDB.NewSelect().Model(&chat).Where("? = ?", bun.Ident("projectId"), projectId).Scan(context.Background())
	if err == nil {
		chat.Messages[0].Result = analysisId.String()
		_, err = codeclarityDB.NewUpdate().Model(&chat).WherePK().Exec(context.Background())
		if err != nil {
			panic(fmt.Sprintf("Failed to add image to chat history: %s", err.Error()))
		}
	}

	requestType := r_config["type"].(string)

	if requestType == "chat" {
		out := ExecuteScript(sourceCodeDir, scriptPath, analysisId, chat)
		chat.Messages[0].Image = out.Result.Image
		chat.Messages[0].Text = out.Result.Text
		chat.Messages[0].Data = out.Result.Data
		_, err = codeclarityDB.NewUpdate().Model(&chat).WherePK().Exec(context.Background())
		if err != nil {
			panic(fmt.Sprintf("Failed to add result content to chat history: %s", err.Error()))
		}
		return out
	}

	scriptPath = "r_scripts/parithera.R" // Replace with the path to your R script
	return ExecuteScript(sourceCodeDir, scriptPath, analysisId, chat)

}

func ExecuteScript(sourceCodeDir string, scriptPath string, analysisId uuid.UUID, chat types.Chat) types.Output {
	start := time.Now()
	// Run Rscript in sourceCodeDir
	cmd := exec.Command("Rscript", scriptPath, sourceCodeDir, "data.h5")
	_, err := cmd.CombinedOutput()
	if err != nil {
		// panic(fmt.Sprintf("Failed to run Rscript: %s", err.Error()))
		codeclarity_error := exceptionManager.Error{
			Private: exceptionManager.ErrorContent{
				Description: err.Error(),
				Type:        exceptionManager.GENERIC_ERROR,
			},
			Public: exceptionManager.ErrorContent{
				Description: "The script failed to execute",
				Type:        exceptionManager.GENERIC_ERROR,
			},
		}
		return generate_output(start, "", nil, "", codeclarity.FAILURE, []exceptionManager.Error{codeclarity_error})
	}

	// We check if a graph was generated and rename it
	oldName := filepath.Join(sourceCodeDir, "graph.png")
	_, statErr := os.Stat(oldName)
	image := ""
	if statErr == nil {
		newName := filepath.Join(sourceCodeDir, analysisId.String()+".png")
		os.Rename(oldName, newName)
		image = analysisId.String()
	}

	// We check if a text was generated and rename it
	oldName = filepath.Join(sourceCodeDir, "result.txt")
	_, statErr = os.Stat(oldName)
	text := ""
	if statErr == nil {
		newName := filepath.Join(sourceCodeDir, analysisId.String()+".txt")
		os.Rename(oldName, newName)
		// Open the renamed text file and put its content in the 'text' variable
		txtFile, err := os.Open(newName)
		if err != nil {
			panic(fmt.Sprintf("Failed to open text file: %s", err.Error()))
		}
		defer txtFile.Close()

		var buffer bytes.Buffer
		scanner := bufio.NewScanner(txtFile)
		for scanner.Scan() {
			buffer.WriteString(scanner.Text() + "\n")
		}
		text = buffer.String()
	}

	// We check if data were generated and rename them
	oldName = filepath.Join(sourceCodeDir, "data.json")
	var data map[string]interface{}
	_, statErr = os.Stat(oldName)
	if statErr == nil {
		newName := filepath.Join(sourceCodeDir, analysisId.String()+".json")
		os.Rename(oldName, newName)
		jsonFile, err := os.Open(newName)
		if err != nil {
			panic(fmt.Sprintf("Failed to open JSON file: %s", err.Error()))
		}
		defer jsonFile.Close()

		decoder := json.NewDecoder(jsonFile)
		err = decoder.Decode(&data)
		if err != nil {
			panic(fmt.Sprintf("Failed to decode JSON data: %s", err.Error()))
		}

	}

	return generate_output(start, image, data, text, codeclarity.SUCCESS, []exceptionManager.Error{})
}

func generate_output(start time.Time, imageName string, data any, text string, status codeclarity.AnalysisStatus, errors []exceptionManager.Error) types.Output {
	formattedStart, formattedEnd, delta := output_generator.GetAnalysisTiming(start)

	output := types.Output{
		Result: types.Result{
			Image: imageName,
			Data:  data,
			Text:  text,
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
