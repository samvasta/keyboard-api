package utils

import (
	"fmt"
	"os"
)

func ServerURL(path string) string {
	return fmt.Sprintf("%s%s", os.Getenv("SERVER_URL"), path)
}
