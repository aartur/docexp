package utils

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func createTempFile(t *testing.T, content []byte) string {
	dir := t.TempDir()
	filename := filepath.Join(dir, "file.txt")
	if err := os.WriteFile(filename, content, 0600); err != nil {
		t.Fatal(err)
	}
	return filename
}

func TestReadingFile_SomeContent(t *testing.T) {
	content := []byte(string("\nthis is\nsome file\n"))
	filename := createTempFile(t, content)
	file := NewFile(filename)
	if err := file.EnsureFullContent(); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(file.FullContent, content) {
		t.Fatal("Read content doesn't not match written content")
	}
}

func TestReadingFile_EmptyContent(t *testing.T) {
	filename := createTempFile(t, []byte{})
	file := NewFile(filename)
	if err := file.EnsureFullContent(); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(file.FullContent, []byte{}) {
		t.Fatal("Read content doesn't not match written content")
	}
}

func TestReadingInvalidFile(t *testing.T) {
	file := NewFile("invalid_file")
	err := file.EnsureFullContent()
	t.Logf("Got error from reading an invalid file: %v", err)
	if err == nil {
		t.Fatal("No error from reading an invalid file")
	}
}

func TestCalculatingFilePosition(t *testing.T) {
	content := []byte(string("a\nbc\n\nde"))

	//for i, char := range content {
	//	t.Logf("At %d: %v %s", i, char, string(char))
	//}

	filename := createTempFile(t, content)
	file := NewFile(filename)
	if err := file.EnsureFullContent(); err != nil {
		t.Fatal(err)
	}

	checkPos := func(byteIndex int, expectedPos FilePosition) {
		pos, err := file.FilePositionFromByteIndex(byteIndex)
		if err != nil {
			t.Fatal(err)
		}
		//t.Logf("Got position %+v for index %d", pos, byteIndex)
		if pos != expectedPos {
			t.Errorf("Invalid position for index %d: %+v, expected: %+v", byteIndex, pos, expectedPos)
		}
	}

	// valid positions
	checkPos(0, FilePosition{Line: 0, Column: 0, Byte: 0})
	checkPos(1, FilePosition{Line: 0, Column: 1, Byte: 1})
	checkPos(2, FilePosition{Line: 1, Column: 0, Byte: 2})
	checkPos(3, FilePosition{Line: 1, Column: 1, Byte: 3})
	checkPos(4, FilePosition{Line: 1, Column: 2, Byte: 4})
	checkPos(5, FilePosition{Line: 2, Column: 0, Byte: 5})
	checkPos(6, FilePosition{Line: 3, Column: 0, Byte: 6})
	checkPos(7, FilePosition{Line: 3, Column: 1, Byte: 7})

	checkPos(0, FilePosition{Line: 0, Column: 0, Byte: 0})

	// invalid positions
	for _, invalidBytePos := range []int{8, -1, 9, -1, 100} {
		_, err := file.FilePositionFromByteIndex(invalidBytePos)
		if err == nil {
			t.Fatalf("no error for invalid position %d", invalidBytePos)
		}
	}
}
