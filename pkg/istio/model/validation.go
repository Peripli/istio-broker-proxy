package model

import (
	"fmt"
	"github.com/gogo/protobuf/proto"
	"github.com/hashicorp/go-multierror"
	networking "istio.io/api/networking/v1alpha3"
)

// wrapper around multierror.Append that enforces the invariant that if all input errors are nil, the output
// error is nil (allowing validation without branching).
func appendErrors(err error, errs ...error) error {
	appendError := func(err, err2 error) error {
		if err == nil {
			return err2
		} else if err2 == nil {
			return err
		}
		return multierror.Append(err, err2)
	}

	for _, err2 := range errs {
		err = appendError(err, err2)
	}
	return err
}

// ValidateGateway checks gateway specifications
func ValidateGateway(name, namespace string, msg proto.Message) (errs error) {
	value, ok := msg.(*networking.Gateway)
	if !ok {
		errs = appendErrors(errs, fmt.Errorf("cannot cast to gateway: %#v", msg))
		return
	}

	if len(value.Servers) == 0 {
		errs = appendErrors(errs, fmt.Errorf("gateway must have at least one server"))
	}
	//FIXME: Skip validation of servers due to dependencies (copy of istio)

	// Ensure unique port names
	portNames := make(map[string]bool)

	for _, s := range value.Servers {
		if s.Port != nil {
			if portNames[s.Port.Name] {
				errs = appendErrors(errs, fmt.Errorf("port names in servers must be unique: duplicate name %s", s.Port.Name))
			}
			portNames[s.Port.Name] = true
		}
	}

	return errs
}
