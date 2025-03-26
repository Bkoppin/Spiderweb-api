package neo

import (
	"context"
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

type QueryBuilder struct {
	createClause []string
	matchClause []string
	with string
	optionalMatch []string
	returnClause string
	params map[string]interface{}
}

func NewDriver() (neo4j.DriverWithContext, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, err
	}

	uri := os.Getenv("NEO4J_URI")
	username := os.Getenv("NEO4J_USER")
	password := os.Getenv("NEO4J_PASSWORD")

	driver, err := neo4j.NewDriverWithContext(uri, neo4j.BasicAuth(username, password, ""))
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	err = driver.VerifyConnectivity(ctx)
	if err != nil {
		return nil, err
	}
	return driver, nil
}


func NewQueryBuilder(queryType string) (*QueryBuilder, error) {
	if queryType != "create" && queryType != "match" {
		return nil, fmt.Errorf("invalid query type, must be 'create' or 'match'")
	}
	return &QueryBuilder{
		params:    make(map[string]interface{}),
	}, nil
}

func (qb *QueryBuilder) With(param string) *QueryBuilder {
	qb.with = param
	return qb
}


func (qb *QueryBuilder) Match(clause string) *QueryBuilder {
	
	qb.matchClause = append(qb.matchClause, clause)
	return qb
}

func (qb *QueryBuilder) Create(clause string) *QueryBuilder {
	qb.createClause = append(qb.createClause, clause)
	return qb
}

func (qb *QueryBuilder) OptionalMatch(clause string) *QueryBuilder {
	qb.optionalMatch = append(qb.optionalMatch, clause)
	return qb
}

func (qb *QueryBuilder) Return(returnClause string) *QueryBuilder {
	qb.returnClause = returnClause
	return qb
}

func (qb *QueryBuilder) WithParam(key string, value interface{}) *QueryBuilder {
	qb.params[key] = value
	return qb
}

func (qb *QueryBuilder) Build() (string, map[string]interface{}) {
	var query string
	for _, clause := range qb.matchClause {
		query += fmt.Sprintf("\nMATCH %s", clause)
	}
	if qb.with != "" {
		query += fmt.Sprintf("\nWITH %s", qb.with)
	}
	for _, clause := range qb.createClause {
		query += fmt.Sprintf("\nCREATE %s", clause)
	}


	for _, clause := range qb.optionalMatch {
		query += fmt.Sprintf("\nOPTIONAL MATCH %s", clause)
	}
	query += fmt.Sprintf("\nRETURN %s", qb.returnClause)
	return query, qb.params
}


