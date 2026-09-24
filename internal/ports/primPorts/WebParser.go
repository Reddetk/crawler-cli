// Package primports stands for web parse interfase
package primports

import "context"

type WebParser interface {
	ProcessRequest(ctx context.Context, URL string, reqChan <- chan string) error
}
