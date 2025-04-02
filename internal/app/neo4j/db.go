/*
package neo is a package that provides an interface to interact with a Neo4j database.

It includes a QueryBuilder for constructing Cypher queries, a Node interface for representing
nodes in the graph, and a BaseNode struct that implements the Node interface.
It also provides functions to create a Neo4j driver, build a node tree from query results,
and establish relationships between nodes.
The package uses the Neo4j Go driver to connect to the database and execute queries.

@exported:
  - @func NewQueryBuilder: creates a new QueryBuilder instance.
  - @func NewDriver: creates a new Neo4j driver instance.
  - @func BuildNodeTree: builds a tree of nodes from query results.
  - @interface Node: defines methods for interacting with nodes in the graph.
  - @struct BaseNode: implements the Node interface and provides methods for managing node properties and relationships.
  - @struct QueryBuilder: provides methods for constructing Cypher queries.
*/
package neo

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/joho/godotenv"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

/*
type QueryBuilder: A struct that provides methods for constructing Cypher queries.
	- @property query: The Cypher query string.
	- @property returnClause: The RETURN clause of the Cypher query.
	- @property params: A map of parameters to be used in the Cypher query.
	- @method With: Adds a WITH clause to the query.
	- @method Match: Adds a MATCH clause to the query.
	- @method Create: Adds a CREATE clause to the query.
	- @method OptionalMatch: Adds an OPTIONAL MATCH clause to the query.
	- @method Return: Sets the RETURN clause of the query.
	- @method WithParam: Adds a parameter to the query.
	- @method Build: Builds the final Cypher query string and returns it along with the parameters.
*/
type QueryBuilder struct {
	query        string
	returnClause string
	params       map[string]interface{}
}

/* 
func NewQueryBuilder: Creates a new QueryBuilder instance.
	- @returns: A pointer to a new QueryBuilder instance.
@example:
	qb := NewQueryBuilder()
	query, params := qb.Create("(u:User {username: $username})").
		With("u").
		Match("((u)-[:OWNS]->(w:World)).
		WithParam("username", "JohnDoe").
		Return("u, w").
		Build()
	fmt.Println(query) // Prints the constructed Cypher query
	fmt.Println(params) // Prints the parameters used in the query
*/
func NewQueryBuilder() *QueryBuilder {
	return &QueryBuilder{
		params: make(map[string]interface{}),
	}
}

/*
 @method With: Adds a WITH clause to the query.

 @param param: The WITH clause to be added.

 @returns: A pointer to the QueryBuilder instance.
*/
func (qb *QueryBuilder) With(param string) *QueryBuilder {
	qb.query += fmt.Sprintf("\nWITH %s", param)
	return qb
}

/*
 @method Match: Adds a MATCH clause to the query.

 @param clause: The MATCH clause to be added.

 @returns: A pointer to the QueryBuilder instance.
*/
func (qb *QueryBuilder) Match(clause string) *QueryBuilder {
	qb.query += fmt.Sprintf("\nMATCH %s", clause)
	return qb
}

/*
 @method Create: Adds a CREATE clause to the query.

 @param clause: The CREATE clause to be added.

 @returns: A pointer to the QueryBuilder instance.
*/
func (qb *QueryBuilder) Create(clause string) *QueryBuilder {
	qb.query += fmt.Sprintf("\nCREATE %s", clause)
	return qb
}

/*
 @method OptionalMatch: Adds an OPTIONAL MATCH clause to the query.

 @param clause: The OPTIONAL MATCH clause to be added.

 @returns: A pointer to the QueryBuilder instance.
*/
func (qb *QueryBuilder) OptionalMatch(clause string) *QueryBuilder {
	qb.query += fmt.Sprintf("\nOPTIONAL MATCH %s", clause)
	return qb
}

/*
 @method Return: Sets the RETURN clause of the query.

 @param returnClause: The RETURN clause to be set.

 @returns: A pointer to the QueryBuilder instance.
*/
func (qb *QueryBuilder) Return(returnClause string) *QueryBuilder {
	qb.returnClause = returnClause
	return qb
}

/*
 @method WithParam: Adds a parameter to the query.

 @param key: The key of the parameter.

 @param value: The value of the parameter.

 @returns: A pointer to the QueryBuilder instance.
*/
func (qb *QueryBuilder) WithParam(key string, value interface{}) *QueryBuilder {
	qb.params[key] = value
	return qb
}

/*
 @method Build: Builds the final Cypher query string and returns it along with the parameters.

 @returns: The constructed Cypher query string and a map of parameters.
*/
func (qb *QueryBuilder) Build() (string, map[string]interface{}) {
	var query string
	query += qb.query
	query += fmt.Sprintf("\nRETURN %s", qb.returnClause)
	return query, qb.params
}

