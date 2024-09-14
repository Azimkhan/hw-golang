package main

import (
	"errors"
	"os"
	"testing"
)

func TestCopy(t *testing.T) {
	testdata := []struct {
		name               string
		inputFile          string
		expectedOutputFile string
		offset             int64
		limit              int64
		expectedError      error
	}{
		{
			name:               "offset 0 limit 0",
			inputFile:          "testdata/input.txt",
			expectedOutputFile: "testdata/out_offset0_limit0.txt",
			offset:             0,
			limit:              0,
			expectedError:      nil,
		},
		{
			name:               "offset 0 limit 10",
			inputFile:          "testdata/input.txt",
			expectedOutputFile: "testdata/out_offset0_limit10.txt",
			offset:             0,
			limit:              10,
			expectedError:      nil,
		},
		{
			name:               "offset 0 limit 1000",
			inputFile:          "testdata/input.txt",
			expectedOutputFile: "testdata/out_offset0_limit1000.txt",
			offset:             0,
			limit:              1000,
			expectedError:      nil,
		},
		{
			name:               "offset 0 limit 10000",
			inputFile:          "testdata/input.txt",
			expectedOutputFile: "testdata/out_offset0_limit10000.txt",
			offset:             0,
			limit:              10000,
			expectedError:      nil,
		},
		{
			name:               "offset 100 limit 1000",
			inputFile:          "testdata/input.txt",
			expectedOutputFile: "testdata/out_offset100_limit1000.txt",
			offset:             100,
			limit:              1000,
			expectedError:      nil,
		},
		{
			name:               "offset 6000 limit 1000",
			inputFile:          "testdata/input.txt",
			expectedOutputFile: "testdata/out_offset6000_limit1000.txt",
			offset:             6000,
			limit:              1000,
			expectedError:      nil,
		},
		{
			name:               "offset 10000 limit 1000",
			inputFile:          "testdata/input.txt",
			expectedOutputFile: "",
			offset:             10000,
			limit:              1000,
			expectedError:      ErrOffsetExceedsFileSize,
		},
		{
			name:               "unsupported file",
			inputFile:          "/dev/urandom",
			expectedOutputFile: "",
			offset:             0,
			limit:              0,
			expectedError:      ErrUnsupportedFile,
		},
	}
	compareFiles := func(file1, file2 string) bool {
		f1Content, err := os.ReadFile(file1)
		if err != nil {
			return false
		}
		f2Content, err := os.ReadFile(file2)
		if err != nil {
			return false
		}
		return string(f1Content) == string(f2Content)
	}

	for _, tc := range testdata {
		t.Run(tc.name, func(t *testing.T) {
			// temporary output file
			tmpFile, err := os.CreateTemp(os.TempDir(), "tmp_output_")
			if err != nil {
				t.Fatal(err)
			}
			defer func() {
				err := tmpFile.Close()
				if err != nil {
					t.Fatal("Unable to close file", err)
				}
				err = os.Remove(tmpFile.Name())
				if err != nil {
					t.Fatal("Unable to remove file", err)
				}
			}()

			err = Copy(tc.inputFile, tmpFile.Name(), tc.offset, tc.limit)

			// check if error is expected
			if tc.expectedError != nil {
				if !errors.Is(err, tc.expectedError) {
					t.Errorf("expected error %v, got %v", tc.expectedError, err)
				}
				return
			}

			// compare output file with expected output file
			if !compareFiles(tmpFile.Name(), tc.expectedOutputFile) {
				if err != nil {
					t.Fatal("Unable to seek file", err)
				}
				t.Errorf("files are not equal")
			}
		})
	}
}
