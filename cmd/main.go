package main

import (
	"encoding/json"
	"fmt"
)

func main() {
	a := "{\"appattempt_1657221318387_0020_000001\":1,\"appattempt_1657221318387_0021_000001\":1,\"appattempt_1657221318387_0022_000001\":1,\"appattempt_1657221318387_0010_000001\":1,\"appattempt_1657221318387_0019_000002\":1,\"appattempt_1657221318387_0017_000001\":1,\"appattempt_1657221318387_0018_000001\":1,\"appattempt_1657221318387_0015_000002\":1}"
	var m map[string]string
	fmt.Println(json.Unmarshal([]byte(a), &m))
	fmt.Println(m)
}
