package neo

import (
	"context"
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

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

func CreateAndReturnNode(ctx context.Context, driver neo4j.DriverWithContext, label string, properties map[string]interface{}) (map[string]interface{}, error) {
	session := driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	result, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
					query := fmt.Sprintf("CREATE (n:%s $props) RETURN n", label)
					res, err := tx.Run(ctx, query, map[string]interface{}{"props": properties})
					if err != nil {
									return nil, err
					}

					if res.Next(ctx) {
									record := res.Record()
									node := record.Values[0].(neo4j.Node)
									return node.Props, nil
					}

					return nil, fmt.Errorf("node creation failed")
	})

	if err != nil {
					return nil, err
	}

	return result.(map[string]interface{}), nil
}

