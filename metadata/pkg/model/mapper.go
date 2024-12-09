package model

import "movieexample.com/gen"

// MetadataToProto converts a Metadata struct to a gen.Metadata proto.
// It copies the fields from the Metadata struct to the gen.Metadata proto.
func MetadataToProto(m *Metadata) *gen.Metadata {
	return &gen.Metadata{
		Id:          m.ID,
		Title:       m.Title,
		Description: m.Description,
		Director:    m.Director,
	}
}

// MetadataFromProto converts a gen.Metadata proto to a Metadata struct.
// It copies the fields from the gen.Metadata proto to the Metadata struct.
func MetadataFromProto(m *gen.Metadata) *Metadata {
	return &Metadata{
		ID:          m.Id,
		Title:       m.Title,
		Description: m.Description,
		Director:    m.Director,
	}
}
