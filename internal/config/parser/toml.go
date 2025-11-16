package parser

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// TOMLValue represents a value in TOML
type TOMLValue interface{}

// TOMLTable represents a TOML table (map)
type TOMLTable map[string]TOMLValue

// TOMLParser parses TOML files
type TOMLParser struct {
	data         TOMLTable
	currentTable TOMLTable
	currentPath  []string
}

// New creates a new TOML parser
func New() *TOMLParser {
	return &TOMLParser{
		data:         make(TOMLTable),
		currentTable: nil,
		currentPath:  []string{},
	}
}

// ParseFile parses a TOML file
func (p *TOMLParser) ParseFile(path string) (TOMLTable, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	p.currentTable = p.data
	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Handle table headers
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			tableName := strings.TrimSpace(line[1 : len(line)-1])
			if err := p.setCurrentTable(tableName); err != nil {
				return nil, fmt.Errorf("line %d: %v", lineNum, err)
			}
			continue
		}

		// Handle key-value pairs
		if err := p.parseKeyValue(line); err != nil {
			return nil, fmt.Errorf("line %d: %v", lineNum, err)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return p.data, nil
}

// setCurrentTable sets the current table context
func (p *TOMLParser) setCurrentTable(tableName string) error {
	parts := strings.Split(tableName, ".")
	p.currentPath = parts
	p.currentTable = p.data

	for i, part := range parts {
		if i == len(parts)-1 {
			// Last part - create or use existing table
			if _, exists := p.currentTable[part]; !exists {
				p.currentTable[part] = make(TOMLTable)
			}
			table, ok := p.currentTable[part].(TOMLTable)
			if !ok {
				return fmt.Errorf("key '%s' is not a table", part)
			}
			p.currentTable = table
		} else {
			// Intermediate parts - ensure they exist
			if _, exists := p.currentTable[part]; !exists {
				p.currentTable[part] = make(TOMLTable)
			}
			table, ok := p.currentTable[part].(TOMLTable)
			if !ok {
				return fmt.Errorf("key '%s' is not a table", part)
			}
			p.currentTable = table
		}
	}

	return nil
}

// parseKeyValue parses a key = value line
func (p *TOMLParser) parseKeyValue(line string) error {
	parts := strings.SplitN(line, "=", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid key-value pair: %s", line)
	}

	key := strings.TrimSpace(parts[0])
	valueStr := strings.TrimSpace(parts[1])

	value, err := p.parseValue(valueStr)
	if err != nil {
		return err
	}

	p.currentTable[key] = value
	return nil
}

// parseValue parses a TOML value
func (p *TOMLParser) parseValue(valueStr string) (TOMLValue, error) {
	valueStr = strings.TrimSpace(valueStr)

	// Boolean
	if valueStr == "true" {
		return true, nil
	}
	if valueStr == "false" {
		return false, nil
	}

	// String (quoted)
	if (strings.HasPrefix(valueStr, `"`) && strings.HasSuffix(valueStr, `"`)) ||
		(strings.HasPrefix(valueStr, `'`) && strings.HasSuffix(valueStr, `'`)) {
		return strings.Trim(valueStr, `"'`), nil
	}

	// Array
	if strings.HasPrefix(valueStr, "[") && strings.HasSuffix(valueStr, "]") {
		return p.parseArray(valueStr)
	}

	// Integer
	if intVal, err := strconv.ParseInt(valueStr, 10, 64); err == nil {
		return int(intVal), nil
	}

	// Float
	if floatVal, err := strconv.ParseFloat(valueStr, 64); err == nil {
		return floatVal, nil
	}

	// Unquoted string (bare key as value - not standard but lenient)
	return valueStr, nil
}

// parseArray parses a TOML array
func (p *TOMLParser) parseArray(arrayStr string) ([]TOMLValue, error) {
	arrayStr = strings.TrimSpace(arrayStr)
	arrayStr = arrayStr[1 : len(arrayStr)-1] // Remove [ ]

	var result []TOMLValue
	var current strings.Builder
	inQuote := false
	quoteChar := rune(0)
	depth := 0

	for _, char := range arrayStr {
		switch {
		case (char == '"' || char == '\'') && !inQuote:
			inQuote = true
			quoteChar = char
			current.WriteRune(char)

		case char == quoteChar && inQuote:
			inQuote = false
			quoteChar = 0
			current.WriteRune(char)

		case char == '[' && !inQuote:
			depth++
			current.WriteRune(char)

		case char == ']' && !inQuote:
			depth--
			current.WriteRune(char)

		case char == ',' && !inQuote && depth == 0:
			// End of element
			value, err := p.parseValue(strings.TrimSpace(current.String()))
			if err != nil {
				return nil, err
			}
			result = append(result, value)
			current.Reset()

		default:
			current.WriteRune(char)
		}
	}

	// Add last element
	if current.Len() > 0 {
		value, err := p.parseValue(strings.TrimSpace(current.String()))
		if err != nil {
			return nil, err
		}
		result = append(result, value)
	}

	return result, nil
}

// Helper methods for TOMLTable

// GetString gets a string value from the table
func (t TOMLTable) GetString(key string) (string, bool) {
	val, ok := t[key]
	if !ok {
		return "", false
	}
	str, ok := val.(string)
	return str, ok
}

// GetInt gets an integer value from the table
func (t TOMLTable) GetInt(key string) (int, bool) {
	val, ok := t[key]
	if !ok {
		return 0, false
	}
	intVal, ok := val.(int)
	return intVal, ok
}

// GetBool gets a boolean value from the table
func (t TOMLTable) GetBool(key string) (bool, bool) {
	val, ok := t[key]
	if !ok {
		return false, false
	}
	boolVal, ok := val.(bool)
	return boolVal, ok
}

// GetTable gets a sub-table
func (t TOMLTable) GetTable(key string) (TOMLTable, bool) {
	val, ok := t[key]
	if !ok {
		return nil, false
	}
	table, ok := val.(TOMLTable)
	return table, ok
}

// GetStringSlice gets a string array value
func (t TOMLTable) GetStringSlice(key string) ([]string, bool) {
	val, ok := t[key]
	if !ok {
		return nil, false
	}
	arr, ok := val.([]TOMLValue)
	if !ok {
		return nil, false
	}
	result := make([]string, 0, len(arr))
	for _, v := range arr {
		if str, ok := v.(string); ok {
			result = append(result, str)
		}
	}
	return result, true
}
