package valkyrie

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type DataTransferObject struct {
	Email         string `json:"email" validate:"required,email"`
	Password      string `json:"password" validate:"required"`
	Today         string `json:"today" validate:"date,required"`
	CreateDate    string `json:"create_date" validate:"datetime,required"`
	CourierID     int    `json:"courier_id" validate:"required,gt=1"`
	PaymentMethod string `json:"payment_method" validate:"enum=ovo gopay virtual"`
}

func TestValidate(t *testing.T) {
	dto := []struct {
		label string
		in    DataTransferObject
		out   interface{}
	}{
		{
			"Test Validate Pass",
			DataTransferObject{
				Email:         "nanang.jobs@gmail.com",
				Password:      "sekret",
				Today:         "2019-09-01",
				CreateDate:    "2019-09-01T16:18:22+00:00",
				CourierID:     17,
				PaymentMethod: "ovo",
			},
			[]ErrorValidator(nil),
		},
		{
			"Test Validate Email Fail",
			DataTransferObject{
				Email:         "nanang.jobs@gmail",
				Password:      "sekret",
				Today:         "2019-09-01",
				CreateDate:    "2019-09-01T16:18:22+00:00",
				CourierID:     17,
				PaymentMethod: "ovo",
			},
			[]ErrorValidator{
				{
					Field:   "email",
					Value:   "nanang.jobs@gmail",
					Tag:     "email",
					Type:    "string",
					Message: "Invalid Type nanang.jobs@gmail for input email",
				},
			},
		},
		{
			"Test Validate Date Fail",
			DataTransferObject{
				Email:         "nanang.jobs@gmail.com",
				Password:      "sekret",
				Today:         "2019-09",
				CreateDate:    "2019-09-01T16:18:22+00:00",
				CourierID:     17,
				PaymentMethod: "gopay",
			},
			[]ErrorValidator{
				{
					Tag:     "date",
					Value:   "2019-09",
					Field:   "today",
					Type:    "string",
					Message: "Invalid Type 2019-09 for input today",
				},
			},
		},
		{
			"Test Validate Courier ID Fail",
			DataTransferObject{
				Email:         "nanang.jobs@gmail.com",
				Password:      "sekret",
				Today:         "2019-09-01",
				CreateDate:    "2019-09-01T16:18:22+00:00",
				CourierID:     0,
				PaymentMethod: "ovo",
			},
			[]ErrorValidator{
				{
					Tag:     "required",
					Value:   "0",
					Field:   "courier_id",
					Type:    "int",
					Message: "Invalid Type 0 for input courier_id",
				},
			},
		},
	}

	for _, tt := range dto {
		tt := tt
		t.Run(tt.label, func(t *testing.T) {
			err := Validate(tt.in)
			if err != nil {
				assert.NotNil(t, err)
			}
			assert.Equal(t, err, tt.out)
		})
	}
}

func TestParseDate(t *testing.T) {
	dt := ParseDate("2019-09-01")
	assert.Equal(t, time.Date(2019, 9, 1, 0, 0, 0, 0, time.UTC), dt)
	dt = ParseDate("2019-09-01T16:18:22Z00:00")
	assert.Equal(t, time.Time{}, dt)
}

func TestParseDatetime(t *testing.T) {
	dt := ParseDatetime("2019-09-01T16:18:22+00:00")
	assert.Equal(t, time.Date(2019, 9, 1, 16, 18, 22, 0, time.UTC), dt.UTC())
	dt = ParseDatetime("2019-09-01T16:18:22Z00:00")
	assert.Equal(t, time.Time{}, dt)
}
