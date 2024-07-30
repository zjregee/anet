package ahttp

func newRouter() *router {
	return &router{}
}

type router struct {
	root   *node
	routes map[string]HandlerFunc
}

type node struct {
	children map[byte]*node
	isEnd    bool
}

func (r *router) add(path string, h HandlerFunc) {
	path = normalizePathSlash(path)
	r.routes[path] = h
	r.insert(path)
}

func (r *router) find(path string) HandlerFunc {
	path = normalizePathSlash(path)
	prefix := r.longestPrefixMatch(path)
	return r.routes[prefix]
}

func (r *router) insert(prefix string) {
	n := r.root
	for i := 0; i < len(prefix); i++ {
		char := prefix[i]
		if _, found := n.children[char]; !found {
			n.children[char] = &node{children: make(map[byte]*node)}
		}
		n = n.children[char]
	}
	n.isEnd = true
}

func (r *router) longestPrefixMatch(query string) string {
	n := r.root
	longestPrefix := ""
	currentPrefix := ""
	for i := 0; i < len(query); i++ {
		char := query[i]
		if _, found := n.children[char]; !found {
			break
		}
		n = n.children[char]
		currentPrefix += string(char)
		if n.isEnd {
			longestPrefix = currentPrefix
		}
	}
	return longestPrefix
}

func normalizePathSlash(path string) string {
	if path == "" {
		path = "/"
	} else if path[0] != '/' {
		path = "/" + path
	}
	return path
}
