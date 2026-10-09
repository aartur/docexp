package utils

import (
	"errors"
	"os"
	"slices"
)

type File struct {
	Filename    *string
	FullContent []byte

	newlineBytePositions     []int
	bytePositionToLineNumber map[int]int
}

func NewFile(filename string) *File {
	return &File{
		Filename: &filename,
	}
}

func (f *File) EnsureFullContent() error {
	if f.FullContent != nil {
		return nil
	}
	if f.Filename == nil {
		return errors.New("neither FullContent nor Filename provided")
	}
	byteContent, err := os.ReadFile(*f.Filename)
	if err != nil {
		return err
	}
	f.FullContent = byteContent
	return nil
}

type FilePosition struct {
	Line   int
	Column int
	Byte   int
}

type FileRange struct {
	FirstPosition FilePosition
	LastPosition  FilePosition
}

func (f *File) buildNewlineIndexes() error {
	if f.newlineBytePositions != nil {
		return nil
	}
	if err := f.EnsureFullContent(); err != nil {
		return err
	}
	f.newlineBytePositions = []int{}
	f.bytePositionToLineNumber = map[int]int{}

	for i, byteVal := range f.FullContent {
		if byteVal == '\n' {
			f.newlineBytePositions = append(f.newlineBytePositions, i)
			f.bytePositionToLineNumber[i] = len(f.newlineBytePositions)
		}
	}
	return nil
}

func (f *File) FilePositionFromByteIndex(byteIndex int) (FilePosition, error) {
	if err := f.buildNewlineIndexes(); err != nil {
		return FilePosition{}, err
	}
	if byteIndex < 0 || byteIndex >= len(f.FullContent) {
		return FilePosition{}, errors.New("invalid byte position")
	}
	pos, _ := slices.BinarySearch(f.newlineBytePositions, byteIndex)
	if pos == 0 {
		return FilePosition{Line: 0, Column: byteIndex, Byte: byteIndex}, nil
	}
	//if pos == len(f.newlineBytePositions) {
	newlineBytePos := f.newlineBytePositions[pos-1]
	line, ok := f.bytePositionToLineNumber[newlineBytePos]
	if !ok {
		return FilePosition{}, errors.New("invalid newline index")
	}
	column := byteIndex - newlineBytePos
	if column > 0 {
		column--
	}
	return FilePosition{
		Line:   line,
		Column: column,
		Byte:   byteIndex,
	}, nil
}
