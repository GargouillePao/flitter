package core

import (
	"errors"
	"strings"
)

type NodeInfo string

func (n NodeInfo) GetLeaderInfo() (NodeInfo, error) {
	nodes := strings.Split(string(n), "/")
	if len(nodes) > 0 {
		nodes = nodes[:len(nodes)-1]
		return NodeInfo(strings.Join(nodes, "/")), nil
	}
	return "", errors.New("Invalid node info")
}

func (n NodeInfo) GetEndpoint() (string, error) {
	nodes := strings.Split(string(n), "/")
	if len(nodes) > 0 {
		return "tcp://" + nodes[len(nodes)-1], nil
	}
	return "", errors.New("Invalid node info")
}
