// Package primadapter stay for CLI interaction model
package primadapter

import primports "github.com/Reddetk/crawler-cli/internal/ports/primPorts"

type SeedProducer struct{
	WevParser primports.WebParser
}

func (reqMan *SeedProducer) configurateRequests(){

}
func (reqMan *SeedProducer) getRequests() {

}
func (reqMan *SeedProducer) parseRequests()
func (reqMan *SeedProducer) processRequests()

type Request struct {
	URL string
	inheritage map[string]bool
}