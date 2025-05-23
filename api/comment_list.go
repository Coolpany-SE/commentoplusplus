package main

import (
	"strings"
	"database/sql"
	"net/http"
)

func commentList(commenterHex string, domain string, path string, includeUnapproved bool) ([]comment, map[string]commenter, error) {
	// path can be empty
	if commenterHex == "" || domain == "" {
		return nil, nil, errorMissingField
	}

	statement := `
		SELECT
			commentHex,
			commenterHex,
			markdown,
			html,
			parentHex,
			score,
			state,
			deleted,
			creationDate
		FROM comments
		WHERE
			canon($1) LIKE canon(comments.domain) AND
			comments.path = $2
	`

	if !includeUnapproved {
		if commenterHex == "anonymous" {
			statement += `AND state = 'approved'`
		} else {
			statement += `AND (state = 'approved' OR commenterHex = $3)`
		}
	}

	statement += `;`

	var rows *sql.Rows
	var err error

	if !includeUnapproved && commenterHex != "anonymous" {
		rows, err = db.Query(statement, domain, path, commenterHex)
	} else {
		rows, err = db.Query(statement, domain, path)
	}

	if err != nil {
		logger.Errorf("cannot get comments: %v", err)
		return nil, nil, errorInternal
	}
	defer rows.Close()

	commenters := make(map[string]commenter)
	commenters["anonymous"] = commenter{CommenterHex: "anonymous", Email: "undefined", Name: "Anonymous", Link: "undefined", Photo: "undefined", Provider: "undefined"}

	comments := []comment{}
	for rows.Next() {
		c := comment{}
		if err = rows.Scan(
			&c.CommentHex,
			&c.CommenterHex,
			&c.Markdown,
			&c.Html,
			&c.ParentHex,
			&c.Score,
			&c.State,
			&c.Deleted,
			&c.CreationDate); err != nil {
			return nil, nil, errorInternal
		}

		if commenterHex != "anonymous" {
			statement = `
				SELECT direction
				FROM votes
				WHERE commentHex=$1 AND commenterHex=$2;
			`
			row := db.QueryRow(statement, c.CommentHex, commenterHex)

			if err = row.Scan(&c.Direction); err != nil {
				// TODO: is the only error here that there is no such entry?
				c.Direction = 0
			}
		}

		if commenterHex != c.CommenterHex {
			c.Markdown = ""
		}

		if !includeUnapproved {
			c.State = ""
		}

		comments = append(comments, c)

		if _, ok := commenters[c.CommenterHex]; !ok {
			commenters[c.CommenterHex], err = commenterGetByHex(c.CommenterHex)
			if err != nil {
				logger.Errorf("cannot retrieve commenter: %v", err)
				return nil, nil, errorInternal
			}
		}
	}

	return comments, commenters, nil
}

