package shared

import "uuid"

func ValidateID(id string) error {
	_, err := uuid.Parse(id)
	return err
}
