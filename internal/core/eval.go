package core

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"slices"
	"strconv"
	"strings"
	"time"
)

var RESP_NIL = []byte("$-1\r\n")

func evalCommand(_ []string, c io.Writer) error {
	_, err := c.Write(RESP_NIL)
	return err
}

func evalPing(args []string, c io.Writer) error {
	if len(args) > 1 {
		return errors.New("ERR wrong number of arguments for 'ping' command")
	}
	var err error
	if len(args) == 0 {
		_, err = c.Write(Encode("PONG", true))
	} else {
		_, err = c.Write(Encode(args[0], false))
	}
	return err
}

func evalSET(args []string, c io.Writer) error {
	if len(args) < 2 {
		return errors.New("ERR wrong number of arguments for 'set' command")
	}
	var key string = args[0]
	var value any = args[1]
	var ttl time.Duration = 0
	var override bool = true

	idx := 2
	for idx < len(args) {
		validDoubleCommands := []string{"EX", "PX"}
		validSingleCommands := []string{"KEEP"}

		currentCommand := args[idx]
		if slices.Contains(validDoubleCommands, currentCommand) && idx+1 < len(args) {
			ttlRaw, err := strconv.Atoi(args[idx+1])
			if err != nil {
				return errors.New("ERR syntax error")
			}
			if currentCommand == "EX" {
				ttl = time.Duration(ttlRaw) * time.Second

			}
			if currentCommand == "PX" {
				ttl = time.Duration(ttlRaw) * time.Millisecond
			}
			idx += 2
		} else if slices.Contains(validSingleCommands, currentCommand) {
			override = false
			idx += 1
		} else {
			return errors.New("ERR syntax error")
		}
	}
	c.Write(set(key, value, ttl, override))
	return nil
}

func evalGET(args []string, c io.Writer) error {
	if len(args) != 1 {
		return errors.New("ERR wrong number of arguments for 'get' command")
	}
	c.Write(get(args[0]))
	return nil
}

func evalTTL(args []string, c io.Writer) error {
	if len(args) != 1 {
		return errors.New("ERR wrong number of arguments for 'ttl' command")
	}
	c.Write(ttl(args[0]))
	return nil
}

func evalDEL(args []string, c io.Writer) error {
	if len(args) == 0 {
		return errors.New("ERR wrong number of arguments for 'del' command")
	}
	c.Write(del(args))
	return nil
}

func evalExpire(args []string, c io.Writer) error {
	if len(args) < 2 {
		return errors.New("ERR wrong number of arguments for 'expire' command")
	}
	// NX: Set expiry only if the key has no existing expiry
	// XX: Set expiry only if the key already has an expiry
	// GT: Set expiry only when the new expiry is greater than the current one
	// LT: Set expiry only when the new expiry is less than the current one
	nx, xx, gl, lt := false, false, false, false
	if len(args) >= 2 {
		for _, val := range args[2:] {
			switch val := strings.ToLower(val); {
			case val == "nx":
				nx = true
			case val == "xx":
				xx = true
			case val == "gl":
				gl = true
			case val == "lt":
				lt = true
			default:
				return errors.New(fmt.Sprint("ERR Unsupported option ", val))
			}
		}
	}

	if nx == true && xx == true || gl == true && lt == true {
		return errors.New("ERR NX and XX, GT or LT options at the same time are not compatible")
	}

	ttl, err := strconv.Atoi(args[1])
	if err != nil {
		return errors.New("ERR syntax error")
	}
	c.Write(expire(args[0], time.Duration(ttl)*time.Second, nx, xx, gl, lt))
	return nil
}

func evalINCR(args []string, c io.Writer) error {
	if len(args) != 1 {
		return errors.New("ERR wrong number of arguments for 'incr' command")
	}

	key := args[0]
	obj, exists := getObj(key)
	if !exists {
		val := "0"
		oType, oEnc := getTypeEnc(val)
		obj = Object{
			value:        val,
			typeEncoding: oType | oEnc,
		}
	}
	oType, oEnc := getTypeEnc(obj.value)
	if err := assertType(oType, OBJ_TYPE_STRING); err != nil {
		return err
	}
	if err := assertEnc(oEnc, OBJ_ENC_INT); err != nil {
		return err
	}

	val, _ := strconv.Atoi(obj.value.(string))
	obj.value = strconv.Itoa(val + 1)
	setObj(key, obj)

	c.Write(Encode(val+1, false))
	return nil
}

func EvalAndRespond(cmd Cmd, c io.Writer) error {
	switch cmd.Cmd {
	case "COMMAND":
		return evalCommand(cmd.Args, c)
	case "PING":
		return evalPing(cmd.Args, c)
	case "SET":
		return evalSET(cmd.Args, c)
	case "GET":
		return evalGET(cmd.Args, c)
	case "TTL":
		return evalTTL(cmd.Args, c)
	case "DEL":
		return evalDEL(cmd.Args, c)
	case "EXPIRE":
		return evalExpire(cmd.Args, c)
	case "BGWRITEAOF":
		return DumpAllAOF()
	case "INCR":
		return evalINCR(cmd.Args, c)
	}
	return errors.ErrUnsupported
}

func EvalAndRespondCmds(cmds Cmds, c io.ReadWriter) error {
	var response []byte
	buf := bytes.NewBuffer(response)

	var errs error
	for _, cmd := range cmds {
		if err := EvalAndRespond(*cmd, buf); err != nil {
			errors.Join(errs, err)
		}
	}

	c.Write(buf.Bytes())
	return errs
}
