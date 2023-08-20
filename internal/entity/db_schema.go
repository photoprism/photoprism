package entity

import (
	"reflect"
	"strings"
	"time"
)

type DbFieldInfo struct {
	FieldName string
	FieldType string
}

// Returns a map["jsonFieldName"]"database_field_name" for all json-tagged fields in the given entity type
func GetDbFieldMap(entityType interface{}) map[string]DbFieldInfo {
	m := make(map[string]DbFieldInfo)
	for _, field := range reflect.VisibleFields(reflect.TypeOf(entityType)) {
		jsonFieldName := strings.Split(field.Tag.Get("json"), ",")[0]
		if len(jsonFieldName) > 0 {
			columnName := getGormColumnName(field)
			if len(columnName) == 0 {
				columnName = getDbFieldName(field.Name)
			}
			m[jsonFieldName] = DbFieldInfo{FieldName: columnName, FieldType: field.Type.String()}
		}
	}
	return m
}

func getGormColumnName(field reflect.StructField) string {
	for _, v := range strings.Split(field.Tag.Get("gorm"), ";") {
		if strings.HasPrefix(v, "column:") {
			return v[7:]
		}
	}
	return ""
}

// Return entity map with substituted field names and parsed values
func SubstDbFields(entity map[string]interface{}, substituteMap map[string]DbFieldInfo) map[string]interface{} {
	ret := make(map[string]interface{}, len(entity))
	for key, val := range entity {
		if dbFieldInfo, ok := substituteMap[key]; ok {
			ret[dbFieldInfo.FieldName] = parseDbFieldValue(val, dbFieldInfo.FieldType)
		}
	}
	return ret
}

// Parse field values that are not automatially unmarshalled from JSON to DB field compatible types
func parseDbFieldValue(value interface{}, dbFieldType string) interface{} {
	switch dbFieldType {
	case "time.Time":
		if v, ok := value.(string); ok {
			if val, err := time.Parse(time.RFC3339, v); err == nil {
				return val
			}
		}
	}
	return value
}

// Code below is borrowed from a newer GORM's schema package: https://github.com/go-gorm/gorm/blob/master/schema/naming.go (MIT License)

var (
	// https://github.com/golang/lint/blob/master/lint.go#L770
	commonInitialisms         = []string{"API", "ASCII", "CPU", "CSS", "DNS", "EOF", "GUID", "HTML", "HTTP", "HTTPS", "ID", "IP", "JSON", "LHS", "QPS", "RAM", "RHS", "RPC", "SLA", "SMTP", "SSH", "TLS", "TTL", "UID", "UI", "UUID", "URI", "URL", "UTF8", "VM", "XML", "XSRF", "XSS"}
	commonInitialismsReplacer *strings.Replacer
)

func init() {
	commonInitialismsForReplacer := make([]string, 0, len(commonInitialisms))
	for _, initialism := range commonInitialisms {
		commonInitialismsForReplacer = append(commonInitialismsForReplacer, initialism, strings.Title(strings.ToLower(initialism)))
	}
	commonInitialismsReplacer = strings.NewReplacer(commonInitialismsForReplacer...)
}

// Do GORM's snake_case mapping to get the database column name.
func getDbFieldName(name string) string {

	if name == "" {
		return ""
	}

	var (
		value                          = commonInitialismsReplacer.Replace(name)
		buf                            strings.Builder
		lastCase, nextCase, nextNumber bool // upper case == true
		curCase                        = value[0] <= 'Z' && value[0] >= 'A'
	)

	for i, v := range value[:len(value)-1] {
		nextCase = value[i+1] <= 'Z' && value[i+1] >= 'A'
		nextNumber = value[i+1] >= '0' && value[i+1] <= '9'

		if curCase {
			if lastCase && (nextCase || nextNumber) {
				buf.WriteRune(v + 32)
			} else {
				if i > 0 && value[i-1] != '_' && value[i+1] != '_' {
					buf.WriteByte('_')
				}
				buf.WriteRune(v + 32)
			}
		} else {
			buf.WriteRune(v)
		}

		lastCase = curCase
		curCase = nextCase
	}

	if curCase {
		if !lastCase && len(value) > 1 {
			buf.WriteByte('_')
		}
		buf.WriteByte(value[len(value)-1] + 32)
	} else {
		buf.WriteByte(value[len(value)-1])
	}
	ret := buf.String()
	return ret
}
