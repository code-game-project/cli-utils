package modules

import (
	"fmt"
	"io"
	"os"

	"google.golang.org/protobuf/proto"
)

func GetCreateData() *ActionCreateData {
	data, err := readActionData()
	if err != nil {
		panic(err)
	}
	message := &ActionCreateData{}
	err = proto.Unmarshal(data, message)
	if err != nil {
		panic(err)
	}
	return message
}

func readActionData() ([]byte, error) {
	path := os.Getenv("CG_MODULE_ACTION_DATA_FILE")
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open action data file '%s': %w", path, err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read action data file '%s': %w", path, err)
	}
	return data, nil
}
