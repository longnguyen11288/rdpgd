package cfsb

type Service struct {
	Id          string        `json:"id"`
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Tags        []string      `json:"tags"`
	Metadata    Metadata      `json:"metadata"`
	Plans       []*plans.Plan `json:"plans"`
	Bindable    bool          `json:"bindable"`
}

type Metadata struct {
	Label       string       `json:"label"`
	Description string       `json:"description"`
	Provider    string       `json:"provider"`
	Version     string       `json:"version"`
	Requires    []string     `json:"requires"`
	Tags        []string     `json:"tags"`
	Metadata    TileMetadata `json:"metadata"`
}

type TileMetadata struct {
	DisplayName         string `json:"displayname"`
	ImageUrl            string `json:"imageurl"`
	LongDescription     string `json:"longdescription"`
	ProviderDisplayName string `json:"providerdisplayname"`
	DocumentationUrl    string `json:"documentationurl"`
	SupportUrl          string `json:"supporturl"`
}

