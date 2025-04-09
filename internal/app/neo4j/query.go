package neo

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

type PopulateOptions struct {
	Depth int
	Limit int
}

type PopulateQuery[T any] struct {
	baseModel *NeoBaseModel[T]
	model     *T
	models    *[]T
	field     string
	value     interface{}
	options   PopulateOptions
}

// @method Populate
//
// @description Populates a single model or a slice of models with related nodes from Neo4j.
//
// @param options PopulateOptions
//
// @return error
//
// @example
//
//	// Populate a single model
//	var user User
//	user := User{}
//	err := user.Find(&user, "userID", 123).Populate(PopulateOptions{Depth: 2})
//	if err != nil {
//		log.Fatal(err)
//	}
//	fmt.Println(user)
func (q *PopulateQuery[T]) Populate(options PopulateOptions) error {
	q.options = options
	if q.model != nil {
		return q.executeSingle()
	}
	if q.models != nil {
		return q.executeMultiple()
	}
	return fmt.Errorf("no model or models provided")
}

func (q *PopulateQuery[T]) executeSingle() error {
	if err := q.baseModel.initDriver(); err != nil {
		return err
	}

	ctx := context.Background()
	session := q.baseModel.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)
	defer q.baseModel.driver.Close(ctx)

	query, params := q.buildQuery()
	records, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		res, err := tx.Run(ctx, query, params)
		if err != nil {
			return nil, err
		}

		var recordList []neo4j.Record
		for res.Next(ctx) {
			recordList = append(recordList, *res.Record())
		}
		if err := res.Err(); err != nil {
			return nil, err
		}

		return recordList, nil
	})

	if err != nil {
		return err
	}

	recordList, ok := records.([]neo4j.Record)
	if !ok {
		return fmt.Errorf("failed to convert result to []neo4j.Record")
	}

	mappedNodes, err := buildNodeTree[T](recordList)
	if err != nil {
		return err
	}

	if len(mappedNodes) == 0 {
		return fmt.Errorf("no nodes found")
	}

	*q.model = *mappedNodes[0]
	return nil
}

func (q *PopulateQuery[T]) executeMultiple() error {
	if err := q.baseModel.initDriver(); err != nil {
		return err
	}

	ctx := context.Background()
	session := q.baseModel.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)
	defer q.baseModel.driver.Close(ctx)

	query, params := q.buildQuery()
	records, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		res, err := tx.Run(ctx, query, params)
		if err != nil {
			return nil, err
		}

		var recordList []neo4j.Record
		for res.Next(ctx) {
			recordList = append(recordList, *res.Record())
		}
		if err := res.Err(); err != nil {
			return nil, err
		}

		return recordList, nil
	})

	if err != nil {
		return err
	}

	recordList, ok := records.([]neo4j.Record)
	if !ok {
		return fmt.Errorf("failed to convert result to []neo4j.Record")
	}

	mappedNodes, err := buildNodeTree[T](recordList)
	if err != nil {
		return err
	}

	if len(mappedNodes) == 0 {
		return fmt.Errorf("no nodes found")
	}

	*q.models = make([]T, len(mappedNodes))
	for i, node := range mappedNodes {
		(*q.models)[i] = *node
	}
	return nil
}

func (q *PopulateQuery[T]) buildQuery() (string, map[string]interface{}) {
	if q.baseModel.Label == "" {
		panic("baseModel.Label is not set. Ensure the model's Label field is initialized.")
	}

	query := fmt.Sprintf("MATCH (n:%s {%s: $%s})", q.baseModel.Label, q.field, q.field)
	if q.field == "elementID" {
		query = fmt.Sprintf("MATCH (n:%s) WHERE elementId(n) = $%s", q.baseModel.Label, q.field)
	}
	relationships := q.buildRelationships(reflect.TypeOf(*q.model), q.options.Depth)
	for _, rel := range relationships {
		query += fmt.Sprintf(" OPTIONAL MATCH %s", rel)
	}

	if q.options.Limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", q.options.Limit)
	}

	query += " RETURN n, collect(r) as relatedNodes"

	params := map[string]interface{}{
		q.field: q.value,
	}

	fmt.Printf("Query -> Neo4j: %s\n", query)

	return query, params
}

func (q *PopulateQuery[T]) buildRelationships(modelType reflect.Type, depth int) []string {
	if depth == 0 {
		depth = -1
	}

	var paths []string
	for i := 0; i < modelType.NumField(); i++ {
		field := modelType.Field(i)
		relTag := field.Tag.Get("rel")
		if relTag == "" {
			continue
		}

		tagParts := strings.Split(relTag, ",")
		if len(tagParts) != 2 {
			continue
		}
		relType := tagParts[0]
		relDirection := tagParts[1]

		relatedNodeLabel := ""
		if field.Type.Kind() == reflect.Ptr {
			relatedNodeLabel = field.Type.Elem().Name()
		} else if field.Type.Kind() == reflect.Slice && field.Type.Elem().Kind() == reflect.Ptr {
			relatedNodeLabel = field.Type.Elem().Elem().Name()
		} else {
			relatedNodeLabel = field.Type.Name()
		}

		path := fmt.Sprintf("(n)-[:%s]%s(r:%s)", relType, relDirection, relatedNodeLabel)
		paths = append(paths, path)

		if depth != 1 && field.Type.Kind() == reflect.Struct {
			nestedPaths := q.buildRelationships(field.Type, depth-1)
			paths = append(paths, nestedPaths...)
		} else if depth != 1 && field.Type.Kind() == reflect.Ptr && field.Type.Elem().Kind() == reflect.Struct {
			nestedPaths := q.buildRelationships(field.Type.Elem(), depth-1)
			paths = append(paths, nestedPaths...)
		} else if depth != 1 && field.Type.Kind() == reflect.Slice && field.Type.Elem().Kind() == reflect.Ptr && field.Type.Elem().Elem().Kind() == reflect.Struct {
			nestedPaths := q.buildRelationships(field.Type.Elem().Elem(), depth-1)
			paths = append(paths, nestedPaths...)
		}
	}

	return paths
}
