package main

import "testing"

func BenchmarkSerializeToJSON(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := serializeToJSON(metadata)
		if err != nil {
			b.Fatal(err)
		}
	}
}
func BenchmarkSerializeToXML(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := serializeToXML(metadata)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkSerializeToProto(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := serializeToProto(genMetadata)
		if err != nil {
			b.Fatal(err)
		}
	}
}
