package main

import (
	"encoding/json"
	"reflect"
)

func UnmarshalJson(body []byte, target interface{}) error {
	if err := json.Unmarshal(body, target); err != nil {
		logger.Errorf("cannot unmarshal JSON: %v\n", err)
		return errorInvalidJSONBody
	}

	return nil
}

func UnmarshalJsonRequired(body []byte, target interface{}) error {
	return UnmarshalAndValidateJson(body, target, true, []string{})
}

func UnmarshalAndValidateJson(body []byte, target interface{}, requireFields bool, optionalFields []string) error {
	if err := UnmarshalJson(body, target); err != nil {
		return err
	}

	// If fields are required, create a set of provided optional fields for quick lookup
	optional := make(map[string]bool)
	if requireFields {
		for _, field := range optionalFields {
			optional[field] = true
		}
	}

	xv := reflect.Indirect(reflect.ValueOf(target))
	xt := xv.Type()
	for i := 0; i < xv.NumField(); i++ {
		fieldName := xt.Field(i).Name
		if requireFields && xv.Field(i).IsNil() && !optional[fieldName] {
			return errorMissingField
		}
	}

	return nil
}

func MarshalJson(body interface{}) ([]byte, error) {
	json, err := json.Marshal(body)

	if err != nil {
		logger.Errorf("cannot marshal JSON: %v\n", err)
		fallback := []byte(`{"success":false,"message":"Some internal error occurred"}`)

		return fallback, err
	}

	return json, nil
}
