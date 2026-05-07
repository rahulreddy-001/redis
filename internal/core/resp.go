package core

import (
	"errors"
	"fmt"
	"strings"
)

func DecoreArrayString(data []byte) ([]string, error) {
	value, err := Decode(data)
	if err != nil {
		return []string{}, err
	}
	values := value.([]any)
	tokens := make([]string, len(values))
	for i, value := range values {
		tokens[i] = strings.Trim(value.(string), " ")
	}
	return tokens, nil
}

func Decode(data []byte) (any, error) {
	if len(data) == 0 {
		return nil, errors.New("no data")
	}
	value, _, err := decodeOne(data)
	return value, err
}

func decodeOne(data []byte) (any, int, error) {
	switch data[0] {
	case '+':
		return decodeSimpleString(data)
	case '-':
		return decodeSimpleString(data)
	case ':':
		return decodeInt64(data)
	case '$':
		return decodeBulkString(data)
	case '*':
		return decodeArray(data)
	}
	return nil, 0, nil
}

func decodeArray(data []byte) ([]any, int, error) {
	value := []any{}
	count, pos, _ := decodeInt64(data)
	for i := 0; i < int(count); i++ {
		val, delta, err := decodeOne(data[pos:])
		if err != nil {
			return []any{}, 0, err
		}
		value = append(value, val)
		pos += delta
	}
	return value, pos, nil
}

func decodeBulkString(data []byte) (string, int, error) {
	strLen, posStart, err := decodeInt64(data)
	posEnd := posStart
	for strLen > 0 {
		posEnd++
		strLen--
	}
	return string(data[posStart:posEnd]), posEnd + 2, err
}

func decodeInt64(data []byte) (int64, int, error) {
	offset := 1
	var value int64
	for data[offset] != '\r' {
		value = value*10 + int64(data[offset]-'0')
		offset++
	}
	return value, offset + 2, nil
}

func decodeSimpleString(data []byte) (string, int, error) {
	fmt.Println(string(data))
	offset := 1
	for data[offset] != '\r' {
		offset++
	}
	return string(data[1:offset]), offset + 2, nil
}

func decodeError(data []byte) (string, int, error) {
	return decodeSimpleString(data)
}

func Encode(value any, isSimple bool) []byte {
	switch v := value.(type) {

	case string:
		if isSimple {
			return []byte(fmt.Sprintf("+%s\r\n", v))
		}
		return []byte(fmt.Sprintf("$%d\r\n%s\r\n", len(v), v))
	case int64, int, int32:
		return []byte(fmt.Sprintf(":%d\r\n", v))
	}
	return RESP_NIL
}
