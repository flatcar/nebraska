package api

import (
	"database/sql/driver"
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"time"
	"unicode"

	log "github.com/mgutz/logxi/v1"
)

var (
	numericPlaceHolderRegexp = regexp.MustCompile(`\$\d+`)
	sqlRegexp                = regexp.MustCompile(`\?`)
)

// This type satisfies gorm's internal logger interface
type gormLogger struct {
	logger log.Logger
}

func newGORMLogger() gormLogger {
	return gormLogger{
		logger: log.New("gorm"),
	}
}

func (l gormLogger) Print(values ...interface{}) {
	if len(values) == 0 {
		return
	}
	message := "log from GORM"
	if len(values) == 1 {
		pairs := []interface{}{"message", values[0]}
		level := logLevelFromGORMLevel(values[0])
		l.logger.Log(level, message, pairs)
		return
	}

	var (
		sql             string
		formattedValues []string
		level           = values[0]
	)

	pairs := []interface{}{"source", values[1]}

	if level == "sql" {
		message = "SQL log from GORM"
		pairs = append(pairs, "duration", values[2])
		// sql

		for _, value := range values[4].([]interface{}) {
			indirectValue := reflect.Indirect(reflect.ValueOf(value))
			if indirectValue.IsValid() {
				value = indirectValue.Interface()
				if t, ok := value.(time.Time); ok {
					if t.IsZero() {
						formattedValues = append(formattedValues, fmt.Sprintf("'%v'", "0000-00-00 00:00:00"))
					} else {
						formattedValues = append(formattedValues, fmt.Sprintf("'%v'", t.Format("2006-01-02 15:04:05")))
					}
				} else if b, ok := value.([]byte); ok {
					if str := string(b); isPrintable(str) {
						formattedValues = append(formattedValues, fmt.Sprintf("'%v'", str))
					} else {
						formattedValues = append(formattedValues, "'<binary>'")
					}
				} else if r, ok := value.(driver.Valuer); ok {
					if value, err := r.Value(); err == nil && value != nil {
						formattedValues = append(formattedValues, fmt.Sprintf("'%v'", value))
					} else {
						formattedValues = append(formattedValues, "NULL")
					}
				} else {
					switch value.(type) {
					case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, bool:
						formattedValues = append(formattedValues, fmt.Sprintf("%v", value))
					default:
						formattedValues = append(formattedValues, fmt.Sprintf("'%v'", value))
					}
				}
			} else {
				formattedValues = append(formattedValues, "NULL")
			}
		}

		// differentiate between $n placeholders or else treat like ?
		if numericPlaceHolderRegexp.MatchString(values[3].(string)) {
			sql = values[3].(string)
			for index, value := range formattedValues {
				placeholder := fmt.Sprintf(`\$%d([^\d]|$)`, index+1)
				sql = regexp.MustCompile(placeholder).ReplaceAllString(sql, value+"$1")
			}
		} else {
			formattedValuesLength := len(formattedValues)
			for index, value := range sqlRegexp.Split(values[3].(string), -1) {
				sql += value
				if index < formattedValuesLength {
					sql += formattedValues[index]
				}
			}
		}

		pairs = append(pairs, "sql", sql)
		pairs = append(pairs, "rows affected/returned", values[5])
	} else {
		for idx, value := range values[2:] {
			pairs = append(pairs, fmt.Sprintf("value%d", idx), value)
		}
	}
	l.logger.Log(logLevelFromGORMLevel(values[0]), message, pairs)
}

func isPrintable(s string) bool {
	for _, r := range s {
		if !unicode.IsPrint(r) {
			return false
		}
	}
	return true
}

func logLevelFromGORMLevel(level interface{}) int {
	switch level {
	case "sql":
		return log.LevelDebug
	case "log":
		return log.LevelInfo
	case "error":
		return log.LevelError
	}
	if str, ok := level.(string); ok {
		if strings.HasPrefix(str, "[info]") {
			return log.LevelInfo
		}
		if strings.HasPrefix(str, "[warning]") {
			return log.LevelWarn
		}
	}
	return log.LevelDebug
}
