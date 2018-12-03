package logstash

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/vend/logrus"
)

// Formatter generates json in logstash format.
// Logstash site: http://logstash.net/
type LogstashFormatter struct {
	Type string // if not empty use for logstash type field.

	// TimestampFormat sets the format used for timestamps.
	TimestampFormat string
}

func (f *LogstashFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	fields := make(logrus.Fields)
	for k, v := range entry.Data {
		switch v := v.(type) {
		case error:
			// Otherwise errors are ignored by `encoding/json`
			// https://github.com/Sirupsen/logrus/issues/377
			fields[k] = v.Error()
		default:
			fields[k] = v
		}
	}

	fields["@version"] = 1

	timeStampFormat := f.TimestampFormat

	if timeStampFormat == "" {
		timeStampFormat = time.RFC3339
	}

	fields["@timestamp"] = entry.Time.Format(timeStampFormat)

	// If a message was provided at time of logging, set that as the `message`
	// and move the message provided as a field to `field.message`
	if entry.Message != "" {
		v, ok := entry.Data["message"]
		if ok {
			fields["fields.message"] = v
		}

		// A hack.
		// If the string is encapsulated in brackets, trim the brackets
		if string(entry.Message[0]) == "[" {
			entry.Message = strings.Trim(entry.Message, "[]")
		}

		fields["message"] = entry.Message
	}

	// set level field
	v, ok := entry.Data["level"]
	if ok {
		fields["fields.level"] = v
	}
	fields["level"] = entry.Level.String()

	// set type field
	if f.Type != "" {
		v, ok = entry.Data["type"]
		if ok {
			fields["fields.type"] = v
		}
		fields["type"] = f.Type
	}

	serialized, err := json.Marshal(fields)
	if err != nil {
		return nil, fmt.Errorf("Failed to marshal fields to JSON, %v", err)
	}
	return append(serialized, '\n'), nil
}