func commentListHandler(w http.ResponseWriter, r *http.Request) {
	type request struct {
		CommenterToken *string `json:"CommenterToken"`
		Domain         *string `json:"domain"`
		Path           *string `json:"path"`
	}

	var x request
	if err := bodyUnmarshal(r, &x); err != nil {
		bodyMarshal(w, response{"success": false, "message": err.Error()})
		return
	}

	domain := domainStrip(*x.Domain)
	path := *x.Path

	d, err := domainGet(domain)
	if err != nil {
		bodyMarshal(w, response{"success": false, "message": err.Error()})
		return
	}

	p, err := pageGet(domain, path)
	if err != nil {
		bodyMarshal(w, response{"success": false, "message": err.Error()})
		return
	}

	commenterHex := "anonymous"
	isModerator := false
	modList := map[string]bool{}

	if *x.CommenterToken != "anonymous" {
		c, err := commenterGetByCommenterToken(*x.CommenterToken)
		if err != nil {
			if err == errorNoSuchToken {
				commenterHex = "anonymous"
			} else {
				bodyMarshal(w, response{"success": false, "message": err.Error()})
				return
			}
		} else {
			commenterHex = c.CommenterHex
		}

		for _, mod := range d.Moderators {
			modList[mod.Email] = true
			if mod.Email == c.Email {
				isModerator = true
			}
		}
	} else {
		for _, mod := range d.Moderators {
			modList[mod.Email] = true
		}
	}

	domainViewRecord(domain, commenterHex)

	comments, commenters, err := commentList(commenterHex, domain, path, isModerator)
	if err != nil {
		bodyMarshal(w, response{"success": false, "message": err.Error()})
		return
	}

	_commenters := map[string]commenter{}
	for commenterHex, cr := range commenters {
		if _, ok := modList[cr.Email]; ok {
			cr.IsModerator = true
		}
		cr.Email = ""
		_commenters[commenterHex] = cr
	}

	bodyMarshal(w, response{
		"success":               true,
		"domain":                domain,
		"comments":              comments,
		"commenters":            _commenters,
		"requireModeration":     d.RequireModeration,
		"requireIdentification": d.RequireIdentification,
		"isFrozen":              d.State == "frozen",
		"isModerator":           isModerator,
		"defaultSortPolicy":     d.DefaultSortPolicy,
		"attributes":            p,
		"configuredOauths": map[string]bool{
			"commento": d.CommentoProvider,
			"google":   googleConfigured && d.GoogleProvider,
			"twitter":  twitterConfigured && d.TwitterProvider,
			"github":   githubConfigured && d.GithubProvider,
			"gitlab":   gitlabConfigured && d.GitlabProvider,
			"sso":      d.SsoProvider,
		},
	})
}

func commentListApprovals(domain string) ([]comment, map[string]commenter, error) {
	if domain == "" {
		return nil, nil, errorMissingField
	}

	statement := `
		SELECT
			path,
			commentHex,
			commenterHex,
			markdown,
			html,
			parentHex,
			score,
			state,
			deleted,
			creationDate
		FROM comments
		WHERE
		canon(comments.domain) LIKE canon($1) AND deleted = false AND
			( state = 'unapproved' OR state = 'flagged' );
	`

	var rows *sql.Rows
	var err error

	rows, err = db.Query(statement, domain)

	if err != nil {
		logger.Errorf("cannot get comments: %v", err)
		return nil, nil, errorInternal
	}
	defer rows.Close()

	commenters := make(map[string]commenter)
	commenters["anonymous"] = commenter{CommenterHex: "anonymous", Email: "undefined", Name: "Anonymous", Link: "undefined", Photo: "undefined", Provider: "undefined"}

	comments := []comment{}
	for rows.Next() {
		c := comment{}
		if err = rows.Scan(
			&c.Path,
			&c.CommentHex,
			&c.CommenterHex,
			&c.Markdown,
			&c.Html,
			&c.ParentHex,
			&c.Score,
			&c.State,
			&c.Deleted,
			&c.CreationDate); err != nil {
			return nil, nil, errorInternal
		}

		comments = append(comments, c)

		if _, ok := commenters[c.CommenterHex]; !ok {
			commenters[c.CommenterHex], err = commenterGetByHex(c.CommenterHex)
			if err != nil {
				logger.Errorf("cannot retrieve commenter: %v", err)
				return nil, nil, errorInternal
			}
		}
	}

	return comments, commenters, nil
}

func commentListApprovalsHandler(w http.ResponseWriter, r *http.Request) {
	type request struct {
		OwnerToken *string `json:"ownerToken"`
		Domain     *string `json:"domain"`
	}

	var x request
	if err := bodyUnmarshal(r, &x); err != nil {
		bodyMarshal(w, response{"success": false, "message": err.Error()})
		return
	}

	o, err := ownerGetByOwnerToken(*x.OwnerToken)
	if err != nil {
		bodyMarshal(w, response{"success": false, "message": err.Error()})
		return
	}

	domain := domainStrip(*x.Domain)
	isOwner, err := domainOwnershipVerify(o.OwnerHex, domain)
	if err != nil {
		bodyMarshal(w, response{"success": false, "message": err.Error()})
		return
	}

	if !isOwner {
		bodyMarshal(w, response{"success": false, "message": errorNotAuthorised.Error()})
		return
	}

	comments, commenters, err := commentListApprovals(domain)
	if err != nil {
		bodyMarshal(w, response{"success": false, "message": err.Error()})
		return
	}

	_commenters := map[string]commenter{}
	for commenterHex, cr := range commenters {
		cr.Email = ""
		_commenters[commenterHex] = cr
	}

	bodyMarshal(w, response{
		"success":    true,
		"domain":     domain,
		"comments":   comments,
		"commenters": _commenters,
	})

}

