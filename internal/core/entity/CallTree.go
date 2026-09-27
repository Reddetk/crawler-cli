package entity

import (
	"net/url"
)

type CallTree struct {
	Root  *Page
	Host  string
	Um    *UniqMap
	Error error
}

func NewCallTree(root *Page) (*CallTree, error) {
	hostURL, err := url.Parse(root.Resource)
	if err != nil {
		return nil, err
	}
	um := NewUniqMap()
	return &CallTree{
		Root: root,
		Host: hostURL.Host,
		Um:   um,
	}, nil
}


func(ct *CallTree) ConnectBranch(parentPage *Page){
	
}