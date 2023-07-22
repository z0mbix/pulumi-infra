package common

type Environment struct {
	Name        string
	Description string
}

type Region struct {
	Id        string
	Name      string
	ShortName string
}

type Bucket struct {
	Name               string
	Prefix             string
	DataClassification DataClassification
}

var BucketPrefix = "zombix"

type Cloud struct {
	Name        string
	Description string
}

var AWSCloud = Cloud{
	Name:        "aws",
	Description: "Amazon Web Services",
}

var GCPCloud = Cloud{
	Name:        "gcp",
	Description: "Google Cloud Platform",
}
