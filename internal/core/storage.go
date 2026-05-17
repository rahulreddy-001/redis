package core

import (
	"fmt"
	"time"
)

var memory map[string]Object

func init() {
	memory = map[string]Object{}
}

func setObj(key string, o Object) {
	memory[key] = o
}

func set(key string, value any, ttl time.Duration, override bool) []byte {
	ttlTime := time.Time{}
	if ttl != 0 {
		ttlTime = time.Now().Add(ttl)
	}
	if val, ok := memory[key]; ok && !override {
		val.value = val
	} else {
		oType, oEnc := getTypeEnc(value)
		memory[key] = Object{
			value:        value,
			ttl:          ttlTime,
			typeEncoding: oType | oEnc,
		}
	}
	return Encode("OK", true)
}

func get(key string) []byte {
	if val, ok := memory[key]; ok {
		if val.ttl.IsZero() || time.Now().Before(val.ttl) {
			return Encode(val.value, false)
		}
		delete(memory, key)
	}
	return RESP_NIL
}

func getObj(key string) (Object, bool) {
	if val, ok := memory[key]; ok {
		if val.ttl.IsZero() || time.Now().Before(val.ttl) {
			return val, true
		}
		delete(memory, key)
	}
	return Object{}, false
}

func ttl(key string) []byte {
	if val, ok := memory[key]; ok {
		if !val.ttl.IsZero() && time.Now().Before(val.ttl) {
			return Encode(int64(-time.Since(val.ttl).Seconds()), true)
		}
		if !val.ttl.IsZero() {
			delete(memory, key)
			return Encode(-2, false)
		}
		return Encode(-1, false)
	}
	return Encode(-2, false)
}

func del(keys []string) []byte {
	count := 0
	for _, key := range keys {
		if val, ok := memory[key]; ok {
			if val.ttl.IsZero() || time.Now().Before(val.ttl) {
				count++
			}
			delete(memory, key)
		}
	}
	return Encode(count, false)
}

// NX: Set expiry only if the key has no existing expiry
// XX: Set expiry only if the key already has an expiry
// GT: Set expiry only when the new expiry is greater than the current one
// LT: Set expiry only when the new expiry is less than the current one
func expire(key string, ttl time.Duration, nx, xx, gl, lt bool) []byte {
	val, ok := memory[key]
	if !ok {
		return Encode(0, false)
	}
	if !val.ttl.IsZero() && time.Now().After(val.ttl) {
		delete(memory, key)
		return Encode(0, false)
	}

	currentTTL := val.ttl
	newTTL := time.Now().Add(ttl)

	setExpire := true
	if nx && !currentTTL.IsZero() {
		setExpire = false
	}
	if xx && currentTTL.IsZero() {
		setExpire = false
	}
	if gl {
		if currentTTL.IsZero() || !newTTL.After(currentTTL) {
			setExpire = false
		}
	}
	if lt {
		if currentTTL.IsZero() || !newTTL.Before(currentTTL) {
			setExpire = false
		}
	}

	if setExpire {
		val.ttl = newTTL
		memory[key] = val
		return Encode(1, false)
	}

	return Encode(0, false)
}

func Clean() bool {
	fmt.Println("cleanup initiated")

	numValues := 20
	numValuesRef := numValues
	delValues := 0
	for key, val := range memory {
		numValues--
		if !val.ttl.IsZero() && time.Now().After(val.ttl) {
			delete(memory, key)
			delValues++
		}
		if numValues == 0 {
			break
		}
	}
	if int(100*(delValues/numValuesRef)) > 25 {
		fmt.Println("cleanup done, more than 25% have been cleaned")
		return true
	}
	fmt.Println("cleanup done, less than 25% have been cleaned")
	return false
}
