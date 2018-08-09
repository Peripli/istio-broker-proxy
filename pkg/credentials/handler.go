package credentials

import "errors"

func Update(in []byte) ([]byte, error) {
	asString := string(in)
	if !isValidUpdateRequestBody(asString) {
		return nil, errors.New("Invalid request")
	}
	out := translateCredentials(asString)
	return []byte(out), nil
}

