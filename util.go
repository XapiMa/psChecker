package pschecker

import (
	"fmt"
	"os"
	"strings"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

func parseTypes(typesString string) (int, error) {
	typeStrings := strings.Split(typesString, "|")
	types := 0
	for i, typeString := range typeStrings {
		switch typeString {
		case execSentence:
			types |= execFlag
		case cmdSentence:
			types |= cmdFlag
		case openSentence:
			types |= openFlag
		case userSentence:
			types |= userFlag
		case pidSentence:
			types |= pidFlag
		case "":
		default:
			return types, fmt.Errorf("%dth type is invalid", i+1)
		}
	}
	return types, nil
}

func appendFile(outputPath, outputString string) error {
	if outputPath == "" {
		fmt.Printf("%s", outputString)
	} else {
		file, err := os.OpenFile(outputPath, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
		if err != nil {
			return err
		}
		defer file.Close()
		file.Write(([]byte)(outputString))
	}
	return nil
}

func ymlUnmarshal(fileBuffer []byte) ([]map[interface{}]interface{}, error) {
	errorWrap := func(err error) error {
		return errors.Wrap(err, "cause in ymlUnmarshal")
	}
	data := make([]map[interface{}]interface{}, 100)
	err := yaml.Unmarshal(fileBuffer, &data)
	if err != nil {
		return nil, errorWrap(err)
	}
	return data, nil
}
