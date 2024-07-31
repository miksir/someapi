package user

import (
	"github.com/gofrs/uuid"
	"testing"
)

func TestUser_Validate(t *testing.T) {
	type fields struct {
		ID       string
		Name     string
		Email    string
		Birthday string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name:    "Normal user",
			fields:  fields{
				ID:       "9d225408-6a94-49e7-8e04-69ff654e0ff4",
				Name:     "Alice",
				Email:    "test1@example.com",
				Birthday: "1999-12-31",
			},
			wantErr: false,
		},
		{
			name:    "Normal user",
			fields:  fields{
				ID:       "64f4327a-645f-4cf7-a45a-ab1b90d1a497",
				Name:     "John",
				Email:    "test2@example.com",
				Birthday: "31/12/1999",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uuid1, err := uuid.FromString(tt.fields.ID)
			if err != nil {
				t.Errorf("UUID format error = %v", err)
			}
			u := User{
				ID:       uuid1,
				Name:     tt.fields.Name,
				Email:    tt.fields.Email,
				Birthday: tt.fields.Birthday,
			}
			if err := u.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
