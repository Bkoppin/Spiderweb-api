package neo

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)


type NeoBaseModel[T any] struct {
	Label  string `json:"-"`
  driver neo4j.DriverWithContext
}

type FindOptions struct {
	Field string
	Value interface{}
}

type CreateOptions struct {
	Field string // Field name you want to target ie: id
	Value interface{} // Value you want to target ie: 123
	Label string // Label of node you want to establish a relationship with ie: World
	Rel  string // Relationship type you want to establish ie: OWNS
	RelDirection string // Relationship direction you want to establish ie: ->
}

// Initialize the Neo4j driver (singleton pattern)
func (b *NeoBaseModel[T]) initDriver() error {
	if b.Label == "" {
		b.Label = reflect.TypeOf(*new(T)).Name()
	}
    if b.driver == nil {
        var err error
        b.driver, err = NewDriver()
        if err != nil {
            return fmt.Errorf("failed to initialize Neo4j driver: %w", err)
        }
    }
    return nil
}

// Close the Neo4j driver (optional, for cleanup)
func (b *NeoBaseModel[T]) CloseDriver() {
    if b.driver != nil {
        b.driver.Close(context.Background())
    }
}

// Find a single record and populate the model
func (b *NeoBaseModel[T]) Find(model *T, field string, value interface{}) *PopulateQuery[T] {
    return &PopulateQuery[T]{
        baseModel: b,
        model:     model,
        field:     field,
        value:     value,
    }
}

// Find multiple records and populate the slice
func (b *NeoBaseModel[T]) FindAll(models *[]T, options FindOptions) *PopulateQuery[T] {
    return &PopulateQuery[T]{
        baseModel: b,
        models:    models,
        field:     options.Field,
        value:     options.Value,
    }
}

// Create a node and link relationships
func (b *NeoBaseModel[T]) Create(model *T, options CreateOptions) error {
  if err := b.initDriver(); err != nil {
    return err
  }

  ctx := context.Background()
  session := b.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
  defer session.Close(ctx)

  query, params := b.buildCreateQuery(model, options)

  // Execute the query and return the created node
  result, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
    records, err := tx.Run(ctx, query+" RETURN n", params)
    if err != nil {
      return nil, err
    }

    if records.Next(ctx) {
      value, ok := records.Record().Get("n")
      if !ok {
        return nil, fmt.Errorf("failed to retrieve 'n' from record")
      }
      node, ok := value.(neo4j.Node)
      if !ok {
        return nil, fmt.Errorf("failed to cast result to neo4j.Node")
      }
      return node, nil
    }

    return nil, fmt.Errorf("failed to create node")
  })

  if err != nil {
    return err
  }

  // Map the created node back to the model
  createdNode, ok := result.(neo4j.Node)
  if !ok {
    return fmt.Errorf("unexpected result type: %T", result)
  }

  return mapNodeToModel(createdNode, model)
}

func (b *NeoBaseModel[T]) buildCreateQuery(model *T, options CreateOptions) (string, map[string]interface{}) {
  modelType := reflect.TypeOf(*model)
  modelValue := reflect.ValueOf(*model)

  var queryBuilder strings.Builder
  params := make(map[string]interface{})

  // Build the CREATE clause for the primary node
  queryBuilder.WriteString(fmt.Sprintf("CREATE (n:%s {", b.Label))
  for i := 0; i < modelType.NumField(); i++ {
    field := modelType.Field(i)
    nodeTag := field.Tag.Get("node")
    if nodeTag == "" {
      continue
    }

    fieldValue := modelValue.Field(i).Interface()
    queryBuilder.WriteString(fmt.Sprintf("%s: $%s, ", nodeTag, nodeTag))
    params[nodeTag] = fieldValue
  }

  // Remove the trailing comma and space
  query := queryBuilder.String()
  query = strings.TrimSuffix(query, ", ")
  queryBuilder.Reset()
  queryBuilder.WriteString(query)
  queryBuilder.WriteString("})")

  // Add optional relationship if CreateOptions are provided
  if options.Field != "" && options.Value != nil && options.Label != "" {
    queryBuilder.WriteString(fmt.Sprintf(" MERGE (r:%s {%s: $relatedValue})", options.Label, options.Field))
    if options.RelDirection == "->" {
      queryBuilder.WriteString(fmt.Sprintf(" CREATE (n)-[:%s]->(r)", options.Rel))
    } else if options.RelDirection == "<-" {
      queryBuilder.WriteString(fmt.Sprintf(" CREATE (n)<-[:%s]-(r)", options.Rel))
    }
    params["relatedValue"] = options.Value
  }

  return queryBuilder.String(), params
}