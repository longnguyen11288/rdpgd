package services

import (
	"github.com/wayneeseguin/rdpg-agent/plans"
)

type Service struct {
	Id          string          `json:"id"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Tags        []string        `json:"tags"`
	Metadata    ServiceMetadata `json:"metadata"`
	Plans []*plans.Plan `json:"plans"`
	Bindable bool `json:"bindable"`
}

type ServiceMetadata struct {
	Label       string                  `json:"label"`
	Description string                  `json:"description"`
	Provider    string                  `json:"provider"`
	Version     string                  `json:"version"`
	Requires    []string                `json:"requires"`
	Tags        []string                `json:"tags"`
	Metadata    ServiceMetadataMetadata `json:"metadata"`
}

type ServiceMetadataMetadata struct {
	DisplayName         string `json:"displayname"`
	ImageUrl            string `json:"imageurl"`
	LongDescription     string `json:"longdescription"`
	ProviderDisplayName string `json:"providerdisplayname"`
	DocumentationUrl    string `json:"documentationurl"`
	SupportUrl          string `json:"supporturl"`
}
