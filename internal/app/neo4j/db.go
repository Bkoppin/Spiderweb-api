package neo

import (
	"api/internal/app/models"
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

type QueryBuilder struct {
	query        string
	returnClause string
	params       map[string]interface{}
}

func NewQueryBuilder() *QueryBuilder {
	return &QueryBuilder{
		params: make(map[string]interface{}),
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

type Node interface {
	AddChild(child Node)
	GetChildren() []Node
	SetProps(props map[string]interface{})
	GetProps() map[string]interface{}
	GetLabel() string
	MarshalJSON() ([]byte, error)
	buildObject() interface{}
	BuildTree() interface{}
}

type BaseNode struct {
	Label    string
	Name     string
	Props    map[string]interface{}
	Children []Node
}

func (n *BaseNode) AddChild(child Node) {
	n.Children = append(n.Children, child)
}

func (n *BaseNode) GetChildren() []Node {
	return n.Children
}

func (n *BaseNode) SetProps(props map[string]interface{}) {
	n.Props = props
	if name, ok := props["name"].(string); ok {
		n.Name = name
	}
}

func (n *BaseNode) GetProps() map[string]interface{} {
	return n.Props
}

func (n *BaseNode) GetLabel() string {
	return n.Label
}

func (n *BaseNode) MarshalJSON() ([]byte, error) {
	props := map[string]interface{}{
		"label": n.Label,
		"name":  n.Name,
		"props": n.Props,
	}
	return json.Marshal(props)
}

func (n *BaseNode) buildObject() interface{} {
	if n.Label == "" {
		return nil
	}

	if n.Label == "User" {
		return &models.NeoUser{
			Username: n.Props["username"].(string),
			UserID:   n.Props["userID"].(int64),
			Worlds:   []models.World{},
		}
	} else if n.Label == "World" {
		return &models.World{
			Name:        n.Props["name"].(string),
			ID:          n.Props["id"].(string),
			Type:        n.Props["type"].(string),
			Description: n.Props["description"].(string),
			Continents:  []models.Continent{},
			Oceans:      []models.Ocean{},
		}
	} else if n.Label == "Continent" {
		return &models.Continent{
			Name:        n.Props["name"].(string),
			ID:          n.Props["id"].(string),
			Description: n.Props["description"].(string),
			Type:        n.Props["type"].(string),
			Zones:       []models.Zone{},
		}
	} else if n.Label == "Ocean" {
		return &models.Ocean{
			Name:        n.Props["name"].(string),
			Description: n.Props["description"].(string),
		}
	} else if n.Label == "Zone" {
		return &models.Zone{
			Name:        n.Props["name"].(string),
			Type:        n.Props["type"].(string),
			Description: n.Props["description"].(string),
			Locations:   []models.Location{},
			Cities:      []models.City{},
			Biome:       n.Props["biome"].(string),
		}
	} else if n.Label == "Location" {
		return &models.Location{
			Name:        n.Props["name"].(string),
			Type:        n.Props["type"].(string),
			Description: n.Props["description"].(string),
		}
	} else if n.Label == "City" {
		return &models.City{
			Name:        n.Props["name"].(string),
			Type:        n.Props["type"].(string),
			Description: n.Props["description"].(string),
			Capital:     n.Props["capital"].(bool),
		}
	}
	return nil
}

func (n *BaseNode) BuildTree() interface{} {
	root := n.buildObject()
	if root == nil {
		return nil
	}

	if len(n.Children) == 0 {
		return root
	}

	switch node := root.(type) {
	case *models.NeoUser:
		for _, child := range n.Children {
			childObj := child.BuildTree()
			if childObj == nil {
				continue
			}
			if world, ok := childObj.(*models.World); ok {
				node.Worlds = append(node.Worlds, *world)
			}
		}
	case *models.World:
		for _, child := range n.Children {
			childObj := child.BuildTree()
			if childObj == nil {
				continue
			}
			if continent, ok := childObj.(*models.Continent); ok {
				node.Continents = append(node.Continents, *continent)
			} else if ocean, ok := childObj.(*models.Ocean); ok {
				node.Oceans = append(node.Oceans, *ocean)
			}
		}
	case *models.Continent:
		for _, child := range n.Children {
			childObj := child.BuildTree()
			if childObj == nil {
				continue
			}
			if zone, ok := childObj.(*models.Zone); ok {
				node.Zones = append(node.Zones, *zone)
			}
		}
	case *models.Zone:
		for _, child := range n.Children {
			childObj := child.BuildTree()
			if childObj == nil {
				continue
			}
			if location, ok := childObj.(*models.Location); ok {
				node.Locations = append(node.Locations, *location)
			} else if city, ok := childObj.(*models.City); ok {
				node.Cities = append(node.Cities, *city)
			}
		}
	}

	return root
}

func newBaseNode(label string) *BaseNode {
	return &BaseNode{
		Label:    label,
		Children: []Node{},
	}
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

func BuildNodeTree(records []neo4j.Record) *BaseNode {
	nodeMap := make(map[string]*BaseNode)
	var root *BaseNode
	populateNodeMap(records, nodeMap, &root)
	establishRelationships(records, nodeMap)
	return root
}

func populateNodeMap(records []neo4j.Record, nodeMap map[string]*BaseNode, root **BaseNode) {
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
				if *root == nil && nodeValue.Labels[0] == "User" {
					*root = newNode
				}
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
			return
		}
	}
	parentNode.AddChild(childNode)
}
