package main

import (
	"net/http"
)

func ownerUpdate(ownerHex string, email string, name string, confirmedEmail bool) error {
	if ownerHex == "" || email == "" || name == "" {
		return errorMissingField
	}

	currentOwner, currentOwnerErr := ownerGetByOwnerHex(ownerHex)

	// Check if owner exists
	if currentOwnerErr != nil {
		return errorNoSuchOwner
	}

	// If the email is being changed, check if the new email already exists for another owner
	if currentOwner.Email != email {
		if _, err := ownerGetByEmail(email); err == nil {
			return errorEmailAlreadyExists
		}

		if err := emailNew(email); err != nil {
			return errorInternal
		}
	}

	statement := `
		UPDATE owners
		SET email = $2, name = $3, confirmedEmail = $4
		WHERE ownerHex = $1;
	`
	_, err := db.Exec(statement, ownerHex, email, name, confirmedEmail)
	if err != nil {
		logger.Errorf("cannot update owner: %v", err)
		return errorInternal
	}

	return nil
}

func ownerUpdateHandler(w http.ResponseWriter, r *http.Request) {
	type request struct {
		OwnerToken     *string `json:"ownerToken"`
		OwnerHex       *string `json:"ownerHex"`
		Email          *string `json:"email"`
		Name           *string `json:"name"`
		ConfirmedEmail *bool   `json:"confirmedEmail"`
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

	if err := ownerUpdate(*x.OwnerHex, *x.Email, *x.Name, *x.ConfirmedEmail); err != nil {
		bodyMarshal(w, response{"success": false, "message": err.Error()})
		return
	}

	bodyMarshal(w, response{"success": true})
}
