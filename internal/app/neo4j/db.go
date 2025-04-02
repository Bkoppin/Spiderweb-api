package neo

import (
	"context"
	"fmt"
	"os"
	"reflect"

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

func BuildNodeTree[T any](records []neo4j.Record) ([]*T, error) {
  var results []*T

  for _, record := range records {
    node, ok := record.Get("n")
    if !ok {
      continue
    }

    relatedNodes, _ := record.Get("relatedNodes")

    model := new(T)
    err := mapNodeToModel(node.(neo4j.Node), model)
    if err != nil {
      return nil, err
    }

    if relatedNodes != nil {
      err := mapRelatedNodesToModel(relatedNodes.([]interface{}), model)
      if err != nil {
        return nil, err
      }
    }

    results = append(results, model)
  }

  return results, nil
}

func mapNodeToModel[T any](node neo4j.Node, model *T) error {
  modelValue := reflect.ValueOf(model).Elem()
  modelType := reflect.TypeOf(*model)

  for i := 0; i < modelType.NumField(); i++ {
    field := modelType.Field(i)
    nodeTag := field.Tag.Get("node")

    if field.Name == "ID" && nodeTag == "id" {
      fieldValue := modelValue.FieldByName(field.Name)
      if fieldValue.IsValid() && fieldValue.CanSet() {
        if node.ElementId != "" {
          fieldValue.Set(reflect.ValueOf(node.ElementId))
        } else {
          fieldValue.Set(reflect.Zero(fieldValue.Type()))
        }
      }
      continue
    }

    if field.Name == "Label" {
      continue
    }

    if nodeTag == "" {
      continue
    }

    value, ok := node.Props[nodeTag]
    fieldValue := modelValue.FieldByName(field.Name)
    if fieldValue.IsValid() && fieldValue.CanSet() {
      if ok {
        fieldValue.Set(reflect.ValueOf(value))
      } else {
        fieldValue.Set(reflect.Zero(fieldValue.Type()))
      }
    }
  }

  return nil
}

func mapRelatedNodesToModel[T any](relatedNodes []interface{}, model *T) error {
  modelValue := reflect.ValueOf(model).Elem()
  modelType := reflect.TypeOf(*model)

  for i := 0; i < modelType.NumField(); i++ {
    field := modelType.Field(i)
    relTag := field.Tag.Get("rel")
    if relTag == "" {
      continue
    }

    fieldValue := modelValue.FieldByName(field.Name)
    if fieldValue.Kind() == reflect.Slice {
      slice := reflect.MakeSlice(field.Type, 0, len(relatedNodes))

      for _, relatedNode := range relatedNodes {
        node, ok := relatedNode.(neo4j.Node)
        if !ok {
          continue
        }

        relatedType, err := resolveTypeFromLabels(node.Labels)
        if err != nil {
          return err
        }

        expectedType := field.Type.Elem().Elem()
        if relatedType != expectedType {
          return fmt.Errorf("type mismatch: expected %v, got %v", expectedType, relatedType)
        }

        relatedModel := reflect.New(relatedType).Interface()
        mapNodeToModelReflect(node, relatedModel)
        slice = reflect.Append(slice, reflect.ValueOf(relatedModel))
      }

      fieldValue.Set(slice)
    }
  }

  return nil
}

func resolveTypeFromLabels(labels []string) (reflect.Type, error) {
  for _, label := range labels {
    if typ, ok := modelRegistry[label]; ok {
      return typ, nil
    }
  }
  return nil, fmt.Errorf("unknown label: %v", labels)
}

var modelRegistry = make(map[string]reflect.Type)

func RegisterModel(modelName string, model interface{}) {
  modelType := reflect.TypeOf(model)
  if modelType.Kind() != reflect.Ptr {
    panic(fmt.Sprintf("model %s must be a pointer to a struct", modelName))
  }
  modelRegistry[modelName] = modelType.Elem()
}

func mapNodeToModelReflect(node neo4j.Node, model interface{}) error {
  modelValue := reflect.ValueOf(model).Elem()
  modelType := reflect.TypeOf(model).Elem()

  for i := 0; i < modelType.NumField(); i++ {
    field := modelType.Field(i)
    nodeTag := field.Tag.Get("node")

    if field.Name == "ID" && nodeTag == "id" {
      fieldValue := modelValue.FieldByName(field.Name)
      if fieldValue.IsValid() && fieldValue.CanSet() {
        if node.ElementId != "" {
          fieldValue.Set(reflect.ValueOf(node.ElementId))
        } else {
          fieldValue.Set(reflect.Zero(fieldValue.Type()))
        }
      }
      continue
    }

    if field.Name == "Label" {
      continue
    }

    if nodeTag == "" {
      continue
    }

    value, ok := node.Props[nodeTag]
    fieldValue := modelValue.FieldByName(field.Name)
    if fieldValue.IsValid() && fieldValue.CanSet() {
      if ok {
        fieldValue.Set(reflect.ValueOf(value))
      } else {
        fieldValue.Set(reflect.Zero(fieldValue.Type()))
      }
    }
  }

  return nil
}