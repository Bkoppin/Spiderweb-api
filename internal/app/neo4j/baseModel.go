package neo

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

/*
NeoBaseModel is a base model for Neo4j database operations.

It provides methods for creating, finding, and managing nodes in the Neo4j database.

It uses generics to allow for any model type to be used.

Example:

	// Define your model, embedding NeoBaseModel
	type User struct {
		NeoBaseModel[User]
		ID   int    `json:"id" node:"id"`
		Name string `json:"name" node:"name"`
	}

	user := &User{
		ID:   1,
		Name: "John Doe",
	}

	dbUser := &User{}
	err := dbUser.Find(user, FindOptions{
	Field}).Populate(PopulateOptions{
			Limit: 1,
			Depth:1})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(user)
*/
type NeoBaseModel[T any] struct {
	Label  string `json:"-"`
	driver neo4j.DriverWithContext
}

/*
CreateOptions is a struct that holds options for creating a node in the Neo4j database.
It includes the field name, value, label of the node to establish a relationship with,
the relationship type, and the relationship direction.
Example:

	// CreateOptions for creating a node
	options := CreateOptions{
		Field:        "id",
		Value:        123,
		Label:        "World",
		Rel:          "OWNS",
		RelDirection: "->",
	}

	// Create a new node
	err := dbUser.Create(user, options)
	if err != nil {
		log.Fatal(err)
	}
*/
type CreateOptions struct {
	Field        string      // Field name you want to target ie: id
	Value        interface{} // Value you want to target ie: 123
	Label        string      // Label of node you want to establish a relationship with ie: World
	Rel          string      // Relationship type you want to establish ie: OWNS
	RelDirection string      // Relationship direction you want to establish ie: ->
}

type DeleteOptions struct {
	Detach bool // Whether to detach the node from relationships before deletion
}

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

/*
CloseDriver closes the Neo4j driver connection.
This should be called when the application is shutting down to release resources.
Example:

	// Close the Neo4j driver connection
	defer dbUser.CloseDriver()
*/
func (b *NeoBaseModel[T]) CloseDriver() {
	if b.driver != nil {
		b.driver.Close(context.Background())
	}
}

/*
@method Find

@description Find a single node in the Neo4j database by a specific field and value.

@params model *T - The model to populate with the found node data.

@params field string - The field name to search for in the database.

@params value interface{} - The value to search for in the database.

@returns *PopulateQuery[T] - A pointer to a PopulateQuery struct that can be used to further refine the query.

@example

	// Find a single node in the Neo4j database
	user := &User{}
	err := dbUser.Find(user, "id", 123).Populate(PopulateOptions{
		Limit: 1,
		Depth: 1,
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(user)
*/
func (b *NeoBaseModel[T]) Find(model *T, field string, value interface{}) *PopulateQuery[T] {
	return &PopulateQuery[T]{
		baseModel: b,
		model:     model,
		field:     field,
		value:     value,
	}
}

/*
@method FindAll

@description Find all nodes in the Neo4j database by a specific field and value.

@params models *[]T - A pointer to a slice of models to populate with the found nodes data.

@params field string - The field name to search for in the database.

@params value interface{} - The value to search for in the database.

@returns *PopulateQuery[T] - A pointer to a PopulateQuery struct that can be used to further refine the query.

@example

	// Find all nodes in the Neo4j database
	users := []User{}
	err := dbUser.FindAll(&users, "id", 123).Populate(PopulateOptions{
		Limit: 1,
		Depth: 1,
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(users)
*/
func (b *NeoBaseModel[T]) FindAll(models *[]T, field string, value interface{}) *PopulateQuery[T] {
	return &PopulateQuery[T]{
		baseModel: b,
		models:    models,
		field:     field,
		value:     value,
	}
}

/*
@method Create

@description Create a new node in the Neo4j database.

@params model *T - The model to create in the database.

@params options CreateOptions - Options for creating the node, including field, value, label, relationship type, and direction.

@example

	// Create a new node in the Neo4j database
	user := &User{
		ID:   1,
		Name: "John Doe",
	}
	options := CreateOptions{
		Field:        "id",
		Value:        123,
		Label:        "World",
		Rel:          "OWNS",
		RelDirection: "->",
	}
	err := dbUser.Create(user, options)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(user)
*/
func (b *NeoBaseModel[T]) Create(model *T, options CreateOptions) error {
	if err := b.initDriver(); err != nil {
		return err
	}

	ctx := context.Background()
	session := b.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	query, params := b.buildCreateQuery(model, options)

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

	createdNode, ok := result.(neo4j.Node)
	if !ok {
		return fmt.Errorf("unexpected result type: %T", result)
	}

	return mapNodeToModel(createdNode, model)
}

