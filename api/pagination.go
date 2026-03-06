package main

import (
	"fmt"
	"net/http"
	"strings"
)

type PaginationRequest struct {
	Page  int `json:"page"`
	Limit int `json:"limit"`

	GetOffset      func() int                                     `json:"-"`
	PaginateQuery  func(sb *strings.Builder, args *[]interface{}) `json:"-"`
	CreateResponse func(count *int) PaginationResponse            `json:"-"`
}

type PaginationResponse struct {
	Page  int  `json:"page"`
	Limit int  `json:"limit"`
	Count *int `json:"count"`
}

const DEFAULT_LIMIT = 20

func EnsurePaginationRequestIsValid(pagination *PaginationRequest) {
	if pagination.Page <= 0 {
		pagination.Page = 1
	}

	if pagination.Limit <= 0 {
		pagination.Limit = DEFAULT_LIMIT
	}
}

func CreatePaginationRequest(page int, limit int) PaginationRequest {
	pagination := PaginationRequest{
		Page:  page,
		Limit: limit,
	}

	EnsurePaginationRequestIsValid(&pagination)

	pagination.GetOffset = func() int {
		return GetPaginationOffset(pagination.Page, pagination.Limit)
	}

	pagination.PaginateQuery = func(sb *strings.Builder, args *[]interface{}) {
		PaginateQuery(&pagination, sb, args)
	}

	pagination.CreateResponse = func(count *int) PaginationResponse {
		return CreatePaginationResponse(&pagination, count)
	}

	return pagination
}

func CreatePaginationRequestFromBody(r *http.Request) (PaginationRequest, error) {
	pagination := CreatePaginationRequest(1, DEFAULT_LIMIT)

	if err := bodyUnmarshalOptional(r, &pagination); err != nil {
		return PaginationRequest{}, errorInvalidJSONBody
	}

	EnsurePaginationRequestIsValid(&pagination)

	return pagination, nil
}

func CreatePaginationResponse(pagination *PaginationRequest, count *int) PaginationResponse {
	return PaginationResponse{
		Page:  pagination.Page,
		Limit: pagination.Limit,
		Count: count,
	}
}

func GetPaginationOffset(page, limit int) int {
	if limit <= 0 || page <= 1 {
		return 0
	}
	return (page - 1) * limit
}

func PaginateQuery(pagination *PaginationRequest, sb *strings.Builder, args *[]interface{}) {
	argsIndex := len(*args)
	*args = append(*args, pagination.Limit, pagination.GetOffset())

	sb.WriteString(fmt.Sprintf("LIMIT $%d OFFSET $%d", argsIndex+1, argsIndex+2))
}
