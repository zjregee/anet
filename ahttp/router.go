package ahttp

import (
	"strings"
)

type node struct {
	part     string
	isParam  bool
	isWild   bool
	isEnd    bool
	children map[string]*node
	handler  HandlerFunc
}

func newNode(part string, isParam, isWild bool) *node {
	return &node{
		part:     part,
		isParam:  isParam,
		isWild:   isWild,
		children: make(map[string]*node),
	}
}

func (n *node) insert(parts []string, handler HandlerFunc) {
	node := n
	for i := 0; i < len(parts); i++ {
		part := parts[i]
		isParam := strings.HasPrefix(part, ":")
		isWild := strings.HasPrefix(part, "*")
		child, ok := node.children[part]
		if !ok {
			child = newNode(part, isParam, isWild)
			node.children[part] = child
		}
		node = child
	}
	node.isEnd = true
	node.handler = handler
}

func (n *node) search(parts []string) (HandlerFunc, map[string]string) {
	node := n
	params := make(map[string]string)
	for _, part := range parts {
		child, ok := node.children[part]
		if !ok {
			for _, childNode := range node.children {
				if childNode.isParam {
					params[childNode.part[1:]] = part
					child = childNode
					break
				}
				if childNode.isWild {
					params[childNode.part[1:]] = strings.Join(parts, "/")
					child = childNode
					break
				}
			}
		}
		if child == nil {
			return nil, nil
		}
		node = child
	}
	if node.isEnd {
		return node.handler, params
	}
	return nil, nil
}

func newRouter() *router {
	return &router{
		routes: make(map[string]*node),
	}
}

type router struct {
	routes map[string]*node
}

func (r *router) add(method, path string, h HandlerFunc) {
	path = normalizePathSlash(path)
	node, ok := r.routes[method]
	if !ok {
		node = newNode("/", false, false)
		r.routes[method] = node
	}
	parts := strings.Split(path, "/")
	node.insert(parts, h)
}

func (r *router) find(method, path string) (HandlerFunc, map[string]string) {
	path = normalizePathSlash(path)
	node, ok := r.routes[method]
	if !ok {
		return nil, nil
	}
	parts := strings.Split(path, "/")
	h, params := node.search(parts)
	return h, params
}

func normalizePathSlash(path string) string {
	if path == "" {
		path = "/"
	} else if path[0] != '/' {
		path = "/" + path
	}
	return path
}
