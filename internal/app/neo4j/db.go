package neo

import (
	"context"
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

type QueryBuilder struct {
	query string
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


func NewQueryBuilder() (*QueryBuilder) {
	return &QueryBuilder{
		params:    make(map[string]interface{}),
		
	}
}

func (qb *QueryBuilder) With(param string) *QueryBuilder {
	qb.query += fmt.Sprintf("\nWITH %s", param)
	return qb
}


func (qb *QueryBuilder) Match(clause string) *QueryBuilder {
	
	qb.query += fmt.Sprintf("\nMATCH %s", clause)
	return qb
}

func (qb *QueryBuilder) Create(clause string) *QueryBuilder {
	qb.query += fmt.Sprintf("\nCREATE %s", clause)
	return qb
}

func (qb *QueryBuilder) OptionalMatch(clause string) *QueryBuilder {
	qb.query += fmt.Sprintf("\nOPTIONAL MATCH %s", clause)
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
	query += qb.query
	query += fmt.Sprintf("\nRETURN %s", qb.returnClause)
	return query, qb.params
}


