package common

type DataClassification string

const (
	CONFIDENTIAL DataClassification = "confidential"
	PUBLIC       DataClassification = "public"
	RESTRICTED   DataClassification = "restricted"
	SENSITIVE    DataClassification = "sensitive"
)