/*

@interface Node: Defines methods for interacting with nodes in the graph.
	- @method AddChild: Adds a child node to the current node.
	- @method GetChildren: Returns the child nodes of the current node.
	- @method SetProps: Sets the properties of the current node.
	- @method GetProps: Returns the properties of the current node.
	- @method GetLabel: Returns the label of the current node.
	- @method MarshalJSON: Marshals the current node to JSON format.
*/
type Node interface {
	AddChild(child Node)
	GetChildren() []Node
	SetProps(props map[string]interface{})
	GetProps() map[string]interface{}
	GetLabel() string
	MarshalJSON() ([]byte, error)
}

/*
@struct BaseNode: Implements the Node interface and provides methods for managing node properties and relationships.

	- @property Label: The label of the node.
	- @property Name: The name of the node.
	- @property Props: A map of properties associated with the node.
	- @property Children: A slice of child nodes.
*/	
type BaseNode struct {
	Label    string
	Name     string
	Props    map[string]interface{}
	Children []Node
}

/*
@method AddChild: Adds a child node to the current node.
	- @param child: The child node to be added.
*/
func (n *BaseNode) AddChild(child Node) {
	n.Children = append(n.Children, child)
}

/*
@method GetChildren: Returns the child nodes of the current node.
	- @returns: A slice of child nodes.
*/
func (n *BaseNode) GetChildren() []Node {
	return n.Children
}

/*
@method SetProps: Sets the properties of the current node.
	- @param props: A map of properties to be set.
*/
func (n *BaseNode) SetProps(props map[string]interface{}) {
	n.Props = props
	if name, ok := props["name"].(string); ok {
		n.Name = name
	}
}

/*
@method GetProps: Returns the properties of the current node.
	- @returns: A map of properties associated with the node.
*/
func (n *BaseNode) GetProps() map[string]interface{} {
	return n.Props
}

/*
@method GetLabel: Returns the label of the current node.
	- @returns: The label of the node.
*/
func (n *BaseNode) GetLabel() string {
	return n.Label
}

/*
@method MarshalJSON: Marshals the current node to JSON format.
	- @returns: A byte slice containing the JSON representation of the node and an error if any.
*/
func (n *BaseNode) MarshalJSON() ([]byte, error) {
	props := map[string]interface{}{
		"label": n.Label,
		"name":  n.Name,
		"props": n.Props,
	}
	return json.Marshal(props)
}

/*
@private func buildTree maps the current BaseNode and its children to a strongly-typed model.

This method uses reflection to dynamically map the properties and relationships
of the BaseNode to the fields of the specified model type `T`. The method ensures
type safety by returning an error if the mapping fails or if the type casting
to `*T` is unsuccessful.

@type T: The type of the model to which the BaseNode should be mapped.

@returns:
  - (*T, error): A pointer to the mapped model of type `T` if successful, or an error
    if the mapping or type casting fails.

@example:
    // Assuming `root` is a *BaseNode representing a "World" node:
    world, err := root.buildTree[models.World]()
    if err != nil {
        log.Fatalf("Failed to build tree: %v", err)
    }
    fmt.Printf("Mapped model: %+v\n", world)
*/
func buildTree[T any](n *BaseNode) (*T, error) {
	model, ok := mapModels(n).(*T)
	if !ok {
		fmt.Printf("Failed to cast node with label '%s' to type '%T'\n", n.Label, model) // Debug log
		return nil, fmt.Errorf("failed to cast node to the desired model type")
}
	return model, nil
}

func newBaseNode(label string) *BaseNode {
	return &BaseNode{
		Label:    label,
		Children: []Node{},
	}
}

/*
@func NewDriver: Creates a new Neo4j driver instance.
	- @returns: A pointer to a Neo4j driver instance and an error if any.
	- @note: The driver is created using the URI, username, and password from environment variables.
@example:
		driver, err := NewDriver()
		if err != nil {
			log.Fatal(err)
		}
		defer driver.Close(ctx)
		session := driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
		defer session.Close(ctx)
		res, err := session.Run(ctx, "MATCH (n) RETURN n", nil)
		if err != nil {
			log.Fatal(err)
		}
		for res.Next(ctx) {
			record := res.Record()
			node, ok := record.Get("n")
			if !ok {
				log.Fatal("Failed to retrieve node from record")
			}
			fmt.Println(node)
		}
		if err = res.Err(); err != nil {
			log.Fatal(err)
		}
@note: The driver is verified for connectivity after creation.

@returns: A pointer to a Neo4j driver instance and an error if any.
*/
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

/*
@func BuildNodeTree: Builds a tree of nodes from a slice of neo4j records returned from a query.
	- @param records: A slice of neo4j records containing the nodes and their relationships.
	- @returns: A pointer to the root node of the tree and an error if any.
	- @note: The function populates a map with nodes, establishes relationships, and returns the root node.
	- @note: The function uses a generic type `T` to return a strongly-typed model.
@example:
		records, err := session.Run(ctx, "MATCH (n) RETURN n", nil)
		if err != nil {
			log.Fatal(err)
		}
		var recordList []neo4j.Record
		for records.Next(ctx) {
			recordList = append(recordList, *records.Record())
		}
		model, err := BuildNodeTree[models.World](recordList)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Mapped model: %+v\n", model)

*/
func BuildNodeTree[T any](records []neo4j.Record) ([]*T, error) {
	nodeMap := make(map[string]*BaseNode)

	populateNodeMap(records, nodeMap)
	establishRelationships(records, nodeMap)

	expectedLabel := getLabelForType[T]()

	var results []*T
	for _, node := range nodeMap {
			if node.Label != expectedLabel {
					continue
			}
			model, err := buildTree[T](node)
			if err != nil {
					return nil, err
			}
			results = append(results, model)
	}

	return results, nil
}

