package logger

import (
    "github.com/sirupsen/logrus"
    "os"
)

// Logger is a custom logger that wraps the logrus.Logger.
type Logger struct {
    *logrus.Logger
}

// NewLogger initializes a new Logger instance with settings based on config.
func NewLogger() *Logger {
    logger := logrus.New()

    // Set log level based on configuration
    level, err := logrus.ParseLevel("debug")
    if err != nil {
        logger.Warnf("Invalid log level: %s, defaulting to InfoLevel", "debug")
        level = logrus.InfoLevel
    }
    logger.SetLevel(level)

    // Set the log output (default is os.Stdout)
    logger.SetOutput(os.Stdout)

    // Optionally, set a log formatter (e.g., JSONFormatter, TextFormatter)
    logger.SetFormatter(&logrus.TextFormatter{
        FullTimestamp:   true,
        TimestampFormat: "2006-01-02 15:04:05",
    })

    return &Logger{logger}
}
