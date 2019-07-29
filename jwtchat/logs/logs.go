package logs

import (
	"database/sql/driver"
	"fmt"
	"math"
	"os"
	"reflect"
	"regexp"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

var sqlRegexp = regexp.MustCompile(`(\$\d+)|\?`)

type GormLogger struct {
	Name   string
	Logger *logrus.Logger
}

func (l *GormLogger) Print(values ...interface{}) {
	entry := l.Logger.WithField("name", l.Name)
	if len(values) > 1 {
		level := values[0]
		source := values[1]
		entry = l.Logger.WithField("source", source)
		if level == "sql" {
			duration := values[2]
			// sql
			var formattedValues []interface{}
			for _, value := range values[4].([]interface{}) {
				indirectValue := reflect.Indirect(reflect.ValueOf(value))
				if indirectValue.IsValid() {
					value = indirectValue.Interface()
					if t, ok := value.(time.Time); ok {
						formattedValues = append(formattedValues, fmt.Sprintf("'%v'", t.Format(time.RFC3339)))
					} else if b, ok := value.([]byte); ok {
						formattedValues = append(formattedValues, fmt.Sprintf("'%v'", string(b)))
					} else if r, ok := value.(driver.Valuer); ok {
						if value, err := r.Value(); err == nil && value != nil {
							formattedValues = append(formattedValues, fmt.Sprintf("'%v'", value))
						} else {
							formattedValues = append(formattedValues, "NULL")
						}
					} else {
						formattedValues = append(formattedValues, fmt.Sprintf("'%v'", value))
					}
				} else {
					formattedValues = append(formattedValues, fmt.Sprintf("'%v'", value))
				}
			}
			entry.WithField("took", duration).Debug(fmt.Sprintf(sqlRegexp.ReplaceAllString(values[3].(string), "%v"), formattedValues...))
		} else {
			entry.Error(values[2:]...)
		}
	} else {
		entry.Error(values...)
	}

}

var timeFormat = "02/02/2006 15:04:05"

// Logger middleware
func Logger(log *logrus.Logger) gin.HandlerFunc {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknow"
	}
	return func(c *gin.Context) {
		// other handler can change c.Path so:
		path := c.Request.URL.Path
		start := time.Now()
		c.Next()
		stop := time.Since(start)
		latency := int(math.Ceil(float64(stop.Nanoseconds()) / 1000000.0))
		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()
		clientUserAgent := c.Request.UserAgent()
		referer := c.Request.Referer()
		dataLength := c.Writer.Size()
		if dataLength < 0 {
			dataLength = 0
		}

		entry := logrus.NewEntry(log).WithFields(logrus.Fields{
			"hostname":   hostname,
			"statusCode": statusCode,
			"latency":    latency, // time to process
			"clientIP":   clientIP,
			"method":     c.Request.Method,
			"path":       path,
			"referer":    referer,
			"dataLength": dataLength,
			"userAgent":  clientUserAgent,
		})

		if len(c.Errors) > 0 {
			entry.Error(c.Errors.ByType(gin.ErrorTypePrivate).String())
		} else {
			msg := fmt.Sprintf("%s - %s [%s] \"%s %s\" %d %d \"%s\" \"%s\" (%dms)", clientIP, hostname, time.Now().Format(timeFormat), c.Request.Method, path, statusCode, dataLength, referer, clientUserAgent, latency)
			if statusCode > 499 {
				entry.Error(msg)
			} else if statusCode > 399 {
				entry.Warn(msg)
			} else {
				entry.Info(msg)
			}
		}
	}
}
