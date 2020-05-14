package meta

import "testing"

func TestGpsToLat(t *testing.T) {
	lat := GpsToDecimal("51 deg 15' 17.47\" N")
	exp := float32(51.254852)

	if lat-exp > 0 {
		t.Fatalf("lat is %f, should be %f", lat, exp)
	}
}

func TestGpsToLng(t *testing.T) {
	lng := GpsToDecimal("7 deg 23' 22.09\" E")
	exp := float32(7.389470)

	if lng-exp > 0 {
		t.Fatalf("lng is %f, should be %f", lng, exp)
	}
}

func TestGpsToLatLng(t *testing.T) {
	lat, lng := GpsToLatLng("51 deg 15' 17.47\" N, 7 deg 23' 22.09\" E")
	expLat, expLng := float32(51.254852), float32(7.389470)

	if lat-expLat > 0 {
		t.Fatalf("lat is %f, should be %f", lat, expLat)
	}

	if lng-expLng > 0 {
		t.Fatalf("lng is %f, should be %f", lng, expLng)
	}
}
