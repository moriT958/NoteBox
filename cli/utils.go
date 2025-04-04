package cli

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func validateArgs(args []string) bool {
	l := len(args)
	if l <= 1 {
		return true
	} else {
		return false
	}
}

func getIdArg(args []string) (int, error) {
	var idStr string

	if !validateArgs(args) {
		return 0, errors.New("too many arguments: expected only one")
	}

	if len(args) > 0 {
		idStr = args[0]
	} else {
		fmt.Fprint(os.Stdout, "Enter ID: ")
		input, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil {
			return 0, err
		}
		idStr = strings.TrimSpace(input)
		if idStr == "" {
			return 0, errors.New("ID cannot be empty")
		}
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		return 0, err
	}

	return id, nil
}
