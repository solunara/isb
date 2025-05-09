package utils

import (
	"log"

	"github.com/bwmarrin/snowflake"
	"github.com/spf13/viper"
)

var node *snowflake.Node

func init() {
	var err error
	node, err = snowflake.NewNode(viper.GetInt64("app.node_id"))
	if err != nil {
		log.Fatalf("Failed to initialize Snowflake node: %v", err)
	}
}

func GenerateUinqueID() string {
	return node.Generate().String()
}
