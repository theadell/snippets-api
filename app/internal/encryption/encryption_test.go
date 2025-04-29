package encryption

import (
	"bytes"
	"testing"
)

func TestNewService(t *testing.T) {
	type args struct {
		systemKey string
	}
	tests := []struct {
		name    string
		args    args
		want    *Service
		wantErr bool
	}{
		{
			name:    "valid 16-byte key",
			args:    args{systemKey: "MTIzNDU2Nzg5MDEyMzQ1Ng=="},
			wantErr: false,
		},
		{
			name:    "invalid base64",
			args:    args{systemKey: "invalid-base64!"},
			wantErr: true,
		},
		{
			name:    "wrong length",
			args:    args{systemKey: "MTIzNDU2Nzg5MA=="},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewService(tt.args.systemKey)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewService() error = %s, wantErr %v", err.Error(), tt.wantErr)
				return
			}
		})
	}
}

func TestService_EncryptDecrypt(t *testing.T) {
	validKey := []byte("1234567890123456")

	tests := []struct {
		name string
		data []byte
	}{
		{
			name: "regular data",
			data: []byte("hello world!"),
		},
		{
			name: "empty data",
			data: []byte{},
		},
		{
			name: "binary data",
			data: []byte{0, 255, 34, 55, 0, 8, 42},
		},
	}

	s := &Service{systemKey: validKey}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encrypted, err := s.Encrypt(tt.data)
			if err != nil {
				t.Fatalf("Encrypt() error: %v", err)
			}

			decrypted, err := s.Decrypt(encrypted)
			if err != nil {
				t.Fatalf("Decrypt() error: %v", err)
			}

			if !bytes.Equal(decrypted, tt.data) {
				t.Errorf("Encrypt/Decrypt roundtrip failed: got %v, want %v", decrypted, tt.data)
			}
		})
	}
}
