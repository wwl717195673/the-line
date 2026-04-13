package domain

const (
	StatusDisabled int8 = 0
	StatusEnabled  int8 = 1
)

func IsValidStatus(status int8) bool {
	return status == StatusDisabled || status == StatusEnabled
}
