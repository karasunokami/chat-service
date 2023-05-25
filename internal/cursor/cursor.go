package cursor

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
)

func Encode(data any) (string, error) {
	bData, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("marshal data, err=%v", err)
	}

	return base64.URLEncoding.EncodeToString(bData), nil
}

func Decode(in string, to any) error {
	dst, err := base64.URLEncoding.DecodeString(in)
	if err != nil {
		return fmt.Errorf("base64 decode input, err=%v", err)
	}

	return json.Unmarshal(dst, to)
}
