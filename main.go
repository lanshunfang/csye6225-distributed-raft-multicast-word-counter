package distributed

import (
	"fmt"
	"os"
)

func main() {

	// if os.Getenv("ENV_MULTICAST_GROUP") == "" {
	// 	os.Setenv("ENV_MULTICAST_GROUP", defaultmulticast)
	// }

	fmt.Println("ENV_MULTICAST_GROUP:", os.Getenv("ENV_MULTICAST_GROUP"))

}
