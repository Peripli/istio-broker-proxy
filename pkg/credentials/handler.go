package credentials

import "errors"

func Update(in []byte) ([]byte, error) {
	asString := string(in)
	ok, err := isValidUpdateRequestBody(asString)
	if !ok {
		return nil, errors.New("Invalid request. " + err.Error())
	}

	out := translateCredentials(asString)
	return []byte(out), nil
}