/*
@method @private buildCreateQuery

@description Build the Cypher query for creating a new node in the Neo4j database.

@params model *T - The model to create in the database.

@params options CreateOptions - Options for creating the node, including field, value, label, relationship type, and direction.

@returns (string, map[string]interface{}) - The Cypher query string and a map of parameters to be used in the query.
*/
func (b *NeoBaseModel[T]) buildCreateQuery(model *T, options CreateOptions) (string, map[string]interface{}) {
	modelType := reflect.TypeOf(*model)
	modelValue := reflect.ValueOf(*model)

	var queryBuilder strings.Builder
	params := make(map[string]interface{})

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

	query := queryBuilder.String()
	query = strings.TrimSuffix(query, ", ")
	queryBuilder.Reset()
	queryBuilder.WriteString(query)
	queryBuilder.WriteString("})")

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

/*
@method Delete

@description Delete a node in the Neo4j database by a specific field and value.

@params model *T - The model to delete from the database.

@params field string - The field name to search for in the database.

@params value interface{} - The value to search for in the database.

@params options DeleteOptions - Options for deleting the node, including whether to detach it from relationships.
@example

	// Delete a node in the Neo4j database
	user := &User{}
	err := dbUser.Delete(user, "id", 123, DeleteOptions{
		Detach: true,
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Node deleted")
*/
func (b *NeoBaseModel[T]) Delete(model *T, field string, value interface{}, options DeleteOptions) error {
	if err := b.initDriver(); err != nil {
		return err
	}

	ctx := context.Background()
	session := b.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	// Step 1: Retrieve the node before deletion
	queryRetrieve := fmt.Sprintf("MATCH (n:%s {%s: $value}) RETURN n", b.Label, field)
	if field == "elementID" {
		queryRetrieve = fmt.Sprintf("MATCH (n:%s) WHERE elementId(n) = $value RETURN n", b.Label)
	}

	params := map[string]interface{}{
		"value": value,
	}

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		res, err := tx.Run(ctx, queryRetrieve, params)
		if err != nil {
			return nil, err
		}

		if res.Next(ctx) {
			node, ok := res.Record().Get("n")
			if !ok {
				return nil, fmt.Errorf("failed to retrieve node before deletion")
			}
			return node, nil
		}

		return nil, fmt.Errorf("node not found for deletion")
	})

	if err != nil {
		return err
	}

	// Map the retrieved node to the model
	node, ok := result.(neo4j.Node)
	if !ok {
		return fmt.Errorf("unexpected result type: %T", result)
	}

	if err := mapNodeToModel(node, model); err != nil {
		return fmt.Errorf("failed to map node to model: %w", err)
	}

	// Step 2: Delete the node
	queryDelete := fmt.Sprintf("MATCH (n:%s {%s: $value}) DELETE n", b.Label, field)

	if field == "elementID" {
		queryDelete = fmt.Sprintf("MATCH (n:%s) WHERE elementId(n) = $value DELETE n", b.Label)
	}

	if options.Detach {
		detachDelete := "DETACH DELETE n"
		queryDelete = strings.Replace(queryDelete, "DELETE n", detachDelete, 1)
	}

	_, err = session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		result, err := tx.Run(ctx, queryDelete, params)
		if err != nil {
			return nil, err
		}
		return result.Consume(ctx)
	})

	return err
}

func (b *NeoBaseModel[T]) Update(model *T, options CreateOptions) error {
	if err := b.initDriver(); err != nil {
		return err
	}

	ctx := context.Background()
	session := b.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	query, params := b.buildUpdateQuery(model, options)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		result, err := tx.Run(ctx, query, params)
		if err != nil {
			return nil, err
		}
		return result.Consume(ctx)
	})

	return err
}

func (b *NeoBaseModel[T]) buildUpdateQuery(model *T, options CreateOptions) (string, map[string]interface{}) {
	modelType := reflect.TypeOf(*model)
	modelValue := reflect.ValueOf(*model)

	var queryBuilder strings.Builder
	params := make(map[string]interface{})

	// Build the MATCH clause for the node to update
	// Use "elementId" for matching instead of "id"
	queryBuilder.WriteString(fmt.Sprintf("MATCH (n:%s WHERE elementId(n) = $value) ", b.Label))
	params["value"] = modelValue.FieldByName("ID").Interface()

	// Build the SET clause for updating the node's properties
	queryBuilder.WriteString("SET ")
	for i := 0; i < modelType.NumField(); i++ {
		field := modelType.Field(i)
		nodeTag := field.Tag.Get("node")
		if nodeTag == "" {
			continue
		}

		fieldValue := modelValue.Field(i).Interface()

		// Skip the "id" field in the update
		if nodeTag == "id" {
			continue
		}

		// Default behavior for other fields
		queryBuilder.WriteString(fmt.Sprintf("n.%s = $%s, ", nodeTag, nodeTag))
		params[nodeTag] = fieldValue
	}

	// Remove the trailing comma and space from the SET clause
	query := queryBuilder.String()
	query = strings.TrimSuffix(query, ", ")
	queryBuilder.Reset()
	queryBuilder.WriteString(query)

	// Add optional relationship handling if CreateOptions are provided
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