func getLabelForType[T any]() string {
	t := reflect.TypeOf((*T)(nil)).Elem()
	return t.Name()
}

func populateNodeMap(records []neo4j.Record, nodeMap map[string]*BaseNode) {
	for _, record := range records {
			for _, key := range record.Keys {
					node, ok := record.Get(key)
					if !ok || node == nil {
							continue
					}
					nodeValue, ok := node.(neo4j.Node)
					if !ok {
							continue
					}
					nodeID := nodeValue.ElementId
					if _, exists := nodeMap[nodeID]; !exists {
							newNode := createNode(nodeValue)
							nodeMap[nodeID] = newNode
					}
			}
	}
}

func createNode(nodeValue neo4j.Node) *BaseNode {
	newNode := newBaseNode(nodeValue.Labels[0])
	if nodeValue.Labels[0] == "User" {
		newNode.SetProps(map[string]interface{}{
			"username": nodeValue.Props["username"],
			"userID":   nodeValue.Props["userID"],
		})
	} else {
		nodeValue.Props["id"] = nodeValue.ElementId
		newNode.SetProps(nodeValue.Props)
	}
	return newNode
}

func establishRelationships(records []neo4j.Record, nodeMap map[string]*BaseNode) {
	for _, record := range records {
			var parentNode *BaseNode
			for _, key := range record.Keys {
					node, ok := record.Get(key)
					if !ok || node == nil {
							continue
					}
					nodeValue, ok := node.(neo4j.Node)
					if !ok {
							continue
					}
					nodeID := nodeValue.ElementId
					currentNode, exists := nodeMap[nodeID]
					if !exists {
							continue
					}
					if parentNode != nil {
							addChildIfNotExists(parentNode, currentNode)
					}
					parentNode = currentNode
			}
	}
}

func addChildIfNotExists(parentNode *BaseNode, childNode *BaseNode) {
	existingChildren := parentNode.GetChildren()
	for _, child := range existingChildren {
			if child == childNode {
				return // Child already exists, skip adding it again
			}
	}
	parentNode.AddChild(childNode) // Add the child if it doesn't already exist
}


var modelRegistry = make(map[string]reflect.Type)

func RegisterModel(modelName string, model interface{}) {
	modelType := reflect.TypeOf(model)
	if modelType.Kind() != reflect.Ptr {
		panic(fmt.Sprintf("model %s must be a pointer to a struct", modelName))
	}
	modelRegistry[modelName] = modelType.Elem()
}

func mapModels(node *BaseNode) interface{} {
	mapping, ok := modelRegistry[node.Label]
	if !ok {
			fmt.Printf("Label '%s' not found in modelRegistry\n", node.Label) // Debug log
			return nil
	}

	modelInstance := reflect.New(mapping).Interface()

	props := node.GetProps()
	modelValue := reflect.ValueOf(modelInstance).Elem()
	modelType := modelValue.Type()
	for i := 0; i < modelValue.NumField(); i++ {
			field := modelType.Field(i)

			// Map node properties
			if tag := field.Tag.Get("node"); tag != "" {
					if value, ok := props[tag]; ok {
							fieldValue := modelValue.FieldByName(field.Name)
							if fieldValue.IsValid() && fieldValue.CanSet() {
									fieldValue.Set(reflect.ValueOf(value))
							}
					}
			}

			// Map relationships
			if tag := field.Tag.Get("rel"); tag != "" {
					tagParts := strings.Split(tag, ",")
					if len(tagParts) != 2 {
							continue
					}
					relType := tagParts[0]
					relDirection := tagParts[1]

					fieldValue := modelValue.FieldByName(field.Name)
					if fieldValue.IsValid() && fieldValue.CanSet() && fieldValue.Kind() == reflect.Slice {
							childSlice := reflect.MakeSlice(fieldValue.Type(), 0, 0)
							for _, child := range node.GetChildren() {
									childNode := child.(*BaseNode)
									if isValidRelationship(relDirection) && childNode.Label == relType {
											childValue := mapModels(childNode)
											if childValue != nil {
													childSlice = reflect.Append(childSlice, reflect.ValueOf(childValue))
											}
									}
							}
							fieldValue.Set(childSlice)
					}
			}
	}

	return modelInstance
}

func isValidRelationship(direction string) bool {
	switch direction {
	case "->":
		return true
	case "<-":
		return true
	case "<->":
		return true
	default:
		return false
	}
}