package annotations

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnMashallingAnnotation(t *testing.T) {
	annotation := Annotation{}
	jason := `{
	            "thing": {
			          "id": "http://api.ft.com/things/2384fa7a-d514-3d6a-a0ea-3a711f66d0d8"
                }
            }
						`
	err := json.Unmarshal([]byte(jason), &annotation)
	if err != nil {
		panic(err)
	}
	assertion := assert.New(t)
	assertion.Nil(err)
}