func commentListAll(domain string, includeDeleted bool, includeUnapproved bool) ([]comment, map[string]commenter, error) {
	if domain == "" {
		return nil, nil, errorMissingField
	}

	var sb strings.Builder

	sb.WriteString(`
		SELECT
			path,
			commentHex,
			commenterHex,
			markdown,
			html,
			parentHex,
			score,
			state,
			deleted,
			creationDate
		FROM comments
		WHERE
		canon(comments.domain) LIKE canon($1) 
	`)

	if !includeDeleted {
		sb.WriteString("AND deleted = false ")
	}

	if !includeUnapproved {
		sb.WriteString("AND ( state = 'approved'  ) ")
	}

	sb.WriteString("ORDER BY creationDate DESC;")

	statement := sb.String()

	var rows *sql.Rows
	var err error

	rows, err = db.Query(statement, domain)

	if err != nil {
		logger.Errorf("cannot get comments: %v", err)
		return nil, nil, errorInternal
	}
	defer rows.Close()

	commenters := make(map[string]commenter)
	commenters["anonymous"] = commenter{CommenterHex: "anonymous", Email: "undefined", Name: "Anonymous", Link: "undefined", Photo: "undefined", Provider: "undefined"}

	comments := []comment{}
	for rows.Next() {
		c := comment{}
		if err = rows.Scan(
			&c.Path,
			&c.CommentHex,
			&c.CommenterHex,
			&c.Markdown,
			&c.Html,
			&c.ParentHex,
			&c.Score,
			&c.State,
			&c.Deleted,
			&c.CreationDate); err != nil {
			return nil, nil, errorInternal
		}

		comments = append(comments, c)

		if _, ok := commenters[c.CommenterHex]; !ok {
			commenters[c.CommenterHex], err = commenterGetByHex(c.CommenterHex)
			if err != nil {
				logger.Errorf("cannot retrieve commenter: %v", err)
				return nil, nil, errorInternal
			}
		}
	}

	return comments, commenters, nil
}

func commentListAllHandler(w http.ResponseWriter, r *http.Request) {
	type request struct {
		OwnerToken *string `json:"ownerToken"`
		Domain     *string `json:"domain"`
		IncludeDeleted    *bool `json:"includeDeleted"`
		IncludeUnapproved *bool `json:"includeUnapproved"`
	}

	var x request
	if err := bodyUnmarshal(r, &x); err != nil {
		bodyMarshal(w, response{"success": false, "message": err.Error()})
		return
	}

	o, err := ownerGetByOwnerToken(*x.OwnerToken)
	if err != nil {
		bodyMarshal(w, response{"success": false, "message": err.Error()})
		return
	}

	domain := domainStrip(*x.Domain)
	isOwner, err := domainOwnershipVerify(o.OwnerHex, domain)
	if err != nil {
		bodyMarshal(w, response{"success": false, "message": err.Error()})
		return
	}

	if !isOwner {
		bodyMarshal(w, response{"success": false, "message": errorNotAuthorised.Error()})
		return
	}

	comments, commenters, err := commentListAll(domain, *x.IncludeDeleted, *x.IncludeUnapproved)
	if err != nil {
		bodyMarshal(w, response{"success": false, "message": err.Error()})
		return
	}

	_commenters := map[string]commenter{}
	for commenterHex, cr := range commenters {
		cr.Email = ""
		_commenters[commenterHex] = cr
	}

	bodyMarshal(w, response{
		"success":    true,
		"domain":     domain,
		"comments":   comments,
		"commenters": _commenters,
	})

}
