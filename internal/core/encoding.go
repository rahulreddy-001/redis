package core

import (
	"errors"
	"strconv"
	"time"
)

type Object struct {
	typeEncoding uint8
	value        any
	ttl          time.Time
}

const OBJ_TYPE_STRING uint8 = 1 << 4

const OBJ_ENC_RAW uint8 = 1
const OBJ_ENC_STR uint8 = 2
const OBJ_ENC_INT uint8 = 3

func getTypeEnc(value any) (uint8, uint8) {
	v := value.(string)
	oType := OBJ_TYPE_STRING

	if _, err := strconv.ParseInt(v, 10, 64); err == nil {
		return oType, OBJ_ENC_INT
	}

	if len(v) <= 44 {
		return oType, OBJ_ENC_STR
	}
	return oType, OBJ_ENC_RAW
}

func assertType(s, t uint8) error {
	if s != t {
		return errors.New("the operation is not permitted on this type")
	}
	return nil
}

func assertEnc(s, t uint8) error {
	if s != t {
		return errors.New("the operation is not permitted on this type")
	}
	return nil
}
