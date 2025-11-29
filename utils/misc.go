package utils

import (
	"fmt"
	"os"
	"strings"
)

type QueryBuilder struct {
	Query string
	Args  []any
}

type PaginatorWrapper[T any] struct {
	TotalPages uint `json:"totalPages"`
	Data       T    `json:"data"`
}

type SearchParams struct {
	Q     *string
	Page  *uint
	Limit *uint
}

func MustEnv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		panic(fmt.Sprintf("No se encontro %s en las variables de entorno", key))
	}
	return val
}

func (qb QueryBuilder) ToString() string {
	var sb strings.Builder

	sb.WriteString("Query:\n")
	sb.WriteString(qb.Query)
	sb.WriteString("\n\nArgs:\n")

	if len(qb.Args) == 0 {
		sb.WriteString("nil")
	} else {
		for i, arg := range qb.Args {
			if arg == nil {
				sb.WriteString(fmt.Sprintf("[%d] nil\n", i))
			} else {
				sb.WriteString(fmt.Sprintf("[%d] %v (%T)\n", i, arg, arg))
			}
		}
	}

	return sb.String()
}
