package api

import (
	"fmt"

	"github.com/aartur/docexp/internal/utils"
)

func OpenDocumentFile(sourceFilePath string) (*utils.File, error) {
	sourceFile := utils.NewFile(sourceFilePath)
	if sourceFile == nil {
		return nil, fmt.Errorf("cannot create input document object for %q", sourceFilePath)
	}
	err := sourceFile.EnsureFullContent()
	if err != nil {
		return nil, fmt.Errorf("cannot open input document %q: %v", sourceFilePath, err)
	}
	return sourceFile, nil
}
