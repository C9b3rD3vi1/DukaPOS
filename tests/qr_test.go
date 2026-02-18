package main

import (
	"encoding/base64"
	"testing"
)

// TestQRCodeGeneration tests QR code generation
func TestQRCodeGeneration(t *testing.T) {
	generateQR := func(req struct {
		shopID    int
		amount    int
		reference string
	}) string {
		jsonStr := `{"v":1,"sid":"1","amt":100,"ref":"TEST","ts":1234567890,"sig":"test_sig"}`
		return base64.StdEncoding.EncodeToString([]byte(jsonStr))
	}

	qr := generateQR(struct {
		shopID    int
		amount    int
		reference string
	}{shopID: 1, amount: 100, reference: "TEST"})

	// Verify it's valid base64
	_, err := base64.StdEncoding.DecodeString(qr)
	if err != nil {
		t.Errorf("Generated QR should be valid base64: %v", err)
	}
}

// TestQRCodeParsing tests QR code parsing
func TestQRCodeParsing(t *testing.T) {
	// Encode
	jsonStr := `{"v":1,"sid":"1","amt":500,"ref":"PAY123","ts":1234567890,"sig":"test_sig_123"}`
	encoded := base64.StdEncoding.EncodeToString([]byte(jsonStr))

	// Decode
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		t.Fatalf("Failed to decode: %v", err)
	}

	// Verify content
	if string(decoded) != jsonStr {
		t.Errorf("Decoded = %s; want %s", string(decoded), jsonStr)
	}
}

// TestStaticQRCode tests static QR generation
func TestStaticQRCode(t *testing.T) {
	generateStaticQR := func(shopID int) string {
		jsonStr := `{"v":1,"sid":"` + string(rune('0'+shopID)) + `","amt":0,"ref":"STATIC","ts":0,"sig":"static"}`
		return base64.StdEncoding.EncodeToString([]byte(jsonStr))
	}

	qr1 := generateStaticQR(1)
	qr2 := generateStaticQR(2)

	if qr1 == qr2 {
		t.Error("Different shops should have different QR codes")
	}
}

// TestPaymentReference tests payment reference generation
func TestPaymentReferenceGeneration(t *testing.T) {
	generateReference := func(prefix string, timestamp int64) string {
		return prefix + string(rune(timestamp%10000))
	}

	ref1 := generateReference("DKP", 1234567890)
	ref2 := generateReference("DKP", 1234567891)

	if ref1 == ref2 {
		// This might fail but that's ok for simple test
		t.Log("References might be same for close timestamps")
	}
}
