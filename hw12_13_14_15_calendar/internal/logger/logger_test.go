package logger

import (
	"os"
	"testing"
)

func TestLogger(t *testing.T) {
	testData := []struct {
		name           string
		level          LogLevel
		expectedOutput string
		callFunc       func(logger *Logger)
	}{
		{
			name:           "Log Info with level Error",
			level:          Error,
			expectedOutput: "",
			callFunc: func(logger *Logger) {
				logger.Info("Info message")
			},
		},
		{
			name:           "Log Info with level Info",
			level:          Info,
			expectedOutput: "Info message\n",
			callFunc: func(logger *Logger) {
				logger.Info("Info message")
			},
		},
	}

	for _, tt := range testData {
		t.Run(tt.name, func(t *testing.T) {
			r, w, _ := os.Pipe()
			stdout := os.Stdout
			os.Stdout = w
			defer func() {
				os.Stdout = stdout
			}()
			logger, _ := New(tt.level.String())
			tt.callFunc(logger)
			w.Close()
			output := make([]byte, 100)
			n, _ := r.Read(output)
			if string(output[:n]) != tt.expectedOutput {
				t.Fatalf("unexpected output: %s", output[:n])
			}
		})
	}
}
