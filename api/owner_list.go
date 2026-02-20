package main

import (
	"net/http"
)

func ownerList() ([]owner, error) {
	statement := `
		SELECT ` + ownersRowColumns + `
		FROM owners;
	`
	rows, err := db.Query(statement)
	if err != nil {
		logger.Errorf("cannot query owners: %v", err)
		return nil, errorInternal
	}
	defer rows.Close()

	owners := []owner{}
	for rows.Next() {
		var o owner
		if err = ownersRowScan(rows, &o); err != nil {
			logger.Errorf("cannot scan owner: %v", err)
			return nil, errorInternal
		}

		owners = append(owners, o)
	}

	return owners, rows.Err()
}

func ownerListHandler(w http.ResponseWriter, r *http.Request) {
	type request struct {
		OwnerToken *string `json:"ownerToken"`
	}

	var x request
	if err := bodyUnmarshal(r, &x); err != nil {
		bodyMarshal(w, response{"success": false, "message": err.Error()})
		return
	}

	// Verify the requesting owner is authenticated
	if _, err := ownerGetByOwnerToken(*x.OwnerToken); err != nil {
		bodyMarshal(w, response{"success": false, "message": err.Error()})
		return
	}

	owners, err := ownerList()
	if err != nil {
		bodyMarshal(w, response{"success": false, "message": err.Error()})
		return
	}

	bodyMarshal(w, response{"success": true, "owners": owners})
}
