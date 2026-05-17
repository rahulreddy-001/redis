package core

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
)

func dumpKey(fp *os.File, k string, v Object) {
	cmd := fmt.Sprintf("SET %s %s", k, v.value)
	tokens := strings.Split(cmd, " ")
	fp.Write(Encode(tokens, false))
}

func DumpAllAOF() error {
	fp, err := os.OpenFile("./redis.aof", os.O_CREATE|os.O_WRONLY, os.ModeAppend)

	if err != nil {
		return errors.New(fmt.Sprint("failed opening AOF file", err))
	}

	log.Println("rewriting AOF file at ./redis.aof")
	for k, v := range memory {
		dumpKey(fp, k, v)
	}
	log.Println("AOF file updated")
	return nil
}
